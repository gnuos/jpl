// Package format 实现 JPL 代码格式化器
package format

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gnuos/jpl/lexer"
	"github.com/gnuos/jpl/parser"
	"github.com/gnuos/jpl/token"
)

const defaultIndent = "    "

// Formatter 代码格式化器
type Formatter struct {
	program    *parser.Program
	comments   []token.Token // 收集的注释 token（按行号排序）
	leadingIdx int           // 注释消费位置
	src        string
	out        strings.Builder
	indent     int
	lastLine   int // 上次输出的 AST 节点所在行
}

// Format 格式化 JPL 源代码
func Format(src string, filename string) (string, error) {
	// 1. 收集注释 token
	comments := collectComments(src, filename)

	// 2. 解析 AST
	l := lexer.NewLexer(src, filename)
	p := parser.NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		return "", fmt.Errorf("syntax error:\n%s", strings.Join(p.Errors(), "\n"))
	}

	// 3. 格式化
	f := &Formatter{
		program:  program,
		comments: comments,
		src:      src,
		indent:   0,
		lastLine: -1,
	}
	return f.format(), nil
}

// collectComments 从源码中收集所有注释 token
func collectComments(src string, filename string) []token.Token {
	l := lexer.NewLexer(src, filename)
	var comments []token.Token
	for {
		tok := l.NextToken()
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.COMMENT || tok.Type == token.BLOCK_COMMENT {
			comments = append(comments, tok)
		}
	}
	return comments
}

// ============================================================================
// 核心输出方法
// ============================================================================

func (f *Formatter) write(s string) {
	f.out.WriteString(s)
}

func (f *Formatter) writeLine(s string) {
	f.out.WriteString(s)
	f.out.WriteByte('\n')
}

func (f *Formatter) writeIndent() {
	for i := 0; i < f.indent; i++ {
		f.out.WriteString(defaultIndent)
	}
}

func (f *Formatter) writeNewline() {
	f.out.WriteByte('\n')
}

func (f *Formatter) writeSpace() {
	f.out.WriteByte(' ')
}

func (f *Formatter) incIndent() {
	f.indent++
}

func (f *Formatter) decIndent() {
	f.indent--
}

// ============================================================================
// 格式化入口
// ============================================================================

func (f *Formatter) format() string {
	for i, stmt := range f.program.Statements {
		line := stmt.Pos().Line

		// 输出 leading comments（注释会自带换行）
		f.flushLeadingComments(line)

		f.writeStmt(stmt)

		// 检查行尾注释（同一行上的注释）
		if !f.flushTrailingComments(line) {
			f.writeNewline()
		}

		// 在语句之间加空行（除最后一个语句外）
		_ = i // 空行逻辑通过注释判断，不依赖索引

		f.lastLine = line
	}

	// 输出文件末尾的注释
	f.flushRemainingComments()

	return f.out.String()
}

// ============================================================================
// 注释处理
// ============================================================================

// flushLeadingComments 输出在指定行之前出现的注释
func (f *Formatter) flushLeadingComments(beforeLine int) {
	for f.leadingIdx < len(f.comments) {
		c := f.comments[f.leadingIdx]
		if c.Pos.Line >= beforeLine {
			break
		}
		f.writeComment(c)
		f.leadingIdx++
	}
}

// flushTrailingComments 输出行尾注释（同一行上），返回是否写入了注释
func (f *Formatter) flushTrailingComments(afterLine int) bool {
	found := false
	for f.leadingIdx < len(f.comments) {
		c := f.comments[f.leadingIdx]
		if c.Pos.Line != afterLine {
			break
		}
		f.writeSpace()
		f.write(c.Literal)
		f.writeNewline()
		f.leadingIdx++
		found = true
	}
	return found
}

// flushRemainingComments 输出剩余注释
func (f *Formatter) flushRemainingComments() {
	for f.leadingIdx < len(f.comments) {
		f.writeComment(f.comments[f.leadingIdx])
		f.leadingIdx++
	}
}

func (f *Formatter) writeComment(c token.Token) {
	if c.Type == token.BLOCK_COMMENT {
		// 多行注释：保持原有格式
		f.writeIndent()
		f.writeLine(c.Literal)
	} else {
		// 单行注释
		f.writeIndent()
		f.writeLine(c.Literal)
	}
}

// ============================================================================
// 语句格式化
// ============================================================================

func (f *Formatter) writeStmt(stmt parser.Statement) {
	switch s := stmt.(type) {
	case *parser.ExprStmt:
		f.writeIndent()
		f.writeExpr(s.Expression)
	case *parser.VarDecl:
		f.writeVarDecl(s)
	case *parser.ConstDecl:
		f.writeConstDecl(s)
	case *parser.FuncDecl:
		f.writeFuncDecl(s)
	case *parser.ReturnStmt:
		f.writeReturnStmt(s)
	case *parser.BreakStmt:
		f.writeIndent()
		f.write("break")
	case *parser.ContinueStmt:
		f.writeIndent()
		f.write("continue")
	case *parser.IfStmt:
		f.writeIfStmt(s)
	case *parser.WhileStmt:
		f.writeWhileStmt(s)
	case *parser.ForStmt:
		f.writeForStmt(s)
	case *parser.ForeachStmt:
		f.writeForeachStmt(s)
	case *parser.ImportStmt:
		f.writeImportStmt(s)
	case *parser.IncludeStmt:
		f.writeIncludeStmt(s)
	case *parser.TryCatchStmt:
		f.writeTryCatchStmt(s)
	case *parser.ThrowStmt:
		f.writeIndent()
		f.write("throw ")
		f.writeExpr(s.Value)
	case *parser.GlobalDecl:
		f.writeGlobalDecl(s)
	case *parser.StaticDecl:
		f.writeStaticDecl(s)
	case *parser.BlockStmt:
		f.writeBlock(s)
	default:
		f.writeIndent()
		f.write(stmt.String())
	}
}

func (f *Formatter) writeVarDecl(v *parser.VarDecl) {
	f.writeIndent()
	f.write(v.Name.Value)
	if v.Value != nil {
		f.write(" = ")
		f.writeExpr(v.Value)
	}
}

func (f *Formatter) writeConstDecl(c *parser.ConstDecl) {
	f.writeIndent()
	f.write("const ")
	f.write(c.Name.Value)
	f.write(" = ")
	f.writeExpr(c.Value)
}

func (f *Formatter) writeFuncDecl(fn *parser.FuncDecl) {
	f.writeIndent()
	f.write("fn ")
	f.write(fn.Name.Value)
	f.write("(")
	for i, p := range fn.Parameters {
		if i > 0 {
			f.write(", ")
		}
		f.write(p.Value)
	}
	f.write(") ")
	f.writeBlock(fn.Body)
}

func (f *Formatter) writeReturnStmt(r *parser.ReturnStmt) {
	f.writeIndent()
	if r.Value != nil {
		f.write("return ")
		f.writeExpr(r.Value)
	} else {
		f.write("return")
	}
}

func (f *Formatter) writeIfStmt(s *parser.IfStmt) {
	f.writeIndent()
	f.write("if (")
	f.writeExpr(s.Cond)
	f.write(") ")
	f.writeBlock(s.Body)

	if s.Else != nil {
		if elseIf, ok := s.Else.(*parser.IfStmt); ok {
			f.write(" else ")
			// else if 不换行缩进，直接接在后面
			f.writeInlineIfStmt(elseIf)
		} else {
			f.write(" else ")
			if block, ok := s.Else.(*parser.BlockStmt); ok {
				f.writeBlock(block)
			} else {
				f.incIndent()
				f.writeStmt(s.Else)
				f.decIndent()
			}
		}
	}
}

func (f *Formatter) writeInlineIfStmt(s *parser.IfStmt) {
	f.write("if (")
	f.writeExpr(s.Cond)
	f.write(") ")
	f.writeBlock(s.Body)
	if s.Else != nil {
		if elseIf, ok := s.Else.(*parser.IfStmt); ok {
			f.write(" else ")
			f.writeInlineIfStmt(elseIf)
		} else {
			f.write(" else ")
			if block, ok := s.Else.(*parser.BlockStmt); ok {
				f.writeBlock(block)
			}
		}
	}
}

func (f *Formatter) writeWhileStmt(w *parser.WhileStmt) {
	f.writeIndent()
	f.write("while (")
	f.writeExpr(w.Cond)
	f.write(") ")
	f.writeBlock(w.Body)
}

func (f *Formatter) writeForStmt(fo *parser.ForStmt) {
	f.writeIndent()
	f.write("for (")
	if fo.Init != nil {
		f.writeForInit(fo.Init)
	}
	f.write("; ")
	if fo.Cond != nil {
		f.writeExpr(fo.Cond)
	}
	f.write("; ")
	if fo.Post != nil {
		f.writeExpr(fo.Post)
	}
	f.write(") ")
	f.writeBlock(fo.Body)
}

func (f *Formatter) writeForInit(stmt parser.Statement) {
	switch s := stmt.(type) {
	case *parser.VarDecl:
		f.write(s.Name.Value)
		if s.Value != nil {
			f.write(" = ")
			f.writeExpr(s.Value)
		}
	case *parser.ExprStmt:
		f.writeExpr(s.Expression)
	default:
		f.write(stmt.String())
	}
}

func (f *Formatter) writeForeachStmt(fo *parser.ForeachStmt) {
	f.writeIndent()
	f.write("foreach (")
	if fo.Key != nil {
		f.write(fo.Key.Value)
		f.write(" => ")
	}
	f.write(fo.Value.Value)
	f.write(" in ")
	f.writeExpr(fo.Array)
	f.write(") ")
	f.writeBlock(fo.Body)
}

func (f *Formatter) writeImportStmt(imp *parser.ImportStmt) {
	f.writeIndent()
	if imp.Alias != nil {
		f.write(fmt.Sprintf("import %q as %s", imp.Source, imp.Alias.Value))
	} else if len(imp.Names) > 0 {
		var names []string
		for _, n := range imp.Names {
			names = append(names, n.Value)
		}
		f.write(fmt.Sprintf("from %q import %s", imp.Source, strings.Join(names, ", ")))
	} else {
		f.write(fmt.Sprintf("import %q", imp.Source))
	}
}

func (f *Formatter) writeIncludeStmt(inc *parser.IncludeStmt) {
	f.writeIndent()
	if inc.Once {
		f.write(fmt.Sprintf("include_once %q", inc.Source))
	} else {
		f.write(fmt.Sprintf("include %q", inc.Source))
	}
}

func (f *Formatter) writeTryCatchStmt(tc *parser.TryCatchStmt) {
	f.writeIndent()
	f.write("try ")
	f.writeBlock(tc.TryBody)
	f.write(" catch (")
	f.write(tc.CatchVar.Value)
	if tc.CatchCondition != nil {
		f.write(" when ")
		f.writeExpr(tc.CatchCondition)
	}
	f.write(") ")
	f.writeBlock(tc.CatchBody)
}

func (f *Formatter) writeGlobalDecl(g *parser.GlobalDecl) {
	f.writeIndent()
	f.write("global ")
	for i, n := range g.Names {
		if i > 0 {
			f.write(", ")
		}
		f.write(n.Value)
	}
}

func (f *Formatter) writeStaticDecl(s *parser.StaticDecl) {
	f.writeIndent()
	f.write("static ")
	f.write(s.Name.Value)
	if s.Value != nil {
		f.write(" = ")
		f.writeExpr(s.Value)
	}
}

// ============================================================================
// 代码块格式化
// ============================================================================

func (f *Formatter) writeBlock(block *parser.BlockStmt) {
	f.write("{")
	if len(block.Statements) == 0 {
		f.write("}")
		return
	}
	f.writeNewline()
	f.incIndent()

	for _, stmt := range block.Statements {
		line := stmt.Pos().Line

		f.flushLeadingComments(line)
		f.writeStmt(stmt)

		// 检查行尾注释
		if !f.flushTrailingComments(line) {
			f.writeNewline()
		}
		f.lastLine = line
	}

	f.decIndent()
	f.writeIndent()
	f.write("}")
	// 注意：不在结尾写换行，由调用者控制
}

// ============================================================================
// 表达式格式化
// ============================================================================

func (f *Formatter) writeExpr(expr parser.Expression) {
	switch e := expr.(type) {
	case *parser.Identifier:
		f.write(e.Value)
	case *parser.NumberLiteral:
		f.write(e.Value)
	case *parser.StringLiteral:
		f.writeQuotedString(e)
	case *parser.BoolLiteral:
		if e.Value {
			f.write("true")
		} else {
			f.write("false")
		}
	case *parser.NullLiteral:
		f.write("null")
	case *parser.BinaryExpr:
		f.writeBinaryExpr(e)
	case *parser.UnaryExpr:
		f.writeUnaryExpr(e)
	case *parser.CallExpr:
		f.writeCallExpr(e)
	case *parser.IndexExpr:
		f.writeIndexExpr(e)
	case *parser.MemberExpr:
		f.writeMemberExpr(e)
	case *parser.TernaryExpr:
		f.writeTernaryExpr(e)
	case *parser.ArrayLiteral:
		f.writeArrayLiteral(e)
	case *parser.ObjectLiteral:
		f.writeObjectLiteral(e)
	case *parser.AssignExpr:
		f.writeAssignExpr(e)
	case *parser.LambdaExpr:
		f.writeLambdaExpr(e)
	case *parser.ArrowExpr:
		f.writeArrowExpr(e)
	case *parser.ConcatExpr:
		f.writeExpr(e.Left)
		f.write(" .. ")
		f.writeExpr(e.Right)
	case *parser.RangeExpr:
		f.writeExpr(e.Start)
		if e.Inclusive {
			f.write(" ..= ")
		} else {
			f.write(" ... ")
		}
		f.writeExpr(e.End)
	case *parser.RegexLiteral:
		f.write(fmt.Sprintf("#/%s/%s#", e.Pattern, e.Flags))
	case *parser.TypeCast:
		f.write(e.Type)
		f.write("(")
		f.writeExpr(e.Expr)
		f.write(")")
	case *parser.IfStmt:
		// if 作为表达式
		f.writeInlineIfStmt(e)
	case *parser.MatchStmt:
		f.writeMatchExpr(e)
	case *parser.PipeExpr:
		f.writePipeExpr(e)
	default:
		f.write(expr.String())
	}
}

func (f *Formatter) writeQuotedString(s *parser.StringLiteral) {
	// 使用双引号，保持原始字符串内容
	f.write(fmt.Sprintf("%q", s.Value))
}

func (f *Formatter) writeBinaryExpr(b *parser.BinaryExpr) {
	op := b.Operator
	// 确定运算符两侧是否需要空格
	needsSpace := needsSpaceAround(op)

	f.writeExpr(b.Left)
	if needsSpace {
		f.write(" ")
	}
	f.write(op)
	if needsSpace {
		f.write(" ")
	}
	f.writeExpr(b.Right)
}

func needsSpaceAround(op string) bool {
	switch op {
	case ".", "[", "(":
		return false
	default:
		return true
	}
}

func (f *Formatter) writeUnaryExpr(u *parser.UnaryExpr) {
	f.write(u.Operator)
	f.writeExpr(u.Operand)
}

func (f *Formatter) writeCallExpr(c *parser.CallExpr) {
	f.writeExpr(c.Function)
	f.write("(")
	for i, arg := range c.Arguments {
		if i > 0 {
			f.write(", ")
		}
		f.writeExpr(arg)
	}
	f.write(")")
}

func (f *Formatter) writeIndexExpr(idx *parser.IndexExpr) {
	f.writeExpr(idx.Left)
	f.write("[")
	f.writeExpr(idx.Index)
	f.write("]")
}

func (f *Formatter) writeMemberExpr(m *parser.MemberExpr) {
	f.writeExpr(m.Object)
	f.write(".")
	f.write(m.Member.Value)
}

func (f *Formatter) writeTernaryExpr(t *parser.TernaryExpr) {
	f.writeExpr(t.Condition)
	f.write(" ? ")
	f.writeExpr(t.TrueExpr)
	f.write(" : ")
	f.writeExpr(t.FalseExpr)
}

func (f *Formatter) writeArrayLiteral(a *parser.ArrayLiteral) {
	if len(a.Elements) == 0 {
		f.write("[]")
		return
	}

	// 判断是否需要多行
	multiline := f.shouldMultiline(a.Elements)

	if multiline {
		f.write("[")
		f.writeNewline()
		f.incIndent()
		for i, elem := range a.Elements {
			f.writeIndent()
			f.writeExpr(elem)
			if i < len(a.Elements)-1 {
				f.write(",")
			}
			f.writeNewline()
		}
		f.decIndent()
		f.writeIndent()
		f.write("]")
	} else {
		f.write("[")
		for i, elem := range a.Elements {
			if i > 0 {
				f.write(", ")
			}
			f.writeExpr(elem)
		}
		f.write("]")
	}
}

func (f *Formatter) writeObjectLiteral(o *parser.ObjectLiteral) {
	if len(o.Pairs) == 0 {
		f.write("{}")
		return
	}

	// 收集键值对
	type kv struct {
		key   parser.Expression
		value parser.Expression
		name  string // 排序用的键名
	}
	var pairs []kv
	for k, v := range o.Pairs {
		name := objectKeyName(k)
		pairs = append(pairs, kv{k, v, name})
	}

	// 按键名排序（保证幂等性）
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].name < pairs[j].name
	})

	multiline := len(pairs) > 3

	if multiline {
		f.write("{")
		f.writeNewline()
		f.incIndent()
		for i, p := range pairs {
			f.writeIndent()
			f.writeObjectKey(p.key)
			f.write(": ")
			f.writeExpr(p.value)
			if i < len(pairs)-1 {
				f.write(",")
			}
			f.writeNewline()
		}
		f.decIndent()
		f.writeIndent()
		f.write("}")
	} else {
		f.write("{")
		for i, p := range pairs {
			if i > 0 {
				f.write(", ")
			}
			f.writeObjectKey(p.key)
			f.write(": ")
			f.writeExpr(p.value)
		}
		f.write("}")
	}
}

func (f *Formatter) writeObjectKey(key parser.Expression) {
	switch k := key.(type) {
	case *parser.StringLiteral:
		f.write(fmt.Sprintf("%q", k.Value))
	case *parser.Identifier:
		f.write(k.Value)
	default:
		f.writeExpr(key)
	}
}

func (f *Formatter) writeAssignExpr(a *parser.AssignExpr) {
	f.writeExpr(a.Left)
	f.write(" = ")
	f.writeExpr(a.Value)
}

func (f *Formatter) writeLambdaExpr(l *parser.LambdaExpr) {
	f.write("fn(")
	for i, p := range l.Parameters {
		if i > 0 {
			f.write(", ")
		}
		f.write(p.Value)
	}
	f.write(") ")
	f.writeBlock(l.Body)
}

func (f *Formatter) writeArrowExpr(a *parser.ArrowExpr) {
	if len(a.Parameters) == 1 {
		f.write(a.Parameters[0].Value)
	} else {
		f.write("(")
		for i, p := range a.Parameters {
			if i > 0 {
				f.write(", ")
			}
			f.write(p.Value)
		}
		f.write(")")
	}
	f.write(" -> ")
	if a.Body != nil {
		f.writeExpr(a.Body)
	} else if a.BlockBody != nil {
		f.writeBlock(a.BlockBody)
	}
}

func (f *Formatter) writeMatchExpr(m *parser.MatchStmt) {
	f.write("match (")
	f.writeExpr(m.Value)
	f.write(") {")
	if len(m.Cases) == 0 {
		f.write("}")
		return
	}
	f.writeNewline()
	f.incIndent()
	for _, c := range m.Cases {
		f.writeIndent()
		f.write("case ")
		f.writePattern(c.Pattern)
		if c.Guard != nil {
			f.write(" if ")
			f.writeExpr(c.Guard)
		}

		// Body 可能是 BlockStmt（多行）或 ExprStmt（表达式体）
		if block, ok := c.Body.(*parser.BlockStmt); ok {
			f.write(" ")
			f.writeBlock(block)
		} else if exprStmt, ok := c.Body.(*parser.ExprStmt); ok {
			f.write(" => ")
			f.writeExpr(exprStmt.Expression)
		} else {
			f.write(" => ")
			f.writeStmt(c.Body)
		}
		f.writeNewline()
	}
	f.decIndent()
	f.writeIndent()
	f.write("}")
}

func (f *Formatter) writePattern(p parser.Pattern) {
	switch pat := p.(type) {
	case *parser.LiteralPattern:
		f.writeExpr(pat.Value)
	case *parser.IdentifierPattern:
		f.write(pat.Name.Value)
	case *parser.WildcardPattern:
		f.write("_")
	case *parser.ArrayPattern:
		f.write("[")
		for i, elem := range pat.Elements {
			if i > 0 {
				f.write(", ")
			}
			f.writePattern(elem)
		}
		f.write("]")
	case *parser.ObjectPattern:
		f.write("{")
		i := 0
		for key, pat := range pat.Pairs {
			if i > 0 {
				f.write(", ")
			}
			f.write(key)
			f.write(": ")
			f.writePattern(pat)
			i++
		}
		if pat.Rest != nil {
			if len(pat.Pairs) > 0 {
				f.write(", ")
			}
			f.write("...")
			f.write(pat.Rest.Value)
		}
		f.write("}")
	case *parser.OrPattern:
		for i, alt := range pat.Patterns {
			if i > 0 {
				f.write(" | ")
			}
			f.writePattern(alt)
		}
	case *parser.RangePattern:
		f.writeExpr(pat.Start)
		f.write(" ... ")
		f.writeExpr(pat.End)
	case *parser.RegexPattern:
		f.write(fmt.Sprintf("#/%s/%s#", pat.Pattern, pat.Flags))
		if pat.Binding != nil {
			f.write(" as ")
			f.write(pat.Binding.Value)
		}
	default:
		f.write(p.String())
	}
}

func (f *Formatter) writePipeExpr(p *parser.PipeExpr) {
	f.writeExpr(p.Left)
	f.write(" |> ")
	f.writeExpr(p.Right)
}

// ============================================================================
// 辅助方法
// ============================================================================

// objectKeyName 提取对象字面量键的排序名
func objectKeyName(key parser.Expression) string {
	switch k := key.(type) {
	case *parser.Identifier:
		return k.Value
	case *parser.StringLiteral:
		return k.Value
	default:
		return key.String()
	}
}

// shouldMultiline 判断表达式列表是否应该多行显示
func (f *Formatter) shouldMultiline(exprs []parser.Expression) bool {
	if len(exprs) > 4 {
		return true
	}
	for _, e := range exprs {
		// 如果元素本身是复杂表达式，多行显示
		switch e.(type) {
		case *parser.ObjectLiteral, *parser.ArrayLiteral,
			*parser.LambdaExpr, *parser.CallExpr:
			return true
		}
	}
	return false
}
