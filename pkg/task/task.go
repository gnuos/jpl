package task

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ============================================================================
// 任务定义 (jpl.json tasks 字段)
// ============================================================================

// TaskDef 单个任务定义，支持两种格式：
//   - 简单字符串: "task-name": "rm -rf build"
//   - 带依赖的对象: "task-name": {"cmd": "jpl run build.jpl", "deps": ["clean", "lint"]}
type TaskDef struct {
	Cmd  string   `json:"cmd"`  // 要执行的命令
	Deps []string `json:"deps"` // 依赖的其他任务名
}

// UnmarshalJSON 自定义 JSON 解析，支持字符串和对象两种格式
func (t *TaskDef) UnmarshalJSON(data []byte) error {
	// 尝试解析为字符串
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		t.Cmd = s
		t.Deps = nil
		return nil
	}

	// 尝试解析为对象
	type taskDefAlias struct {
		Cmd  string   `json:"cmd"`
		Deps []string `json:"deps"`
	}
	var obj taskDefAlias
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("task must be a string or object with 'cmd' field")
	}
	if obj.Cmd == "" {
		return fmt.Errorf("task object must have a non-empty 'cmd' field")
	}
	t.Cmd = obj.Cmd
	t.Deps = obj.Deps
	return nil
}

// MarshalJSON 序列化时，如果无 deps 则输出字符串格式
func (t TaskDef) MarshalJSON() ([]byte, error) {
	if len(t.Deps) == 0 {
		return json.Marshal(t.Cmd)
	}
	type alias struct {
		Cmd  string   `json:"cmd"`
		Deps []string `json:"deps"`
	}
	return json.Marshal(alias{Cmd: t.Cmd, Deps: t.Deps})
}

// ============================================================================
// 任务执行计划
// ============================================================================

// TaskExecutionPlan 任务执行计划
type TaskExecutionPlan struct {
	Order []string           // 按依赖顺序排列的任务名
	Tasks map[string]TaskDef // 任务定义
}

// ResolveTaskOrder 解析任务执行顺序，检测循环依赖
func ResolveTaskOrder(tasks map[string]TaskDef, target string) (*TaskExecutionPlan, error) {
	if _, ok := tasks[target]; !ok {
		return nil, fmt.Errorf("task %q not found", target)
	}

	plan := &TaskExecutionPlan{
		Tasks: tasks,
	}

	visited := make(map[string]bool) // 已完成访问
	inStack := make(map[string]bool) // 当前递归栈中（用于检测循环）
	order := make([]string, 0)

	var dfs func(name string) error
	dfs = func(name string) error {
		if visited[name] {
			return nil
		}
		if inStack[name] {
			return fmt.Errorf("circular dependency detected: task %q depends on itself", name)
		}

		inStack[name] = true

		task, ok := tasks[name]
		if !ok {
			return fmt.Errorf("task %q not found (referenced as dependency)", name)
		}

		for _, dep := range task.Deps {
			if err := dfs(dep); err != nil {
				return err
			}
		}

		inStack[name] = false
		visited[name] = true
		order = append(order, name)

		return nil
	}

	if err := dfs(target); err != nil {
		return nil, err
	}

	plan.Order = order
	return plan, nil
}

// ============================================================================
// 任务执行器
// ============================================================================

// TaskRunner 任务执行器
type TaskRunner struct {
	ProjectDir string // 项目根目录
	Verbose    bool   // 详细输出
}

// NewTaskRunner 创建任务执行器
func NewTaskRunner(projectDir string, verbose bool) *TaskRunner {
	return &TaskRunner{
		ProjectDir: projectDir,
		Verbose:    verbose,
	}
}

// Run 执行任务计划中的所有任务
func (r *TaskRunner) Run(plan *TaskExecutionPlan) error {
	for _, name := range plan.Order {
		task := plan.Tasks[name]
		if err := r.runSingle(name, task); err != nil {
			return fmt.Errorf("task %q failed: %w", name, err)
		}
	}
	return nil
}

// runSingle 执行单个任务
func (r *TaskRunner) runSingle(name string, task TaskDef) error {
	cmdStr := strings.TrimSpace(task.Cmd)

	if r.Verbose {
		fmt.Fprintf(os.Stderr, "[verbose] running task %q: %s\n", name, cmdStr)
	}

	// 判断是 JPL 脚本还是 shell 命令
	if r.isJPLCommand(cmdStr) {
		return r.runJPLCommand(name, cmdStr)
	}
	return r.runShellCommand(name, cmdStr)
}

// isJPLCommand 判断命令是否是 JPL 命令（以 "jpl run" 开头或首个参数以 .jpl 结尾）
func (r *TaskRunner) isJPLCommand(cmdStr string) bool {
	// "jpl run xxx.jpl" 形式
	if strings.HasPrefix(cmdStr, "jpl run ") {
		return true
	}
	// 直接引用 .jpl 文件: "scripts/build.jpl" 或 "scripts/build.jpl --watch"
	if !strings.ContainsAny(cmdStr, "|&;><") {
		parts := strings.Fields(cmdStr)
		if len(parts) > 0 && strings.HasSuffix(parts[0], ".jpl") {
			return true
		}
	}
	return false
}

// runJPLCommand 执行 JPL 脚本命令（内部调用，无需子进程）
func (r *TaskRunner) runJPLCommand(name, cmdStr string) error {
	parts := strings.Fields(cmdStr)

	var scriptPath string
	var scriptArgs []string

	if parts[0] == "jpl" && len(parts) >= 3 && parts[1] == "run" {
		scriptPath = parts[2]
		scriptArgs = parts[3:]
	} else {
		scriptPath = parts[0]
		if len(parts) > 1 {
			scriptArgs = parts[1:]
		}
	}

	// 解析为相对于项目目录的路径
	if !filepath.IsAbs(scriptPath) {
		scriptPath = filepath.Join(r.ProjectDir, scriptPath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	// 使用外部 jpl 命令执行（保持与 `jpl run` 一致的行为）
	args := append([]string{"run", scriptPath}, scriptArgs...)
	cmd := exec.Command("jpl", args...)
	cmd.Dir = r.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if r.Verbose {
		fmt.Fprintf(os.Stderr, "[verbose] executing: jpl %s\n", strings.Join(args, " "))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}
	return nil
}

// runShellCommand 执行 shell 命令
func (r *TaskRunner) runShellCommand(name, cmdStr string) error {
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Dir = r.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if r.Verbose {
		fmt.Fprintf(os.Stderr, "[verbose] executing: sh -c %q\n", cmdStr)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}
	return nil
}

// DryRun 模拟执行，仅输出执行顺序
func (r *TaskRunner) DryRun(plan *TaskExecutionPlan) {
	for i, name := range plan.Order {
		task := plan.Tasks[name]
		marker := "→"
		if i == len(plan.Order)-1 {
			marker = "→"
		}
		fmt.Printf("  %s %s (%s)\n", marker, name, task.Cmd)
	}
}

// FormatTaskList 格式化输出任务列表
func FormatTaskList(tasks map[string]TaskDef) {
	if len(tasks) == 0 {
		fmt.Println("No tasks defined.")
		return
	}

	for name, task := range tasks {
		depsStr := ""
		if len(task.Deps) > 0 {
			depsStr = " → deps: " + strings.Join(task.Deps, ", ")
		}
		fmt.Printf("  %-20s %s%s\n", name, task.Cmd, depsStr)
	}
}
