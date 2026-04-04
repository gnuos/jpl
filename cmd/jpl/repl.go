package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	prompt "github.com/elk-language/go-prompt"
	pstrings "github.com/elk-language/go-prompt/strings"
	"github.com/gnuos/jpl"
	"github.com/gnuos/jpl/engine"
	"github.com/gnuos/jpl/pkg/stdlib"
	"github.com/gnuos/jpl/token"
	"github.com/spf13/cobra"
)

const evalTimeout = 10 * time.Second

// REPL 表示交互式解释器
type REPL struct {
	Engine      *jpl.Engine
	VM          *engine.VM
	Program     *engine.Program
	DebugMode   bool
	HistoryFile string
	CmdHistory  []string
	CodeHistory []string        // 累积的代码历史（用于函数/变量持久化）
	multiLine   bool            // 是否处于多行续输模式
	multiBuf    strings.Builder // 多行续输缓冲区
}

// NewREPL 创建新的 REPL 实例
func NewREPL() *REPL {
	return &REPL{
		Engine: jpl.NewEngine(),
	}
}

// Init 初始化 REPL（注册内置函数、加载历史）
func (r *REPL) Init() {
	stdlib.RegisterAll(r.Engine)
	r.initHistory()
}

// initHistory 初始化历史记录
func (r *REPL) initHistory() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	r.HistoryFile = filepath.Join(homeDir, ".jpl", "history")
	os.MkdirAll(filepath.Dir(r.HistoryFile), 0755)

	// 加载历史记录
	if f, err := os.Open(r.HistoryFile); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				r.CmdHistory = append(r.CmdHistory, line)
			}
		}
	}
}

// saveHistory 保存命令到历史
func (r *REPL) saveHistory(input string) {
	if r.HistoryFile == "" || input == "" {
		return
	}

	// 检查是否与最后一条重复
	if len(r.CmdHistory) > 0 && r.CmdHistory[len(r.CmdHistory)-1] == input {
		return
	}

	r.CmdHistory = append(r.CmdHistory, input)

	// 追加到文件
	f, err := os.OpenFile(r.HistoryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintln(f, input)
}

// Executor 执行输入的代码或调试指令
func (r *REPL) Executor(input string) {
	// 处理多行续输的退出（空行提交）
	if r.multiLine && strings.TrimSpace(input) == "" {
		r.multiLine = false
		code := r.multiBuf.String()
		r.multiBuf.Reset()
		if strings.TrimSpace(code) == "" {
			return
		}
		r.saveHistory(code)
		r.ExecCode(code)
		return
	}

	// 多行续输模式：累积输入
	if r.multiLine {
		r.multiBuf.WriteString(input)
		r.multiBuf.WriteString("\n")
		if !r.needsContinuation() {
			r.multiLine = false
			code := r.multiBuf.String()
			r.multiBuf.Reset()
			r.saveHistory(code)
			r.ExecCode(code)
		}
		return
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// 调试指令以 : 开头
	if strings.HasPrefix(input, ":") {
		r.HandleCommand(input)
		return
	}

	// 检查是否需要多行续输
	r.multiBuf.Reset()
	r.multiBuf.WriteString(input)
	r.multiBuf.WriteString("\n")
	if r.needsContinuation() {
		r.multiLine = true
		return
	}

	// 单行直接执行
	r.saveHistory(input)
	r.ExecCode(input)
}

// needsContinuation 检查当前缓冲区是否需要继续输入（括号/引号未闭合）
func (r *REPL) needsContinuation() bool {
	code := r.multiBuf.String()
	return !isBalanced(code)
}

// isBalanced 检查代码中的括号、引号是否平衡
func isBalanced(code string) bool {
	var stack []rune
	inSingleQuote := false
	inDoubleQuote := false
	inTripleSingle := false
	inTripleDouble := false
	escape := false
	i := 0
	runes := []rune(code)
	n := len(runes)

	for i < n {
		ch := runes[i]

		if escape {
			escape = false
			i++
			continue
		}

		if ch == '\\' {
			escape = true
			i++
			continue
		}

		// 检查三引号
		if !inSingleQuote && !inDoubleQuote && i+2 < n {
			if runes[i] == '\'' && runes[i+1] == '\'' && runes[i+2] == '\'' {
				inTripleSingle = !inTripleSingle
				i += 3
				continue
			}
			if runes[i] == '"' && runes[i+1] == '"' && runes[i+2] == '"' {
				inTripleDouble = !inTripleDouble
				i += 3
				continue
			}
		}

		// 在三引号内，忽略其他字符
		if inTripleSingle || inTripleDouble {
			i++
			continue
		}

		// 单引号
		if ch == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			i++
			continue
		}

		// 双引号
		if ch == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			i++
			continue
		}

		// 在字符串内，忽略括号
		if inSingleQuote || inDoubleQuote {
			i++
			continue
		}

		// 跳过注释
		if ch == '/' && i+1 < n && runes[i+1] == '/' {
			// 跳过到行尾
			for i < n && runes[i] != '\n' {
				i++
			}
			continue
		}

		// 括号平衡
		switch ch {
		case '(', '{', '[':
			stack = append(stack, ch)
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				return false
			}
			stack = stack[:len(stack)-1]
		case '}':
			if len(stack) == 0 || stack[len(stack)-1] != '{' {
				return false
			}
			stack = stack[:len(stack)-1]
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != '[' {
				return false
			}
			stack = stack[:len(stack)-1]
		}

		i++
	}

	return len(stack) == 0 && !inSingleQuote && !inDoubleQuote && !inTripleSingle && !inTripleDouble
}

// HandleCommand 处理调试指令
func (r *REPL) HandleCommand(cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case ":help":
		fmt.Println("调试指令:")
		fmt.Println("  :debug on/off  - 切换调试模式")
		fmt.Println("  :globals       - 显示全局变量")
		fmt.Println("  :locals        - 显示局部变量")
		fmt.Println("  :vars          - 显示所有变量")
		fmt.Println("  :funcs         - 显示所有内置函数")
		fmt.Println("  :consts        - 显示预设常量")
		fmt.Println("  :doc <name>    - 查看函数签名")
		fmt.Println("  :help          - 显示此帮助")
		fmt.Println("  :quit          - 退出 REPL")
		fmt.Println()
		fmt.Println("多行续输:")
		fmt.Println("  输入未闭合的括号或引号时，自动进入多行模式")
		fmt.Println("  提示符变为 '... '，输入空行提交代码")
		fmt.Println()
		fmt.Println("快捷键:")
		fmt.Println("  Ctrl+C         - 中断执行")
		fmt.Println("  Ctrl+D         - 退出")
		fmt.Println("  Tab            - 自动补全")
		fmt.Println("  Up/Down        - 历史导航")

	case ":debug":
		if len(parts) > 1 && parts[1] == "on" {
			r.DebugMode = true
			if r.VM != nil {
				r.VM.SetDebugMode(true)
			}
			fmt.Println("调试模式已开启")
		} else if len(parts) > 1 && parts[1] == "off" {
			r.DebugMode = false
			if r.VM != nil {
				r.VM.SetDebugMode(false)
			}
			fmt.Println("调试模式已关闭")
		} else {
			fmt.Printf("调试模式: %v\n", r.DebugMode)
		}

	case ":globals":
		if r.VM == nil {
			fmt.Println("（无变量）")
			return
		}
		globals := r.VM.GetGlobals()
		fmt.Println(FormatVars(globals.Vars, true))

	case ":locals":
		fmt.Println("（局部变量暂不支持）")

	case ":vars":
		if r.VM == nil {
			fmt.Println("（无变量）")
			return
		}
		globals := r.VM.GetGlobals()
		fmt.Println(FormatVars(globals.Vars, false))

	case ":funcs":
		names := stdlib.FunctionNames()
		fmt.Println(strings.Join(names, ", "))

	case ":consts":
		consts := []string{
			"INF=+Infinity", "NaN=NaN", "PI=3.14159",
			"TAU=6.28318", "E=2.71828", "SQRT2=1.41421",
			"LN2=0.69315", "LN10=2.30259",
		}
		fmt.Println(strings.Join(consts, ", "))

	case ":doc":
		if len(parts) < 2 {
			fmt.Println("用法: :doc <函数名>")
			return
		}
		doc := GetFunctionDoc(parts[1])
		fmt.Println(doc)

	case ":quit", ":exit":
		fmt.Println("再见!")
		os.Exit(0)

	default:
		fmt.Printf("未知指令: %s，输入 :help 查看帮助\n", parts[0])
	}
}

// ExecCode 执行 JPL 代码
func (r *REPL) ExecCode(input string) {
	// 过滤掉纯空白或纯注释的代码（避免累积后污染后续代码）
	trimmed := strings.TrimSpace(input)
	if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
		// 添加到代码历史（只添加有效代码）
		r.CodeHistory = append(r.CodeHistory, input)
	}

	// 合并所有历史代码（用分号分隔，使它们在同一作用域）
	var fullScript strings.Builder
	for i, code := range r.CodeHistory {
		if i > 0 {
			fullScript.WriteString("; ")
		}
		fullScript.WriteString(code)
	}

	// 编译合并后的代码
	prog, err := engine.CompileStringWithGlobals(fullScript.String(), "<repl>", nil)
	if err != nil {
		if ce, ok := err.(*engine.CompileError); ok {
			fmt.Printf("编译错误: %s\n", ce.Message)
		} else {
			fmt.Printf("编译错误: %v\n", err)
		}
		// 编译失败，回滚代码历史
		r.CodeHistory = r.CodeHistory[:len(r.CodeHistory)-1]
		return
	}

	// 创建新 VM
	r.VM = engine.NewVMWithProgram(r.Engine, prog)
	r.Program = prog
	r.VM.SetDebugMode(r.DebugMode)

	// 带超时的执行
	done := make(chan error, 1)
	go func() {
		done <- r.VM.Execute()
	}()

	select {
	case err := <-done:
		if err != nil {
			if re, ok := err.(*engine.RuntimeError); ok {
				if re.Line > 0 && r.Program != nil {
					fmt.Printf("%s", re.FormatWithContext(r.Program.SourceLines))
				} else {
					fmt.Printf("运行时错误: %s\n", re.Message)
				}
			} else {
				fmt.Printf("运行时错误: %v\n", err)
			}
			return
		}

		// 显示结果
		result := r.VM.GetResult()
		if result != nil && result.String() != "null" {
			fmt.Println(result.String())
		}

	case <-time.After(evalTimeout):
		r.VM.Interrupt()
		fmt.Printf("执行超时（超过 %v），已中断\n", evalTimeout)
	}
}

// Completer 提供自动补全
func (r *REPL) Completer(d prompt.Document) ([]prompt.Suggest, pstrings.RuneNumber, pstrings.RuneNumber) {
	word := d.GetWordBeforeCursor()
	if word == "" {
		return nil, 0, 0
	}

	var suggestions []prompt.Suggest
	prefix := strings.ToLower(word)

	// 关键字补全
	for _, kw := range token.Keywords() {
		if strings.HasPrefix(strings.ToLower(kw), prefix) {
			suggestions = append(suggestions, prompt.Suggest{Text: kw})
		}
	}

	// 内置函数补全
	for _, fn := range stdlib.FunctionNames() {
		if strings.HasPrefix(strings.ToLower(fn), prefix) {
			suggestions = append(suggestions, prompt.Suggest{Text: fn})
		}
	}

	// 变量名补全（从 VM 获取）
	if r.VM != nil {
		globals := r.VM.GetGlobals()
		for _, v := range globals.Vars {
			if strings.HasPrefix(strings.ToLower(v.Name), prefix) {
				suggestions = append(suggestions, prompt.Suggest{Text: v.Name})
			}
		}
	}

	// 计算替换范围：从单词开始到光标位置
	wordLen := pstrings.RuneNumber(len([]rune(word)))
	cursorPos := d.CurrentRuneIndex()
	startChar := cursorPos - wordLen
	endChar := cursorPos

	return suggestions, startChar, endChar
}

// FormatVars 格式化变量显示（导出用于测试）
func FormatVars(vars []engine.VarInfo, skipDollar bool) string {
	if len(vars) == 0 {
		return "（无变量）"
	}

	var parts []string
	for _, v := range vars {
		// 跳过函数类型的变量（避免显示内置函数）
		if v.Type == engine.TypeFunc {
			continue
		}
		// 跳过以 $ 开头的特殊变量（如参数）
		if skipDollar && strings.HasPrefix(v.Name, "$") {
			continue
		}
		val := v.Value.String()
		if len(val) > 20 {
			val = val[:17] + "..."
		}
		parts = append(parts, fmt.Sprintf("%s=%s", v.Name, val))
	}

	if len(parts) == 0 {
		return "（无变量）"
	}

	return strings.Join(parts, ", ")
}

// GetFunctionDoc 获取函数文档（导出用于测试）
func GetFunctionDoc(name string) string {
	doc := stdlib.GetFunctionDoc(name)
	if doc != "" {
		return doc
	}
	return fmt.Sprintf("未知函数: %s", name)
}

// ============================================================================
// CLI 命令
// ============================================================================

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "启动交互式 REPL",
	Long:  `启动 JPL 交互式 Read-Eval-Print Loop (REPL) 环境。`,
	Run:   runREPL,
}

func init() {
	rootCmd.AddCommand(replCmd)
}

// replInstance 全局 REPL 实例（用于 TUI）
var replInstance *REPL

func runREPL(cmd *cobra.Command, args []string) {
	// 创建并初始化 REPL
	replInstance = NewREPL()
	replInstance.Init()

	fmt.Println("JPL REPL - 输入 :help 查看指令，Ctrl+D 退出")
	fmt.Println()

	// 创建 go-prompt
	p := prompt.New(
		replInstance.Executor,
		prompt.WithPrefixCallback(func() string {
			if replInstance.multiLine {
				return "... "
			}
			return "> "
		}),
		prompt.WithHistory(replInstance.CmdHistory),
		prompt.WithTitle("JPL REPL"),
		prompt.WithCompleter(replInstance.Completer),
	)
	p.Run()
}
