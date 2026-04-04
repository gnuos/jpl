package stdlib

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/gnuos/jpl/engine"
)

// ==============================================================================
// 异步文件 IO 函数
// 提供非阻塞的文件操作，通过回调通知完成
// 所有函数底层使用 goroutine 执行，避免阻塞事件循环
// ==============================================================================

// fileAccessTracker 文件访问追踪器
// 用于检测批量操作中的同文件冲突
type fileAccessTracker struct {
	mu     sync.Mutex
	active map[string]int // path -> 引用计数
}

// 全局访问追踪器
var globalFileTracker = &fileAccessTracker{
	active: make(map[string]int),
}

// acquire 获取文件访问权
func (t *fileAccessTracker) acquire(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.active[path]++
}

// release 释放文件访问权
func (t *fileAccessTracker) release(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if count, ok := t.active[path]; ok {
		if count <= 1 {
			delete(t.active, path)
		} else {
			t.active[path] = count - 1
		}
	}
}

// isBusy 检查文件是否正在被访问
func (t *fileAccessTracker) isBusy(path string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.active[path] > 0
}

// checkConflicts 检测路径列表中的冲突
func (t *fileAccessTracker) checkConflicts(paths []string) []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	var conflicts []string
	seen := make(map[string]bool)
	for _, path := range paths {
		if seen[path] || t.active[path] > 0 {
			conflicts = append(conflicts, path)
		}
		seen[path] = true
	}
	return conflicts
}

// RegisterFileIOAsync 注册异步文件 IO 函数到引擎
func RegisterFileIOAsync(e *engine.Engine) {
	// 一次性读写
	e.RegisterFunc("file_get_async", builtinFileGetAsync)
	e.RegisterFunc("file_put_async", builtinFilePutAsync)
	e.RegisterFunc("file_append_async", builtinFileAppendAsync)
	e.RegisterFunc("file_get_bytes", builtinFileGetBytes)
	e.RegisterFunc("file_put_bytes", builtinFilePutBytes)

	// 流式读取
	e.RegisterFunc("file_read_lines", builtinFileReadLines)
	e.RegisterFunc("file_read_chunks", builtinFileReadChunks)

	// 批量操作
	e.RegisterFunc("file_get_batch", builtinFileGetBatch)
	e.RegisterFunc("file_put_batch", builtinFilePutBatch)
	e.RegisterFunc("file_parallel", builtinFileParallel)

	// 文件锁
	e.RegisterFunc("file_with_lock", builtinFileWithLock)

	// 模块注册 — import "asyncio" 可用（参考 Python asyncio）
	e.RegisterModule("asyncio", map[string]engine.GoFunction{
		"file_get":         builtinFileGetAsync,
		"file_put":         builtinFilePutAsync,
		"file_append":      builtinFileAppendAsync,
		"file_bytes":       builtinFileGetBytes,
		"file_write_bytes": builtinFilePutBytes,
		"read_lines":       builtinFileReadLines,
		"read_chunks":      builtinFileReadChunks,
		"get_batch":        builtinFileGetBatch,
		"put_batch":        builtinFilePutBatch,
		"parallel":         builtinFileParallel,
		"with_lock":        builtinFileWithLock,
	})
}

// FileIOAsyncNames 返回异步文件 IO 函数名称列表
func FileIOAsyncNames() []string {
	return []string{
		"file_get_async", "file_put_async", "file_append_async",
		"file_get_bytes", "file_put_bytes",
		"file_read_lines", "file_read_chunks",
		"file_get_batch", "file_put_batch", "file_parallel",
		"file_with_lock",
	}
}

// ==============================================================================
// 一次性读写
// ==============================================================================

// builtinFileGetAsync 异步读取整个文件（文本模式）
// file_get_async($path, fn($data) { ... }) → null
func builtinFileGetAsync(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("file_get_async() expects 2 arguments: path, callback")
	}

	path := args[0].String()
	callback := args[1]

	if callback.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_get_async() expects function callback")
	}

	// 异步执行
	go func() {
		data, err := os.ReadFile(path)
		if err != nil {
			// 回调传入 null 表示错误
			_, _ = ctx.VM().CallValue(callback, engine.NewNull())
			return
		}
		_, _ = ctx.VM().CallValue(callback, engine.NewString(string(data)))
	}()

	return engine.NewNull(), nil
}

// builtinFilePutAsync 异步写入整个文件（文本模式）
// file_put_async($path, $data, fn() { ... }) → null
func builtinFilePutAsync(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("file_put_async() expects 3 arguments: path, data, callback")
	}

	path := args[0].String()
	data := []byte(args[1].String())
	callback := args[2]

	if callback.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_put_async() expects function callback")
	}

	go func() {
		err := os.WriteFile(path, data, 0644)
		if err != nil {
			_, _ = ctx.VM().CallValue(callback, engine.NewBool(false))
			return
		}
		_, _ = ctx.VM().CallValue(callback, engine.NewBool(true))
	}()

	return engine.NewNull(), nil
}

// builtinFileAppendAsync 异步追加到文件
// file_append_async($path, $data, fn() { ... }) → null
func builtinFileAppendAsync(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("file_append_async() expects 3 arguments: path, data, callback")
	}

	path := args[0].String()
	data := []byte(args[1].String())
	callback := args[2]

	if callback.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_append_async() expects function callback")
	}

	go func() {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			_, _ = ctx.VM().CallValue(callback, engine.NewBool(false))
			return
		}
		defer f.Close()

		_, err = f.Write(data)
		if err != nil {
			_, _ = ctx.VM().CallValue(callback, engine.NewBool(false))
			return
		}
		_, _ = ctx.VM().CallValue(callback, engine.NewBool(true))
	}()

	return engine.NewNull(), nil
}

// builtinFileGetBytes 异步读取二进制文件
// file_get_bytes($path, fn($buffer) { ... }) → null
func builtinFileGetBytes(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("file_get_bytes() expects 2 arguments: path, callback")
	}

	path := args[0].String()
	callback := args[1]

	if callback.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_get_bytes() expects function callback")
	}

	go func() {
		data, err := os.ReadFile(path)
		if err != nil {
			_, _ = ctx.VM().CallValue(callback, engine.NewNull())
			return
		}
		// 返回 buffer 对象
		buf := NewBuffer("big")
		buf.data.Write(data)
		_, _ = ctx.VM().CallValue(callback, buf)
	}()

	return engine.NewNull(), nil
}

// builtinFilePutBytes 异步写入二进制文件
// file_put_bytes($path, $buffer, fn() { ... }) → null
func builtinFilePutBytes(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("file_put_bytes() expects 3 arguments: path, buffer, callback")
	}

	path := args[0].String()
	callback := args[2]

	if callback.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_put_bytes() expects function callback")
	}

	// 获取 buffer 数据
	var data []byte
	if buf, ok := args[1].(*BufferValue); ok {
		data = buf.Bytes()
	} else {
		data = []byte(args[1].String())
	}

	go func() {
		err := os.WriteFile(path, data, 0644)
		if err != nil {
			_, _ = ctx.VM().CallValue(callback, engine.NewBool(false))
			return
		}
		_, _ = ctx.VM().CallValue(callback, engine.NewBool(true))
	}()

	return engine.NewNull(), nil
}

// ==============================================================================
// 流式读取
// ==============================================================================

// builtinFileReadLines 逐行读取文件
// file_read_lines($path, fn($line) { ... }, fn() { ... }) → null
func builtinFileReadLines(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("file_read_lines() expects 3 arguments: path, onLine, onDone")
	}

	path := args[0].String()
	onLine := args[1]
	onDone := args[2]

	if onLine.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_read_lines() expects function for onLine")
	}
	if onDone.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_read_lines() expects function for onDone")
	}

	go func() {
		f, err := os.Open(path)
		if err != nil {
			// 错误时调用 onDone
			_, _ = ctx.VM().CallValue(onDone, engine.NewBool(false))
			return
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			_, _ = ctx.VM().CallValue(onLine, engine.NewString(line))
		}

		_, _ = ctx.VM().CallValue(onDone, engine.NewBool(true))
	}()

	return engine.NewNull(), nil
}

// builtinFileReadChunks 分块读取文件
// file_read_chunks($path, $chunkSize, fn($chunk) { ... }, fn() { ... }) → null
func builtinFileReadChunks(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("file_read_chunks() expects 4 arguments: path, chunkSize, onChunk, onDone")
	}

	path := args[0].String()
	chunkSize := int(args[1].Int())
	onChunk := args[2]
	onDone := args[3]

	if chunkSize <= 0 {
		chunkSize = 4096
	}
	if onChunk.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_read_chunks() expects function for onChunk")
	}
	if onDone.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_read_chunks() expects function for onDone")
	}

	go func() {
		f, err := os.Open(path)
		if err != nil {
			_, _ = ctx.VM().CallValue(onDone, engine.NewBool(false))
			return
		}
		defer f.Close()

		buf := make([]byte, chunkSize)
		for {
			n, err := f.Read(buf)
			if n > 0 {
				chunk := engine.NewString(string(buf[:n]))
				_, _ = ctx.VM().CallValue(onChunk, chunk)
			}
			if err != nil {
				break
			}
		}

		_, _ = ctx.VM().CallValue(onDone, engine.NewBool(true))
	}()

	return engine.NewNull(), nil
}

// ==============================================================================
// 批量操作
// ==============================================================================

// builtinFileGetBatch 批量读取文件
// file_get_batch($paths, fn($results) { ... }) → null
// $results 是字符串数组，与 $paths 对应
func builtinFileGetBatch(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("file_get_batch() expects 2 arguments: paths, callback")
	}

	pathsArray := args[0].Array()
	callback := args[1]

	if callback.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_get_batch() expects function callback")
	}

	paths := make([]string, len(pathsArray))
	for i, p := range pathsArray {
		paths[i] = p.String()
	}

	go func() {
		results := make([]engine.Value, len(paths))
		var wg sync.WaitGroup

		for i, path := range paths {
			wg.Add(1)
			go func(idx int, p string) {
				defer wg.Done()
				data, err := os.ReadFile(p)
				if err != nil {
					results[idx] = engine.NewNull()
				} else {
					results[idx] = engine.NewString(string(data))
				}
			}(i, path)
		}

		wg.Wait()
		_, _ = ctx.VM().CallValue(callback, engine.NewArray(results))
	}()

	return engine.NewNull(), nil
}

// builtinFilePutBatch 批量写入文件
// file_put_batch($items, fn($results) { ... }) → null
// $items 是对象数组：[{path: "...", data: "..."}, ...]
func builtinFilePutBatch(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("file_put_batch() expects 2 arguments: items, callback")
	}

	itemsArray := args[0].Array()
	callback := args[1]

	if callback.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_put_batch() expects function callback")
	}

	type writeItem struct {
		path string
		data []byte
	}

	items := make([]writeItem, len(itemsArray))
	for i, item := range itemsArray {
		obj := item.Object()
		items[i] = writeItem{
			path: obj["path"].String(),
			data: []byte(obj["data"].String()),
		}
	}

	go func() {
		results := make([]engine.Value, len(items))
		var wg sync.WaitGroup

		for i, item := range items {
			wg.Add(1)
			go func(idx int, p string, d []byte) {
				defer wg.Done()
				err := os.WriteFile(p, d, 0644)
				results[idx] = engine.NewBool(err == nil)
			}(i, item.path, item.data)
		}

		wg.Wait()
		_, _ = ctx.VM().CallValue(callback, engine.NewArray(results))
	}()

	return engine.NewNull(), nil
}

// builtinFileParallel 并行执行多个文件操作
// file_parallel($ops, fn($results) { ... }) → null
// $ops 是操作数组：[{op: "read", path: "..."}, {op: "write", path: "...", data: "..."}]
func builtinFileParallel(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("file_parallel() expects 2 arguments: ops, callback")
	}

	opsArray := args[0].Array()
	callback := args[1]

	if callback.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_parallel() expects function callback")
	}

	type fileOp struct {
		op   string
		path string
		data string
	}

	ops := make([]fileOp, len(opsArray))
	for i, op := range opsArray {
		obj := op.Object()
		ops[i] = fileOp{
			op:   obj["op"].String(),
			path: obj["path"].String(),
		}
		if d, ok := obj["data"]; ok {
			ops[i].data = d.String()
		}
	}

	// 检测冲突
	var paths []string
	for _, op := range ops {
		paths = append(paths, op.path)
	}
	conflicts := globalFileTracker.checkConflicts(paths)

	go func() {
		results := make([]engine.Value, len(ops))
		var wg sync.WaitGroup

		for i, op := range ops {
			wg.Add(1)
			go func(idx int, o fileOp) {
				defer wg.Done()
				defer globalFileTracker.release(o.path)

				// 检查冲突
				if slices.Contains(conflicts, o.path) {
					results[idx] = engine.NewObject(map[string]engine.Value{
						"success": engine.NewBool(false),
						"error":   engine.NewString("conflict: file is being accessed"),
					})
					return
				}

				globalFileTracker.acquire(o.path)

				switch o.op {
				case "read":
					data, err := os.ReadFile(o.path)
					if err != nil {
						results[idx] = engine.NewObject(map[string]engine.Value{
							"success": engine.NewBool(false),
							"error":   engine.NewString(err.Error()),
						})
					} else {
						results[idx] = engine.NewObject(map[string]engine.Value{
							"success": engine.NewBool(true),
							"data":    engine.NewString(string(data)),
						})
					}
				case "write":
					err := os.WriteFile(o.path, []byte(o.data), 0644)
					results[idx] = engine.NewObject(map[string]engine.Value{
						"success": engine.NewBool(err == nil),
					})
				default:
					results[idx] = engine.NewObject(map[string]engine.Value{
						"success": engine.NewBool(false),
						"error":   engine.NewString("unknown op: " + o.op),
					})
				}
			}(i, op)
		}

		wg.Wait()
		_, _ = ctx.VM().CallValue(callback, engine.NewArray(results))
	}()

	return engine.NewNull(), nil
}

// ==============================================================================
// 文件锁
// ==============================================================================

// fileLockValue 文件锁对象
type fileLockValue struct {
	path   string
	locked bool
	mu     sync.Mutex
}

// Type 返回类型标识
func (l *fileLockValue) Type() engine.ValueType           { return engine.TypeObject }
func (l *fileLockValue) IsNull() bool                     { return false }
func (l *fileLockValue) Bool() bool                       { return l.locked }
func (l *fileLockValue) Int() int64                       { return 0 }
func (l *fileLockValue) Float() float64                   { return 0 }
func (l *fileLockValue) String() string                   { return fmt.Sprintf("FileLock(%s)", l.path) }
func (l *fileLockValue) Stringify() string                { return l.String() }
func (l *fileLockValue) Array() []engine.Value            { return nil }
func (l *fileLockValue) Len() int                         { return 0 }
func (l *fileLockValue) Equals(v engine.Value) bool       { return false }
func (l *fileLockValue) Less(v engine.Value) bool         { return false }
func (l *fileLockValue) Greater(v engine.Value) bool      { return false }
func (l *fileLockValue) LessEqual(v engine.Value) bool    { return false }
func (l *fileLockValue) GreaterEqual(v engine.Value) bool { return false }
func (l *fileLockValue) ToBigInt() engine.Value           { return engine.NewInt(0) }
func (l *fileLockValue) ToBigDecimal() engine.Value       { return engine.NewFloat(0) }
func (l *fileLockValue) Add(v engine.Value) engine.Value  { return l }
func (l *fileLockValue) Sub(v engine.Value) engine.Value  { return l }
func (l *fileLockValue) Mul(v engine.Value) engine.Value  { return l }
func (l *fileLockValue) Div(v engine.Value) engine.Value  { return l }
func (l *fileLockValue) Mod(v engine.Value) engine.Value  { return l }
func (l *fileLockValue) Negate() engine.Value             { return l }

// Object 返回对象值
func (l *fileLockValue) Object() map[string]engine.Value {
	return map[string]engine.Value{
		"release": engine.NewFunc("release", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			l.mu.Lock()
			defer l.mu.Unlock()
			if l.locked {
				globalFileTracker.release(l.path)
				l.locked = false
			}
			return engine.NewBool(true), nil
		}),
		"path": engine.NewString(l.path),
	}
}

// builtinFileWithLock 获取文件锁并执行操作
// file_with_lock($path, fn($lock) { ... }) → null
func builtinFileWithLock(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("file_with_lock() expects 2 arguments: path, callback")
	}

	path := args[0].String()
	callback := args[1]

	if callback.Type() != engine.TypeFunc {
		return nil, fmt.Errorf("file_with_lock() expects function callback")
	}

	// 获取锁
	globalFileTracker.acquire(path)

	lock := &fileLockValue{
		path:   path,
		locked: true,
	}

	// 调用回调
	_, _ = ctx.VM().CallValue(callback, lock)

	return engine.NewNull(), nil
}

// FileIOAsyncSigs returns function signatures for REPL :doc command.
func FileIOAsyncSigs() map[string]string {
	return map[string]string{
		"file_get_async":    "file_get_async(path, fn(data)) → null  — Async read entire file",
		"file_put_async":    "file_put_async(path, data, fn()) → null  — Async write entire file",
		"file_append_async": "file_append_async(path, data, fn()) → null  — Async append to file",
		"file_get_bytes":    "file_get_bytes(path, fn(buffer)) → null  — Async read binary file",
		"file_put_bytes":    "file_put_bytes(path, buffer, fn()) → null  — Async write binary file",
		"file_read_lines":   "file_read_lines(path, fn(line), fn()) → null  — Stream read line by line",
		"file_read_chunks":  "file_read_chunks(path, chunkSize, fn(chunk), fn()) → null  — Stream read in chunks",
		"file_get_batch":    "file_get_batch(paths, fn(results)) → null  — Batch read multiple files",
		"file_put_batch":    "file_put_batch(items, fn(results)) → null  — Batch write multiple files",
		"file_parallel":     "file_parallel(ops, fn(results)) → null  — Parallel file operations",
		"file_with_lock":    "file_with_lock(path, fn(lock)) → null  — Acquire file lock and execute",
	}
}
