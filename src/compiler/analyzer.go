package compiler

import (
	"bytes"
	"error"
	. "parser"
	"path"
	"strconv"
	"strings"
)

type SemanticAnalyzer struct {
	Symbols        *SymbolTable
	Imports        map[string]*SymbolTable
	ImportPrefixes map[string][]byte
	Exports        *SymbolTable
	Path           string
	Index          int
}

func AnalyzeFile(ast File, pathh string) (*SymbolTable, map[string]*SymbolTable, map[string][]byte, *SymbolTable, int) {
	n := num
	num++
	s := SemanticAnalyzer{
		Symbols:        &SymbolTable{},
		Exports:        &SymbolTable{},
		Imports:        map[string]*SymbolTable{},
		ImportPrefixes: map[string][]byte{},
		Path:           pathh,
		Index:          num,
	}
	s.addSymbol(I8Token, I8Type)
	s.addSymbol(I16Token, I16Type)
	s.addSymbol(I32Token, I32Type)
	s.addSymbol(I64Token, I64Type)

	s.addSymbol(U8Token, U8Type)
	s.addSymbol(U16Token, U16Type)
	s.addSymbol(U32Token, U32Type)
	s.addSymbol(U64Token, U64Type)

	s.addSymbol(F32Token, F32Type)
	s.addSymbol(F64Token, F64Type)

	s.addSymbol(UptrToken, UptrType)

	s.addSymbol(VoidToken, VoidType)
	s.addSymbol(SizeTToken, SizeTType)
	s.addSymbol(BoolToken, BoolType)

	s.addSymbol(True.Value, BoolType.Type)
	s.addSymbol(False.Value, BoolType.Type)
	s.addSymbol(Null.Value, VoidType.Type)

	for _, statement := range ast.Statements {
		s.globalStmt(statement)
	}
	return s.Symbols, s.Imports, s.ImportPrefixes, s.Exports, n
}

func (s *SemanticAnalyzer) error(message string, line, column int) {
	error.New(s.Path+": "+message, line, column)
}

func (s *SemanticAnalyzer) getSymbol(Ident Token, Curr bool) (Node, bool) {
	if Curr {
		return s.Symbols.Find(Ident)
	}
	return s.Symbols.FindInAll(Ident)
}

func (s *SemanticAnalyzer) addSymbol(Ident Token, Type Type) {
	s.Symbols.Add(Node{Identifier: Ident, Type: Type})
}

func (s *SemanticAnalyzer) pushScope() {
	s.Symbols = s.Symbols.NewChild()
}

func (s *SemanticAnalyzer) popScope() {
	s.Symbols = s.Symbols.Parent
}

func (s *SemanticAnalyzer) globalStmt(stmt Statement) {
	switch stmt.(type) {
	case Typedef:
		s.typedef(stmt.(Typedef))
	case Import:
		s.imprt(stmt.(Import))
	case Declaration:
		s.declaration(stmt.(Declaration))
	case ExportStatement:
		st := stmt.(ExportStatement).Stmt
		switch st.(type) {
		case Declaration:
			s.exportDeclaration(st.(Declaration))
		case Typedef:
			s.exportTypedef(st.(Typedef))
		default:
			s.error("Invalid export statement, expected typedef or decleration, got {st}", st.LineM(), st.ColumnM())
		}
	case NullStatement:
		break
	default:
		s.error("Non-declarative statement outside function body.", stmt.LineM(), stmt.ColumnM())
	}
}

func (s *SemanticAnalyzer) stmt(stmt Statement, returnType Type) {
	switch stmt.(type) {
	case Declaration:
		s.declaration(stmt.(Declaration))
	case IfElseBlock:
		s.ifElse(stmt.(IfElseBlock), returnType)
	case Loop:
		s.loop(stmt.(Loop), returnType)
	case Typedef:
		s.typedef(stmt.(Typedef))
	case Switch:
		s.swtch(stmt.(Switch), returnType)
	case Block:
		s.pushScope()
		s.block(stmt.(Block), returnType)
		s.popScope()
	case Assignment:
		s.assignment(stmt.(Assignment))
	case Delete:
		s.delete(stmt.(Delete))
	case Expression:
		s.expr(stmt.(Expression))
	case Return:
		s.rturn(stmt.(Return), returnType)
	}
}

func (s *SemanticAnalyzer) basicStmt(stmt Statement) {
	switch stmt.(type) {
	case Declaration:
		s.declaration(stmt.(Declaration))
	case Assignment:
		s.assignment(stmt.(Assignment))
	case Delete:
		s.delete(stmt.(Delete))
	case Expression:
		s.expr(stmt.(Expression))
	case NullStatement:
		break
	default:
		s.error("Invalid statement {stmt}.", stmt.LineM(), stmt.ColumnM())
	}
}

func (s *SemanticAnalyzer) delete(del Delete) {
	for _, expr := range del.Exprs {
		s.expr(expr)
	}
}

func (s *SemanticAnalyzer) imprt(stmt Import) {
	for _, Path := range stmt.Paths {
		path1 := path.Clean(string(Path.Buff[1 : len(Path.Buff)-1]))
		tmp := ImportFile(path.Dir(s.Path), path1, false, s.Index)

		if path.Ext(path1) != ".h" {
			name := strings.Split(path.Base(path1), ".")[0]

			s.ImportPrefixes[name] = []byte(getLastImportPrefix())
			s.Imports[name] = tmp
		}
	}
}

func (s *SemanticAnalyzer) declaration(dec Declaration) {
	Types := []Type{}

	if len(dec.Types) == 1 {
		Type := dec.Types[0]

		if len(dec.Values) == 0 {
			Types = append(Types, Type)
		}
		for _, val := range dec.Values {
			Types = append(Types, Type)
			Type2 := s.getType(val)
			if s.compareTypes(Type2, Type) {
				continue
			}
			s.error("Type mismatch: val has type {Type2}, expected {Type}.", val.LineM(), val.ColumnM())
		}
	} else if len(dec.Types) == 0 {
		if len(dec.Values) == 0 {
			s.error("Cannot declare variable without type.", dec.Identifiers[0].Line, dec.Identifiers[0].Column)
		}
		for _, val := range dec.Values {
			Typ := s.getType(val)

			switch Typ.(type) {
			case NumberType:
				if bytes.Contains(val.(BasicLit).Value.Buff, []byte(".")) {
					Typ = F32Type
				} else {
					Typ = I32Type.Type
				}
			}
			Types = append(Types, Typ)
		}
	} else if len(dec.Types) != len(dec.Values) {
		s.error("Invalid number of types or values specified", dec.Identifiers[0].Line, dec.Identifiers[0].Column)
	} else {
		Types = dec.Types
	}

	for i, Ident := range dec.Identifiers {
		if _, ok := s.getSymbol(Ident, true); ok {
			s.error(string(Ident.Buff)+" has already been declared.", Ident.Line, Ident.Column)
		} else {
			s.addSymbol(Ident, Types[i])
		}
	}
	for _, typ := range Types {
		s.typ(typ)
	}
	for _, Val := range dec.Values {
		s.expr(Val)
	}
}

func (s *SemanticAnalyzer) exportDeclaration(dec Declaration) {
	Types := []Type{}

	if len(dec.Types) == 1 {
		Type := dec.Types[0]

		if len(dec.Values) == 0 {
			Types = append(Types, Type)
		}
		for _, val := range dec.Values {
			Types = append(Types, Type)
			Type2 := s.getType(val)
			if s.compareTypes(Type2, Type) {
				continue
			}
			s.error("Type mismatch: val has type {Type2}, expected {Type}.", val.LineM(), val.ColumnM())
		}
	} else if len(dec.Types) == 0 {
		if len(dec.Values) == 0 {
			s.error("Cannot declare variable without type.", dec.Identifiers[0].Line, dec.Identifiers[0].Column)
		}
		for _, val := range dec.Values {
			Typ := s.getType(val)
			switch Typ.(type) {
			case NumberType:
				Typ = I32Type.Type
			}
			Types = append(Types, Typ)
		}
	} else if len(dec.Types) != len(dec.Values) {
		s.error("Invalid number of types or values specified", dec.Identifiers[0].Line, dec.Identifiers[0].Column)
	} else {
		Types = dec.Types
	}
	for i, Ident := range dec.Identifiers {
		if _, ok := s.getSymbol(Ident, true); ok {
			s.error(string(Ident.Buff)+" has already been declared.", Ident.Line, Ident.Column)
		} else {
			s.addSymbol(Ident, Types[i])
			s.Exports.Add(Node{Identifier: Ident, Type: Types[i]})
		}
	}

	for _, typ := range Types {
		s.typ(typ)
	}
	for _, Val := range dec.Values {
		s.expr(Val)
	}
}

func (s *SemanticAnalyzer) ifElse(ifElse IfElseBlock, returnType Type) {
	s.pushScope()
	if ifElse.HasInitStmt {
		s.basicStmt(ifElse.InitStatement)
	}
	s.exprArray(ifElse.Conditions)
	for _, block := range ifElse.Blocks {
		s.pushScope()
		s.block(block, returnType)
		s.popScope()
	}
	s.block(ifElse.ElseBlock, returnType)
	s.popScope()
}

func (s *SemanticAnalyzer) loop(loop Loop, returnType Type) {
	s.pushScope()
	if loop.Type&InitLoop == InitLoop {
		s.basicStmt(loop.InitStatement)
	}
	if loop.Type&CondLoop == CondLoop {
		s.expr(loop.Condition)
	}
	if loop.Type&LoopLoop == LoopLoop {
		s.basicStmt(loop.LoopStatement)
	}
	s.block(loop.Block, returnType)
	s.popScope()
}

func (s *SemanticAnalyzer) swtch(swtch Switch, returnType Type) {
	s.popScope()
	if swtch.Type == 1 {
		s.basicStmt(swtch.InitStatement)
		s.expr(swtch.Expr)
	}
	for _, Case := range swtch.Cases {
		s.expr(Case.Condition)
		s.block(Case.Block, returnType)
	}
	if swtch.HasDefaultCase {
		s.block(swtch.DefaultCase, returnType)
	}
	s.popScope()
}

func (s *SemanticAnalyzer) block(block Block, returnType Type) {
	for _, stmt := range block.Statements {
		s.stmt(stmt, returnType)
	}
}

func (s *SemanticAnalyzer) assignment(as Assignment) {
	s.exprArray(as.Variables)
	s.exprArray(as.Values)

	for x, vr := range as.Variables {
		val := as.Values[x]
		s.expr(val)

		Type1 := s.getType(vr)
		Type2 := s.getType(val)

		switch Type1.(type) {
		/*
			case InternalType:
				s.error("Cannot assign variable of type {Type1} to a value of type \"InternalType\". Needs type casting.", val.LineM(), val.ColumnM())
		*/
		case FuncType:
			if !Type1.(FuncType).Mut {
				s.error("Cannot assign to a constant function.", vr.LineM(), vr.ColumnM())
			}
		case ConstType:
			s.error("Cannot assign to constant variable.", vr.LineM(), vr.ColumnM())
		}

		if !s.compareTypes(Type1, Type2) {
			// s.error("Cannot assign variable of type {Type1} to a value of type {Type2}.", val.LineM(), val.ColumnM())
		}
	}
}

func (s *SemanticAnalyzer) expr(expr Expression) {
	switch expr.(type) {
	case IdentExpr:
		tok := expr.(IdentExpr).Value
		if isInternal(tok) {
			break
		}
		_, ok := s.getSymbol(tok, false)
		if !ok {
			s.error("Use of undeclared variable '"+string(tok.Buff)+"'.", tok.Line, tok.Column)
		}
	case UnaryExpr:
		s.expr(expr.(UnaryExpr).Expr)
	case BinaryExpr:
		bExpr := expr.(BinaryExpr)

		s.expr(bExpr.Left)
		s.expr(bExpr.Right)

		lType := s.getType(bExpr.Left)
		rType := s.getType(bExpr.Right)

		if s.compareTypes(lType, rType) {
			return
		}

		lType = s.getRootType(lType)
		rType = s.getRootType(rType)

		switch lType.(type) {
		case PointerType:
			switch rType.(type) {
			case NumberType:
				return
			}
			if (s.compareTypes(rType, BasicType{Expr: IdentExpr{Value: I8Token}}) ||
				s.compareTypes(rType, BasicType{Expr: IdentExpr{Value: I16Token}}) ||
				s.compareTypes(rType, BasicType{Expr: IdentExpr{Value: I32Token}}) ||
				s.compareTypes(rType, BasicType{Expr: IdentExpr{Value: I64Token}}) ||
				s.compareTypes(rType, BasicType{Expr: IdentExpr{Value: U8Token}}) ||
				s.compareTypes(rType, BasicType{Expr: IdentExpr{Value: U16Token}}) ||
				s.compareTypes(rType, BasicType{Expr: IdentExpr{Value: U32Token}}) ||
				s.compareTypes(rType, BasicType{Expr: IdentExpr{Value: U64Token}}) ||
				s.compareTypes(rType, BasicType{Expr: IdentExpr{Value: SizeTToken}})) {
				return
			}
		}
		switch rType.(type) {
		case PointerType:
			switch lType.(type) {
			case NumberType:
				return
			}
			if (s.compareTypes(lType, BasicType{Expr: IdentExpr{Value: I8Token}}) ||
				s.compareTypes(lType, BasicType{Expr: IdentExpr{Value: I16Token}}) ||
				s.compareTypes(lType, BasicType{Expr: IdentExpr{Value: I32Token}}) ||
				s.compareTypes(lType, BasicType{Expr: IdentExpr{Value: I64Token}}) ||
				s.compareTypes(lType, BasicType{Expr: IdentExpr{Value: U8Token}}) ||
				s.compareTypes(lType, BasicType{Expr: IdentExpr{Value: U16Token}}) ||
				s.compareTypes(lType, BasicType{Expr: IdentExpr{Value: U32Token}}) ||
				s.compareTypes(lType, BasicType{Expr: IdentExpr{Value: U64Token}}) ||
				s.compareTypes(lType, BasicType{Expr: IdentExpr{Value: SizeTToken}})) {
				return
			}
		}

		s.error("Type mismatch: expected {lType}, got {rType}", bExpr.LineM(), bExpr.ColumnM())
	case PostfixUnaryExpr:
		s.expr(expr.(PostfixUnaryExpr).Expr)
	case TernaryExpr:
		s.expr(expr.(TernaryExpr).Cond)
		s.expr(expr.(TernaryExpr).Left)
		s.expr(expr.(TernaryExpr).Right)
	case ArrayLiteral:
		s.exprArray(expr.(ArrayLiteral).Exprs)
	case CallExpr:
		s.callExpr(expr.(CallExpr))
	case TypeCast:
		s.typeCast(expr.(TypeCast))
	case ArrayMemberExpr:
		s.arrayMemberExpr(expr.(ArrayMemberExpr))
	case MemberExpr:
		s.memberExpr(expr.(MemberExpr))
	case LenExpr:
		s.lenExpr(expr.(LenExpr))
	case SizeExpr:
		s.sizeExpr(expr.(SizeExpr))
	case CompoundLiteral:
		s.compoundLiteral(expr.(CompoundLiteral))
	case FuncExpr:
		s.pushScope()
		s.typ(expr.(FuncExpr).Type)
		for i, arg := range expr.(FuncExpr).Type.ArgNames {
			s.addSymbol(arg, expr.(FuncExpr).Type.ArgTypes[i])
		}
		s.block(expr.(FuncExpr).Block, expr.(FuncExpr).Type.ReturnTypes[0])
		s.popScope()
	case HeapAlloc:
		s.typ(expr.(HeapAlloc).Type)
	}
}

func (s *SemanticAnalyzer) callExpr(expr CallExpr) {
	s.expr(expr.Function)
	s.exprArray(expr.Args)

	typ := s.getRootType(s.getType(expr.Function))
	Args := expr.Args

	switch typ.(type) {
	case FuncType:
		break
	case PointerType:
		t := s.getRootType(typ.(PointerType).BaseType)
		switch t.(type) {
		case FuncType:
			typ = t
			break
		default:
			s.error("{expr} is a pointer to a non-function type.", expr.LineM(), expr.ColumnM())
		}
	case InternalType:
		return
	default:
		s.error("{expr} is not a function or function pointer.", expr.LineM(), expr.ColumnM())
	}

	l := len(expr.Args)
	l2 := len(typ.(FuncType).ArgTypes)

	first := typ.(FuncType).ArgTypes[0]
	if s.compareTypes(first, VoidType.Type) {
		l2--
	}

	switch expr.Function.(type) {
	case MemberExpr:
		base := expr.Function.(MemberExpr).Base
		typ2 := s.getRootType(s.getType(base))

		switch first.(type) {
		case StructType:
			switch typ2.(type) {
			case StructType:
				l++
				Args = append([]Expression{base}, expr.Args...)
			case PointerType:
				l++
				Args = append([]Expression{UnaryExpr{Expr: base, Op: Token{Buff: []byte("*"), PrimaryType: AirthmaticOperator, SecondaryType: Mul}}}, expr.Args...)
			}
		case PointerType:
			typ3 := s.getRootType(first.(PointerType).BaseType)

			switch typ3.(type) {
			case StructType:
				switch typ2.(type) {
				case StructType:
					l++
					Args = append([]Expression{UnaryExpr{Expr: base, Op: Token{Buff: []byte("&"), PrimaryType: BitwiseOperator, SecondaryType: And}}}, expr.Args...)
				case PointerType:
					l++
					Args = append([]Expression{base}, expr.Args...)
				}
			case VecType:
				l++
				Args = append([]Expression{UnaryExpr{Expr: base, Op: Token{Buff: []byte("*"), PrimaryType: AirthmaticOperator, SecondaryType: Mul}}}, expr.Args...)
			case PromiseType:
				l++
				Args = append([]Expression{UnaryExpr{Expr: base, Op: Token{Buff: []byte("*"), PrimaryType: AirthmaticOperator, SecondaryType: Mul}}}, expr.Args...)
			}
		case VecType:
			l++
			Args = append([]Expression{base}, expr.Args...)
		case PromiseType:
			l++
			Args = append([]Expression{base}, expr.Args...)
		}
	}

	if l2 > l {
		s.error("Too few arguments in function call.", expr.LineM(), expr.ColumnM())
	} else if l2 < l {
		s.error("Too many arguments in function call.", expr.LineM(), expr.ColumnM())
	}

	for x, e := range Args {
		t1 := s.getType(e)
		t2 := typ.(FuncType).ArgTypes[x]

		if !s.compareTypes(t1, t2) {
			// s.error("Cannot pass expression of type {t} as an argument of type {t2}.", e.LineM(), e.ColumnM())
		}
	}
}

func (s *SemanticAnalyzer) arrayMemberExpr(expr ArrayMemberExpr) {
	Typ := s.getRootType(s.getType(expr.Parent))

	switch Typ.(type) {
	case ArrayType:
	case ImplictArrayType:
	case PointerType:
	case VecType:
	case TupleType:
		break
	default:
		s.error("Type mismatch: expected an array type, got {Typ}.", Typ.LineM(), Typ.ColumnM())
	}
}

func (s *SemanticAnalyzer) memberExpr(expr MemberExpr) {
	Typ1 := s.getType(expr.Base)

	if Typ1 == nil {
		switch expr.Base.(type) {
		case IdentExpr:
			if n, ok := s.Imports[string(expr.Base.(IdentExpr).Value.Buff)]; ok {
				if _, ok := n.Find(expr.Prop); !ok {
					s.error("'"+string(expr.Prop.Buff)+"' is not exported from '"+string(expr.Base.(IdentExpr).Value.Buff)+"'.", expr.Prop.Line, expr.Prop.Column)
				}
				return
			}
		}
	}

	isImported := false
	base := Expression(nil)

	switch Typ1.(type) {
	case PointerType:
		Typ1 = Typ1.(PointerType).BaseType
	}

	switch Typ1.(type) {
	case BasicType:
		switch Typ1.(BasicType).Expr.(type) {
		case MemberExpr:
			if _, ok := s.ImportPrefixes[string(Typ1.(BasicType).Expr.(MemberExpr).Base.(IdentExpr).Value.Buff)]; ok {
				isImported = true
				base = Typ1.(BasicType).Expr.(MemberExpr).Base
			}
		}
	}

	Typ := s.getRootType(Typ1)

	switch Typ.(type) {
	case EnumType:
		for _, prop := range Typ.(EnumType).Identifiers {
			if bytes.Compare(prop.Buff, expr.Prop.Buff) == 0 {
				return
			}
		}
		s.error("Enum {expr.Base} has no member called '"+string(expr.Prop.Buff)+"'.", expr.Prop.Line, expr.Prop.Column)
	case StructType:
		if isImported {
			Typ8 := Typ.(StructType)
			Typ9 := StructType{}
			Typ9.Props = Typ8.Props
			Typ9.SuperStructs = make([]Expression, len(Typ8.SuperStructs))

			for i, e := range Typ8.SuperStructs {
				switch e.(type) {
				case IdentExpr:
					Typ9.SuperStructs[i] = MemberExpr{Base: base, Prop: e.(IdentExpr).Value}
				default:
					Typ9.SuperStructs[i] = e
				}
			}
			s.getPropType(expr.Prop, Typ9)
		}
		s.getPropType(expr.Prop, Typ.(StructType))
	}
}

func (s *SemanticAnalyzer) compoundLiteral(cl CompoundLiteral) {
	Typ := s.getRootType(cl.Name)

	switch Typ.(type) {
	case InternalType:
		break
	case StructType:
		// ahh so much work NEEDS WORK AAAAAAAAAAAAAAAAAAAAAAAAAAAAA
	case TupleType:
		tupl := Typ.(TupleType)

		if len(cl.Data.Fields) != 0 {
			field1 := cl.Data.Fields[0]
			s.error("Named properties aren't allowed in tuple compound literals.", field1.Line, field1.Column)
		}

		l1 := len(cl.Data.Values)
		l2 := len(tupl.Types)

		if l1 > l2 {
			s.error("Too many fields in compound literal. Expected "+strconv.Itoa(l2)+", got "+strconv.Itoa(l1)+".", cl.LineM(), cl.ColumnM())
		}
		if l1 < l2 {
			s.error("Too few fields in compound literal. Expected "+strconv.Itoa(l2)+", got "+strconv.Itoa(l1)+".", cl.LineM(), cl.ColumnM())
		}

		for x, val := range cl.Data.Values {
			Type1 := s.getType(val)
			Type2 := Typ.(TupleType).Types[x]

			if !s.compareTypes(Type1, Type2) {
				s.error("Type mismatch: tuple has type {Type2} at index {x} but got {Type1}.", val.LineM(), val.ColumnM())
			}
		}
	case VecType:
	case PromiseType:
	case ArrayType:
	case ImplictArrayType:
		break
	default:
		s.error("Invalid type in compound literal. Expected struct or tuple type, got {Typ}.", Typ.LineM(), Typ.ColumnM())
	}
}

func (s *SemanticAnalyzer) rturn(stmt Return, returnType Type) {
	s.exprArray(stmt.Values)
	typ := s.getType(stmt.Values[0])

	if !s.compareTypes(typ, returnType) {
		s.error("Type mismatch: return statement returns {typ} but function has return type {returnType}", typ.LineM(), typ.ColumnM())
	}
}

func (s *SemanticAnalyzer) typedef(typedef Typedef) {
	s.addSymbol(typedef.Name, typedef)
	switch typedef.Type.(type) {
	case StructType:
		s.pushScope()
		s.typ(typedef.Type)
		s.popScope()
	default:
		s.typ(typedef.Type)
	}
}

func (s *SemanticAnalyzer) exportTypedef(typedef Typedef) {
	s.addSymbol(typedef.Name, typedef)
	switch typedef.Type.(type) {
	case StructType:
		s.pushScope()
		s.typ(typedef.Type)
		s.popScope()
	default:
		s.typ(typedef.Type)
	}
	s.Exports.Add(Node{Identifier: typedef.Name, Type: typedef})
}

func (s *SemanticAnalyzer) exprArray(exprs []Expression) {
	for _, expr := range exprs {
		s.expr(expr)
	}
}

func (s *SemanticAnalyzer) typ(typ Type) {
	switch typ.(type) {
	case BasicType:
		s.expr(typ.(BasicType).Expr)
	case PointerType:
		s.typ(typ.(PointerType).BaseType)
	case CaptureType:
		s.typ(typ.(CaptureType).BaseType)
	case StaticType:
		s.typ(typ.(StaticType).BaseType)
	case ConstType:
		s.typ(typ.(ConstType).BaseType)
	case ImplictArrayType:
		s.typ(typ.(ImplictArrayType).BaseType)
	case ArrayType:
		s.typ(typ.(ArrayType).BaseType)
	case FuncType:
		if len(typ.(FuncType).ReturnTypes) > 1 {
			s.error("Multiple return values are not yet supported.", typ.LineM(), typ.ColumnM())
		}
		for _, t := range typ.(FuncType).ReturnTypes {
			s.typ(t)
		}
		for _, t := range typ.(FuncType).ArgTypes {
			s.typ(t)
		}
	case StructType:
		s.strct(typ.(StructType))
	case UnionType:
		union := typ.(UnionType)
		s.pushScope()
		for i, t := range union.Types {
			s.typ(t)
			ident := getPropName(union.Identifiers[i])
			if _, ok := s.getSymbol(ident, true); ok {
				s.error("Repeated field {ident} in union {union}.", ident.Line, ident.Column)
			}
			s.addSymbol(ident, t)
		}
		s.popScope()
	case EnumType:
		enum := typ.(EnumType)
		s.pushScope()
		for x, ident := range enum.Identifiers {
			val := enum.Values[x]
			if val != nil {
				s.expr(val)
			}
			if _, ok := s.getSymbol(ident, true); ok {
				s.error("Repeated field {ident} in enum {enum}.", ident.Line, ident.Column)
			}
			s.addSymbol(ident, enum)
		}
		s.popScope()
	case TupleType:
		tuple := typ.(TupleType)
		for _, typ := range tuple.Types {
			switch typ.(type) {
			case BasicType:
				typ2 := s.getType(typ.(BasicType).Expr)
				switch typ2.(type) {
				case Typedef:
					break
				default:
					s.error("Burh gib type.", typ2.LineM(), typ2.ColumnM())
				}
			}
		}
	}
}

func (s *SemanticAnalyzer) strct(typ StructType) {
	for x, prop := range typ.Props {
		typ.Props[x].Types = s.propDeclaration(prop)
	}
	for _, prop := range typ.Props {
		for _, Val := range prop.Values {
			s.expr(Val)
		}
	}
	for _, superSt := range typ.SuperStructs {
		Typ1 := s.getType(superSt)

		switch Typ1.(type) {
		case Typedef:
			break
		default:
			s.error("Expected a struct typedef, got {Typ1}.", superSt.LineM(), superSt.ColumnM())
		}
		Typ2 := s.getRootType(Typ1)

		switch Typ2.(type) {
		case StructType:
			break
		default:
			s.error("Expected a struct, got {Typ2}.", superSt.LineM(), superSt.ColumnM())
		}
		s.superStrct(Typ2.(StructType), typ)
	}
}

func (s *SemanticAnalyzer) superStrct(typ StructType, strct StructType) {
	for _, prop := range typ.Props {
		s.superPropDeclaration(prop, typ, strct.Name)
	}

	for _, prop := range typ.Props {
		for _, val := range prop.Values {
			s.expr(val)
		}
	}

	for _, superSt := range typ.SuperStructs {
		Typ1 := s.getType(superSt)

		switch Typ1.(type) {
		case Typedef:
			break
		default:
			s.error("Expected a struct typedef, got {Typ1}.", superSt.LineM(), superSt.ColumnM())
		}
		Typ2 := s.getRootType(Typ1)

		switch Typ2.(type) {
		case StructType:
			break
		default:
			s.error("Expected a struct, got {Typ2}.", superSt.LineM(), superSt.ColumnM())
		}
		s.superStrct(Typ2.(StructType), strct)
	}
}

func (s *SemanticAnalyzer) propDeclaration(dec Declaration) []Type {
	Types := []Type{}

	if len(dec.Types) == 1 {
		Type := dec.Types[0]

		if len(dec.Values) == 0 {
			Types = append(Types, Type)
		}
		for _, val := range dec.Values {
			Types = append(Types, Type)
			Type2 := s.getType(val)

			if s.compareTypes(Type2, Type) {
				continue
			}
			s.error("Type mismatch: val has type {Type2}, expected {Type}.", val.LineM(), val.ColumnM())
		}
	} else if len(dec.Types) == 0 {
		if len(dec.Values) == 0 {
			s.error("Cannot declare variable without type.", dec.Identifiers[0].Line, dec.Identifiers[0].Column)
		}
		for _, val := range dec.Values {
			Typ := s.getType(val)

			switch Typ.(type) {
			case NumberType:
				Typ = I32Type.Type
			}
			Types = append(Types, Typ)
		}
	} else if len(dec.Types) != len(dec.Values) {
		s.error("Invalid number of types or values specified", dec.Identifiers[0].Line, dec.Identifiers[0].Column)
	} else {
		Types = dec.Types
	}

	for i, Ident := range dec.Identifiers {
		if _, ok := s.getSymbol(getPropName(Ident), true); ok {
			s.error(string(Ident.Buff)+" has already been declared.", Ident.Line, Ident.Column)
		} else {
			s.addSymbol(getPropName(Ident), Types[i])
		}
	}

	for _, typ := range Types {
		s.typ(typ)
	}

	return Types
}

func (s *SemanticAnalyzer) superPropDeclaration(dec Declaration, superSt Type, Name Token) []Type {
	Types := make([]Type, len(dec.Types))
	Values := make([]Expression, len(dec.Values))

	copy(Types, dec.Types)
	copy(Values, dec.Values)

	for x, val := range dec.Values {
		switch val.(type) {
		case FuncExpr:
			break
		default:
			Values[x] = val
			continue
		}
		if val.(FuncExpr).Type.Mut {
			Values[x] = val
			continue
		}

		fnc := val.(FuncExpr)
		argTypes := fnc.Type.ArgTypes
		first := argTypes[0]

		switch first.(type) {
		case PointerType:
			first2 := first.(PointerType).BaseType
			if s.compareTypes(first2, superSt) {
				Typ := FuncType{ReturnTypes: fnc.Type.ReturnTypes, ArgNames: fnc.Type.ArgNames, ArgTypes: append([]Type{PointerType{BaseType: BasicType{Expr: IdentExpr{Value: Name}}}}, argTypes[1:]...)}
				Types[x] = Typ
				Values[x] = FuncExpr{Type: Typ, Block: fnc.Block}
				continue
			}
		case BasicType:
			if s.compareTypes(first, superSt) {
				Typ := FuncType{ReturnTypes: fnc.Type.ReturnTypes, ArgNames: fnc.Type.ArgNames, ArgTypes: append([]Type{BasicType{Expr: IdentExpr{Value: Name}}}, argTypes[1:]...)}
				Types[x] = Typ
				Values[x] = FuncExpr{Type: Typ, Block: fnc.Block}
				continue
			}
		}
		Values[x] = val
	}

	for i, Ident := range dec.Identifiers {
		if _, ok := s.getSymbol(getPropName(Ident), true); ok {
			s.error(string(Ident.Buff)+" has already been declared.", Ident.Line, Ident.Column)
		} else {
			s.addSymbol(getPropName(Ident), Types[i])
		}
	}

	return Types
}

func (s *SemanticAnalyzer) getType(expr Expression) Type {
	switch expr.(type) {
	case BasicLit:
		switch expr.(BasicLit).Value.PrimaryType {
		case CharLiteral:
			return NumberType{}
		case NumberLiteral:
			return NumberType{}
		case StringLiteral:
			return ArrayType{
				Size:     Token{Buff: []byte(strconv.Itoa(expr.(BasicLit).Value.Flags)), PrimaryType: NumberLiteral, SecondaryType: DecimalRadix},
				BaseType: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("u8"), PrimaryType: Identifier}}},
			}
		}
	case IdentExpr:
		Ident := expr.(IdentExpr).Value
		if isInternal(Ident) {
			return InternalType{}
		}
		sym, ok := s.getSymbol(Ident, false)
		if ok {
			return sym.Type
		}
		if _, ok := s.Imports[string(Ident.Buff)]; ok {
			return nil
		}
		s.error("Use of undeclared variable '"+string(Ident.Buff)+"'.", Ident.Line, Ident.Column)
	case BinaryExpr:
		if expr.(BinaryExpr).Op.PrimaryType == RelationalOperator {
			return BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("bool"), PrimaryType: Identifier}}}
		}

		lt := s.getType(expr.(BinaryExpr).Left)

		switch lt.(type) {
		case NumberType:
		case InternalType:
			break
		default:
			return lt
		}

		return s.getType(expr.(BinaryExpr).Right)
	case TernaryExpr:
		lt := s.getType(expr.(TernaryExpr).Left)

		switch lt.(type) {
		case NumberType:
		case InternalType:
			break
		default:
			return lt
		}
		return s.getType(expr.(TernaryExpr).Right)
	case TypeCast:
		return expr.(TypeCast).Type
	case UnaryExpr:
		if expr.(UnaryExpr).Op.SecondaryType == Mul {
			Typ := s.getType(expr.(UnaryExpr).Expr)
			switch Typ.(type) {
			case PointerType:
				return Typ.(PointerType).BaseType
			}
			s.error("{expr} is not a pointer.", expr.LineM(), expr.ColumnM())
		} else if expr.(UnaryExpr).Op.SecondaryType == And {
			return PointerType{BaseType: s.getType(expr.(UnaryExpr).Expr)}
		} else {
			return s.getType(expr.(UnaryExpr).Expr)
		}
	case PostfixUnaryExpr:
		return s.getType(expr.(PostfixUnaryExpr).Expr)
	case CallExpr:
		Typ := s.getType(expr.(CallExpr).Function)

		switch Typ.(type) {
		case InternalType:
			return InternalType{}
		}
		return Typ.(FuncType).ReturnTypes[0]
	case ArrayMemberExpr:
		Typ := s.getType(expr.(ArrayMemberExpr).Parent)

		switch Typ.(type) {
		case ArrayType:
			return Typ.(ArrayType).BaseType
		case ImplictArrayType:
			return Typ.(ImplictArrayType).BaseType
		case PointerType:
			return Typ.(PointerType).BaseType
		case VecType:
			return Typ.(VecType).BaseType
		}
		return Typ
	case FuncExpr:
		return expr.(FuncExpr).Type
	case HeapAlloc:
		return PointerType{BaseType: expr.(HeapAlloc).Type}
	case CompoundLiteral:
		return expr.(CompoundLiteral).Name
	case SizeExpr:
		return BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("size_t"), PrimaryType: Identifier}}}
	case ArrayLiteral:
		return InternalType{}
	case MemberExpr:
		switch expr.(MemberExpr).Base.(type) {
		case IdentExpr:
			Ident := expr.(MemberExpr).Base.(IdentExpr).Value
			if _, ok := s.getSymbol(Ident, false); ok {
				break
			}
			if t, ok := s.Imports[string(Ident.Buff)]; ok {
				sym, ok := t.Find(expr.(MemberExpr).Prop)
				if ok {
					return s.ofNamespace(sym.Type, expr.(MemberExpr).Base, t)
				}
				s.error("'"+string(expr.(MemberExpr).Prop.Buff)+"' is not exported from '"+string(Ident.Buff)+"'.", expr.LineM(), expr.ColumnM())
			}
		}

		Typ := s.getType(expr.(MemberExpr).Base)
		isImported := false
		var base Expression
		var table *SymbolTable

		switch Typ.(type) {
		case PointerType:
			Typ = Typ.(PointerType).BaseType
		case Typedef:
			return Typ.(Typedef).Type // only happens with enums
		}

		Typ7 := s.getRootType(Typ)

		switch Typ7.(type) {
		case BasicType:
			switch Typ7.(BasicType).Expr.(type) {
			case MemberExpr:
				t, ok := s.Imports[string(Typ.(BasicType).Expr.(MemberExpr).Base.(IdentExpr).Value.Buff)]
				if !ok {
					break
				}
				isImported = true
				table = t
				base = Typ7.(BasicType).Expr.(MemberExpr).Base
			}
		}

		switch Typ7.(type) {
		case StructType:
			if !isImported {
				return s.getPropType(expr.(MemberExpr).Prop, Typ7.(StructType))
			}
			Typ8 := Typ7.(StructType)
			Typ9 := StructType{}
			Typ9.Props = Typ8.Props
			Typ9.SuperStructs = make([]Expression, len(Typ8.SuperStructs))

			for i, e := range Typ8.SuperStructs {
				switch e.(type) {
				case IdentExpr:
					Typ9.SuperStructs[i] = MemberExpr{Base: base, Prop: e.(IdentExpr).Value}
				default:
					Typ9.SuperStructs[i] = e
				}
			}
			return s.ofNamespace(s.getPropType(expr.(MemberExpr).Prop, Typ9), base, table)
		case UnionType:
			for x, prop := range Typ7.(UnionType).Identifiers {
				if bytes.Compare(prop.Buff, expr.(MemberExpr).Prop.Buff) == 0 {
					return Typ7.(UnionType).Types[x]
				}
			}
		case VecType:
			return s.getVectorPropType(Typ7.(VecType), expr.(MemberExpr).Prop)
		case PromiseType:
			return s.getPromisePropType(Typ7.(PromiseType), expr.(MemberExpr).Prop)
		}
	}

	return nil
}

func (s *SemanticAnalyzer) getPromisePropType(prom PromiseType, prop Token) Type {
	switch string(prop.Buff) {
	case "then":
		return FuncType{
			Type:        OrdFunction,
			ReturnTypes: []Type{VoidType.Type},
			ArgTypes:    []Type{prom, FuncType{Type: OrdFunction, ReturnTypes: []Type{VoidType.Type}, ArgTypes: []Type{prom.BaseType}}},
		}
	case "resolve":
		return FuncType{
			Type:        OrdFunction,
			ReturnTypes: []Type{VoidType.Type},
			ArgTypes:    []Type{prom, prom.BaseType},
		}
	case "pending":
		return BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("bool"), PrimaryType: Identifier}}}
	case "resolved":
		return BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("bool"), PrimaryType: Identifier}}}
	}
	s.error("burh", prop.Line, prop.Column)
	return nil
}

func (s *SemanticAnalyzer) getVectorPropType(vec VecType, prop Token) Type {
	switch string(prop.Buff) {
	case "push":
		return FuncType{
			Type:        OrdFunction,
			ReturnTypes: []Type{vec.BaseType},
			ArgTypes:    []Type{vec, vec.BaseType},
		}
	case "pop":
		return FuncType{
			Type:        OrdFunction,
			ReturnTypes: []Type{vec.BaseType},
			ArgTypes:    []Type{vec},
		}
	case "concat":
		return FuncType{
			Type:        OrdFunction,
			ReturnTypes: []Type{vec},
			ArgTypes:    []Type{vec, vec},
		}
	case "free":
		return FuncType{
			Type:        OrdFunction,
			ReturnTypes: []Type{VoidType},
			ArgTypes:    []Type{vec},
		}
	case "clone":
		return FuncType{
			Type:        OrdFunction,
			ReturnTypes: []Type{vec},
			ArgTypes:    []Type{vec},
		}
	case "length":
		return BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("size_t"), PrimaryType: Identifier}}}
	case "capacity":
		return BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("size_t"), PrimaryType: Identifier}}}
	}
	s.error("burh", prop.Line, prop.Column)
	return nil
}

func (s *SemanticAnalyzer) getPropType(Prop Token, strct StructType) Type {
	for _, prop := range strct.Props {
		for x, ident := range prop.Identifiers {
			if bytes.Compare(ident.Buff, Prop.Buff) != 0 {
				continue
			}
			if len(prop.Types) == 0 {
				return s.getType(prop.Values[x])
			} else if len(prop.Types) == 1 {
				return prop.Types[0]
			} else {
				return prop.Types[x]
			}
		}
	}
	for _, superSt := range strct.SuperStructs {
		Typ1 := s.getType(superSt)

		switch Typ1.(type) {
		case Typedef:
			break
		default:
			s.error("Expected a struct typedef, got {Typ1}.", superSt.LineM(), superSt.ColumnM())
		}
		Typ2 := s.getRootType(Typ1)

		switch Typ2.(type) {
		case StructType:
			break
		default:
			s.error("Expected a struct, got {Typ2}.", superSt.LineM(), superSt.ColumnM())
		}

		return s.getPropType(Prop, Typ2.(StructType))
	}
	s.error("struct {expr.Base} has no member called {expr.Prop}", Prop.Line, Prop.Column)
	return nil
}

func (s *SemanticAnalyzer) compareTypes(Type1 Type, Type2 Type) bool {
	switch Type2.(type) {
	case InternalType:
		return true
	case NumberType:
		root := s.getRootType(Type1)

		switch root.(type) {
		case NumberType:
			return true
		case InternalType:
			return true
		case BasicType:
			switch root.(BasicType).Expr.(type) {
			case IdentExpr:
				for _, buff := range GlobalTypes {
					if bytes.Compare([]byte(buff), root.(BasicType).Expr.(IdentExpr).Value.Buff) == 0 {
						return true
					}
				}
			}
		}
		return false
	case CaptureType:
		switch Type1.(type) {
		case CaptureType:
			return s.compareTypes(Type1.(CaptureType).BaseType, Type2.(CaptureType).BaseType)
		}
		return s.compareTypes(Type1, Type2.(CaptureType).BaseType)
	case StaticType:
		switch Type1.(type) {
		case StaticType:
			return s.compareTypes(Type1.(StaticType).BaseType, Type2.(StaticType).BaseType)
		}
		return s.compareTypes(Type1, Type2.(StaticType).BaseType)
	}

	switch Type1.(type) {
	case InternalType:
		return true
	case NumberType:
		root := s.getRootType(Type2)

		switch root.(type) {
		case NumberType:
			return true
		case InternalType:
			return true
		case BasicType:
			switch root.(BasicType).Expr.(type) {
			case IdentExpr:
				for _, buff := range GlobalTypes {
					if bytes.Compare([]byte(buff), root.(BasicType).Expr.(IdentExpr).Value.Buff) == 0 {
						return true
					}
				}
			}
		}
		return false
	case BasicType:
		switch Type2.(type) {
		case BasicType:
			break
		default:
			return false
		}
		switch Type1.(BasicType).Expr.(type) {
		case IdentExpr:
			switch Type2.(BasicType).Expr.(type) {
			case IdentExpr:
				break
			default:
				return false
			}
			return bytes.Compare(Type1.(BasicType).Expr.(IdentExpr).Value.Buff, Type2.(BasicType).Expr.(IdentExpr).Value.Buff) == 0
		case MemberExpr:
			switch Type2.(BasicType).Expr.(type) {
			case MemberExpr:
				break
			default:
				return false
			}
			return true
		}
		return s.compareTypes(BasicType{Expr: Type2.(BasicType).Expr.(MemberExpr).Base}, BasicType{Expr: Type1.(BasicType).Expr.(MemberExpr).Base}) && bytes.Compare(Type2.(BasicType).Expr.(MemberExpr).Prop.Buff, Type1.(BasicType).Expr.(MemberExpr).Prop.Buff) == 0
	case PointerType:
		switch Type2.(type) {
		case PointerType:
			return s.compareTypes(Type2.(PointerType).BaseType, BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("void"), PrimaryType: Identifier}}}) || s.compareTypes(Type1.(PointerType).BaseType, Type2.(PointerType).BaseType)
		case FuncType:
			return s.compareTypes(Type1.(PointerType).BaseType, BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("void"), PrimaryType: Identifier}}})
		default:
			return false
		}
	case VecType:
		switch Type2.(type) {
		case VecType:
			return s.compareTypes(Type1.(VecType).BaseType, Type2.(VecType).BaseType)
		default:
			return false
		}
	case PromiseType:
		switch Type2.(type) {
		case PromiseType:
			return s.compareTypes(Type1.(PromiseType).BaseType, Type2.(PromiseType).BaseType)
		default:
			return false
		}
	case ArrayType:
		switch Type2.(type) {
		case ArrayType:
			break
		default:
			return false
		}
		return s.compareTypes(Type1.(ArrayType).BaseType, Type2.(ArrayType).BaseType)
	case ConstType:
		switch Type2.(type) {
		case ConstType:
			break
		default:
			return false
		}
		return s.compareTypes(Type1.(ConstType).BaseType, Type2.(ConstType).BaseType)
	case FuncType:
		switch Type2.(type) {
		case FuncType:
			break
		case PointerType:
			return s.compareTypes(Type2.(PointerType).BaseType, BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("void"), PrimaryType: Identifier}}})
		default:
			return false
		}
		type1 := Type1.(FuncType)
		type2 := Type2.(FuncType)

		if len(type1.ArgTypes) != len(type2.ArgTypes) {
			return false
		}
		if type1.Type != type2.Type {
			return false
		}
		for x, arg := range type1.ArgTypes {
			if !s.compareTypes(arg, type2.ArgTypes[x]) {
				return false
			}
		}
		return s.compareTypes(type1.ReturnTypes[0], type2.ReturnTypes[0])
	case CaptureType:
		switch Type2.(type) {
		case CaptureType:
			return s.compareTypes(Type1.(CaptureType).BaseType, Type2.(CaptureType).BaseType)
		}
		return s.compareTypes(Type1.(CaptureType).BaseType, Type2)
	case StaticType:
		switch Type2.(type) {
		case StaticType:
			return s.compareTypes(Type1.(StaticType).BaseType, Type2.(StaticType).BaseType)
		}
		return s.compareTypes(Type1.(StaticType).BaseType, Type2)
	case TupleType:
		switch Type2.(type) {
		case TupleType:
			break
		default:
			return false
		}
		for x, typ := range Type1.(TupleType).Types {
			if !s.compareTypes(typ, Type2.(TupleType).Types[x]) {
				return false
			}
		}
		return true
	}

	return false
}

func (s *SemanticAnalyzer) typeCast(typecast TypeCast) {
	s.expr(typecast.Expr)
	s.typ(typecast.Type)

	type1 := s.getRootType(typecast.Type)
	type2 := s.getRootType(s.getType(typecast.Expr))

	switch type2.(type) {
	case ConstType:
		switch type1.(type) {
		case ConstType:
			break
		default:
			s.error("Cannot typecast a constant expression to a non constant expression.", typecast.LineM(), typecast.ColumnM())
		}
	}

	switch type1.(type) {
	case VecType:
		switch type2.(type) {
		case VecType:
		case InternalType:
			break
		default:
			s.error("Cannot typecast a non-vector type to a vector type.", typecast.LineM(), typecast.ColumnM())
		}
	case BasicType:
	case PointerType:
	case InternalType:
	case FuncType:
		break
	default:
		s.error("Typecasting to non scalar data types isn't allowed.", typecast.LineM(), typecast.ColumnM())
	}
}

func (s *SemanticAnalyzer) getRootType(typ Type) Type {
	Typ := typ

	switch typ.(type) {
	case BasicType:
		Typ = s.getType(typ.(BasicType).Expr)
	case Typedef:
		break
	default:
		return typ
	}

	switch Typ.(type) {
	case Typedef:
		break
	default:
		return Typ
	}
	/*
		Sym := s.getSymbol(Typ.(Typedef).Name, false)
		if Sym == nil {
			s.error("Unknown type {typ}.", typ.LineM(), typ.ColumnM())
		}
	*/
	return s.getRootType(Typ.(Typedef).Type)
}

func (s *SemanticAnalyzer) lenExpr(lenExpr LenExpr) {
	Typ := lenExpr.Type

	switch Typ.(type) {
	case BasicType:
		Typ = s.getType(lenExpr.Type.(BasicType).Expr)
	}

	Typ = s.getRootType(Typ)

	switch Typ.(type) {
	case ArrayType:
		return
	case ImplictArrayType:
		return
	}
	s.error("len keyword cant be used with non-array types", lenExpr.LineM(), lenExpr.ColumnM())
}

func (s *SemanticAnalyzer) sizeExpr(sizeExpr SizeExpr) {
	switch sizeExpr.Expr.(type) {
	case Type:
		s.typ(sizeExpr.Expr.(Type))
	default:
		s.expr(sizeExpr.Expr)
	}
}

func (s *SemanticAnalyzer) appendBase(expr Expression, base Expression) Expression {
	switch expr.(type) {
	case IdentExpr:
		return MemberExpr{Base: base, Prop: expr.(IdentExpr).Value}
	case MemberExpr:
		return MemberExpr{Base: s.appendBase(expr.(MemberExpr).Base.(MemberExpr), base), Prop: expr.(MemberExpr).Prop}
	}
	// s.error("Can't append to this expression.", expr.LineM(), expr.ColumnM())
	return nil
}

func (s *SemanticAnalyzer) ofNamespace(typ Type, name Expression, t *SymbolTable) Type {
	switch typ.(type) {
	case BasicType:
		switch typ.(BasicType).Expr.(type) {
		case IdentExpr:
			if typ.(BasicType).Expr.(IdentExpr).Value.Buff[0] == '$' {
				return typ
			}
			for _, buff := range Globals {
				if bytes.Compare([]byte(buff), typ.(BasicType).Expr.(IdentExpr).Value.Buff) == 0 {
					return typ
				}
			}
			return BasicType{Expr: s.appendBase(typ.(BasicType).Expr, name), Line: typ.LineM(), Column: typ.ColumnM()}
		}
	case PointerType:
		return PointerType{BaseType: s.ofNamespace(typ.(PointerType).BaseType, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case VecType:
		return VecType{BaseType: s.ofNamespace(typ.(VecType).BaseType, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case PromiseType:
		return PromiseType{BaseType: s.ofNamespace(typ.(PromiseType).BaseType, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case ArrayType:
		return ArrayType{BaseType: s.ofNamespace(typ.(ArrayType).BaseType, name, t), Size: typ.(ArrayType).Size, Line: typ.LineM(), Column: typ.ColumnM()}
	case ImplictArrayType:
		return ImplictArrayType{BaseType: s.ofNamespace(typ.(ImplictArrayType).BaseType, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case ConstType:
		return ConstType{BaseType: s.ofNamespace(typ.(ConstType).BaseType, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case FuncType:
		fnc := typ.(FuncType)
		return FuncType{Type: fnc.Type, ArgNames: fnc.ArgNames, ArgTypes: s.ofNamespaceArray(fnc.ArgTypes, name, t), ReturnTypes: s.ofNamespaceArray(fnc.ReturnTypes, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case TupleType:
		return TupleType{Types: s.ofNamespaceArray(typ.(TupleType).Types, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case StructType:
		newProps := make([]Declaration, len(typ.(StructType).Props))
		newSuperStructs := make([]Expression, len(typ.(StructType).SuperStructs))

		for i, prop := range typ.(StructType).Props {
			newProps[i] = Declaration{Types: s.ofNamespaceArray(prop.Types, name, t), Identifiers: prop.Identifiers, Values: prop.Values}
		}
		for i, superStruct := range typ.(StructType).SuperStructs {
			newSuperStructs[i] = s.ofNamespace(BasicType{Expr: superStruct}, name, t).(BasicType).Expr
		}
		return StructType{Props: newProps, SuperStructs: newSuperStructs, Name: typ.(StructType).Name}
	case Typedef:
		for _, buff := range Globals {
			if bytes.Compare([]byte(buff), typ.(Typedef).Name.Buff) == 0 {
				return typ
			}
		}
		return Typedef{Name: typ.(Typedef).Name, DefaultName: typ.(Typedef).DefaultName, Type: s.ofNamespace(typ.(Typedef).Type, name, t), NameSpace: name.(IdentExpr).Value}
	case UnionType:
		return UnionType{Identifiers: typ.(UnionType).Identifiers, Types: s.ofNamespaceArray(typ.(UnionType).Types, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	}
	return typ
}

func (s *SemanticAnalyzer) ofNamespaceArray(types []Type, name Expression, t *SymbolTable) []Type {
	newTypes := make([]Type, len(types))
	for i, typ := range types {
		newTypes[i] = s.ofNamespace(typ, name, t)
	}
	return newTypes
}
