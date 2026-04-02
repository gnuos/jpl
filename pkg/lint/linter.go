// Package lint 实现 JPL 静态分析
package lint

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gnuos/jpl/lexer"
	"github.com/gnuos/jpl/parser"
	"github.com/gnuos/jpl/token"
)

// Severity 诊断级别
type Severity int

const (
	Warning Severity = iota
	Error
)

func (s Severity) String() string {
	if s == Error {
		return "error"
	}
	return "warning"
}

// Diagnostic 诊断信息
type Diagnostic struct {
	Rule     string   // 规则名
	Severity Severity // 级别
	Pos      token.Position
	Message  string
}

func (d *Diagnostic) String() string {
	return fmt.Sprintf("%s:%d:%d: %s: %s (%s)",
		d.Pos.Filename, d.Pos.Line, d.Pos.Column,
		d.Severity, d.Message, d.Rule)
}

// LintResult 分析结果
type LintResult struct {
	Diagnostics []Diagnostic
}

// HasErrors 是否有错误
func (r *LintResult) HasErrors() bool {
	for _, d := range r.Diagnostics {
		if d.Severity == Error {
			return true
		}
	}
	return false
}

// String 格式化输出
func (r *LintResult) String() string {
	var lines []string
	for _, d := range r.Diagnostics {
		lines = append(lines, d.String())
	}
	return strings.Join(lines, "\n")
}

// Lint 分析源代码
func Lint(src string, filename string) *LintResult {
	l := lexer.NewLexer(src, filename)
	p := parser.NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		// 语法错误直接返回
		result := &LintResult{}
		for _, e := range p.Errors() {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Rule:     "syntax",
				Severity: Error,
				Pos:      token.Position{Filename: filename, Line: 1},
				Message:  e,
			})
		}
		return result
	}

	linter := &linter{
		filename: filename,
		result:   &LintResult{},
	}
	linter.pushScope() // 全局作用域
	linter.analyzeProgram(program)
	linter.popScope()

	// 按行号排序
	sort.Slice(linter.result.Diagnostics, func(i, j int) bool {
		a, b := linter.result.Diagnostics[i], linter.result.Diagnostics[j]
		if a.Pos.Line != b.Pos.Line {
			return a.Pos.Line < b.Pos.Line
		}
		return a.Pos.Column < b.Pos.Column
	})

	return linter.result
}

// ============================================================================
// 作用域
// ============================================================================

type varInfo struct {
	name     string
	pos      token.Position
	used     bool
	isGlobal bool // 是否通过 global 声明引用
}

type scope struct {
	vars   map[string]*varInfo
	parent *scope
	isFunc bool // 函数作用域（决定 global 引用的目标）
}

// ============================================================================
// Linter 核心
// ============================================================================

type linter struct {
	filename string
	result   *LintResult
	scopes   []*scope
}

func (l *linter) currentScope() *scope {
	return l.scopes[len(l.scopes)-1]
}

func (l *linter) pushScope() {
	s := &scope{vars: make(map[string]*varInfo)}
	if len(l.scopes) > 0 {
		s.parent = l.scopes[len(l.scopes)-1]
	}
	l.scopes = append(l.scopes, s)
}

func (l *linter) popScope() {
	s := l.currentScope()
	// 检查未使用的变量
	for _, v := range s.vars {
		if !v.used && !v.isGlobal {
			l.result.Diagnostics = append(l.result.Diagnostics, Diagnostic{
				Rule:     "unused-var",
				Severity: Warning,
				Pos:      v.pos,
				Message:  fmt.Sprintf("variable %q declared but never used", v.name),
			})
		}
	}
	l.scopes = l.scopes[:len(l.scopes)-1]
}

// declareVar 声明变量
func (l *linter) declareVar(name string, pos token.Position) {
	s := l.currentScope()
	s.vars[name] = &varInfo{name: name, pos: pos}
}

// useVar 标记变量使用
func (l *linter) useVar(name string, pos token.Position) {
	// 向上查找作用域
	for s := l.currentScope(); s != nil; s = s.parent {
		if v, ok := s.vars[name]; ok {
			v.used = true
			return
		}
	}
	// 未找到声明 — 检查是否是内置函数名（不报错）
	// 全局未声明的 $var 报 undefined-var
	if strings.HasPrefix(name, "$") {
		l.result.Diagnostics = append(l.result.Diagnostics, Diagnostic{
			Rule:     "undefined-var",
			Severity: Error,
			Pos:      pos,
			Message:  fmt.Sprintf("variable %q used but not declared", name),
		})
	}
}

// ============================================================================
// AST 遍历
// ============================================================================

func (l *linter) analyzeProgram(prog *parser.Program) {
	for _, stmt := range prog.Statements {
		l.analyzeStmt(stmt)
	}
}

func (l *linter) analyzeStmt(stmt parser.Statement) {
	switch s := stmt.(type) {
	case *parser.VarDecl:
		if s.Name != nil {
			l.declareVar(s.Name.Value, s.Name.Pos())
		}
		if s.Value != nil {
			l.analyzeExpr(s.Value)
		}

	case *parser.ConstDecl:
		if s.Name != nil {
			l.declareVar(s.Name.Value, s.Name.Pos())
		}
		if s.Value != nil {
			l.analyzeExpr(s.Value)
		}

	case *parser.FuncDecl:
		if s.Name != nil {
			l.declareVar(s.Name.Value, s.Name.Pos())
		}
		l.analyzeFuncBody(s.Parameters, s.Body)

	case *parser.ExprStmt:
		l.analyzeExpr(s.Expression)

	case *parser.ReturnStmt:
		if s.Value != nil {
			l.analyzeExpr(s.Value)
		}

	case *parser.ThrowStmt:
		if s.Value != nil {
			l.analyzeExpr(s.Value)
		}

	case *parser.IfStmt:
		l.analyzeExpr(s.Cond)
		l.analyzeBlock(s.Body)
		if s.Else != nil {
			l.analyzeStmt(s.Else)
		}

	case *parser.WhileStmt:
		l.analyzeExpr(s.Cond)
		l.analyzeBlock(s.Body)

	case *parser.ForStmt:
		if s.Init != nil {
			l.analyzeStmt(s.Init)
		}
		if s.Cond != nil {
			l.analyzeExpr(s.Cond)
		}
		if s.Post != nil {
			l.analyzeExpr(s.Post)
		}
		l.analyzeBlock(s.Body)

	case *parser.ForeachStmt:
		if s.Key != nil {
			l.declareVar(s.Key.Value, s.Key.Pos())
		}
		if s.Value != nil {
			l.declareVar(s.Value.Value, s.Value.Pos())
		}
		l.analyzeExpr(s.Array)
		l.analyzeBlock(s.Body)

	case *parser.TryCatchStmt:
		l.analyzeBlock(s.TryBody)
		if s.CatchVar != nil {
			l.declareVar(s.CatchVar.Value, s.CatchVar.Pos())
		}
		if s.CatchCondition != nil {
			l.analyzeExpr(s.CatchCondition)
		}
		l.analyzeBlock(s.CatchBody)

	case *parser.GlobalDecl:
		for _, name := range s.Names {
			// 在当前作用域标记为 global 引用
			l.currentScope().vars[name.Value] = &varInfo{
				name:     name.Value,
				pos:      name.Pos(),
				used:     true,
				isGlobal: true,
			}
		}

	case *parser.StaticDecl:
		if s.Name != nil {
			l.declareVar(s.Name.Value, s.Name.Pos())
		}
		if s.Value != nil {
			l.analyzeExpr(s.Value)
		}

	case *parser.ImportStmt:
		// import 不需要特殊处理

	case *parser.IncludeStmt:
		// include 不需要特殊处理

	case *parser.BlockStmt:
		l.analyzeBlock(s)

	case *parser.MatchStmt:
		l.analyzeExpr(s.Value)
		for _, c := range s.Cases {
			l.analyzePattern(c.Pattern)
			if c.Guard != nil {
				l.analyzeExpr(c.Guard)
			}
			l.analyzeStmt(c.Body)
		}

	default:
		// 其他语句类型，尝试分析为表达式
	}
}

func (l *linter) analyzeBlock(block *parser.BlockStmt) {
	if block == nil {
		return
	}
	l.pushScope()
	for _, stmt := range block.Statements {
		l.analyzeStmt(stmt)
	}
	l.checkDeadCode(block)
	l.popScope()
}

func (l *linter) analyzeFuncBody(params []*parser.Identifier, body *parser.BlockStmt) {
	l.pushScope()
	l.currentScope().isFunc = true
	// 声明参数
	for _, p := range params {
		l.declareVar(p.Value, p.Pos())
	}
	if body != nil {
		for _, stmt := range body.Statements {
			l.analyzeStmt(stmt)
		}
		l.checkDeadCode(body)
	}
	l.popScope()
}

func (l *linter) analyzeExpr(expr parser.Expression) {
	switch e := expr.(type) {
	case *parser.Identifier:
		l.useVar(e.Value, e.Pos())

	case *parser.NumberLiteral, *parser.StringLiteral, *parser.BoolLiteral, *parser.NullLiteral:
		// 字面量无需分析

	case *parser.BinaryExpr:
		l.analyzeExpr(e.Left)
		l.analyzeExpr(e.Right)

	case *parser.UnaryExpr:
		l.analyzeExpr(e.Operand)

	case *parser.CallExpr:
		l.analyzeExpr(e.Function)
		for _, arg := range e.Arguments {
			l.analyzeExpr(arg)
		}

	case *parser.IndexExpr:
		l.analyzeExpr(e.Left)
		l.analyzeExpr(e.Index)

	case *parser.MemberExpr:
		l.analyzeExpr(e.Object)

	case *parser.TernaryExpr:
		l.analyzeExpr(e.Condition)
		l.analyzeExpr(e.TrueExpr)
		l.analyzeExpr(e.FalseExpr)

	case *parser.ArrayLiteral:
		for _, elem := range e.Elements {
			l.analyzeExpr(elem)
		}

	case *parser.ObjectLiteral:
		for k, v := range e.Pairs {
			l.analyzeExpr(k)
			l.analyzeExpr(v)
		}

	case *parser.AssignExpr:
		l.analyzeExpr(e.Left)
		l.analyzeExpr(e.Value)

	case *parser.ConcatExpr:
		l.analyzeExpr(e.Left)
		l.analyzeExpr(e.Right)

	case *parser.RangeExpr:
		l.analyzeExpr(e.Start)
		l.analyzeExpr(e.End)

	case *parser.LambdaExpr:
		l.analyzeFuncBody(e.Parameters, e.Body)

	case *parser.ArrowExpr:
		l.pushScope()
		l.currentScope().isFunc = true
		for _, p := range e.Parameters {
			l.declareVar(p.Value, p.Pos())
		}
		if e.Body != nil {
			l.analyzeExpr(e.Body)
		}
		if e.BlockBody != nil {
			for _, stmt := range e.BlockBody.Statements {
				l.analyzeStmt(stmt)
			}
			l.checkDeadCode(e.BlockBody)
		}
		l.popScope()

	case *parser.PipeExpr:
		l.analyzeExpr(e.Left)
		l.analyzeExpr(e.Right)

	case *parser.TypeCast:
		l.analyzeExpr(e.Expr)

	case *parser.RegexLiteral:
		// 正则字面量无需分析

	case *parser.IfStmt:
		// if 作为表达式
		l.analyzeExpr(e.Cond)
		l.analyzeBlock(e.Body)
		if e.Else != nil {
			l.analyzeStmt(e.Else)
		}

	case *parser.MatchStmt:
		l.analyzeExpr(e.Value)
		for _, c := range e.Cases {
			l.analyzePattern(c.Pattern)
			if c.Guard != nil {
				l.analyzeExpr(c.Guard)
			}
			l.analyzeStmt(c.Body)
		}

	default:
		// 未知表达式类型
	}
}

func (l *linter) analyzePattern(p parser.Pattern) {
	switch pat := p.(type) {
	case *parser.IdentifierPattern:
		l.declareVar(pat.Name.Value, pat.Name.Pos())
	case *parser.LiteralPattern:
		l.analyzeExpr(pat.Value)
	case *parser.ArrayPattern:
		for _, elem := range pat.Elements {
			l.analyzePattern(elem)
		}
	case *parser.ObjectPattern:
		for _, sub := range pat.Pairs {
			l.analyzePattern(sub)
		}
		if pat.Rest != nil {
			l.declareVar(pat.Rest.Value, pat.Rest.Pos())
		}
	case *parser.OrPattern:
		for _, alt := range pat.Patterns {
			l.analyzePattern(alt)
		}
	case *parser.RangePattern:
		l.analyzeExpr(pat.Start)
		l.analyzeExpr(pat.End)
	case *parser.RegexPattern:
		if pat.Binding != nil {
			l.declareVar(pat.Binding.Value, pat.Binding.Pos())
		}
	}
}

// ============================================================================
// 死代码检测
// ============================================================================

func (l *linter) checkDeadCode(block *parser.BlockStmt) {
	if block == nil || len(block.Statements) < 2 {
		return
	}
	for i := 0; i < len(block.Statements)-1; i++ {
		if isTerminating(block.Statements[i]) {
			// 下一条语句不可达
			next := block.Statements[i+1]
			l.result.Diagnostics = append(l.result.Diagnostics, Diagnostic{
				Rule:     "dead-code",
				Severity: Warning,
				Pos:      next.Pos(),
				Message:  "unreachable code after " + terminatingKind(block.Statements[i]),
			})
		}
	}
}

func isTerminating(stmt parser.Statement) bool {
	switch stmt.(type) {
	case *parser.ReturnStmt, *parser.BreakStmt, *parser.ContinueStmt, *parser.ThrowStmt:
		return true
	}
	return false
}

func terminatingKind(stmt parser.Statement) string {
	switch stmt.(type) {
	case *parser.ReturnStmt:
		return "return"
	case *parser.BreakStmt:
		return "break"
	case *parser.ContinueStmt:
		return "continue"
	case *parser.ThrowStmt:
		return "throw"
	}
	return ""
}
