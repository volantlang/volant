package compiler

import (
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
			c.newline()
			/*
				case Import:
					c.imprt(stmt.(Import))
					c.newline()
			*/
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
		case Typedef:
			c.typedefOnlyDec(stmt.(Typedef))
			c.newline()
		case ExportStatement:
			stmt2 := stmt.(ExportStatement).Stmt
			switch stmt2.(type) {
			case Declaration:
				c.onlyDeclaration(stmt2.(Declaration), true)
				c.newline()
			case Typedef:
				c.typedefOnlyDec(stmt2.(Typedef))
				c.newline()
			}
		case NullStatement:
			c.semicolon()
			c.newline()
		case Import:
			c.imprt(stmt.(Import))
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
	default:
		c.indent()
		c.expression(stmt.(Expression))
		c.semicolon()
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
	c.declarationType(typedef.Type, typedef.Name)
	c.semicolon()
	switch typedef.Type.(type) {
	case StructType:
		c.newline()
		c.strctDefault(typedef)
		c.strctMethodsOnlyDec(typedef.Type.(StructType))
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

	if loop.Type&InitLoop == InitLoop {
		c.indent()
		c.openCurlyBrace()
		c.pushScope()
		c.statement(loop.InitStatement)
	}

	c.newline()
	c.indent()
	c.append([]byte("while"))
	c.openParen()

	if loop.Type&CondLoop == CondLoop {
		c.expression(loop.Condition)
	} else {
		c.append([]byte("1"))
	}

	c.closeParen()

	if loop.Type&LoopLoop == LoopLoop {
		loop.Block.Statements = append(loop.Block.Statements, loop.LoopStatement)
	}

	c.block(loop.Block)

	if loop.Type&InitLoop == InitLoop {
		c.popScope()
		c.newline()
		c.indent()
		c.closeCurlyBrace()
	}
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

func (c *Compiler) expression(expr Expression) {

	switch expr.(type) {
	case Type:
		c.Type(expr.(Type), []byte{})
	case CallExpr:
		c.functionCall(expr.(CallExpr))
	case BasicLit:
		c.identifier(expr.(BasicLit).Value)
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
		c.openParen()
		c.expression(expr.(CompoundLiteral).Name)
		c.closeParen()

		c.openCurlyBrace()
		if len(expr.(CompoundLiteral).Data.Fields) > 0 {
			for i, field := range expr.(CompoundLiteral).Data.Fields {
				c.dot()
				c.identifier(field)
				c.space()
				c.equal()
				c.space()
				c.expression(expr.(CompoundLiteral).Data.Values[i])
				c.comma()
				c.space()
			}
		} else {
			for _, val := range expr.(CompoundLiteral).Data.Values {
				c.expression(val)
				c.comma()
				c.space()
			}
		}
		c.closeCurlyBrace()
	case TypeCast:
		c.openParen()
		c.Type(expr.(TypeCast).Type.(Type), []byte{})
		c.closeParen()
		c.openParen()
		c.expression(expr.(TypeCast).Expr)
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
		c.funcExprType(expr.(FuncExpr).Type, nil, nil, false)
		c.block(expr.(FuncExpr).Block)
	}
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
		c.decType(Typ.(ArrayType).BaseType, expr)
		c.openBrace()
		c.identifier(Typ.(ArrayType).Size)
		c.closeBrace()
	case ImplictArrayType:
		c.decType(Typ.(ImplictArrayType).BaseType, expr)
		c.openBrace()
		c.closeBrace()
	case PointerType:
		c.decType(Typ.(PointerType).BaseType, UnaryExpr{Op: Token{PrimaryType: AirthmaticOperator, SecondaryType: Mul, Buff: []byte("*")}, Expr: expr})
	case DynamicType:
		c.decType(Typ.(DynamicType).BaseType, UnaryExpr{Op: Token{PrimaryType: AirthmaticOperator, SecondaryType: Mul, Buff: []byte("*")}, Expr: expr})
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
	case DynamicType:
		buf = append(buf, '*')
		c.Type(Typ.(DynamicType).BaseType, buf)
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
	}
}

func (c *Compiler) ifElse(ifElse IfElseBlock) {
	if ifElse.HasInitStmt {
		c.indent()
		c.openCurlyBrace()
		c.pushScope()
		c.statement(ifElse.InitStatement)
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

func (c *Compiler) assignment(as Assignment) {
	for i, Var := range as.Variables {
		c.indent()
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
}

func (c *Compiler) swtch(swtch Switch) {
	if swtch.Type == InitCondSwitch {
		c.indent()
		c.openCurlyBrace()
		c.pushScope()
		c.statement(swtch.InitStatement)
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
			break
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
			}
			c.declarationType(prop.Types[i], prop.Identifiers[i])
			// c.block(val.(FuncExpr).Block)
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
				break
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
