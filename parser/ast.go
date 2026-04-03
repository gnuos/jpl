// Package parser 实现 JPL 脚本语言的 Pratt Parser
package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gnuos/jpl/token"
)

// ============================================================================
// 基础接口
// ============================================================================

// Node AST 节点基础接口
type Node interface {
	Pos() token.Position // 节点位置
	String() string      // 字符串表示
}

// Expression 表达式节点接口
type Expression interface {
	Node
	expressionNode()
}

// Statement 语句节点接口
type Statement interface {
	Node
	statementNode()
}

// Pattern 模式节点接口
type Pattern interface {
	Node
	patternNode()
}

// ============================================================================
// 程序根节点
// ============================================================================

// Program 程序根节点
type Program struct {
	Statements []Statement
}

func (p *Program) Pos() token.Position {
	if len(p.Statements) > 0 {
		return p.Statements[0].Pos()
	}
	return token.Position{}
}

func (p *Program) String() string {
	var buf bytes.Buffer
	for _, stmt := range p.Statements {
		buf.WriteString(stmt.String())
		buf.WriteString("\n")
	}
	return buf.String()
}

// ============================================================================
// 表达式节点
// ============================================================================

// Identifier 标识符表达式
type Identifier struct {
	Token token.Token // 标识符 Token
	Value string      // 标识符名称
}

func (i *Identifier) expressionNode()     {}
func (i *Identifier) Pos() token.Position { return i.Token.Pos }
func (i *Identifier) String() string      { return i.Value }

// NumberLiteral 数字字面量
type NumberLiteral struct {
	Token token.Token // 数字 Token
	Value string      // 数字字符串
}

func (n *NumberLiteral) expressionNode()     {}
func (n *NumberLiteral) Pos() token.Position { return n.Token.Pos }
func (n *NumberLiteral) String() string      { return n.Value }

// StringLiteral 字符串字面量
type StringLiteral struct {
	Token token.Token // 字符串 Token
	Value string      // 字符串值
}

func (s *StringLiteral) expressionNode()     {}
func (s *StringLiteral) Pos() token.Position { return s.Token.Pos }
func (s *StringLiteral) String() string      { return fmt.Sprintf("%q", s.Value) }

// BoolLiteral 布尔字面量
type BoolLiteral struct {
	Token token.Token // 布尔 Token
	Value bool        // 布尔值
}

func (b *BoolLiteral) expressionNode()     {}
func (b *BoolLiteral) Pos() token.Position { return b.Token.Pos }
func (b *BoolLiteral) String() string      { return fmt.Sprintf("%t", b.Value) }

// NullLiteral null 字面量
type NullLiteral struct {
	Token token.Token // null Token
}

func (n *NullLiteral) expressionNode()     {}
func (n *NullLiteral) Pos() token.Position { return n.Token.Pos }
func (n *NullLiteral) String() string      { return "null" }

// ArrayLiteral 数组字面量
type ArrayLiteral struct {
	Token    token.Token  // [ Token
	Elements []Expression // 元素列表
}

func (a *ArrayLiteral) expressionNode()     {}
func (a *ArrayLiteral) Pos() token.Position { return a.Token.Pos }
func (a *ArrayLiteral) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	for i, elem := range a.Elements {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(elem.String())
	}
	buf.WriteString("]")
	return buf.String()
}

// ObjectLiteral 对象字面量
type ObjectLiteral struct {
	Token token.Token               // { Token
	Pairs map[Expression]Expression // 键值对
}

func (o *ObjectLiteral) expressionNode()     {}
func (o *ObjectLiteral) Pos() token.Position { return o.Token.Pos }
func (o *ObjectLiteral) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	i := 0
	for key, val := range o.Pairs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(key.String())
		buf.WriteString(": ")
		buf.WriteString(val.String())
		i++
	}
	buf.WriteString("}")
	return buf.String()
}

// BinaryExpr 二元表达式
type BinaryExpr struct {
	Token    token.Token // 运算符 Token
	Left     Expression  // 左操作数
	Operator string      // 运算符
	Right    Expression  // 右操作数
}

func (b *BinaryExpr) expressionNode()     {}
func (b *BinaryExpr) Pos() token.Position { return b.Token.Pos }
func (b *BinaryExpr) String() string      { return fmt.Sprintf("(%s %s %s)", b.Left, b.Operator, b.Right) }

// UnaryExpr 一元表达式
type UnaryExpr struct {
	Token    token.Token // 运算符 Token
	Operator string      // 运算符
	Operand  Expression  // 操作数
}

func (u *UnaryExpr) expressionNode()     {}
func (u *UnaryExpr) Pos() token.Position { return u.Token.Pos }
func (u *UnaryExpr) String() string      { return fmt.Sprintf("(%s%s)", u.Operator, u.Operand) }

// CallExpr 函数调用表达式
type CallExpr struct {
	Token     token.Token  // ( Token
	Function  Expression   // 函数表达式
	Arguments []Expression // 参数列表
}

func (c *CallExpr) expressionNode()     {}
func (c *CallExpr) Pos() token.Position { return c.Token.Pos }
func (c *CallExpr) String() string {
	var buf bytes.Buffer
	buf.WriteString(c.Function.String())
	buf.WriteString("(")
	for i, arg := range c.Arguments {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(arg.String())
	}
	buf.WriteString(")")
	return buf.String()
}

// PipeExpr 管道表达式
// |> 正向管道: a |> f(b,c) = f(a, b, c)，左侧值作为首个参数
// <| 反向管道: f(b,c) <| a = f(b, c, a)，右侧值作为末尾参数
type PipeExpr struct {
	Token   token.Token // |> 或 <| Token
	Left    Expression  // |> 左侧: 值; <| 左侧: 函数
	Right   Expression  // |> 右侧: 函数; <| 右侧: 值
	Forward bool        // true = |>, false = <|
}

func (p *PipeExpr) expressionNode()     {}
func (p *PipeExpr) Pos() token.Position { return p.Token.Pos }
func (p *PipeExpr) String() string {
	if p.Forward {
		return fmt.Sprintf("%s |> %s", p.Left, p.Right)
	}
	return fmt.Sprintf("%s <| %s", p.Left, p.Right)
}

// IndexExpr 索引访问表达式
type IndexExpr struct {
	Token token.Token // [ Token
	Left  Expression  // 被索引对象
	Index Expression  // 索引
}

func (i *IndexExpr) expressionNode()     {}
func (i *IndexExpr) Pos() token.Position { return i.Token.Pos }
func (i *IndexExpr) String() string      { return fmt.Sprintf("%s[%s]", i.Left, i.Index) }

// MemberExpr 成员访问表达式
type MemberExpr struct {
	Token  token.Token // . Token
	Object Expression  // 对象
	Member *Identifier // 成员名
}

func (m *MemberExpr) expressionNode()     {}
func (m *MemberExpr) Pos() token.Position { return m.Token.Pos }
func (m *MemberExpr) String() string      { return fmt.Sprintf("%s.%s", m.Object, m.Member) }

// TernaryExpr 三元表达式
type TernaryExpr struct {
	Token     token.Token // ? Token
	Condition Expression  // 条件
	TrueExpr  Expression  // 真值表达式
	FalseExpr Expression  // 假值表达式
}

func (t *TernaryExpr) expressionNode()     {}
func (t *TernaryExpr) Pos() token.Position { return t.Token.Pos }
func (t *TernaryExpr) String() string {
	return fmt.Sprintf("(%s ? %s : %s)", t.Condition, t.TrueExpr, t.FalseExpr)
}

// TypeCast 类型转换表达式 AST 节点
// 表示 Go 风格的类型转换语法：int(x), float(x), string(x), bool(x)
//
// 与函数调用不同，TypeCast 是语言级别的类型转换语法，
// 在编译期生成专门的 OP_CAST 字节码指令，而非函数调用。
//
// 字段说明：
//   - Token: 类型关键字 token（int/float/string/bool），用于错误报告和位置信息
//   - Type:  目标类型字符串，取值为 "int", "float", "string", "bool"
//   - Expr:  被转换的表达式，可以是任意表达式（如 int($x + 1)）
//
// 示例：
//
//	int("42")  → TypeCast{Token: int, Type: "int", Expr: StringLiteral{"42"}}
//	float($y)  → TypeCast{Token: float, Type: "float", Expr: Identifier{$y}}
//	string(123) → TypeCast{Token: string, Type: "string", Expr: NumberLiteral{123}}
type TypeCast struct {
	Token token.Token // 类型关键字 Token（int/float/string/bool）
	Type  string      // 目标类型: "int", "float", "string", "bool"
	Expr  Expression  // 被转换的表达式
}

func (tc *TypeCast) expressionNode()     {}
func (tc *TypeCast) Pos() token.Position { return tc.Token.Pos }
func (tc *TypeCast) String() string {
	return fmt.Sprintf("%s(%s)", tc.Type, tc.Expr)
}

// LambdaExpr Lambda 表达式
type LambdaExpr struct {
	Token      token.Token   // fn Token
	Parameters []*Identifier // 参数列表
	Body       *BlockStmt    // 函数体
}

func (l *LambdaExpr) expressionNode()     {}
func (l *LambdaExpr) Pos() token.Position { return l.Token.Pos }
func (l *LambdaExpr) String() string {
	var buf bytes.Buffer
	buf.WriteString("fn(")
	for i, param := range l.Parameters {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(param.String())
	}
	buf.WriteString(") ")
	buf.WriteString(l.Body.String())
	return buf.String()
}

// ArrowExpr 箭头函数表达式
type ArrowExpr struct {
	Token      token.Token   // -> Token
	Parameters []*Identifier // 参数列表
	Body       Expression    // 表达式体（单行）
	BlockBody  *BlockStmt    // 块体（多行）
}

func (a *ArrowExpr) expressionNode()     {}
func (a *ArrowExpr) Pos() token.Position { return a.Token.Pos }
func (a *ArrowExpr) String() string {
	var buf bytes.Buffer
	if len(a.Parameters) == 1 {
		buf.WriteString(a.Parameters[0].String())
	} else {
		buf.WriteString("(")
		for i, param := range a.Parameters {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(param.String())
		}
		buf.WriteString(")")
	}
	buf.WriteString(" -> ")
	if a.Body != nil {
		buf.WriteString(a.Body.String())
	} else {
		buf.WriteString(a.BlockBody.String())
	}
	return buf.String()
}

// AssignExpr 赋值表达式
type AssignExpr struct {
	Token token.Token // = Token
	Left  Expression  // 左值
	Value Expression  // 值
}

func (a *AssignExpr) expressionNode()     {}
func (a *AssignExpr) Pos() token.Position { return a.Token.Pos }
func (a *AssignExpr) String() string      { return fmt.Sprintf("%s = %s", a.Left, a.Value) }

// ConcatExpr 字符串连接表达式
type ConcatExpr struct {
	Token token.Token // .. Token
	Left  Expression  // 左操作数
	Right Expression  // 右操作数
}

func (c *ConcatExpr) expressionNode()     {}
func (c *ConcatExpr) Pos() token.Position { return c.Token.Pos }
func (c *ConcatExpr) String() string      { return fmt.Sprintf("(%s .. %s)", c.Left, c.Right) }

// FormatExpr 字符串格式化表达式（用于插值格式化）
// 例如：#{$value:.2f} → FormatExpr{Expr: $value, Format: ".2f"}
type FormatExpr struct {
	Token  token.Token // : Token
	Expr   Expression  // 要格式化的表达式
	Format string      // 格式说明符（如 ".2f", "05d", "+.3e"）
}

func (f *FormatExpr) expressionNode()     {}
func (f *FormatExpr) Pos() token.Position { return f.Token.Pos }
func (f *FormatExpr) String() string      { return fmt.Sprintf("format(%s, %q)", f.Expr, f.Format) }

// RangeExpr 范围表达式
type RangeExpr struct {
	Token     token.Token // ... 或 ..= Token
	Start     Expression  // 起始值
	End       Expression  // 结束值
	Inclusive bool        // true 表示闭区间 ..=
}

func (r *RangeExpr) expressionNode()     {}
func (r *RangeExpr) Pos() token.Position { return r.Token.Pos }
func (r *RangeExpr) String() string {
	if r.Inclusive {
		return fmt.Sprintf("(%s ..= %s)", r.Start, r.End)
	}
	return fmt.Sprintf("(%s ... %s)", r.Start, r.End)
}

// ============================================================================
// 语句节点
// ============================================================================

// BlockStmt 代码块语句
type BlockStmt struct {
	Token      token.Token // { Token
	Statements []Statement // 语句列表
}

func (b *BlockStmt) statementNode()      {}
func (b *BlockStmt) Pos() token.Position { return b.Token.Pos }
func (b *BlockStmt) String() string {
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for _, stmt := range b.Statements {
		buf.WriteString("  ")
		buf.WriteString(stmt.String())
		buf.WriteString("\n")
	}
	buf.WriteString("}")
	return buf.String()
}

// IfStmt 条件语句
type IfStmt struct {
	Token token.Token // if Token
	Cond  Expression  // 条件
	Body  *BlockStmt  // if 分支
	Else  Statement   // else 分支（可能是 BlockStmt 或 IfStmt）
}

func (i *IfStmt) statementNode()      {}
func (i *IfStmt) expressionNode()     {} // if 也可作为表达式
func (i *IfStmt) Pos() token.Position { return i.Token.Pos }
func (i *IfStmt) String() string {
	var buf bytes.Buffer
	buf.WriteString("if (")
	buf.WriteString(i.Cond.String())
	buf.WriteString(") ")
	buf.WriteString(i.Body.String())
	if i.Else != nil {
		buf.WriteString(" else ")
		buf.WriteString(i.Else.String())
	}
	return buf.String()
}

// WhileStmt while 循环语句
type WhileStmt struct {
	Token token.Token // while Token
	Cond  Expression  // 条件
	Body  *BlockStmt  // 循环体
}

func (w *WhileStmt) statementNode()      {}
func (w *WhileStmt) Pos() token.Position { return w.Token.Pos }
func (w *WhileStmt) String() string {
	var buf bytes.Buffer
	buf.WriteString("while (")
	buf.WriteString(w.Cond.String())
	buf.WriteString(") ")
	buf.WriteString(w.Body.String())
	return buf.String()
}

// ForStmt for 循环语句
type ForStmt struct {
	Token token.Token // for Token
	Init  Statement   // 初始化语句
	Cond  Expression  // 条件
	Post  Expression  // 后置表达式
	Body  *BlockStmt  // 循环体
}

func (f *ForStmt) statementNode()      {}
func (f *ForStmt) Pos() token.Position { return f.Token.Pos }
func (f *ForStmt) String() string {
	var buf bytes.Buffer
	buf.WriteString("for (")
	if f.Init != nil {
		buf.WriteString(f.Init.String())
	}
	buf.WriteString("; ")
	if f.Cond != nil {
		buf.WriteString(f.Cond.String())
	}
	buf.WriteString("; ")
	if f.Post != nil {
		buf.WriteString(f.Post.String())
	}
	buf.WriteString(") ")
	buf.WriteString(f.Body.String())
	return buf.String()
}

// ForeachStmt foreach 循环语句
type ForeachStmt struct {
	Token token.Token // foreach Token
	Key   *Identifier // 键变量（可选）
	Value *Identifier // 值变量
	Array Expression  // 数组表达式
	Body  *BlockStmt  // 循环体
}

func (f *ForeachStmt) statementNode()      {}
func (f *ForeachStmt) Pos() token.Position { return f.Token.Pos }
func (f *ForeachStmt) String() string {
	var buf bytes.Buffer
	buf.WriteString("foreach (")
	if f.Key != nil {
		buf.WriteString(f.Key.String())
		buf.WriteString(" => ")
	}
	if f.Value != nil {
		buf.WriteString(f.Value.String())
	}
	buf.WriteString(" in ")
	buf.WriteString(f.Array.String())
	buf.WriteString(") ")
	buf.WriteString(f.Body.String())
	return buf.String()
}

// ReturnStmt return 语句
type ReturnStmt struct {
	Token token.Token // return Token
	Value Expression  // 返回值（可选）
}

func (r *ReturnStmt) statementNode()      {}
func (r *ReturnStmt) Pos() token.Position { return r.Token.Pos }
func (r *ReturnStmt) String() string {
	var buf bytes.Buffer
	buf.WriteString("return")
	if r.Value != nil {
		buf.WriteString(" ")
		buf.WriteString(r.Value.String())
	}
	return buf.String()
}

// BreakStmt break 语句
type BreakStmt struct {
	Token token.Token // break Token
}

func (b *BreakStmt) statementNode()      {}
func (b *BreakStmt) Pos() token.Position { return b.Token.Pos }
func (b *BreakStmt) String() string      { return "break" }

// ContinueStmt continue 语句
type ContinueStmt struct {
	Token token.Token // continue Token
}

func (c *ContinueStmt) statementNode()      {}
func (c *ContinueStmt) Pos() token.Position { return c.Token.Pos }
func (c *ContinueStmt) String() string      { return "continue" }

// ExprStmt 表达式语句
type ExprStmt struct {
	Token      token.Token // 表达式第一个 Token
	Expression Expression  // 表达式
}

func (e *ExprStmt) statementNode()      {}
func (e *ExprStmt) Pos() token.Position { return e.Token.Pos }
func (e *ExprStmt) String() string      { return e.Expression.String() }

// VarDecl 变量声明
type VarDecl struct {
	Token token.Token // $ Token 或标识符 Token
	Name  *Identifier // 变量名
	Value Expression  // 初始值（可选）
}

func (v *VarDecl) statementNode()      {}
func (v *VarDecl) Pos() token.Position { return v.Token.Pos }
func (v *VarDecl) String() string {
	var buf bytes.Buffer
	buf.WriteString(v.Name.String())
	if v.Value != nil {
		buf.WriteString(" = ")
		buf.WriteString(v.Value.String())
	}
	return buf.String()
}

// ConstDecl 常量声明
type ConstDecl struct {
	Token token.Token // const Token
	Name  *Identifier // 常量名
	Value Expression  // 常量值
}

func (c *ConstDecl) statementNode()      {}
func (c *ConstDecl) Pos() token.Position { return c.Token.Pos }
func (c *ConstDecl) String() string {
	return fmt.Sprintf("const %s = %s", c.Name, c.Value)
}

// FuncDecl 函数声明
type FuncDecl struct {
	Token      token.Token   // fn Token
	Name       *Identifier   // 函数名
	Parameters []*Identifier // 参数列表
	Body       *BlockStmt    // 函数体
}

func (f *FuncDecl) statementNode()      {}
func (f *FuncDecl) Pos() token.Position { return f.Token.Pos }
func (f *FuncDecl) String() string {
	var buf bytes.Buffer
	buf.WriteString("fn ")
	buf.WriteString(f.Name.String())
	buf.WriteString("(")
	for i, param := range f.Parameters {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(param.String())
	}
	buf.WriteString(") ")
	buf.WriteString(f.Body.String())
	return buf.String()
}

// ImportStmt import 语句
// import "x"           → Source="x", Names=nil, Alias=nil
// import "x" as foo    → Source="x", Names=nil, Alias=foo
// from "x" import a, b → Source="x", Names=[a, b], Alias=nil
type ImportStmt struct {
	Token  token.Token   // import 或 from Token
	Source string        // 导入源（文件路径或 URL）
	Names  []*Identifier // 选择性导入的名称列表（nil 表示全部导入）
	Alias  *Identifier   // 别名（import "x" as foo）
}

func (i *ImportStmt) statementNode()      {}
func (i *ImportStmt) Pos() token.Position { return i.Token.Pos }
func (i *ImportStmt) String() string {
	if i.Alias != nil {
		return fmt.Sprintf("import %q as %s", i.Source, i.Alias.Value)
	}
	if len(i.Names) > 0 {
		var names []string
		for _, n := range i.Names {
			names = append(names, n.Value)
		}
		return fmt.Sprintf("from %q import %s", i.Source, strings.Join(names, ", "))
	}
	return fmt.Sprintf("import %q", i.Source)
}

// IncludeStmt include/include_once 语句
type IncludeStmt struct {
	Token  token.Token // include 或 include_once Token
	Source string      // 导入源（文件路径）
	Once   bool        // 是否只加载一次
}

func (i *IncludeStmt) statementNode()      {}
func (i *IncludeStmt) Pos() token.Position { return i.Token.Pos }
func (i *IncludeStmt) String() string {
	if i.Once {
		return fmt.Sprintf("include_once %q", i.Source)
	}
	return fmt.Sprintf("include %q", i.Source)
}

// CatchClause 单个 catch 分支
type CatchClause struct {
	CatchVar  *Identifier // catch 变量名
	Condition Expression  // catch 条件（可选，when 后的表达式）
	Body      *BlockStmt  // catch 块
	Token     token.Token // catch Token
}

func (c *CatchClause) String() string {
	var buf bytes.Buffer
	buf.WriteString("catch (")
	buf.WriteString(c.CatchVar.String())
	if c.Condition != nil {
		buf.WriteString(" when ")
		buf.WriteString(c.Condition.String())
	}
	buf.WriteString(") ")
	buf.WriteString(c.Body.String())
	return buf.String()
}

// TryCatchStmt try/catch 语句
type TryCatchStmt struct {
	Token        token.Token    // try Token
	TryBody      *BlockStmt     // try 块
	CatchClauses []*CatchClause // catch 分支列表
}

func (t *TryCatchStmt) statementNode()      {}
func (t *TryCatchStmt) Pos() token.Position { return t.Token.Pos }
func (t *TryCatchStmt) String() string {
	var buf bytes.Buffer
	buf.WriteString("try ")
	buf.WriteString(t.TryBody.String())
	for _, clause := range t.CatchClauses {
		buf.WriteString(" ")
		buf.WriteString(clause.String())
	}
	return buf.String()
}

// ThrowStmt throw 语句
type ThrowStmt struct {
	Token token.Token // throw Token
	Value Expression  // 抛出的值
}

func (t *ThrowStmt) statementNode()      {}
func (t *ThrowStmt) Pos() token.Position { return t.Token.Pos }
func (t *ThrowStmt) String() string      { return fmt.Sprintf("throw %s", t.Value) }

// GlobalDecl global 声明
// global $varname; 或 global $varname, $varname2;
type GlobalDecl struct {
	Token token.Token   // global Token
	Names []*Identifier // 全局变量名列表
}

func (g *GlobalDecl) statementNode()      {}
func (g *GlobalDecl) Pos() token.Position { return g.Token.Pos }
func (g *GlobalDecl) String() string {
	var names []string
	for _, n := range g.Names {
		names = append(names, n.Value)
	}
	return fmt.Sprintf("global %s", names)
}

// StaticDecl static 声明
// static $varname = initialValue;
type StaticDecl struct {
	Token token.Token // static Token
	Name  *Identifier // 变量名
	Value Expression  // 初始值
}

func (s *StaticDecl) statementNode()      {}
func (s *StaticDecl) Pos() token.Position { return s.Token.Pos }
func (s *StaticDecl) String() string {
	if s.Value != nil {
		return fmt.Sprintf("static %s = %s", s.Name, s.Value)
	}
	return fmt.Sprintf("static %s", s.Name)
}

// MatchStmt match/case 语句
type MatchStmt struct {
	Token  token.Token  // match Token
	Value  Expression   // 匹配值
	Cases  []*MatchCase // case 列表
	IsExpr bool         // 是否为表达式模式
}

func (m *MatchStmt) statementNode()      {}
func (m *MatchStmt) expressionNode()     {} // 同时实现 expressionNode
func (m *MatchStmt) Pos() token.Position { return m.Token.Pos }
func (m *MatchStmt) String() string {
	var buf bytes.Buffer
	buf.WriteString("match (")
	buf.WriteString(m.Value.String())
	buf.WriteString(") {\n")
	for _, c := range m.Cases {
		buf.WriteString("  ")
		buf.WriteString(c.String())
		buf.WriteString("\n")
	}
	buf.WriteString("}")
	return buf.String()
}

// MatchCase match 的 case 分支
type MatchCase struct {
	Token   token.Token // case Token
	Pattern Pattern     // 模式
	Guard   Expression  // 守卫条件（可选）
	Body    Statement   // 分支体（表达式或语句块）
}

func (m *MatchCase) String() string {
	var buf bytes.Buffer
	buf.WriteString("case ")
	buf.WriteString(m.Pattern.String())
	if m.Guard != nil {
		buf.WriteString(" if ")
		buf.WriteString(m.Guard.String())
	}
	buf.WriteString(": ")
	buf.WriteString(m.Body.String())
	return buf.String()
}

// ============================================================================
// 模式节点
// ============================================================================

// LiteralPattern 字面量模式
type LiteralPattern struct {
	Token token.Token // 字面量 Token
	Value Expression  // 字面量值
}

func (l *LiteralPattern) patternNode()        {}
func (l *LiteralPattern) Pos() token.Position { return l.Token.Pos }
func (l *LiteralPattern) String() string      { return l.Value.String() }

// IdentifierPattern 标识符模式（变量绑定）
type IdentifierPattern struct {
	Token token.Token // 标识符 Token
	Name  *Identifier // 变量名
}

func (i *IdentifierPattern) patternNode()        {}
func (i *IdentifierPattern) Pos() token.Position { return i.Token.Pos }
func (i *IdentifierPattern) String() string      { return i.Name.String() }

// ArrayPattern 数组解构模式
type ArrayPattern struct {
	Token    token.Token // [ Token
	Elements []Pattern   // 元素模式列表
	Rest     *Identifier // 剩余元素变量（可选）
}

func (a *ArrayPattern) patternNode()        {}
func (a *ArrayPattern) Pos() token.Position { return a.Token.Pos }
func (a *ArrayPattern) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	for i, elem := range a.Elements {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(elem.String())
	}
	if a.Rest != nil {
		if len(a.Elements) > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("...")
		buf.WriteString(a.Rest.String())
	}
	buf.WriteString("]")
	return buf.String()
}

// ObjectPattern 对象解构模式
type ObjectPattern struct {
	Token token.Token        // { Token
	Pairs map[string]Pattern // 键模式对
	Rest  *Identifier        // 剩余属性变量（可选）
}

func (o *ObjectPattern) patternNode()        {}
func (o *ObjectPattern) Pos() token.Position { return o.Token.Pos }
func (o *ObjectPattern) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	i := 0
	for key, pat := range o.Pairs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%q: ", key))
		buf.WriteString(pat.String())
		i++
	}
	if o.Rest != nil {
		if len(o.Pairs) > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("...")
		buf.WriteString(o.Rest.String())
	}
	buf.WriteString("}")
	return buf.String()
}

// OrPattern OR 模式
type OrPattern struct {
	Token    token.Token // , Token
	Patterns []Pattern   // 模式列表
}

func (o *OrPattern) patternNode()        {}
func (o *OrPattern) Pos() token.Position { return o.Token.Pos }
func (o *OrPattern) String() string {
	var patterns []string
	for _, p := range o.Patterns {
		patterns = append(patterns, p.String())
	}
	return strings.Join(patterns, ", ")
}

// RangePattern 范围模式 (如 1...10, 1..=50)
type RangePattern struct {
	Token     token.Token // ... 或 ..= Token
	Start     Expression  // 起始值
	End       Expression  // 结束值
	Inclusive bool        // true 表示闭区间 ..=
}

func (r *RangePattern) patternNode()        {}
func (r *RangePattern) Pos() token.Position { return r.Token.Pos }
func (r *RangePattern) String() string {
	if r.Inclusive {
		return fmt.Sprintf("%s ..= %s", r.Start.String(), r.End.String())
	}
	return fmt.Sprintf("%s ... %s", r.Start.String(), r.End.String())
}

// WildcardPattern 通配符模式
type WildcardPattern struct {
	Token token.Token // _ Token
}

func (w *WildcardPattern) patternNode()        {}
func (w *WildcardPattern) Pos() token.Position { return w.Token.Pos }
func (w *WildcardPattern) String() string      { return "_" }

// RegexLiteral 正则字面量表达式 #/pattern/flags#
type RegexLiteral struct {
	Token   token.Token // #/pattern/flags# Token
	Pattern string      // 原始模式字符串
	Flags   string      // flags 字符串（imsU）
}

func (r *RegexLiteral) expressionNode()     {}
func (r *RegexLiteral) Pos() token.Position { return r.Token.Pos }
func (r *RegexLiteral) String() string      { return fmt.Sprintf("#/%s/%s#", r.Pattern, r.Flags) }

// RegexPattern 正则模式（match/case 中使用）
type RegexPattern struct {
	Token   token.Token // REGEX Token
	Pattern string      // 原始模式字符串
	Flags   string      // flags 字符串
	Binding *Identifier // as $var 可选绑定
}

func (r *RegexPattern) patternNode()        {}
func (r *RegexPattern) Pos() token.Position { return r.Token.Pos }
func (r *RegexPattern) String() string {
	s := fmt.Sprintf("#/%s/%s#", r.Pattern, r.Flags)
	if r.Binding != nil {
		s += " as " + r.Binding.String()
	}
	return s
}
