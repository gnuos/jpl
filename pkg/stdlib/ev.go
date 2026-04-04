package stdlib

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gnuos/jpl/engine"
)

// ==============================================================================
// EvRegistryValue 通用事件注册表
// ==============================================================================

// evHandler 通用事件处理器
type evHandler struct {
	id        int                // 唯一标识符
	eventType string             // 事件类型: "accept", "read", "write", "data", "change" 等
	source    engine.Value       // 事件源: socket, file, path 等
	callback  engine.Value       // 回调函数
	ctx       *engine.Context    // 执行上下文
	cancel    context.CancelFunc // 取消函数（用于停止 goroutine）
}

// EvRegistryValue 表示通用事件注册表
// 提供抽象的事件注册/注销/触发机制
// 各模块（net, fileio 等）通过此接口注册事件处理器
type EvRegistryValue struct {
	mu       sync.RWMutex
	handlers map[int]*evHandler // handler_id -> handler
	nextID   int

	// Event Loop 核心事件（保留）
	timers  map[int]*evTimer
	signals map[int]*evSignalHandler

	// context 管理
	ctx    context.Context
	cancel context.CancelFunc
}

// Type 返回类型标识
func (r *EvRegistryValue) Type() engine.ValueType { return engine.TypeObject }
func (r *EvRegistryValue) IsNull() bool           { return false }
func (r *EvRegistryValue) Bool() bool             { return true }
func (r *EvRegistryValue) Int() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return int64(len(r.handlers) + len(r.timers) + len(r.signals))
}
func (r *EvRegistryValue) Float() float64                   { return float64(r.Int()) }
func (r *EvRegistryValue) String() string                   { return fmt.Sprintf("EvRegistry(%d handlers)", r.Int()) }
func (r *EvRegistryValue) Stringify() string                { return r.String() }
func (r *EvRegistryValue) Array() []engine.Value            { return nil }
func (r *EvRegistryValue) Len() int                         { return int(r.Int()) }
func (r *EvRegistryValue) Equals(v engine.Value) bool       { return false }
func (r *EvRegistryValue) Less(v engine.Value) bool         { return false }
func (r *EvRegistryValue) Greater(v engine.Value) bool      { return false }
func (r *EvRegistryValue) LessEqual(v engine.Value) bool    { return false }
func (r *EvRegistryValue) GreaterEqual(v engine.Value) bool { return false }
func (r *EvRegistryValue) ToBigInt() engine.Value           { return engine.NewInt(0) }
func (r *EvRegistryValue) ToBigDecimal() engine.Value       { return engine.NewFloat(0) }
func (r *EvRegistryValue) Add(v engine.Value) engine.Value  { return r }
func (r *EvRegistryValue) Sub(v engine.Value) engine.Value  { return r }
func (r *EvRegistryValue) Mul(v engine.Value) engine.Value  { return r }
func (r *EvRegistryValue) Div(v engine.Value) engine.Value  { return r }
func (r *EvRegistryValue) Mod(v engine.Value) engine.Value  { return r }
func (r *EvRegistryValue) Negate() engine.Value             { return r }

// Count 返回注册的事件总数
func (r *EvRegistryValue) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.handlers) + len(r.timers) + len(r.signals)
}

// Context 返回 registry 的 context
func (r *EvRegistryValue) Context() context.Context {
	return r.ctx
}

// RegisterHandler 注册事件处理器（供各模块调用）
// 返回 handler_id，可用于后续注销
func (r *EvRegistryValue) RegisterHandler(eventType string, source engine.Value, callback engine.Value, ctx *engine.Context, cancel context.CancelFunc) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID
	r.nextID++

	r.handlers[id] = &evHandler{
		id:        id,
		eventType: eventType,
		source:    source,
		callback:  callback,
		ctx:       ctx,
		cancel:    cancel,
	}

	return id
}

// UnregisterHandler 注销指定 handler（供各模块调用）
func (r *EvRegistryValue) UnregisterHandler(id int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if handler, ok := r.handlers[id]; ok {
		if handler.cancel != nil {
			handler.cancel()
		}
		delete(r.handlers, id)
	}
}

// UnregisterBySource 注销指定 source 的所有事件
func (r *EvRegistryValue) UnregisterBySource(source engine.Value) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, handler := range r.handlers {
		if handler.source == source {
			if handler.cancel != nil {
				handler.cancel()
			}
			delete(r.handlers, id)
		}
	}
}

// UnregisterBySourceAndType 注销指定 source 和类型的事件
func (r *EvRegistryValue) UnregisterBySourceAndType(source engine.Value, eventType string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, handler := range r.handlers {
		if handler.source == source && handler.eventType == eventType {
			if handler.cancel != nil {
				handler.cancel()
			}
			delete(r.handlers, id)
		}
	}
}

// Emit 触发事件（供各模块调用）
// 查找匹配的 handler 并调用回调
func (r *EvRegistryValue) Emit(source engine.Value, eventType string, data ...engine.Value) {
	r.mu.RLock()
	var handlers []*evHandler
	for _, handler := range r.handlers {
		if handler.source == source && handler.eventType == eventType {
			handlers = append(handlers, handler)
		}
	}
	r.mu.RUnlock()

	// 调用回调
	for _, handler := range handlers {
		if handler.ctx != nil && handler.callback != nil {
			args := append([]engine.Value{source}, data...)
			_, _ = handler.ctx.VM().CallValue(handler.callback, args...)
		}
	}
}

// cancelAll 取消所有事件（防止 goroutine 泄漏）
func (r *EvRegistryValue) cancelAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 取消通用事件
	for _, handler := range r.handlers {
		if handler.cancel != nil {
			handler.cancel()
		}
	}
	// 取消定时器
	for _, t := range r.timers {
		if t.goroutine != nil {
			t.goroutine()
		}
	}
	// 取消信号
	for _, s := range r.signals {
		if s.goroutine != nil {
			s.goroutine()
		}
	}
	// 取消 registry 自己的 context
	if r.cancel != nil {
		r.cancel()
	}
}

// Object 返回对象值，包含所有注册表方法
func (r *EvRegistryValue) Object() map[string]engine.Value {
	return map[string]engine.Value{
		// 通用事件接口
		"on": engine.NewFunc("on", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			// on(event_type, source, callback)
			if len(args) != 3 {
				return nil, fmt.Errorf("on() expects 3 arguments: event_type, source, callback")
			}
			eventType := args[0].String()
			source := args[1]
			callback := args[2]

			loopCtx := r.Context()
			if loopCtx == nil {
				loopCtx = context.Background()
			}
			_, gCancel := context.WithCancel(loopCtx)

			id := r.RegisterHandler(eventType, source, callback, ctx, gCancel)

			// 返回 handler_id 和 context（供各模块启动 goroutine）
			return engine.NewInt(int64(id)), nil
		}),
		"off": engine.NewFunc("off", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			// off(source) 或 off(source, event_type)
			if len(args) < 1 {
				return nil, fmt.Errorf("off() expects at least 1 argument: source")
			}
			source := args[0]

			if len(args) >= 2 {
				eventType := args[1].String()
				r.UnregisterBySourceAndType(source, eventType)
			} else {
				r.UnregisterBySource(source)
			}

			return engine.NewBool(true), nil
		}),
		"emit": engine.NewFunc("emit", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			// emit(source, event_type, data...)
			if len(args) < 2 {
				return nil, fmt.Errorf("emit() expects at least 2 arguments: source, event_type")
			}
			source := args[0]
			eventType := args[1].String()
			data := args[2:]

			r.Emit(source, eventType, data...)

			return engine.NewBool(true), nil
		}),

		// 保留的便捷方法（向后兼容）
		"on_accept": engine.NewFunc("on_accept", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			// on_accept(server, callback) - 语法糖
			if len(args) != 2 {
				return nil, fmt.Errorf("on_accept() expects 2 arguments: server, callback")
			}
			return builtinNetOnAccept(ctx, append([]engine.Value{r}, args...))
		}),
		"on_read": engine.NewFunc("on_read", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			// on_read(socket, callback) - 语法糖
			if len(args) != 2 {
				return nil, fmt.Errorf("on_read() expects 2 arguments: socket, callback")
			}
			return builtinNetOnRead(ctx, append([]engine.Value{r}, args...))
		}),
		"on_write": engine.NewFunc("on_write", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			// on_write(socket, callback) - 语法糖
			if len(args) != 2 {
				return nil, fmt.Errorf("on_write() expects 2 arguments: socket, callback")
			}
			return builtinNetOnWrite(ctx, append([]engine.Value{r}, args...))
		}),
		"on_timer": engine.NewFunc("on_timer", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			// on_timer(interval_us, callback)
			if len(args) != 2 {
				return nil, fmt.Errorf("on_timer() expects 2 arguments: interval_us, callback")
			}
			return builtinRegistryOnTimer(ctx, append([]engine.Value{r}, args...))
		}),
		"on_timer_once": engine.NewFunc("on_timer_once", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("on_timer_once() expects 2 arguments: delay_us, callback")
			}
			return builtinRegistryOnTimerOnce(ctx, append([]engine.Value{r}, args...))
		}),
		"on_signal": engine.NewFunc("on_signal", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			// on_signal(signal_num, callback)
			if len(args) != 2 {
				return nil, fmt.Errorf("on_signal() expects 2 arguments: signal_num, callback")
			}
			return builtinRegistryOnSignal(ctx, append([]engine.Value{r}, args...))
		}),
		"off_timer": engine.NewFunc("off_timer", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("off_timer() expects 1 argument: timer_id")
			}
			return builtinRegistryOffTimer(ctx, append([]engine.Value{r}, args...))
		}),
		"off_signal": engine.NewFunc("off_signal", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("off_signal() expects 1 argument: signal_num")
			}
			return builtinRegistryOffSignal(ctx, append([]engine.Value{r}, args...))
		}),
		"clear": engine.NewFunc("clear", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			r.cancelAll()
			r.mu.Lock()
			r.handlers = make(map[int]*evHandler)
			r.timers = make(map[int]*evTimer)
			r.signals = make(map[int]*evSignalHandler)
			r.mu.Unlock()
			return engine.NewBool(true), nil
		}),
		"count": engine.NewFunc("count", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
			return engine.NewInt(int64(r.Count())), nil
		}),
	}
}

// NewEvRegistry 创建新的事件注册表
func NewEvRegistry() *EvRegistryValue {
	ctx, cancel := context.WithCancel(context.Background())
	return &EvRegistryValue{
		handlers: make(map[int]*evHandler),
		timers:   make(map[int]*evTimer),
		signals:  make(map[int]*evSignalHandler),
		nextID:   1,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// builtinEvRegistryNew 创建事件注册表
func builtinEvRegistryNew(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	return NewEvRegistry(), nil
}

// ==============================================================================
// evTimer 和 evSignalHandler 定义（Event Loop 核心）
// ==============================================================================

// evTimer 定时器结构
type evTimer struct {
	id        int
	interval  time.Duration
	periodic  bool
	nextRun   time.Time
	active    bool
	ctx       *engine.Context
	callback  engine.Value
	goroutine context.CancelFunc
}

// evSignalHandler 信号处理器结构
type evSignalHandler struct {
	ctx       *engine.Context
	callback  engine.Value
	goroutine context.CancelFunc
}

// ==============================================================================
// EvLoopValue 事件循环
// ==============================================================================

// EvLoopValue 表示事件循环对象
type EvLoopValue struct {
	mu         sync.RWMutex
	registries []*EvRegistryValue
	running    bool
	ctx        context.Context
	cancel     context.CancelFunc
	signalChan chan os.Signal
}

// Type 返回类型标识
func (l *EvLoopValue) Type() engine.ValueType { return engine.TypeObject }
func (l *EvLoopValue) IsNull() bool           { return false }
func (l *EvLoopValue) Bool() bool             { return l.running }
func (l *EvLoopValue) Int() int64             { return int64(len(l.registries)) }
func (l *EvLoopValue) Float() float64         { return float64(l.Int()) }
func (l *EvLoopValue) String() string {
	status := "stopped"
	if l.running {
		status = "running"
	}
	return fmt.Sprintf("EvLoop(%s, %d registries)", status, len(l.registries))
}
func (l *EvLoopValue) Stringify() string                { return l.String() }
func (l *EvLoopValue) Array() []engine.Value            { return nil }
func (l *EvLoopValue) Object() map[string]engine.Value  { return nil }
func (l *EvLoopValue) Len() int                         { return len(l.registries) }
func (l *EvLoopValue) Equals(v engine.Value) bool       { return false }
func (l *EvLoopValue) Less(v engine.Value) bool         { return false }
func (l *EvLoopValue) Greater(v engine.Value) bool      { return false }
func (l *EvLoopValue) LessEqual(v engine.Value) bool    { return false }
func (l *EvLoopValue) GreaterEqual(v engine.Value) bool { return false }
func (l *EvLoopValue) ToBigInt() engine.Value           { return engine.NewInt(0) }
func (l *EvLoopValue) ToBigDecimal() engine.Value       { return engine.NewFloat(0) }
func (l *EvLoopValue) Add(v engine.Value) engine.Value  { return l }
func (l *EvLoopValue) Sub(v engine.Value) engine.Value  { return l }
func (l *EvLoopValue) Mul(v engine.Value) engine.Value  { return l }
func (l *EvLoopValue) Div(v engine.Value) engine.Value  { return l }
func (l *EvLoopValue) Mod(v engine.Value) engine.Value  { return l }
func (l *EvLoopValue) Negate() engine.Value             { return l }

// NewEvLoop 创建新的事件循环
func NewEvLoop() (*EvLoopValue, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &EvLoopValue{
		registries: make([]*EvRegistryValue, 0),
		ctx:        ctx,
		cancel:     cancel,
		signalChan: make(chan os.Signal, 1),
	}, nil
}

// RegisterEv 注册事件循环函数到引擎
func RegisterEv(e *engine.Engine) {
	e.RegisterFunc("ev_registry_new", builtinEvRegistryNew)
	e.RegisterFunc("ev_loop_new", builtinEvLoopNew)
	e.RegisterFunc("ev_attach", builtinEvAttach)
	e.RegisterFunc("ev_run", builtinEvRun)
	e.RegisterFunc("ev_run_once", builtinEvRunOnce)
	e.RegisterFunc("ev_stop", builtinEvStop)
	e.RegisterFunc("ev_is_running", builtinEvIsRunning)
	e.RegisterFunc("ev_timer_now", builtinEvTimerNow)

	// 模块注册 — import "ev" 可用
	e.RegisterModule("ev", map[string]engine.GoFunction{
		"registry_new": builtinEvRegistryNew,
		"loop_new":     builtinEvLoopNew,
		"attach":       builtinEvAttach,
		"run":          builtinEvRun,
		"run_once":     builtinEvRunOnce,
		"stop":         builtinEvStop,
		"is_running":   builtinEvIsRunning,
		"timer_now":    builtinEvTimerNow,
	})
}

// EvNames 返回 ev 函数名称列表
func EvNames() []string {
	return []string{
		"ev_registry_new", "ev_loop_new", "ev_attach",
		"ev_run", "ev_run_once", "ev_stop", "ev_is_running",
		"ev_timer_now",
	}
}

// builtinEvLoopNew 创建事件循环
func builtinEvLoopNew(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	loop, err := NewEvLoop()
	if err != nil {
		return nil, fmt.Errorf("ev_loop_new() failed: %v", err)
	}
	return loop, nil
}

// builtinEvAttach 附加注册表到循环
func builtinEvAttach(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("ev_attach() expects 2 arguments, got %d", len(args))
	}

	loop, ok := args[0].(*EvLoopValue)
	if !ok {
		return nil, fmt.Errorf("ev_attach() expects EvLoop, got %s", args[0].Type())
	}

	registry, ok := args[1].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("ev_attach() expects EvRegistry, got %s", args[1].Type())
	}

	loop.mu.Lock()
	loop.registries = append(loop.registries, registry)
	loop.mu.Unlock()

	// 更新 registry 的 context 为 loop 的 context
	registry.ctx, registry.cancel = context.WithCancel(loop.ctx)

	// 设置信号处理
	registry.mu.RLock()
	if len(registry.signals) > 0 {
		sigs := make([]os.Signal, 0, len(registry.signals))
		for sigNum := range registry.signals {
			sigs = append(sigs, syscall.Signal(sigNum))
		}
		signal.Notify(loop.signalChan, sigs...)
	}
	registry.mu.RUnlock()

	// 启动定时器处理 goroutine
	go timerLoop(registry)

	// 启动信号处理 goroutine
	go signalLoop(loop, registry)

	return engine.NewBool(true), nil
}

// timerLoop 定时器处理循环
func timerLoop(registry *EvRegistryValue) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-registry.ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			registry.mu.Lock()
			for id, timer := range registry.timers {
				if timer.active && now.After(timer.nextRun) {
					if timer.ctx != nil && timer.callback != nil {
						_, _ = timer.ctx.VM().CallValue(timer.callback)
					}
					if timer.periodic {
						timer.nextRun = now.Add(timer.interval)
					} else {
						timer.active = false
						delete(registry.timers, id)
					}
				}
			}
			registry.mu.Unlock()
		}
	}
}

// signalLoop 信号处理循环
func signalLoop(loop *EvLoopValue, registry *EvRegistryValue) {
	for {
		select {
		case <-registry.ctx.Done():
			return
		case sig := <-loop.signalChan:
			registry.mu.RLock()
			if sigHandler, ok := registry.signals[int(sig.(syscall.Signal))]; ok {
				if sigHandler.ctx != nil && sigHandler.callback != nil {
					_, _ = sigHandler.ctx.VM().CallValue(sigHandler.callback)
				}
			}
			registry.mu.RUnlock()
		}
	}
}

// builtinEvRun 运行事件循环（阻塞模式）
func builtinEvRun(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ev_run() expects 1 argument, got %d", len(args))
	}

	loop, ok := args[0].(*EvLoopValue)
	if !ok {
		return nil, fmt.Errorf("ev_run() expects EvLoop, got %s", args[0].Type())
	}

	loop.mu.Lock()
	if loop.running {
		loop.mu.Unlock()
		return nil, fmt.Errorf("ev_run() loop already running")
	}
	loop.running = true
	loop.mu.Unlock()

	defer func() {
		loop.mu.Lock()
		loop.running = false
		loop.mu.Unlock()
	}()

	// 等待 context 取消
	<-loop.ctx.Done()

	// 清理：取消所有注册表的事件
	loop.mu.RLock()
	for _, registry := range loop.registries {
		registry.cancelAll()
	}
	loop.mu.RUnlock()

	return engine.NewNull(), nil
}

// builtinEvRunOnce 运行一次事件循环（非阻塞模式）
func builtinEvRunOnce(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	return engine.NewBool(true), nil
}

// builtinEvStop 停止事件循环
func builtinEvStop(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ev_stop() expects 1 argument, got %d", len(args))
	}

	loop, ok := args[0].(*EvLoopValue)
	if !ok {
		return nil, fmt.Errorf("ev_stop() expects EvLoop, got %s", args[0].Type())
	}

	if loop.cancel != nil {
		loop.cancel()
	}

	return engine.NewBool(true), nil
}

// builtinEvIsRunning 检查循环是否运行中
func builtinEvIsRunning(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ev_is_running() expects 1 argument, got %d", len(args))
	}

	loop, ok := args[0].(*EvLoopValue)
	if !ok {
		return nil, fmt.Errorf("ev_is_running() expects EvLoop, got %s", args[0].Type())
	}

	loop.mu.RLock()
	running := loop.running
	loop.mu.RUnlock()

	return engine.NewBool(running), nil
}

// builtinEvTimerNow 获取当前时间（微秒级时间戳）
func builtinEvTimerNow(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	now := time.Now()
	microseconds := now.UnixNano() / 1000
	return engine.NewInt(microseconds), nil
}

// EvSigs returns function signatures for REPL :doc command.
func EvSigs() map[string]string {
	return map[string]string{
		"ev_registry_new": "ev_registry_new() → EvRegistry  — Create event registry",
		"ev_loop_new":     "ev_loop_new() → EvLoop  — Create event loop",
		"ev_attach":       "ev_attach(loop, registry) → bool  — Attach registry to loop",
		"ev_run":          "ev_run(loop) → null  — Run event loop (blocking)",
		"ev_run_once":     "ev_run_once() → bool  — Run one event iteration",
		"ev_stop":         "ev_stop(loop) → bool  — Stop event loop",
		"ev_is_running":   "ev_is_running(loop) → bool  — Check if loop is running",
		"ev_timer_now":    "ev_timer_now() → int  — Get current time in microseconds",
	}
}

// ==============================================================================
// 定时器和信号注册函数（内部使用）
// ==============================================================================

// builtinRegistryOnTimer 注册周期性定时器
func builtinRegistryOnTimer(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("on_timer() expects 3 arguments, got %d", len(args))
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("on_timer() expects EvRegistry, got %s", args[0].Type())
	}

	var us int64
	if args[1].Type() == engine.TypeInt {
		us = args[1].Int()
	} else if args[1].Type() == engine.TypeFloat {
		us = int64(args[1].Float())
	} else {
		return nil, fmt.Errorf("on_timer() expects int microseconds, got %s", args[1].Type())
	}

	fn := args[2]
	if !ok {
		return nil, fmt.Errorf("on_timer() expects function handler, got %s", args[2].Type())
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	timerID := registry.nextID
	registry.nextID++

	registry.timers[timerID] = &evTimer{
		id:       timerID,
		interval: time.Duration(us) * time.Microsecond,
		periodic: true,
		nextRun:  time.Now().Add(time.Duration(us) * time.Microsecond),
		active:   true,
		ctx:      ctx,
		callback: fn,
	}

	return engine.NewInt(int64(timerID)), nil
}

// builtinRegistryOnTimerOnce 注册一次性定时器
func builtinRegistryOnTimerOnce(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("on_timer_once() expects 3 arguments, got %d", len(args))
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("on_timer_once() expects EvRegistry, got %s", args[0].Type())
	}

	var us int64
	if args[1].Type() == engine.TypeInt {
		us = args[1].Int()
	} else if args[1].Type() == engine.TypeFloat {
		us = int64(args[1].Float())
	} else {
		return nil, fmt.Errorf("on_timer_once() expects int microseconds, got %s", args[1].Type())
	}

	fn := args[2]
	if !ok {
		return nil, fmt.Errorf("on_timer_once() expects function handler, got %s", args[2].Type())
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	timerID := registry.nextID
	registry.nextID++

	registry.timers[timerID] = &evTimer{
		id:       timerID,
		interval: time.Duration(us) * time.Microsecond,
		periodic: false,
		nextRun:  time.Now().Add(time.Duration(us) * time.Microsecond),
		active:   true,
		ctx:      ctx,
		callback: fn,
	}

	return engine.NewInt(int64(timerID)), nil
}

// builtinRegistryOnSignal 注册信号处理器
func builtinRegistryOnSignal(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("on_signal() expects 3 arguments, got %d", len(args))
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("on_signal() expects EvRegistry, got %s", args[0].Type())
	}

	var sigNum int
	if args[1].Type() == engine.TypeInt {
		sigNum = int(args[1].Int())
	} else {
		return nil, fmt.Errorf("on_signal() expects int signal number, got %s", args[1].Type())
	}

	fn := args[2]
	if !ok {
		return nil, fmt.Errorf("on_signal() expects function handler, got %s", args[2].Type())
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.signals[sigNum] = &evSignalHandler{
		ctx:      ctx,
		callback: fn,
	}

	return engine.NewBool(true), nil
}

// builtinRegistryOffTimer 注销指定定时器
func builtinRegistryOffTimer(_ *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("off_timer() expects 2 arguments, got %d", len(args))
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("off_timer() expects EvRegistry, got %s", args[0].Type())
	}

	var timerID int
	if args[1].Type() == engine.TypeInt {
		timerID = int(args[1].Int())
	} else {
		return nil, fmt.Errorf("off_timer() expects int timer_id, got %s", args[1].Type())
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if timer, ok := registry.timers[timerID]; ok {
		if timer.goroutine != nil {
			timer.goroutine()
		}
		delete(registry.timers, timerID)
	}

	return engine.NewBool(true), nil
}

// builtinRegistryOffSignal 注销信号处理器
func builtinRegistryOffSignal(_ *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("off_signal() expects 2 arguments, got %d", len(args))
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("off_signal() expects EvRegistry, got %s", args[0].Type())
	}

	var sigNum int
	if args[1].Type() == engine.TypeInt {
		sigNum = int(args[1].Int())
	} else {
		return nil, fmt.Errorf("off_signal() expects int signal number, got %s", args[1].Type())
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if sig, ok := registry.signals[sigNum]; ok {
		if sig.goroutine != nil {
			sig.goroutine()
		}
		delete(registry.signals, sigNum)
	}

	return engine.NewBool(true), nil
}
