package stdlib

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gnuos/jpl/engine"
)

// ==============================================================================
// 进程扩展函数
// 提供系统命令执行、环境变量管理、进程信息查询等功能
// ==============================================================================

// RegisterProcessExt 注册进程扩展函数到引擎
func RegisterProcessExt(e *engine.Engine) {
	// 命令执行（P0）
	e.RegisterFunc("exec", builtinExec)
	e.RegisterFunc("system", builtinSystem)
	e.RegisterFunc("shell_exec", builtinShellExec)

	// 环境变量（P0）
	e.RegisterFunc("getenv", builtinGetenv)
	e.RegisterFunc("setenv", builtinSetenv)

	// 进程信息（P0）
	e.RegisterFunc("getppid", builtinGetppid)
	e.RegisterFunc("tmpdir", builtinTmpdir)
	e.RegisterFunc("hostname", builtinHostname)

	// 进程管道（P1）
	e.RegisterFunc("proc_open", builtinProcOpen)
	e.RegisterFunc("proc_close", builtinProcClose)
	e.RegisterFunc("proc_wait", builtinProcWait)
	e.RegisterFunc("proc_status", builtinProcStatus)

	// 其他（P1）
	e.RegisterFunc("getlogin", builtinGetlogin)
	e.RegisterFunc("usleep", builtinUsleep)
	e.RegisterFunc("putenv", builtinPutenv)

	// 进阶功能（P2）
	e.RegisterFunc("spawn", builtinSpawn)
	e.RegisterFunc("kill", builtinKill)
	e.RegisterFunc("waitpid", builtinWaitpid)
	e.RegisterFunc("fork", builtinFork)
	e.RegisterFunc("pipe", builtinPipe)

	// 高级功能（P3）
	e.RegisterFunc("sigwait", builtinSigwait)

	// 模块注册 — import "process" 可用
	e.RegisterModule("process", map[string]engine.GoFunction{
		// P0
		"exec":       builtinExec,
		"system":     builtinSystem,
		"shell_exec": builtinShellExec,
		"getenv":     builtinGetenv,
		"setenv":     builtinSetenv,
		"getppid":    builtinGetppid,
		"tmpdir":     builtinTmpdir,
		"hostname":   builtinHostname,
		// P1
		"proc_open":   builtinProcOpen,
		"proc_close":  builtinProcClose,
		"proc_wait":   builtinProcWait,
		"proc_status": builtinProcStatus,
		"getlogin":    builtinGetlogin,
		"usleep":      builtinUsleep,
		"putenv":      builtinPutenv,
		// P2
		"spawn":   builtinSpawn,
		"kill":    builtinKill,
		"waitpid": builtinWaitpid,
		"fork":    builtinFork,
		"pipe":    builtinPipe,
		// P3
		"sigwait": builtinSigwait,
	})
}

// ProcessExtNames 返回进程扩展函数名称列表
func ProcessExtNames() []string {
	return []string{
		// P0
		"exec", "system", "shell_exec",
		"getenv", "setenv",
		"getppid", "tmpdir", "hostname",
		// P1
		"proc_open", "proc_close", "proc_wait", "proc_status",
		"getlogin", "usleep", "putenv",
		// P2
		"spawn", "kill", "waitpid", "fork", "pipe",
		// P3
		"sigwait",
	}
}

// ==============================================================================
// 命令执行函数
// ==============================================================================

// builtinExec 执行系统命令，返回输出字符串
// exec($cmd) → string
// exec($cmd, $args) → string
//
// 参数：
//   - args[0]: 命令字符串
//   - args[1]: 可选，参数数组
//
// 返回值：
//   - 命令输出（去除末尾换行符）
//   - null：执行失败
//
// 示例：
//
//	exec("ls -la")              // 返回 ls 输出
//	exec("echo", ["hello"])     // 返回 "hello"
//	exec("cat /etc/hosts")      // 返回文件内容
func builtinExec(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("exec() expects 1-2 arguments, got %d", len(args))
	}

	cmdStr := args[0].String()

	var cmd *exec.Cmd
	if len(args) == 2 && args[1].Type() == engine.TypeArray {
		// exec("echo", ["hello", "world"])
		cmdArgs := make([]string, 0)
		for _, arg := range args[1].Array() {
			cmdArgs = append(cmdArgs, arg.String())
		}
		cmd = exec.Command(cmdStr, cmdArgs...)
	} else {
		// exec("ls -la") - 使用 shell 解析
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	output, err := cmd.Output()
	if err != nil {
		// 返回 null 表示执行失败
		return engine.NewNull(), nil
	}

	// 去除末尾换行符
	result := strings.TrimRight(string(output), "\n\r")
	return engine.NewString(result), nil
}

// builtinSystem 执行系统命令，返回退出码
// system($cmd) → int
//
// 参数：
//   - args[0]: 命令字符串
//
// 返回值：
//   - 退出码（0 表示成功）
//   - -1：执行失败
//
// 示例：
//
//	system("ls /tmp")           // 返回 0（成功）
//	system("ls /nonexistent")   // 返回非 0（失败）
func builtinSystem(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("system() expects 1 argument, got %d", len(args))
	}

	cmdStr := args[0].String()
	cmd := exec.Command("sh", "-c", cmdStr)

	// 连接标准输出和标准错误
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return engine.NewInt(int64(exitErr.ExitCode())), nil
		}
		return engine.NewInt(-1), nil
	}

	return engine.NewInt(0), nil
}

// builtinShellExec 通过 shell 执行命令并返回完整输出
// shell_exec($cmd) → string
//
// 参数：
//   - args[0]: 命令字符串
//
// 返回值：
//   - 命令完整输出（包含末尾换行符）
//   - null：执行失败
//
// 与 exec() 的区别：
//   - shell_exec 保留末尾换行符
//   - shell_exec 返回 null 而非空字符串
//
// 示例：
//
//	shell_exec("ls -la /tmp")
//	shell_exec("cat /etc/passwd | head -5")
func builtinShellExec(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("shell_exec() expects 1 argument, got %d", len(args))
	}

	cmdStr := args[0].String()
	cmd := exec.Command("sh", "-c", cmdStr)

	output, err := cmd.Output()
	if err != nil {
		return engine.NewNull(), nil
	}

	return engine.NewString(string(output)), nil
}

// ==============================================================================
// 环境变量函数
// ==============================================================================

// builtinGetenv 获取环境变量
// getenv($name) → string
// getenv($name, $default) → string
//
// 参数：
//   - args[0]: 环境变量名
//   - args[1]: 可选，默认值（变量不存在时返回）
//
// 返回值：
//   - 环境变量值
//   - 默认值（如果变量不存在且指定了默认值）
//   - null（如果变量不存在且未指定默认值）
//
// 示例：
//
//	getenv("HOME")              // → "/home/user"
//	getenv("PATH")              // → "/usr/bin:/bin:..."
//	getenv("MY_VAR", "default") // → "default"（如果 MY_VAR 不存在）
func builtinGetenv(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("getenv() expects 1-2 arguments, got %d", len(args))
	}

	name := args[0].String()
	value, exists := os.LookupEnv(name)

	if exists {
		return engine.NewString(value), nil
	}

	// 返回默认值
	if len(args) == 2 {
		return args[1], nil
	}

	return engine.NewNull(), nil
}

// builtinSetenv 设置环境变量
// setenv($name, $value) → bool
//
// 参数：
//   - args[0]: 环境变量名
//   - args[1]: 环境变量值
//
// 返回值：
//   - true：设置成功
//   - false：设置失败
//
// 示例：
//
//	setenv("MY_VAR", "hello")
//	$value = getenv("MY_VAR")  // → "hello"
func builtinSetenv(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("setenv() expects 2 arguments, got %d", len(args))
	}

	name := args[0].String()
	value := args[1].String()

	err := os.Setenv(name, value)
	return engine.NewBool(err == nil), nil
}

// ==============================================================================
// 进程信息函数
// ==============================================================================

// builtinGetppid 获取父进程 ID
// getppid() → int
//
// 返回值：
//   - 父进程 ID
//
// 示例：
//
//	$pid = getpid()
//	$ppid = getppid()
//	println "PID: #{$pid}, Parent: #{$ppid}"
func builtinGetppid(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("getppid() expects 0 arguments, got %d", len(args))
	}

	ppid := os.Getppid()
	return engine.NewInt(int64(ppid)), nil
}

// builtinTmpdir 获取系统临时目录
// tmpdir() → string
//
// 返回值：
//   - 临时目录路径（如 "/tmp"）
//
// 示例：
//
//	$tmp = tmpdir()
//	println "Temp dir: #{$tmp}"  // → "Temp dir: /tmp"
//
//	// 创建临时文件
//	$file = joinPath(tmpdir(), "myapp_#{$pid}.tmp")
func builtinTmpdir(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("tmpdir() expects 0 arguments, got %d", len(args))
	}

	tmpDir := os.TempDir()
	return engine.NewString(tmpDir), nil
}

// builtinHostname 获取主机名
// hostname() → string
//
// 返回值：
//   - 主机名
//
// 示例：
//
//	$host = hostname()
//	println "Running on: #{$host}"
func builtinHostname(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("hostname() expects 0 arguments, got %d", len(args))
	}

	name, err := os.Hostname()
	if err != nil {
		return engine.NewNull(), nil
	}

	return engine.NewString(name), nil
}

// ==============================================================================
// 进程管道对象和函数（P1）
// ==============================================================================

// ProcessValue 表示一个运行中的进程对象
// 用于 proc_open 创建的子进程，支持等待、获取状态、关闭等操作
type ProcessValue struct {
	cmd      *exec.Cmd
	stdin    *os.File
	stdout   *os.File
	stderr   *os.File
	pid      int
	exited   bool
	exitCode int
}

// Type 返回类型标识
func (p *ProcessValue) Type() engine.ValueType { return engine.TypeObject }
func (p *ProcessValue) IsNull() bool           { return false }
func (p *ProcessValue) Bool() bool             { return !p.exited }
func (p *ProcessValue) Int() int64             { return int64(p.pid) }
func (p *ProcessValue) Float() float64         { return float64(p.pid) }
func (p *ProcessValue) String() string {
	if p.exited {
		return fmt.Sprintf("Process(pid=%d, exited=%d)", p.pid, p.exitCode)
	}
	return fmt.Sprintf("Process(pid=%d, running)", p.pid)
}
func (p *ProcessValue) Stringify() string                { return p.String() }
func (p *ProcessValue) Array() []engine.Value            { return nil }
func (p *ProcessValue) Len() int                         { return 0 }
func (p *ProcessValue) Equals(v engine.Value) bool       { return false }
func (p *ProcessValue) Less(v engine.Value) bool         { return false }
func (p *ProcessValue) Greater(v engine.Value) bool      { return false }
func (p *ProcessValue) LessEqual(v engine.Value) bool    { return false }
func (p *ProcessValue) GreaterEqual(v engine.Value) bool { return false }
func (p *ProcessValue) ToBigInt() engine.Value           { return engine.NewInt(0) }
func (p *ProcessValue) ToBigDecimal() engine.Value       { return engine.NewFloat(0) }
func (p *ProcessValue) Add(v engine.Value) engine.Value  { return p }
func (p *ProcessValue) Sub(v engine.Value) engine.Value  { return p }
func (p *ProcessValue) Mul(v engine.Value) engine.Value  { return p }
func (p *ProcessValue) Div(v engine.Value) engine.Value  { return p }
func (p *ProcessValue) Mod(v engine.Value) engine.Value  { return p }
func (p *ProcessValue) Negate() engine.Value             { return p }

// Object 返回对象值，包含进程的方法和属性
func (p *ProcessValue) Object() map[string]engine.Value {
	obj := map[string]engine.Value{
		"pid": engine.NewInt(int64(p.pid)),
	}

	if p.stdin != nil {
		obj["stdin"] = engine.NewInt(int64(p.stdin.Fd()))
	}
	if p.stdout != nil {
		obj["stdout"] = engine.NewInt(int64(p.stdout.Fd()))
	}
	if p.stderr != nil {
		obj["stderr"] = engine.NewInt(int64(p.stderr.Fd()))
	}

	return obj
}

// builtinProcOpen 执行命令并创建进程管道
// proc_open($cmd, $opts) → Process
//
// 参数：
//   - args[0]: 命令字符串
//   - args[1]: 可选，选项对象
//   - stdin: "pipe" | "null" (默认不创建)
//   - stdout: "pipe" | "null" (默认不创建)
//   - stderr: "pipe" | "null" (默认不创建)
//
// 返回值：
//   - Process 对象，包含 pid
//   - null：执行失败
//
// 示例：
//
//	$proc = proc_open("sort", {stdout: "pipe"})
//	$code = proc_wait($proc)
func builtinProcOpen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("proc_open() expects 1-2 arguments, got %d", len(args))
	}

	cmdStr := args[0].String()
	cmd := exec.Command("sh", "-c", cmdStr)

	// 解析选项
	opts := make(map[string]string)
	if len(args) == 2 && args[1].Type() == engine.TypeObject {
		for k, v := range args[1].Object() {
			opts[k] = v.String()
		}
	}

	// 设置 stdin
	if mode, ok := opts["stdin"]; ok {
		if mode == "null" {
			cmd.Stdin = nil
		}
	}

	// 设置 stdout
	if mode, ok := opts["stdout"]; ok {
		if mode == "null" {
			cmd.Stdout = nil
		}
	}

	// 设置 stderr
	if mode, ok := opts["stderr"]; ok {
		if mode == "null" {
			cmd.Stderr = nil
		}
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return engine.NewNull(), nil
	}

	return &ProcessValue{
		cmd: cmd,
		pid: cmd.Process.Pid,
	}, nil
}

// builtinProcClose 关闭进程管道
// proc_close($proc) → int
//
// 参数：
//   - args[0]: Process 对象
//
// 返回值：
//   - 进程退出码
//
// 示例：
//
//	$proc = proc_open("echo hello", {stdout: "pipe"})
//	$code = proc_close($proc)
func builtinProcClose(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("proc_close() expects 1 argument, got %d", len(args))
	}

	proc, ok := args[0].(*ProcessValue)
	if !ok {
		return nil, fmt.Errorf("proc_close() expects Process, got %s", args[0].Type())
	}

	// 关闭管道
	if proc.stdin != nil {
		proc.stdin.Close()
	}
	if proc.stdout != nil {
		proc.stdout.Close()
	}
	if proc.stderr != nil {
		proc.stderr.Close()
	}

	// 等待进程结束
	err := proc.cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			proc.exitCode = exitErr.ExitCode()
		} else {
			proc.exitCode = -1
		}
	} else {
		proc.exitCode = 0
	}
	proc.exited = true

	return engine.NewInt(int64(proc.exitCode)), nil
}

// builtinProcWait 等待进程结束
// proc_wait($proc) → int
//
// 参数：
//   - args[0]: Process 对象
//
// 返回值：
//   - 进程退出码
//
// 示例：
//
//	$proc = proc_open("sleep 1")
//	$code = proc_wait($proc)
func builtinProcWait(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("proc_wait() expects 1 argument, got %d", len(args))
	}

	proc, ok := args[0].(*ProcessValue)
	if !ok {
		return nil, fmt.Errorf("proc_wait() expects Process, got %s", args[0].Type())
	}

	if proc.exited {
		return engine.NewInt(int64(proc.exitCode)), nil
	}

	// 等待进程结束
	state, err := proc.cmd.Process.Wait()
	if err != nil {
		proc.exited = true
		proc.exitCode = -1
		return engine.NewInt(-1), nil
	}

	proc.exited = true
	if state.Exited() {
		proc.exitCode = state.ExitCode()
	} else {
		proc.exitCode = -1
	}

	return engine.NewInt(int64(proc.exitCode)), nil
}

// builtinProcStatus 获取进程状态
// proc_status($proc) → object
//
// 参数：
//   - args[0]: Process 对象
//
// 返回值：
//   - 状态对象：
//   - pid: 进程 ID
//   - running: 是否正在运行
//   - exited: 是否已退出
//   - exit_code: 退出码（仅在退出后有效）
//
// 示例：
//
//	$proc = proc_open("sleep 10")
//	$status = proc_status($proc)
//	println "PID: #{$status.pid}, Running: #{$status.running}"
func builtinProcStatus(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("proc_status() expects 1 argument, got %d", len(args))
	}

	proc, ok := args[0].(*ProcessValue)
	if !ok {
		return nil, fmt.Errorf("proc_status() expects Process, got %s", args[0].Type())
	}

	result := map[string]engine.Value{
		"pid":     engine.NewInt(int64(proc.pid)),
		"running": engine.NewBool(!proc.exited),
		"exited":  engine.NewBool(proc.exited),
	}

	if proc.exited {
		result["exit_code"] = engine.NewInt(int64(proc.exitCode))
	}

	return engine.NewObject(result), nil
}

// ==============================================================================
// 其他 P1 函数
// ==============================================================================

// builtinGetlogin 获取当前登录用户名
// getlogin() → string
//
// 返回值：
//   - 登录用户名
//   - null：获取失败
//
// 示例：
//
//	$user = getlogin()
//	println "Logged in as: #{$user}"
func builtinGetlogin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("getlogin() expects 0 arguments, got %d", len(args))
	}

	// 尝试通过环境变量获取
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("LOGNAME")
	}
	if user == "" {
		// 尝试通过 whoami 命令获取
		cmd := exec.Command("whoami")
		output, err := cmd.Output()
		if err == nil {
			user = strings.TrimSpace(string(output))
		}
	}

	if user == "" {
		return engine.NewNull(), nil
	}

	return engine.NewString(user), nil
}

// builtinUsleep 暂停执行指定微秒数
// usleep($microseconds) → null
//
// 参数：
//   - args[0]: 微秒数（1秒 = 1000000微秒）
//
// 示例：
//
//	usleep(500000)  // 暂停 0.5 秒
//	usleep(1000000) // 暂停 1 秒
func builtinUsleep(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("usleep() expects 1 argument, got %d", len(args))
	}

	var us int64
	if args[0].Type() == engine.TypeInt {
		us = args[0].Int()
	} else if args[0].Type() == engine.TypeFloat {
		us = int64(args[0].Float())
	} else {
		return nil, fmt.Errorf("usleep() expects int, got %s", args[0].Type())
	}

	if us > 0 {
		time.Sleep(time.Duration(us) * time.Microsecond)
	}

	return engine.NewNull(), nil
}

// builtinPutenv 设置环境变量（KEY=VALUE 格式）
// putenv($expr) → bool
//
// 参数：
//   - args[0]: 环境变量表达式（"KEY=VALUE" 格式）
//
// 返回值：
//   - true：设置成功
//   - false：设置失败
//
// 示例：
//
//	putenv("MY_VAR=hello")
//	$value = getenv("MY_VAR")  // → "hello"
func builtinPutenv(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("putenv() expects 1 argument, got %d", len(args))
	}

	expr := args[0].String()
	parts := strings.SplitN(expr, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("putenv() expects KEY=VALUE format, got: %s", expr)
	}

	name := parts[0]
	value := parts[1]

	err := os.Setenv(name, value)
	return engine.NewBool(err == nil), nil
}

// ==============================================================================
// P2 - 进阶功能
// ==============================================================================

// builtinSpawn 创建子进程（不等待完成）
// spawn($cmd) → Process
// spawn($cmd, $args) → Process
//
// 参数：
//   - args[0]: 命令字符串
//   - args[1]: 可选，参数数组
//
// 返回值：
//   - Process 对象
//   - null：创建失败
//
// 与 exec() 的区别：
//   - spawn 立即返回，不等待命令完成
//   - exec 等待命令完成并返回输出
//
// 示例：
//
//	$proc = spawn("sleep", ["10"])
//	println "Spawned PID: #{$proc.pid}"
//	// 可以继续做其他事情
//	$code = waitpid($proc)
func builtinSpawn(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("spawn() expects 1-2 arguments, got %d", len(args))
	}

	cmdStr := args[0].String()

	var cmd *exec.Cmd
	if len(args) == 2 && args[1].Type() == engine.TypeArray {
		// spawn("ls", ["-la", "/tmp"])
		cmdArgs := make([]string, 0)
		for _, arg := range args[1].Array() {
			cmdArgs = append(cmdArgs, arg.String())
		}
		cmd = exec.Command(cmdStr, cmdArgs...)
	} else {
		// spawn("ls -la /tmp") - 使用 shell
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return engine.NewNull(), nil
	}

	return &ProcessValue{
		cmd: cmd,
		pid: cmd.Process.Pid,
	}, nil
}

// builtinKill 向进程发送信号
// kill($pid) → bool
// kill($pid, $signal) → bool
//
// 参数：
//   - args[0]: 进程 ID
//   - args[1]: 可选，信号编号（默认 SIGTERM=15）
//
// 返回值：
//   - true：发送成功
//   - false：发送失败
//
// 常用信号：
//   - 1  SIGHUP  挂起
//   - 2  SIGINT  中断（Ctrl+C）
//   - 9  SIGKILL 强制终止
//   - 15 SIGTERM 终止（默认）
//   - 18 SIGCONT 继续
//   - 19 SIGSTOP 暂停
//
// 示例：
//
//	kill(12345)      // 发送 SIGTERM
//	kill(12345, 9)   // 发送 SIGKILL
//	kill(12345, 2)   // 发送 SIGINT
func builtinKill(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("kill() expects 1-2 arguments, got %d", len(args))
	}

	pid := int(args[0].Int())
	sig := syscall.SIGTERM // 默认信号

	if len(args) == 2 {
		sig = syscall.Signal(args[1].Int())
	}

	err := syscall.Kill(pid, sig)
	return engine.NewBool(err == nil), nil
}

// builtinWaitpid 等待指定子进程结束
// waitpid($proc) → int
// waitpid($pid) → int
//
// 参数：
//   - args[0]: Process 对象或进程 ID
//
// 返回值：
//   - 进程退出码
//   - -1：等待失败
//
// 示例：
//
//	$proc = spawn("sleep 1")
//	$code = waitpid($proc)
//	println "Exit code: #{$code}"
//
//	// 或者等待指定 PID
//	$code = waitpid(12345)
func builtinWaitpid(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("waitpid() expects 1 argument, got %d", len(args))
	}

	// 如果是 Process 对象
	if proc, ok := args[0].(*ProcessValue); ok {
		if proc.exited {
			return engine.NewInt(int64(proc.exitCode)), nil
		}

		state, err := proc.cmd.Process.Wait()
		if err != nil {
			proc.exited = true
			proc.exitCode = -1
			return engine.NewInt(-1), nil
		}

		proc.exited = true
		if state.Exited() {
			proc.exitCode = state.ExitCode()
		} else {
			proc.exitCode = -1
		}

		return engine.NewInt(int64(proc.exitCode)), nil
	}

	// 如果是 PID（整数）
	pid := int(args[0].Int())
	var status syscall.WaitStatus
	wpid, err := syscall.Wait4(pid, &status, 0, nil)
	if err != nil {
		return engine.NewInt(-1), nil
	}

	if wpid == pid {
		if status.Exited() {
			return engine.NewInt(int64(status.ExitStatus())), nil
		}
	}

	return engine.NewInt(-1), nil
}

// builtinFork 创建子进程（Unix）
// fork() → int
//
// 返回值：
//   - 0：在子进程中返回
//   - 正数：在父进程中返回（子进程 PID）
//   - -1：创建失败
//
// 注意：
//   - 此函数仅在 Unix/Linux 系统上可用
//   - 子进程会复制父进程的内存空间
//   - 通常配合 exec() 使用
//
// 示例：
//
//	$pid = fork()
//	if ($pid == 0) {
//	    // 子进程
//	    println "I am child"
//	    exit(0)
//	} else if ($pid > 0) {
//	    // 父进程
//	    println "Child PID: #{$pid}"
//	    $code = waitpid($pid)
//	}
func builtinFork(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("fork() expects 0 arguments, got %d", len(args))
	}

	pid, _, errno := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if errno != 0 {
		return engine.NewInt(-1), nil
	}

	return engine.NewInt(int64(pid)), nil
}

// PipeValue 表示管道对象
type PipeValue struct {
	readFd  int
	writeFd int
}

// Type 返回类型标识
func (p *PipeValue) Type() engine.ValueType { return engine.TypeObject }
func (p *PipeValue) IsNull() bool           { return false }
func (p *PipeValue) Bool() bool             { return true }
func (p *PipeValue) Int() int64             { return int64(p.readFd) }
func (p *PipeValue) Float() float64         { return float64(p.readFd) }
func (p *PipeValue) String() string {
	return fmt.Sprintf("Pipe(read=%d, write=%d)", p.readFd, p.writeFd)
}
func (p *PipeValue) Stringify() string                { return p.String() }
func (p *PipeValue) Array() []engine.Value            { return nil }
func (p *PipeValue) Len() int                         { return 2 }
func (p *PipeValue) Equals(v engine.Value) bool       { return false }
func (p *PipeValue) Less(v engine.Value) bool         { return false }
func (p *PipeValue) Greater(v engine.Value) bool      { return false }
func (p *PipeValue) LessEqual(v engine.Value) bool    { return false }
func (p *PipeValue) GreaterEqual(v engine.Value) bool { return false }
func (p *PipeValue) ToBigInt() engine.Value           { return engine.NewInt(0) }
func (p *PipeValue) ToBigDecimal() engine.Value       { return engine.NewFloat(0) }
func (p *PipeValue) Add(v engine.Value) engine.Value  { return p }
func (p *PipeValue) Sub(v engine.Value) engine.Value  { return p }
func (p *PipeValue) Mul(v engine.Value) engine.Value  { return p }
func (p *PipeValue) Div(v engine.Value) engine.Value  { return p }
func (p *PipeValue) Mod(v engine.Value) engine.Value  { return p }
func (p *PipeValue) Negate() engine.Value             { return p }

// Object 返回对象值
func (p *PipeValue) Object() map[string]engine.Value {
	return map[string]engine.Value{
		"read":  engine.NewInt(int64(p.readFd)),
		"write": engine.NewInt(int64(p.writeFd)),
	}
}

// builtinPipe 创建管道对
// pipe() → Pipe
//
// 返回值：
//   - Pipe 对象，包含 read 和 write 文件描述符
//   - null：创建失败
//
// 示例：
//
//	$p = pipe()
//	println "Read FD: #{$p.read}"
//	println "Write FD: #{$p.write}"
//
//	// 写入数据
//	$file = fdopen($p.write, "w")
//	fwrite($file, "hello")
//	fclose($file)
//
//	// 读取数据
//	$file = fdopen($p.read, "r")
//	$data = fread($file, 1024)
//	fclose($file)
func builtinPipe(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("pipe() expects 0 arguments, got %d", len(args))
	}

	r, w, err := os.Pipe()
	if err != nil {
		return engine.NewNull(), nil
	}

	return &PipeValue{
		readFd:  int(r.Fd()),
		writeFd: int(w.Fd()),
	}, nil
}

// ==============================================================================
// P3 - 高级功能
// ==============================================================================

// builtinSigwait 阻塞等待信号
// sigwait($signal) → int
// sigwait([$signals]) → int
//
// 参数：
//   - args[0]: 信号编号（整数）或信号编号数组
//
// 返回值：
//   - 收到的信号编号
//
// 常用信号：
//   - 1  SIGHUP  挂起
//   - 2  SIGINT  中断（Ctrl+C）
//   - 10 SIGUSR1 用户定义信号1
//   - 12 SIGUSR2 用户定义信号2
//   - 15 SIGTERM 终止
//
// 示例：
//
//	// 等待单个信号
//	println "Waiting for SIGINT (Ctrl+C)..."
//	$sig = sigwait(2)
//	println "Received signal: #{$sig}"
//
//	// 等待多个信号
//	println "Waiting for SIGINT or SIGTERM..."
//	$sig = sigwait([2, 15])
//	println "Received signal: #{$sig}"
//
//	// 配合 fork 使用
//	$pid = fork()
//	if ($pid == 0) {
//	    // 子进程等待信号
//	    $sig = sigwait(10)
//	    println "Child received SIGUSR1"
//	    exit(0)
//	} else {
//	    // 父进程发送信号
//	    usleep(100000)
//	    kill($pid, 10)
//	    waitpid($pid)
//	}
func builtinSigwait(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sigwait() expects 1 argument, got %d", len(args))
	}

	// 解析信号列表
	var sigs []os.Signal
	if args[0].Type() == engine.TypeArray {
		// sigwait([2, 15])
		for _, v := range args[0].Array() {
			sigs = append(sigs, syscall.Signal(v.Int()))
		}
	} else {
		// sigwait(2)
		sigs = append(sigs, syscall.Signal(args[0].Int()))
	}

	if len(sigs) == 0 {
		return nil, fmt.Errorf("sigwait() expects at least one signal")
	}

	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)
	defer signal.Stop(sigChan)

	// 阻塞等待信号
	sig := <-sigChan

	return engine.NewInt(int64(sig.(syscall.Signal))), nil
}

// ProcessExtSigs returns function signatures for REPL :doc command.
func ProcessExtSigs() map[string]string {
	return map[string]string{
		"exec":        "exec(cmd, [args]) → string  — Execute command, return output",
		"system":      "system(cmd) → int  — Execute command, return exit code",
		"shell_exec":  "shell_exec(cmd) → string  — Execute via shell, return full output",
		"proc_open":   "proc_open(cmd, [opts]) → Process  — Open process with pipes",
		"proc_close":  "proc_close(proc) → int  — Close process, return exit code",
		"proc_wait":   "proc_wait(proc) → int  — Wait for process to end",
		"proc_status": "proc_status(proc) → object  — Get process status",
		"spawn":       "spawn(cmd, [args]) → Process  — Spawn background process",
		"kill":        "kill(pid, [signal]) → bool  — Send signal to process",
		"waitpid":     "waitpid(proc_or_pid) → int  — Wait for process, return exit code",
		"usleep":      "usleep(us) → null  — Sleep for microseconds",
		"pipe":        "pipe() → Pipe  — Create pipe",
		"sigwait":     "sigwait(signal_or_signals) → int  — Wait for signal",
		"getpid":      "getpid() → int  — Get current process ID",
		"getppid":     "getppid() → int  — Get parent process ID",
		"getlogin":    "getlogin() → string  — Get login username",
		"hostname":    "hostname() → string  — Get hostname",
		"tmpdir":      "tmpdir() → string  — Get temp directory",
		"getenv":      "getenv(name, [default]) → string  — Get environment variable",
		"setenv":      "setenv(name, value) → bool  — Set environment variable",
		"putenv":      "putenv(expr) → bool  — Set env var (KEY=VALUE format)",
	}
}
