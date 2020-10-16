package compiler

import (
	"bytes"
	. "parser"
	"strconv"
)

type Compiler struct {
	Buff       []byte
	ScopeCount int
}

func CompileFile(ast File) []byte {
	c := Compiler{ScopeCount: 0}
	for _, statement := range ast.Statements {
		c.globalStatement(statement)
	}
	return c.Buff
}

func CompileOnlyInitializations(ast File) []byte {
	c := Compiler{ScopeCount: 0}
	for _, stmt := range ast.Statements {
		switch stmt.(type) {
		case Declaration:
			c.declaration(stmt.(Declaration))
			c.newline()
		case Typedef:
			c.typedefOnlyInit(stmt.(Typedef))
			c.newline()
		case ExportStatement:
			stmt2 := stmt.(ExportStatement).Stmt
			switch stmt2.(type) {
			case Declaration:
				c.declarationOnlyFunc(stmt2.(Declaration))
			case Typedef:
				c.typedefOnlyInit(stmt2.(Typedef))
			}
		case NullStatement:
			c.semicolon()
			/*
				case Import:
					c.imprt(stmt.(Import))
					c.newline()
			*/
		}
	}
	return c.Buff
}

func CompileImports(ast File) []byte {
	c := Compiler{ScopeCount: 0}
	for _, stmt := range ast.Statements {
		switch stmt.(type) {
		case Import:
			c.imprt(stmt.(Import))
		}
	}
	return c.Buff
}

func CompileTypedefs(ast File) []byte {
	c := Compiler{ScopeCount: 0}
	var stuff []Statement

	for _, stmt := range ast.Statements {
		var tdef Typedef
		switch stmt.(type) {
		case Typedef:
			tdef = stmt.(Typedef)
		case ExportStatement:
			stmt2 := stmt.(ExportStatement).Stmt
			switch stmt2.(type) {
			case Typedef:
				tdef = stmt2.(Typedef)
			default:
				continue
			}
		default:
			continue
		}
		switch tdef.Type.(type) {
		case StructType:
			c.append([]byte("typedef struct dummy_"))
			c.identifier(tdef.Name)
			c.space()
			c.identifier(tdef.Name)
			c.semicolon()
			c.newline()
		case UnionType:
			c.append([]byte("typedef union dummy_"))
			c.identifier(tdef.Name)
			c.space()
			c.identifier(tdef.Name)
			c.semicolon()
			c.newline()
		}
		stuff = append(stuff, stmt)
	}

	newStmts := c.SortDec(stuff)
	for _, stmt := range newStmts {
		var tdef Typedef
		switch stmt.(type) {
		case Typedef:
			tdef = stmt.(Typedef)
		case ExportStatement:
			stmt2 := stmt.(ExportStatement).Stmt
			switch stmt2.(type) {
			case Typedef:
				tdef = stmt2.(Typedef)
			default:
				continue
			}
		default:
			continue
		}
		switch tdef.Type.(type) {
		case StructType:
			c.typedefOnlyDec(tdef)
		case UnionType:
			c.typedefOnlyDec(tdef)
			c.newline()
		default:
			c.typedef(tdef)
			c.newline()
		}
	}
	return c.Buff
}

func CompileOnlyDeclarations(ast File) []byte {
	c := Compiler{ScopeCount: 0}
	for _, stmt := range ast.Statements {
		switch stmt.(type) {
		/*
			case Declaration:
				c.onlyDeclaration(stmt.(Declaration), false)
		*/
		case ExportStatement:
			stmt2 := stmt.(ExportStatement).Stmt
			switch stmt2.(type) {
			case Declaration:
				c.onlyDeclaration(stmt2.(Declaration), true)
				c.newline()
			}
		case NullStatement:
			c.semicolon()
			c.newline()
		}
	}
	return c.Buff
}

func (c *Compiler) append(buff []byte) {
	c.Buff = append(c.Buff, []byte(buff)...)
}
func (c *Compiler) colon() {
	c.append([]byte(":"))
}
func (c *Compiler) space() {
	c.append([]byte(" "))
}
func (c *Compiler) comma() {
	c.append([]byte(","))
}
func (c *Compiler) semicolon() {
	c.append([]byte(";"))
}
func (c *Compiler) newline() {
	c.append([]byte("\n"))
}
func (c *Compiler) openParen() {
	c.append([]byte("("))
}
func (c *Compiler) closeParen() {
	c.append([]byte(")"))
}
func (c *Compiler) openCurlyBrace() {
	c.append([]byte("{"))
}
func (c *Compiler) closeCurlyBrace() {
	c.append([]byte("}"))
}
func (c *Compiler) openBrace() {
	c.append([]byte("["))
}
func (c *Compiler) closeBrace() {
	c.append([]byte("]"))
}
func (c *Compiler) dot() {
	c.append([]byte("."))
}
func (c *Compiler) equal() {
	c.append([]byte("="))
}
func (c *Compiler) pushScope() {
	c.ScopeCount++
}
func (c *Compiler) popScope() {
	c.ScopeCount--
}
func (c *Compiler) operator(op Token) {
	c.append(op.Buff)
}

func (c *Compiler) identifier(identifer Token) {
	c.append(identifer.Buff)
}

func (c *Compiler) indent() {
	for i := 0; i < c.ScopeCount; i++ {
		c.append([]byte("	"))
	}
}

func (c *Compiler) SortDec(arr []Statement) []Statement {
	newArr := make([]Statement, len(arr))
	for i, v := range arr {
		newArr[i] = v
	}
	c.sortDec(newArr, 0, len(arr)-1)
	return newArr
}

func (c *Compiler) usesType(t Type, t2 BasicType) bool {
	switch t.(type) {
	case StructType:
		for _, prop := range t.(StructType).Props {
			for _, typ := range prop.Types {
				if c.usesType(typ, t2) {
					return true
				}
			}
		}
	case UnionType:
		// fmt.Println(string(t2.Expr.(IdentExpr).Value.Buff))
		for _, typ := range t.(UnionType).Types {
			if c.usesType(typ, t2) {
				return true
			}
		}
	default:
		if c.compareTypes(t, t2) {
			return true
		}
	}
	return false
}

func (c *Compiler) sortDec(arr []Statement, start, end int) {
	if start >= end {
		return
	}

	pivot := arr[end]
	var tdef Typedef

	switch pivot.(type) {
	case ExportStatement:
		tdef = pivot.(ExportStatement).Stmt.(Typedef)
	case Typedef:
		tdef = pivot.(Typedef)
	}

	t := BasicType{Expr: IdentExpr{Value: tdef.Name}}
	splitIndex := start

	for i := start; i < end; i++ {
		v := arr[i]

		switch v.(type) {
		case ExportStatement:
			tdef = v.(ExportStatement).Stmt.(Typedef)
		case Typedef:
			tdef = v.(Typedef)
		}
		if !c.usesType(tdef.Type, t) {
			arr[i] = arr[splitIndex]
			arr[splitIndex] = v
			splitIndex++
		}
	}

	arr[end] = arr[splitIndex]
	arr[splitIndex] = pivot

	c.sortDec(arr, start, splitIndex-1)
	c.sortDec(arr, splitIndex+1, end)
}

func (c *Compiler) globalStatement(stmt Statement) {
	c.newline()
	switch stmt.(type) {
	case Declaration:
		c.globalDeclaration(stmt.(Declaration), false)
	case Typedef:
		c.typedef(stmt.(Typedef))
	case ExportStatement:
		c.exportStatement(stmt.(ExportStatement).Stmt)
	case NullStatement:
		c.semicolon()
	case Import:
		c.imprt(stmt.(Import))
	}
}

func (c *Compiler) onlyDeclaration(dec Declaration, isExported bool) {
	for i, Var := range dec.Identifiers {
		if !isExported {
			c.append([]byte("static"))
			c.space()
		}
		Typ := dec.Types[i]

		c.declarationType(Typ, Var)
		switch Typ.(type) {
		case FuncType:
			break
		default:
			c.space()
			c.equal()
			c.space()
			c.expression(dec.Values[i])
		}
		c.semicolon()
	}
}

func (c *Compiler) declarationOnlyFunc(dec Declaration) {
	for i, Var := range dec.Identifiers {
		Typ := dec.Types[i]
		switch Typ.(type) {
		case FuncType:
			break
		default:
			continue
		}
		c.declarationType(Typ, Var)
		c.space()
		c.equal()
		c.expression(dec.Values[i])
		c.semicolon()
	}
}

func (c *Compiler) imprt(stmt Import) {
	for _, path := range stmt.Paths {
		c.append([]byte("#include \""))
		c.append(path.Buff)
		c.append([]byte("\""))
		c.newline()
	}
}

func (c *Compiler) exportStatement(stmt Statement) {
	switch stmt.(type) {
	case Declaration:
		c.globalDeclaration(stmt.(Declaration), true)
	case Typedef:
		c.typedef(stmt.(Typedef))
	case ExportStatement:
		c.exportStatement(stmt.(ExportStatement).Stmt)
	case NullStatement:
		c.semicolon()
	}
}

func (c *Compiler) statement(stmt Statement) {
	c.newline()
	switch stmt.(type) {
	case Declaration:
		c.declaration(stmt.(Declaration))
	case Typedef:
		c.typedef(stmt.(Typedef))
	case Return:
		c.rturn(stmt.(Return))
	case IfElseBlock:
		c.ifElse(stmt.(IfElseBlock))
	case Loop:
		c.loop(stmt.(Loop))
	case Assignment:
		c.assignment(stmt.(Assignment))
		c.semicolon()
	case Switch:
		c.swtch(stmt.(Switch))
	case Break:
		c.indent()
		c.append([]byte("break;"))
	case Continue:
		c.indent()
		c.append([]byte("continue;"))
	case NullStatement:
		c.semicolon()
	case Block:
		c.indent()
		c.block(stmt.(Block))
	case Defer:
		c.defr(stmt.(Defer))
	case Delete:
		c.delete(stmt.(Delete))
	case Label:
		c.identifier(stmt.(Label).Name)
		c.colon()
		c.semicolon()
	case Goto:
		c.append([]byte("goto "))
		c.identifier(stmt.(Goto).Name)
		c.semicolon()
	default:
		c.indent()
		c.expression(stmt.(Expression))
		c.semicolon()
	}
}

func (c *Compiler) statementNoSemicolon(stmt Statement) {
	c.newline()
	switch stmt.(type) {
	case Declaration:
		c.declaration(stmt.(Declaration))
	case Typedef:
		c.typedef(stmt.(Typedef))
	case Return:
		c.rturn(stmt.(Return))
	case IfElseBlock:
		c.ifElse(stmt.(IfElseBlock))
	case Loop:
		c.loop(stmt.(Loop))
	case Assignment:
		c.assignment(stmt.(Assignment))
	case Switch:
		c.swtch(stmt.(Switch))
	case Break:
		c.indent()
		c.append([]byte("break"))
	case Continue:
		c.indent()
		c.append([]byte("continue"))
	case NullStatement:
		c.semicolon()
	case Block:
		c.indent()
		c.block(stmt.(Block))
	case Defer:
		c.defr(stmt.(Defer))
	case Delete:
		c.delete(stmt.(Delete))
	default:
		c.indent()
		c.expression(stmt.(Expression))
	}
}

func (c *Compiler) typedef(typedef Typedef) {
	c.append([]byte("typedef"))
	c.space()
	c.declarationType(typedef.Type, typedef.Name)
	c.semicolon()
	switch typedef.Type.(type) {
	case StructType:
		c.newline()
		c.strctDefault(typedef)
		c.strctMethods(typedef.Type.(StructType))
	}
}

func (c *Compiler) typedefOnlyDec(typedef Typedef) {
	c.append([]byte("typedef"))
	c.space()

	switch typedef.Type.(type) {
	case StructType:
		c.append([]byte("struct dummy_"))
		c.identifier(typedef.Name)
		c.space()

		c.openCurlyBrace()
		c.pushScope()
		c.newline()
		for _, prop := range typedef.Type.(StructType).Props {
			c.strctPropDeclaration(prop)
		}
		c.popScope()
		c.closeCurlyBrace()
		c.space()
		c.identifier(typedef.Name)
		c.semicolon()
		c.newline()

		c.strctDefault(typedef)
		c.strctMethodsOnlyDec(typedef.Type.(StructType))
	case UnionType:
		c.append([]byte("union dummy_"))
		c.identifier(typedef.Name)
		c.space()

		c.openCurlyBrace()
		c.newline()
		c.pushScope()

		for x, prop := range typedef.Type.(UnionType).Identifiers {
			c.indent()
			c.declarationType(typedef.Type.(UnionType).Types[x], prop)
			c.semicolon()
			c.newline()
		}

		c.popScope()
		c.closeCurlyBrace()
		c.space()
		c.identifier(typedef.Name)
		c.semicolon()
	default:
		c.declarationType(typedef.Type, typedef.Name)
		c.semicolon()
	}
}

func (c *Compiler) typedefOnlyInit(typedef Typedef) {
	switch typedef.Type.(type) {
	case StructType:
		c.newline()
		c.strctMethods(typedef.Type.(StructType))
	}
}

func (c *Compiler) delete(delete Delete) {
	for _, Expr := range delete.Exprs {
		c.indent()
		c.append([]byte("delete"))
		c.openParen()
		c.expression(Expr)
		c.closeParen()
		c.semicolon()
		c.newline()
	}
}

func (c *Compiler) asyncDelete(delete Delete) int {
	a := 0
	for _, Expr := range delete.Exprs {
		c.indent()
		c.append([]byte("delete"))
		c.openParen()
		a += c.asyncExpr(Expr)
		c.closeParen()
		c.semicolon()
		c.newline()
	}
	return a
}

func (c *Compiler) defr(defr Defer) {
	c.indent()
	c.append([]byte("defer"))
	c.space()
	c.openCurlyBrace()
	c.pushScope()
	c.statement(defr.Stmt)
	c.newline()
	c.popScope()
	c.indent()
	c.closeCurlyBrace()
	c.semicolon()
}

func (c *Compiler) loop(loop Loop) {
	c.indent()
	c.append([]byte("for"))
	c.openParen()

	if loop.Type&InitLoop == InitLoop {
		c.statement(loop.InitStatement)
	} else {
		c.semicolon()
	}

	if loop.Type&CondLoop == CondLoop {
		c.expression(loop.Condition)
	}
	c.semicolon()

	if loop.Type&LoopLoop == LoopLoop {
		c.statementNoSemicolon(loop.LoopStatement)
	}

	c.closeParen()
	c.block(loop.Block)
}

func (c *Compiler) asyncLoop(loop Loop) ([]Declaration, int) {
	decs, a := []Declaration{}, 0

	c.indent()
	c.append([]byte("for"))
	c.openParen()

	if loop.Type&InitLoop == InitLoop {
		decs, a = c.asyncStmt(loop.InitStatement)
	} else {
		c.semicolon()
	}

	if loop.Type&CondLoop == CondLoop {
		a += c.asyncExpr(loop.Condition)
	}
	c.semicolon()

	if loop.Type&LoopLoop == LoopLoop {
		decs2, a2 := c.asyncStmtNoSemicolon(loop.LoopStatement)
		decs = append(decs, decs2...)
		a += a2
	}

	c.closeParen()
	decs2, a2 := c.asyncBlock(loop.Block)

	decs = append(decs, decs2...)
	a += a2

	return decs, a
}

func (c *Compiler) globalDeclaration(dec Declaration, isExported bool) {
	hasValues := len(dec.Values) > 0

	for i, Var := range dec.Identifiers {
		if !isExported {
			c.append([]byte("static"))
			c.space()
		}

		c.declarationType(dec.Types[i], Var)

		if hasValues {
			c.space()
			c.equal()
			c.space()
			c.expression(dec.Values[i])
		}
		c.semicolon()
	}
}

func (c *Compiler) declaration(dec Declaration) {
	hasValues := len(dec.Values) > 0

	for i, Var := range dec.Identifiers {
		c.indent()
		c.declarationType(dec.Types[i], Var)

		if hasValues {
			c.space()
			c.equal()
			c.space()
			c.expression(dec.Values[i])
		}
		c.semicolon()
		c.newline()
	}
}

func (c *Compiler) rturn(rtrn Return) {
	c.indent()
	c.append([]byte("return"))
	c.space()

	if len(rtrn.Values) > 0 {
		c.expression(rtrn.Values[0])
	}

	c.semicolon()
}

func (c *Compiler) block(block Block) {
	c.openCurlyBrace()
	c.pushScope()
	for _, statement := range block.Statements {
		c.statement(statement)
	}
	c.popScope()
	c.newline()
	c.indent()
	c.closeCurlyBrace()
}

func (c *Compiler) asyncBlock(block Block) ([]Declaration, int) {
	c.openCurlyBrace()
	c.pushScope()
	decs, a := []Declaration{}, 0

	for _, statement := range block.Statements {
		dec, a2 := c.asyncStmt(statement)
		a += a2

		if dec != nil {
			decs = append(decs, dec...)
		}
	}

	c.popScope()
	c.newline()
	c.indent()
	c.closeCurlyBrace()

	return decs, a
}

func (c *Compiler) expression(expr Expression) {

	switch expr.(type) {
	case Type:
		c.Type(expr.(Type), []byte{})
	case CallExpr:
		c.functionCall(expr.(CallExpr))
	case BasicLit:
		i := expr.(BasicLit).Value
		c.identifier(i)
	case IdentExpr:
		c.identifier(expr.(IdentExpr).Value)
	case BinaryExpr:
		switch expr.(BinaryExpr).Left.(type) {
		case BasicLit:
			c.expression(expr.(BinaryExpr).Left)
		case IdentExpr:
			c.expression(expr.(BinaryExpr).Left)
		default:
			c.openParen()
			c.expression(expr.(BinaryExpr).Left)
			c.closeParen()
		}

		c.operator(expr.(BinaryExpr).Op)

		switch expr.(BinaryExpr).Right.(type) {
		case BasicLit:
			c.expression(expr.(BinaryExpr).Right)
		case IdentExpr:
			c.expression(expr.(BinaryExpr).Right)
		default:
			c.openParen()
			c.expression(expr.(BinaryExpr).Right)
			c.closeParen()
		}
	case UnaryExpr:
		c.openParen()
		c.operator(expr.(UnaryExpr).Op)

		switch expr.(UnaryExpr).Expr.(type) {
		case BasicLit:
			c.expression(expr.(UnaryExpr).Expr)
		case IdentExpr:
			c.expression(expr.(UnaryExpr).Expr)
		default:
			c.openParen()
			c.expression(expr.(UnaryExpr).Expr)
			c.closeParen()
		}
		c.closeParen()
	case PostfixUnaryExpr:
		switch expr.(PostfixUnaryExpr).Expr.(type) {
		case BasicLit:
			c.expression(expr.(PostfixUnaryExpr).Expr)
		case IdentExpr:
			c.expression(expr.(PostfixUnaryExpr).Expr)
		default:
			c.openParen()
			c.expression(expr.(PostfixUnaryExpr).Expr)
			c.closeParen()
		}
		c.operator(expr.(PostfixUnaryExpr).Op)
	case ArrayMemberExpr:
		switch expr.(ArrayMemberExpr).Parent.(type) {
		case MemberExpr:
			c.expression(expr.(ArrayMemberExpr).Parent)
		case IdentExpr:
			c.expression(expr.(ArrayMemberExpr).Parent)
		default:
			c.openParen()
			c.expression(expr.(ArrayMemberExpr).Parent)
			c.closeParen()
		}
		c.openBrace()
		c.expression(expr.(ArrayMemberExpr).Index)
		c.closeBrace()
	case MemberExpr:
		c.expression(expr.(MemberExpr).Base)
		c.append([]byte("."))
		c.identifier(expr.(MemberExpr).Prop)
	case TernaryExpr:
		c.expression(expr.(TernaryExpr).Cond)
		c.space()
		c.append([]byte("?"))
		c.space()
		c.expression(expr.(TernaryExpr).Left)
		c.space()
		c.colon()
		c.space()
		c.expression(expr.(TernaryExpr).Right)
	case PointerMemberExpr:
		c.expression(expr.(PointerMemberExpr).Base)
		c.append([]byte("->"))
		c.identifier(expr.(PointerMemberExpr).Prop)
	case CompoundLiteral:
		c.compoundLiteral(expr.(CompoundLiteral))
	case TypeCast:
		c.openParen()
		c.openParen()
		c.Type(expr.(TypeCast).Type.(Type), []byte{})
		c.closeParen()
		c.openParen()
		c.expression(expr.(TypeCast).Expr)
		c.closeParen()
		c.closeParen()
	case HeapAlloc:
		c.heapAlloc(expr.(HeapAlloc))
	/*
		case LenExpr:
			c.lenExpr(expr.(LenExpr))
		case SizeExpr:
			c.sizeExpr(expr.(SizeExpr))
	*/
	case ArrayLiteral:
		c.openCurlyBrace()
		for _, expr2 := range expr.(ArrayLiteral).Exprs {
			c.expression(expr2)
			c.comma()
			c.space()
		}
		c.closeCurlyBrace()
	case FuncExpr:
		if expr.(FuncExpr).Type.Type == AsyncFunction {
			c.asyncFunction(expr.(FuncExpr))
		} else {
			c.funcExprType(expr.(FuncExpr).Type, nil, nil, false)
			c.block(expr.(FuncExpr).Block)
		}
	}
}

func (c *Compiler) asyncExpr(expr Expression) int {

	switch expr.(type) {
	case Type:
		c.Type(expr.(Type), []byte{})
	case CallExpr:
		return c.asyncFunctionCall(expr.(CallExpr))
	case BasicLit:
		i := expr.(BasicLit).Value
		c.identifier(i)
	case IdentExpr:
		c.identifier(expr.(IdentExpr).Value)
	case BinaryExpr:
		a := 0
		switch expr.(BinaryExpr).Left.(type) {
		case BasicLit:
			a += c.asyncExpr(expr.(BinaryExpr).Left)
		case IdentExpr:
			a += c.asyncExpr(expr.(BinaryExpr).Left)
		default:
			c.openParen()
			a += c.asyncExpr(expr.(BinaryExpr).Left)
			c.closeParen()
		}

		c.operator(expr.(BinaryExpr).Op)

		switch expr.(BinaryExpr).Right.(type) {
		case BasicLit:
			a += c.asyncExpr(expr.(BinaryExpr).Right)
		case IdentExpr:
			a += c.asyncExpr(expr.(BinaryExpr).Right)
		default:
			c.openParen()
			a += c.asyncExpr(expr.(BinaryExpr).Right)
			c.closeParen()
		}
		return a
	case UnaryExpr:
		a := 0
		c.openParen()
		c.operator(expr.(UnaryExpr).Op)

		switch expr.(UnaryExpr).Expr.(type) {
		case BasicLit:
			a += c.asyncExpr(expr.(UnaryExpr).Expr)
		case IdentExpr:
			a += c.asyncExpr(expr.(UnaryExpr).Expr)
		default:
			c.openParen()
			a += c.asyncExpr(expr.(UnaryExpr).Expr)
			c.closeParen()
		}
		c.closeParen()
		return a
	case PostfixUnaryExpr:
		a := 0
		switch expr.(PostfixUnaryExpr).Expr.(type) {
		case BasicLit:
			a += c.asyncExpr(expr.(PostfixUnaryExpr).Expr)
		case IdentExpr:
			a += c.asyncExpr(expr.(PostfixUnaryExpr).Expr)
		default:
			c.openParen()
			a += c.asyncExpr(expr.(PostfixUnaryExpr).Expr)
			c.closeParen()
		}
		c.operator(expr.(PostfixUnaryExpr).Op)
		return a
	case ArrayMemberExpr:
		a := 0
		switch expr.(ArrayMemberExpr).Parent.(type) {
		case MemberExpr:
			a += c.asyncExpr(expr.(ArrayMemberExpr).Parent)
		case IdentExpr:
			a += c.asyncExpr(expr.(ArrayMemberExpr).Parent)
		default:
			c.openParen()
			a += c.asyncExpr(expr.(ArrayMemberExpr).Parent)
			c.closeParen()
		}
		c.openBrace()
		a += c.asyncExpr(expr.(ArrayMemberExpr).Index)
		c.closeBrace()
		return a
	case MemberExpr:
		a := c.asyncExpr(expr.(MemberExpr).Base)
		c.append([]byte("."))
		c.identifier(expr.(MemberExpr).Prop)
		return a
	case TernaryExpr:
		a := c.asyncExpr(expr.(TernaryExpr).Cond)
		c.space()
		c.append([]byte("?"))
		c.space()
		a += c.asyncExpr(expr.(TernaryExpr).Left)
		c.space()
		c.colon()
		c.space()
		a += c.asyncExpr(expr.(TernaryExpr).Right)
		return a
	case PointerMemberExpr:
		a := c.asyncExpr(expr.(PointerMemberExpr).Base)
		c.append([]byte("->"))
		c.identifier(expr.(PointerMemberExpr).Prop)
		return a
	case CompoundLiteral:
		return c.asyncCompoundLiteral(expr.(CompoundLiteral))
	case TypeCast:
		c.openParen()
		c.Type(expr.(TypeCast).Type.(Type), []byte{})
		c.closeParen()
		c.openParen()
		x := c.asyncExpr(expr.(TypeCast).Expr)
		c.closeParen()
		return x
	case AwaitExpr:
		c.append([]byte("VO_AWAIT"))
		c.openParen()
		x := c.asyncExpr(expr.(AwaitExpr).Expr)
		c.closeParen()
		return x
	case HeapAlloc:
		return c.asyncHeapAlloc(expr.(HeapAlloc))
	/*
		case LenExpr:
			c.lenExpr(expr.(LenExpr))
		case SizeExpr:
			c.sizeExpr(expr.(SizeExpr))
	*/
	case ArrayLiteral:
		a := 0
		c.openCurlyBrace()
		for _, expr2 := range expr.(ArrayLiteral).Exprs {
			a += c.asyncExpr(expr2)
			c.comma()
			c.space()
		}
		c.closeCurlyBrace()
		return a
	case FuncExpr:
		if expr.(FuncExpr).Type.Type == AsyncFunction {
			c.asyncFunction(expr.(FuncExpr))
		} else {
			c.funcExprType(expr.(FuncExpr).Type, nil, nil, false)
			c.block(expr.(FuncExpr).Block)
		}
	}
	return 0
}

func (c *Compiler) asyncHeapAlloc(expr HeapAlloc) int {
	switch expr.Type.(type) {
	case ArrayType:
		if expr.Val != nil {
			c.append([]byte("new3"))
			c.openParen()
			c.Type(expr.Type.(ArrayType).BaseType, []byte{})
			c.comma()
			c.Type(expr.Type, []byte{})
			c.comma()
			c.openParen()
			a := c.asyncExpr(expr.Val)
			c.closeParen()
			return a
		} else {
			c.append([]byte("new"))
			c.openParen()
			c.Type(expr.Type.(ArrayType).BaseType, []byte{})
			c.comma()
			c.Type(expr.Type, []byte{})
		}
	default:
		if expr.Val != nil {
			c.append([]byte("new2"))
			c.openParen()
			c.Type(expr.Type, []byte{})
			c.comma()
			c.Type(expr.Type, []byte{})
			c.comma()
			c.openParen()
			a := c.asyncExpr(expr.Val)
			c.closeParen()
			return a
		} else {
			c.append([]byte("new"))
			c.openParen()
			c.Type(expr.Type, []byte{})
			c.comma()
			c.Type(expr.Type, []byte{})
		}
	}
	c.closeParen()
	return 0
}

func (c *Compiler) asyncFunction(expr FuncExpr) {
	c1 := Compiler{ScopeCount: c.ScopeCount + 1}
	c2 := Compiler{ScopeCount: c.ScopeCount + 1}

	c1.append([]byte("typedef struct {\n int state;\nPROMISE_TYPE("))
	c1.Type(expr.Type.ReturnTypes[0], []byte{})
	c1.closeParen()
	c1.append([]byte("return_promise;\n"))

	c1.pushScope()

	c2.funcExprType(expr.Type, nil, nil, false)
	decs, a := c2.asyncBlock(expr.Block)

	for _, dec := range decs {
		c1.strctPropDeclaration(dec)
	}
	c1.append([]byte("PROMISE_TYPE("))
	c1.Type(expr.Type.ReturnTypes[0], []byte{})
	c1.closeParen()
	c1.append([]byte("promises["))
	c1.append([]byte(strconv.Itoa(a)))
	c1.closeBrace()
	c1.semicolon()
	c1.closeCurlyBrace()
	c1.append([]byte("a_ctx_"))
	c1.semicolon()
	c1.newline()

	c.openParen()
	c.openCurlyBrace()
	c.newline()
	c.append(c1.Buff)
	c.append(c2.Buff)
	c.semicolon()
	c.closeCurlyBrace()
	c.closeParen()
}

func (c *Compiler) asyncStmt(stmt Statement) ([]Declaration, int) {
	c.newline()
	switch stmt.(type) {
	case Declaration:
		return []Declaration{stmt.(Declaration)}, c.asyncDeclaration(stmt.(Declaration))
	case Typedef:
		c.typedef(stmt.(Typedef))
	case Return:
		return nil, c.asyncReturn(stmt.(Return))
	case IfElseBlock:
		return c.asyncIfElse(stmt.(IfElseBlock))
	case Loop:
		return c.asyncLoop(stmt.(Loop))
	case Assignment:
		a := c.asyncAssignment(stmt.(Assignment))
		c.semicolon()
		return nil, a
	case Switch:
		return c.asyncSwitch(stmt.(Switch))
	case Break:
		c.indent()
		c.append([]byte("break;"))
	case Continue:
		c.indent()
		c.append([]byte("continue;"))
	case NullStatement:
		c.semicolon()
	case Block:
		c.indent()
		return c.asyncBlock(stmt.(Block))
	case Delete:
		return nil, c.asyncDelete(stmt.(Delete))
	case Label:
		c.identifier(stmt.(Label).Name)
		c.colon()
		c.semicolon()
	case Goto:
		c.append([]byte("goto "))
		c.identifier(stmt.(Goto).Name)
		c.semicolon()
	default:
		c.indent()
		a := c.asyncExpr(stmt.(Expression))
		c.semicolon()
		return nil, a
	}
	return nil, 0
}

func (c *Compiler) asyncStmtNoSemicolon(stmt Statement) ([]Declaration, int) {
	c.newline()
	switch stmt.(type) {
	case Declaration:
		return []Declaration{stmt.(Declaration)}, c.asyncDeclaration(stmt.(Declaration))
	case Typedef:
		c.typedef(stmt.(Typedef))
	case Return:
		return nil, c.asyncReturn(stmt.(Return))
	case IfElseBlock:
		return c.asyncIfElse(stmt.(IfElseBlock))
	case Loop:
		return c.asyncLoop(stmt.(Loop))
	case Assignment:
		a := c.asyncAssignment(stmt.(Assignment))
		c.semicolon()
		return nil, a
	case Switch:
		return c.asyncSwitch(stmt.(Switch))
	case Break:
		c.indent()
		c.append([]byte("break;"))
	case Continue:
		c.indent()
		c.append([]byte("continue;"))
	case NullStatement:
		c.semicolon()
	case Block:
		c.indent()
		return c.asyncBlock(stmt.(Block))
	case Delete:
		return nil, c.asyncDelete(stmt.(Delete))
	case Label:
		c.identifier(stmt.(Label).Name)
		c.colon()
	case Goto:
		c.append([]byte("goto "))
		c.identifier(stmt.(Goto).Name)
	default:
		c.indent()
		a := c.asyncExpr(stmt.(Expression))
		return nil, a
	}
	return nil, 0
}

func (c *Compiler) asyncReturn(rturn Return) int {
	c.indent()
	if len(rturn.Values) > 0 {
		c.append([]byte("VO_ASYNC_RETURN"))
		c.openParen()
		a := c.asyncExpr(rturn.Values[0])
		c.closeParen()
		c.semicolon()
		return a
	}
	c.append([]byte("VO_ASYNC_VOID_RETURN"))
	c.semicolon()
	return 0
}

func (c *Compiler) asyncDeclaration(dec Declaration) int {
	c.indent()
	c.openParen()
	c.openCurlyBrace()
	a := 0

	for i, Var := range dec.Identifiers {
		c.append([]byte("VO_ASYNC_VAR("))
		c.identifier(Var)
		c.closeParen()
		c.space()
		c.equal()
		c.space()
		if len(dec.Values) > 1 {
			a += c.asyncExpr(dec.Values[i])
		} else {
			a += c.asyncExpr(dec.Values[0])
		}
		c.semicolon()
	}
	c.closeCurlyBrace()
	c.closeParen()

	return a
}

func (c *Compiler) compoundLiteral(expr CompoundLiteral) {

	switch expr.Name.(type) {
	case VecType:
		c.append([]byte("new4"))
		c.openParen()
		c.Type(expr.Name.(VecType).BaseType, []byte{})
		c.comma()
		c.space()
		c.openParen()
		c.expression(CompoundLiteral{Name: ImplictArrayType{BaseType: expr.Name.(VecType).BaseType}, Data: CompoundLiteralData{Fields: expr.Data.Fields, Values: expr.Data.Values}})
		c.closeParen()
		c.comma()
		c.space()
		c.append([]byte(strconv.Itoa(len(expr.Data.Values))))
		c.closeParen()
		return
	case PromiseType:
		c.append([]byte("new5"))
		c.openParen()
		c.Type(expr.Name.(PromiseType).BaseType, []byte{})
		c.closeParen()
		return
	}

	c.openParen()
	c.openParen()
	c.Type(expr.Name, []byte{})
	c.closeParen()

	c.openCurlyBrace()
	if len(expr.Data.Fields) > 0 {
		for i, field := range expr.Data.Fields {
			c.dot()
			c.identifier(field)
			c.space()
			c.equal()
			c.space()
			c.expression(expr.Data.Values[i])
			c.comma()
			c.space()
		}
	} else {
		for _, val := range expr.Data.Values {
			c.expression(val)
			c.comma()
			c.space()
		}
	}
	c.closeCurlyBrace()
	c.closeParen()
}

func (c *Compiler) asyncCompoundLiteral(expr CompoundLiteral) int {

	switch expr.Name.(type) {
	case VecType:
		c.append([]byte("new4"))
		c.openParen()
		c.Type(expr.Name.(VecType).BaseType, []byte{})
		c.comma()
		c.space()
		c.openParen()
		a := c.asyncExpr(CompoundLiteral{Name: ImplictArrayType{BaseType: expr.Name.(VecType).BaseType}, Data: CompoundLiteralData{Fields: expr.Data.Fields, Values: expr.Data.Values}})
		c.closeParen()
		c.comma()
		c.space()
		c.append([]byte(strconv.Itoa(len(expr.Data.Values))))
		c.closeParen()
		return a
	case PromiseType:
		c.append([]byte("new5"))
		c.openParen()
		c.Type(expr.Name.(PromiseType).BaseType, []byte{})
		c.closeParen()
		return 0
	}

	c.openParen()
	c.Type(expr.Name, []byte{})
	c.closeParen()

	a := 0
	c.openCurlyBrace()
	if len(expr.Data.Fields) > 0 {
		for i, field := range expr.Data.Fields {
			c.dot()
			c.identifier(field)
			c.space()
			c.equal()
			c.space()
			a += c.asyncExpr(expr.Data.Values[i])
			c.comma()
			c.space()
		}
	} else {
		for _, val := range expr.Data.Values {
			a += c.asyncExpr(val)
			c.comma()
			c.space()
		}
	}
	c.closeCurlyBrace()
	return a
}

func (c *Compiler) functionCall(call CallExpr) {
	c.expression(call.Function)
	c.openParen()

	if len(call.Args) > 0 {
		c.expression(call.Args[0])

		for i := 1; i < len(call.Args); i++ {
			c.comma()
			c.space()
			c.expression(call.Args[i])
		}
	}
	c.closeParen()
}

func (c *Compiler) asyncFunctionCall(call CallExpr) int {
	c.expression(call.Function)
	c.openParen()
	a := 0

	if len(call.Args) > 0 {
		a = c.asyncExpr(call.Args[0])

		for i := 1; i < len(call.Args); i++ {
			c.comma()
			c.space()
			a += c.asyncExpr(call.Args[i])
		}
	}
	c.closeParen()
	return a
}

func (c *Compiler) declarationType(Typ Type, Name Token) {
	c.decType(Typ, IdentExpr{Value: Name})
	/*
		typ := Typ
		sizes := []Token{}
		pointers := 0

		for {
			switch typ.(type) {
			case DynamicType:
				pointers++
				typ = typ.(DynamicType).BaseType
				continue
			}
			break
		}

		for {
			switch typ.(type) {
			case PointerType:
				pointers++
				typ = typ.(PointerType).BaseType
				continue
			}
			break
		}

		for {
			switch typ.(type) {
			case ArrayType:
				sizes = append(sizes, typ.(ArrayType).Size)
				typ = typ.(ArrayType).BaseType
				continue
			case ImplictArrayType:
				sizes = append(sizes, Token{
					Buff: []byte(""),
				})
				typ = typ.(ImplictArrayType).BaseType
				continue
			}
			break
		}

		for {
			switch typ.(type) {
			case FuncType:
				c.Type(typ.(FuncType).ReturnTypes[0])
				c.space()
				c.openParen()
				c.append([]byte("^"))
				for i := 0; i < pointers; i++ {
					c.append([]byte("*"))
				}
				c.identifier(Name)
				for _, size := range sizes {
					c.openBrace()
					c.identifier(size)
					c.closeBrace()
				}
				c.closeParen()

				argNames := typ.(FuncType).ArgNames
				argTypes := typ.(FuncType).ArgTypes

				c.openParen()
				if len(argNames) > 0 {
					c.declarationType(argTypes[0], argNames[0])

					for i := 1; i < len(argNames); i++ {
						c.comma()
						c.space()
						c.declarationType(argTypes[i], argNames[i])
					}
				} else {
					c.Type(argTypes[0])
					for i := 1; i < len(argNames); i++ {
						c.comma()
						c.space()
						c.Type(argTypes[i])
					}
				}

				c.closeParen()
			case StructType:
				c.strct(typ.(StructType))
				c.space()
				c.identifier(Name)
			case EnumType:
				c.enum(typ.(EnumType))
				c.space()
				c.identifier(Name)
			case TupleType:
				c.tupl(typ.(TupleType))
				c.space()
				c.identifier(Name)
			case UnionType:
				c.union(typ.(UnionType))
				c.space()
				c.identifier(Name)
			default:
				c.Type(typ)
				c.space()
				c.openParen()
				for i := 0; i < pointers; i++ {
					c.append([]byte("*"))
				}
				c.identifier(Name)
				c.closeParen()
				for _, size := range sizes {
					c.openBrace()
					c.identifier(size)
					c.closeBrace()
				}
			}
			break
		}
	*/
}

func (c *Compiler) decType(Typ Type, expr Expression) {
	switch Typ.(type) {
	case ArrayType:
		c.decType(Typ.(ArrayType).BaseType, ArrayMemberExpr{Parent: expr, Index: IdentExpr{Value: Typ.(ArrayType).Size}})
	case ImplictArrayType:
		c.decType(Typ.(ImplictArrayType).BaseType, expr)
		c.openBrace()
		c.closeBrace()
	case PointerType:
		c.decType(Typ.(PointerType).BaseType, UnaryExpr{Op: Token{PrimaryType: AirthmaticOperator, SecondaryType: Mul, Buff: []byte("*")}, Expr: expr})
	case BasicType:
		c.expression(Typ.(BasicType).Expr)
		if expr != nil {
			c.space()
			c.expression(expr)
		}
	case FuncType:
		c.funcDec(Typ, expr, nil, false)
	case StructType:
		c.strct(Typ.(StructType))
		if expr != nil {
			c.space()
			c.expression(expr)
		}
	case EnumType:
		c.enum(Typ.(EnumType))
		if expr != nil {
			c.space()
			c.expression(expr)
		}
	case TupleType:
		c.tupl(Typ.(TupleType))
		if expr != nil {
			c.space()
			c.expression(expr)
		}
	case UnionType:
		c.union(Typ.(UnionType))
		if expr != nil {
			c.space()
			c.expression(expr)
		}
	case ConstType:
		c.append([]byte("const "))
		c.decType(Typ.(ConstType).BaseType, expr)
	case CaptureType:
		c.append([]byte("__block "))
		c.decType(Typ.(CaptureType).BaseType, expr)
	case StaticType:
		c.append([]byte("static "))
		c.decType(Typ.(StaticType).BaseType, expr)
	case VecType:
		c.append([]byte("VECTOR_TYPE("))
		c.Type(Typ.(VecType).BaseType, []byte{})
		c.closeParen()
		c.expression(expr)
	case PromiseType:
		c.append([]byte("PROMISE_TYPE("))
		c.Type(Typ.(PromiseType).BaseType, []byte{})
		c.closeParen()
		c.expression(expr)
	}
}

func (c *Compiler) funcExprType(Typ Expression, expr Expression, expr2 Expression, noReturn bool) {
	t := Typ.(FuncType)

	if noReturn {
		switch expr.(type) {
		case FuncType:
			c.openParen()
			c.append([]byte("^"))
			c.funcExprType(expr, expr2, nil, true)
			c.closeParen()
		}

		argNames := t.ArgNames
		argTypes := t.ArgTypes

		c.openParen()
		if len(argNames) > 0 {
			c.decType(argTypes[0], IdentExpr{Value: argNames[0]})

			for i := 1; i < len(argNames); i++ {
				c.comma()
				c.space()
				c.decType(argTypes[i], IdentExpr{Value: argNames[i]})
			}
		} else {
			c.Type(argTypes[0], []byte{})
			for i := 1; i < len(argNames); i++ {
				c.comma()
				c.space()
				c.Type(argTypes[i], []byte{})
			}
		}
		c.closeParen()
		return
	}

	rType := t.ReturnTypes[0]

	switch rType.(type) {
	case FuncType:
		rt := rType.(FuncType).ReturnTypes[0]
		switch rt.(type) {
		case FuncType:
			break
		default:
			c.append([]byte("^"))
			c.Type(rt, []byte{})
		}
		c.funcExprType(rType, Typ, expr, true)
	default:
		c.append([]byte("^"))
		c.Type(rType, []byte{})
		c.space()
		c.funcExprType(t, nil, nil, true)
	}
}
func (c *Compiler) funcDec(Typ Expression, expr Expression, expr2 Expression, noReturn bool) {
	t := Typ.(FuncType)

	if noReturn {
		c.openParen()
		c.append([]byte("^"))

		switch expr.(type) {
		case FuncType:
			c.funcDec(expr, expr2, nil, true)
		default:
			c.expression(expr)
		}
		c.closeParen()

		argTypes := t.ArgTypes

		c.openParen()
		c.Type(argTypes[0], []byte{})
		for i := 1; i < len(argTypes); i++ {
			c.comma()
			c.space()
			c.Type(argTypes[i], []byte{})
		}
		c.closeParen()
		return
	}

	rType := t.ReturnTypes[0]

	switch rType.(type) {
	case FuncType:
		rt := rType.(FuncType).ReturnTypes[0]
		switch rt.(type) {
		case FuncType:
			break
		default:
			c.Type(rt, []byte{})
		}
		c.funcDec(rType, Typ, expr, true)
	default:
		c.Type(rType, []byte{})
		c.space()
		c.funcDec(Typ, expr, nil, true)
	}
}

func (c *Compiler) Type(Typ Type, buf []byte) {
	switch Typ.(type) {
	case ArrayType:
		c.Type(Typ.(ArrayType).BaseType, buf)
		c.openBrace()
		c.identifier(Typ.(ArrayType).Size)
		c.closeBrace()
	case ImplictArrayType:
		c.Type(Typ.(ImplictArrayType).BaseType, buf)
		c.openBrace()
		c.closeBrace()
	case PointerType:
		buf = append(buf, '*')
		c.Type(Typ.(PointerType).BaseType, buf)
	case BasicType:
		c.expression(Typ.(BasicType).Expr)
		c.append(buf)
	case FuncType:
		c.funcDec(Typ, IdentExpr{Value: Token{Buff: buf, PrimaryType: Identifier}}, nil, false)
	case StructType:
		c.strct(Typ.(StructType))
	case EnumType:
		c.enum(Typ.(EnumType))
	case TupleType:
		c.tupl(Typ.(TupleType))
	case UnionType:
		c.union(Typ.(UnionType))
	case ConstType:
		c.append([]byte("const "))
		c.Type(Typ.(ConstType).BaseType, buf)
	case CaptureType:
		c.append([]byte("__block "))
		c.Type(Typ.(CaptureType).BaseType, buf)
	case StaticType:
		c.append([]byte("static "))
		c.Type(Typ.(StaticType).BaseType, buf)
	case VecType:
		c.append([]byte("void *"))
		// c.Type(Typ.(VecType).BaseType, buf)
		// c.closeParen()
	case PromiseType:
		c.append([]byte("PROMISE_TYPE("))
		c.Type(Typ.(PromiseType).BaseType, buf)
		c.closeParen()
	}
}

func (c *Compiler) TypeVoidVec(Typ Type, buf []byte) {
	switch Typ.(type) {
	case ArrayType:
		c.Type(Typ.(ArrayType).BaseType, buf)
		c.openBrace()
		c.identifier(Typ.(ArrayType).Size)
		c.closeBrace()
	case ImplictArrayType:
		c.Type(Typ.(ImplictArrayType).BaseType, buf)
		c.openBrace()
		c.closeBrace()
	case PointerType:
		buf = append(buf, '*')
		c.Type(Typ.(PointerType).BaseType, buf)
	case BasicType:
		c.expression(Typ.(BasicType).Expr)
		c.append(buf)
	case FuncType:
		c.funcDec(Typ, IdentExpr{Value: Token{Buff: buf, PrimaryType: Identifier}}, nil, false)
	case StructType:
		c.strct(Typ.(StructType))
	case EnumType:
		c.enum(Typ.(EnumType))
	case TupleType:
		c.tupl(Typ.(TupleType))
	case UnionType:
		c.union(Typ.(UnionType))
	case ConstType:
		c.append([]byte("const "))
		c.Type(Typ.(ConstType).BaseType, buf)
	case CaptureType:
		c.append([]byte("__block "))
		c.Type(Typ.(CaptureType).BaseType, buf)
	case StaticType:
		c.append([]byte("static "))
		c.Type(Typ.(StaticType).BaseType, buf)
	case VecType:
		c.append([]byte("void *"))
		// c.Type(Typ.(VecType).BaseType, buf)
		// c.closeParen()
	case PromiseType:
		c.append([]byte("PROMISE_TYPE("))
		c.Type(Typ.(PromiseType).BaseType, buf)
		c.closeParen()
	}
}

func (c *Compiler) ifElse(ifElse IfElseBlock) {
	if ifElse.HasInitStmt {
		c.indent()
		c.openCurlyBrace()
		c.pushScope()
		c.statement(ifElse.InitStatement)
		c.newline()
	}

	c.indent()
	for i, condition := range ifElse.Conditions {
		c.append([]byte("if"))
		c.openParen()
		c.expression(condition)
		c.closeParen()
		c.block(ifElse.Blocks[i])
		c.append([]byte(" else "))
	}

	c.block(ifElse.ElseBlock)

	if ifElse.HasInitStmt {
		c.popScope()
		c.newline()
		c.indent()
		c.closeCurlyBrace()
	}
}

func (c *Compiler) asyncIfElse(ifElse IfElseBlock) ([]Declaration, int) {
	decs := []Declaration{}
	a := 0

	if ifElse.HasInitStmt {
		c.indent()
		c.openCurlyBrace()
		c.pushScope()
		decs, a = c.asyncStmt(ifElse.InitStatement)
		c.newline()
	}

	c.indent()
	for i, condition := range ifElse.Conditions {
		c.append([]byte("if"))
		c.openParen()
		a += c.asyncExpr(condition)
		c.closeParen()

		decs2, a2 := c.asyncBlock(ifElse.Blocks[i])

		decs = append(decs, decs2...)
		a += a2

		c.append([]byte(" else "))
	}

	decs2, a2 := c.asyncBlock(ifElse.ElseBlock)

	decs = append(decs, decs2...)
	a += a2

	if ifElse.HasInitStmt {
		c.popScope()
		c.newline()
		c.indent()
		c.closeCurlyBrace()
	}
	return decs, a
}

func (c *Compiler) assignment(as Assignment) {
	c.indent()
	c.openParen()
	c.openCurlyBrace()
	for i, Var := range as.Variables {
		c.expression(Var)
		c.space()
		c.operator(as.Op)
		c.space()
		if len(as.Values) > 1 {
			c.expression(as.Values[i])
		} else {
			c.expression(as.Values[0])
		}
		c.semicolon()
	}
	c.closeCurlyBrace()
	c.closeParen()
}

func (c *Compiler) asyncAssignment(as Assignment) int {
	c.indent()
	c.openParen()
	c.openCurlyBrace()

	a := 0

	for i, Var := range as.Variables {
		c.asyncExpr(Var)
		c.space()
		c.operator(as.Op)
		c.space()
		if len(as.Values) > 1 {
			a += c.asyncExpr(as.Values[i])
		} else {
			a += c.asyncExpr(as.Values[0])
		}
		c.semicolon()
	}
	c.closeCurlyBrace()
	c.closeParen()

	return a
}

func (c *Compiler) swtch(swtch Switch) {
	if swtch.Type == InitCondSwitch {
		c.indent()
		c.openCurlyBrace()
		c.pushScope()
		c.statement(swtch.InitStatement)
		c.newline()
	}

	c.indent()
	c.append([]byte("switch"))

	c.openParen()
	if swtch.Type == NoneSwtch {
		c.append([]byte("1"))
	} else {
		c.expression(swtch.Expr)
	}
	c.closeParen()
	c.openCurlyBrace()

	for _, Case := range swtch.Cases {
		c.newline()
		c.indent()
		c.append([]byte("case"))
		c.space()
		c.expression(Case.Condition)
		c.colon()

		c.pushScope()
		for _, stmt := range Case.Block.Statements {
			c.statement(stmt)
		}
		c.popScope()
	}

	if swtch.HasDefaultCase {
		c.newline()
		c.indent()
		c.append([]byte("default"))
		c.colon()

		c.pushScope()
		for _, stmt := range swtch.DefaultCase.Statements {
			c.statement(stmt)
		}
		c.popScope()
	}

	c.newline()
	c.indent()
	c.closeCurlyBrace()

	if swtch.Type == InitCondSwitch {
		c.popScope()
		c.newline()
		c.indent()
		c.closeCurlyBrace()
	}
}

func (c *Compiler) asyncSwitch(swtch Switch) ([]Declaration, int) {
	decs, a := []Declaration{}, 0

	if swtch.Type == InitCondSwitch {
		c.indent()
		c.openCurlyBrace()
		c.pushScope()
		decs, a = c.asyncStmt(swtch.InitStatement)
		c.newline()
	}

	c.indent()
	c.append([]byte("switch"))

	c.openParen()
	if swtch.Type == NoneSwtch {
		c.append([]byte("1"))
	} else {
		a += c.asyncExpr(swtch.Expr)
	}
	c.closeParen()
	c.openCurlyBrace()

	for _, Case := range swtch.Cases {
		c.newline()
		c.indent()
		c.append([]byte("case"))
		c.space()
		a += c.asyncExpr(Case.Condition)
		c.colon()

		c.pushScope()
		for _, stmt := range Case.Block.Statements {
			decs2, a2 := c.asyncStmt(stmt)

			decs = append(decs, decs2...)
			a += a2
		}
		c.popScope()
	}

	if swtch.HasDefaultCase {
		c.newline()
		c.indent()
		c.append([]byte("default"))
		c.colon()

		c.pushScope()
		for _, stmt := range swtch.DefaultCase.Statements {
			decs2, a2 := c.asyncStmt(stmt)
			decs = append(decs, decs2...)
			a += a2
		}
		c.popScope()
	}

	c.newline()
	c.indent()
	c.closeCurlyBrace()

	if swtch.Type == InitCondSwitch {
		c.popScope()
		c.newline()
		c.indent()
		c.closeCurlyBrace()
	}

	return decs, a
}

func (c *Compiler) strctPropDeclaration(dec Declaration) {
	for i, Var := range dec.Identifiers {
		t := dec.Types[i]
		switch t.(type) {
		case FuncType:
			if !t.(FuncType).Mut {
				continue
			}
		}
		c.indent()
		c.declarationType(dec.Types[i], Var)
		c.semicolon()
		c.newline()
	}
}

func (c *Compiler) strct(typ StructType) {
	c.append([]byte("struct "))
	c.openCurlyBrace()
	c.pushScope()
	c.newline()
	for _, prop := range typ.Props {
		c.strctPropDeclaration(prop)
	}
	c.popScope()
	c.indent()
	c.closeCurlyBrace()
}

func (c *Compiler) strctDefault(strct Typedef) {
	c.identifier(strct.Name)
	c.space()
	c.identifier(strct.DefaultName)

	c.space()
	c.equal()
	c.space()

	c.openParen()
	c.identifier(strct.Name)
	c.closeParen()

	c.openCurlyBrace()
	for _, prop := range strct.Type.(StructType).Props {
		if len(prop.Values) == 0 {
			continue
		}
		for x, Ident := range prop.Identifiers {
			t := prop.Types[x]
			switch t.(type) {
			case FuncType:
				if !t.(FuncType).Mut {
					continue
				}
			}
			c.dot()
			c.identifier(Ident)
			c.space()
			c.equal()
			c.space()
			c.expression(prop.Values[x])
			c.comma()
			c.space()
		}
	}
	c.closeCurlyBrace()
	c.semicolon()
	c.newline()
}

func (c *Compiler) strctMethods(strct StructType) {
	for _, prop := range strct.Props {
		for i, val := range prop.Values {
			t := prop.Types[i]
			switch t.(type) {
			case FuncType:
				if t.(FuncType).Mut {
					continue
				}
			default:
				continue
			}
			c.declarationType(prop.Types[i], prop.Identifiers[i])
			c.space()
			c.equal()
			c.space()
			c.expression(val)
			c.semicolon()
			c.newline()
		}
	}
}

func (c *Compiler) strctMethodsOnlyDec(strct StructType) {
	for _, prop := range strct.Props {
		for i := range prop.Values {
			switch prop.Types[i].(type) {
			case FuncType:
				if prop.Types[i].(FuncType).Mut {
					break
				}
			default:
				continue
			}
			c.declarationType(prop.Types[i], prop.Identifiers[i])
			c.semicolon()
			c.newline()
		}
	}
}

func (c *Compiler) enum(en EnumType) {
	c.append([]byte("enum {"))
	c.newline()
	c.pushScope()

	for x, prop := range en.Identifiers {
		c.indent()
		c.identifier(prop)
		val := en.Values[x]

		if val != nil {
			c.space()
			c.equal()
			c.space()
			c.expression(val)
		}
		c.comma()
		c.newline()
	}

	c.popScope()
	c.closeCurlyBrace()
}

func (c *Compiler) union(union UnionType) {
	c.append([]byte("union {"))
	c.newline()
	c.pushScope()

	for x, prop := range union.Identifiers {
		c.indent()
		c.declarationType(union.Types[x], prop)
		c.semicolon()
		c.newline()
	}
	c.popScope()
	c.closeCurlyBrace()
}

func (c *Compiler) tupl(tupl TupleType) {
	c.append([]byte("struct {"))
	c.newline()
	c.pushScope()

	for x, prop := range tupl.Types {
		c.indent()
		c.declarationType(prop, Token{
			Buff:        []byte("_" + strconv.Itoa(x)),
			PrimaryType: Identifier,
		})
		c.semicolon()
		c.newline()
	}

	c.popScope()
	c.closeCurlyBrace()
}

func (c *Compiler) heapAlloc(expr HeapAlloc) {
	switch expr.Type.(type) {
	case ArrayType:
		if expr.Val != nil {
			c.append([]byte("new3"))
			c.openParen()
			c.Type(expr.Type.(ArrayType).BaseType, []byte{})
			c.comma()
			c.Type(expr.Type, []byte{})
			c.comma()
			c.openParen()
			c.expression(expr.Val)
			c.closeParen()
		} else {
			c.append([]byte("new"))
			c.openParen()
			c.Type(expr.Type.(ArrayType).BaseType, []byte{})
			c.comma()
			c.Type(expr.Type, []byte{})
		}
	default:
		if expr.Val != nil {
			c.append([]byte("new2"))
			c.openParen()
			c.Type(expr.Type, []byte{})
			c.comma()
			c.Type(expr.Type, []byte{})
			c.comma()
			c.openParen()
			c.expression(expr.Val)
			c.closeParen()
		} else {
			c.append([]byte("new"))
			c.openParen()
			c.Type(expr.Type, []byte{})
			c.comma()
			c.Type(expr.Type, []byte{})
		}
	}
	c.closeParen()
}

/*
func (c *Compiler) lenExpr(expr LenExpr) {

	switch expr.Type.(type) {
	case DynamicType:
		Typ := expr.Type.(DynamicType).BaseType
		c.append([]byte("len"))
		c.openParen()
		c.expression(expr.Expr)
		c.comma()
		switch Typ.(type) {
		case ImplictArrayType:
			c.Type(Typ.(ImplictArrayType).BaseType)
		default:
			c.Type(Typ)
		}
		c.closeParen()
	case ArrayType:
		c.append([]byte("len2"))
		c.openParen()
		c.Type(expr.Type)
		c.comma()
		c.Type(expr.Type.(ArrayType).BaseType)
		c.closeParen()
	default:
		c.append([]byte("len3"))
		c.openParen()
		c.Type(expr.Type)
		c.closeParen()
	}
}

func (c *Compiler) sizeExpr(expr SizeExpr) {
	switch expr.Type.(type) {
	case DynamicType:
		c.append([]byte("size"))
	default:
		c.append([]byte("size2"))
	}
	c.openParen()
	c.expression(expr.Expr)
	c.closeParen()
}

*/

func (c *Compiler) compareTypes(Type1 Type, Type2 BasicType) bool {
	switch Type1.(type) {
	case BasicType:
		switch Type1.(BasicType).Expr.(type) {
		case IdentExpr:
			switch Type2.Expr.(type) {
			case IdentExpr:
				break
			default:
				return false
			}
			return bytes.Compare(Type1.(BasicType).Expr.(IdentExpr).Value.Buff, Type2.Expr.(IdentExpr).Value.Buff) == 0
		}
	case PromiseType:
		return c.compareTypes(Type1.(PromiseType).BaseType, Type2)
	case ArrayType:
		return c.compareTypes(Type1.(ArrayType).BaseType, Type2)
	case ConstType:
		return c.compareTypes(Type1.(ConstType).BaseType, Type2)
	}
	return false
}
