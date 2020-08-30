package compiler

import (
	"bytes"
	. "parser"
	"path"
	"strconv"
)

type Formatter struct {
	Symbols  *SymbolTable
	Imports  map[string]*SymbolTable
	Prefixes map[string][]byte
	NameSp   Namespace
}

func FormatFile(ast File, s *SymbolTable, n map[string]*SymbolTable, p map[string][]byte, num int) File {
	f := Formatter{s, n, p, Namespace{}}
	f.NameSp.Init(num)
	newAst := File{}
	newAst.Statements = make([]Statement, len(ast.Statements))
	for i, statement := range ast.Statements {
		newAst.Statements[i] = f.statement(statement)
	}
	return newAst
}

func (f *Formatter) getSymbol(Ident Token, Curr bool) (Node, bool) {
	if Curr {
		return f.Symbols.Find(Ident)
	}
	return f.Symbols.FindInAll(Ident)
}

func (s *Formatter) pushScope() {
	s.Symbols.Flag++
	s.Symbols = s.Symbols.Children[s.Symbols.Flag-1]
}

func (s *Formatter) popScope() {
	s.Symbols = s.Symbols.Parent
}

func (f *Formatter) statement(stmt Statement) Statement {
	switch stmt.(type) {
	case Typedef:
		return f.typedef(stmt.(Typedef))
	case Import:
		return f.imprt(stmt.(Import))
	case Declaration:
		return f.declaration(stmt.(Declaration))
	case IfElseBlock:
		return f.ifElse(stmt.(IfElseBlock))
	case Loop:
		return f.loop(stmt.(Loop))
	case Switch:
		return f.swtch(stmt.(Switch))
	case Block:
		f.pushScope()
		block := f.block(stmt.(Block))
		f.popScope()
		return block
	case Assignment:
		return f.assignment(stmt.(Assignment))
	case Return:
		return f.rturn(stmt.(Return))
	case Delete:
		return f.delete(stmt.(Delete))
	case ExportStatement:
		return ExportStatement{Stmt: f.statement(stmt.(ExportStatement).Stmt)}
	case Expression:
		return f.expr(stmt.(Expression))
	}
	return stmt
}

func (f *Formatter) ifElse(ifElse IfElseBlock) IfElseBlock {
	f.pushScope()
	if ifElse.HasInitStmt {
		ifElse.InitStatement = f.statement(ifElse.InitStatement)
	}
	ifElse.Conditions = f.exprArray(ifElse.Conditions)

	for x, block := range ifElse.Blocks {
		f.pushScope()
		ifElse.Blocks[x] = f.block(block)
		f.popScope()
	}
	ifElse.ElseBlock = f.block(ifElse.ElseBlock)
	f.popScope()
	return ifElse
}

func (f *Formatter) loop(loop Loop) Loop {
	f.pushScope()
	if loop.Type&InitLoop == InitLoop {
		loop.InitStatement = f.statement(loop.InitStatement)
	}
	if loop.Type&CondLoop == CondLoop {
		loop.Condition = f.expr(loop.Condition)
	}
	if loop.Type&LoopLoop == LoopLoop {
		loop.LoopStatement = f.statement(loop.LoopStatement)
	}
	loop.Block = f.block(loop.Block)
	f.popScope()
	return loop
}

func (f *Formatter) swtch(swtch Switch) Switch {
	newSwitch := Switch{Cases: make([]CaseStruct, len(swtch.Cases))}
	f.pushScope()
	if swtch.Type == 1 {
		newSwitch.InitStatement = f.statement(swtch.InitStatement)
		newSwitch.Expr = f.expr(swtch.Expr)
	}
	for x, Case := range swtch.Cases {
		newSwitch.Cases[x].Condition = f.expr(Case.Condition)
		newSwitch.Cases[x].Block = f.block(Case.Block)
	}
	if swtch.HasDefaultCase {
		newSwitch.HasDefaultCase = true
		newSwitch.DefaultCase = f.block(swtch.DefaultCase)
	}
	f.popScope()
	return swtch
}

func (f *Formatter) imprt(stmt Import) Import {
	imprt := Import{}
	n := strconv.Itoa(f.NameSp.Num + 1)

	for _, Path := range stmt.Paths {
		pth := path.Clean(string(Path.Buff[1 : len(Path.Buff)-1]))
		pth = path.Join(path.Dir(pth), n+path.Base(pth))

		if path.Ext(pth) != ".h" {
			pth += ".h"
		}
		imprt.Paths = append(imprt.Paths, Token{Buff: []byte(pth), PrimaryType: StringLiteral})
	}
	return imprt
}

func (f *Formatter) enum(enum EnumType, Name Token) EnumType {
	newEnum := EnumType{Identifiers: make([]Token, len(enum.Identifiers)), Values: make([]Expression, len(enum.Values))}
	f.pushScope()
	for i, Ident := range enum.Identifiers {
		newEnum.Identifiers[i] = f.NameSp.getEnumProp(Name.Buff, Ident)
		newEnum.Values[i] = f.expr(enum.Values[i])
	}
	f.popScope()
	return newEnum
}

func (f *Formatter) typedef(typedef Typedef) Typedef {
	Typ := typedef.Type
	DefaultName := Token{}

	switch Typ.(type) {
	case StructType:
		Typ = f.strct(Typ.(StructType), typedef.Name)
		DefaultName = f.NameSp.getStrctDefaultName(typedef.Name)
	case EnumType:
		Typ = f.enum(Typ.(EnumType), typedef.Name)
	}
	return Typedef{Type: f.typ(Typ), Name: f.NameSp.getNewVarName(typedef.Name), DefaultName: DefaultName}
}

func (f *Formatter) tupl(typ TupleType) TupleType {
	return TupleType{Types: f.typeArray(typ.Types)}
}

func (f *Formatter) union(typ UnionType) UnionType {
	for x, prop := range typ.Identifiers {
		typ.Identifiers[x] = f.NameSp.getPropName(prop)
	}
	return UnionType{Identifiers: typ.Identifiers, Types: f.typeArray(typ.Types)}
}

func (f *Formatter) delete(delete Delete) Delete {
	return Delete{Exprs: f.exprArray(delete.Exprs)}
}

func (f *Formatter) rturn(rturn Return) Return {
	return Return{Values: f.exprArray(rturn.Values)}
}

func (f *Formatter) assignment(as Assignment) Assignment {
	return Assignment{Variables: f.exprArray(as.Variables), Op: as.Op, Values: f.exprArray(as.Values)}
}

func (f *Formatter) block(block Block) Block {
	newBlock := Block{}
	newBlock.Statements = make([]Statement, len(block.Statements))
	for i, stmt := range block.Statements {
		newBlock.Statements[i] = f.statement(stmt)
	}
	return newBlock
}

func (f *Formatter) declaration(dec Declaration) Declaration {
	newDec := Declaration{}

	if len(dec.Types) == 1 {
		Type := f.typ(dec.Types[0])
		for range dec.Identifiers {
			newDec.Types = append(newDec.Types, Type)
		}
	} else if len(dec.Types) == 0 {
		for _, val := range dec.Values {
			newDec.Types = append(newDec.Types, f.typ(f.getType(val)))
		}
	} else {
		for _, typ := range dec.Types {
			newDec.Types = append(newDec.Types, f.typ(typ))
		}
	}

	for _, Val := range dec.Values {
		newDec.Values = append(newDec.Values, f.expr(Val))
	}
	for _, Ident := range dec.Identifiers {
		newDec.Identifiers = append(newDec.Identifiers, f.NameSp.getNewVarName(Ident))
	}
	return newDec
}

func (f *Formatter) strctProp(prop Declaration, Name Token) Declaration {
	newProp := Declaration{}

	if len(prop.Types) == 1 {
		Typ := f.typ(prop.Types[0])
		for range prop.Identifiers {
			newProp.Types = append(newProp.Types, Typ)
		}
	} else if len(prop.Types) == 0 {
		for _, Val := range prop.Values {
			newProp.Types = append(newProp.Types, f.typ(f.getType(Val)))
		}
	} else {
		for _, Type := range prop.Types {
			newProp.Types = append(newProp.Types, f.typ(Type))
		}
	}
	for _, val := range prop.Values {
		newProp.Values = append(newProp.Values, f.expr(val))
	}
	for i, Ident := range prop.Identifiers {
		switch newProp.Types[i].(type) {
		case FuncType:
			newProp.Identifiers = append(newProp.Identifiers, f.NameSp.getStrctMethodName(Ident, Name))
		default:
			newProp.Identifiers = append(newProp.Identifiers, f.NameSp.getPropName(Ident))
		}
	}
	return newProp
}

func (f *Formatter) superStrctProp(superSt BasicType, prop Declaration, Name Token) Declaration {
	newProp := Declaration{}

	for _, val := range prop.Values {
		switch val.(type) {
		case FuncExpr:
			break
		default:
			newProp.Values = append(newProp.Values, f.expr(val))
			continue
		}
		if val.(FuncExpr).Type.Mut {
			newProp.Values = append(newProp.Values, f.expr(val))
			continue
		}

		fnc := val.(FuncExpr)
		argTypes := fnc.Type.ArgTypes
		first := argTypes[0]

		switch first.(type) {
		case PointerType:
			first2 := first.(PointerType).BaseType
			if f.compareTypes(first2, superSt) {
				newVal := FuncExpr{Type: FuncType{ReturnTypes: fnc.Type.ReturnTypes, ArgNames: fnc.Type.ArgNames, ArgTypes: append([]Type{PointerType{BaseType: BasicType{Expr: IdentExpr{Value: Name}}}}, argTypes[1:]...)}, Block: fnc.Block}
				newProp.Values = append(newProp.Values, f.expr(newVal))
				continue
			}
		case BasicType:
			if f.compareTypes(first, superSt) {
				newVal := FuncExpr{Type: FuncType{ReturnTypes: fnc.Type.ReturnTypes, ArgNames: fnc.Type.ArgNames, ArgTypes: append([]Type{BasicType{Expr: IdentExpr{Value: Name}}}, argTypes[1:]...)}, Block: fnc.Block}
				newProp.Values = append(newProp.Values, f.expr(newVal))
				continue
			}
		}
		newProp.Values = append(newProp.Values, f.expr(val))
	}

	if len(prop.Types) == 1 {
		Typ := f.typ(prop.Types[0])

		switch Typ.(type) {
		case FuncType:
			if len(prop.Values) > 0 {
				Typ = f.getType(newProp.Values[0])
			}
		}

		for range prop.Identifiers {
			newProp.Types = append(newProp.Types, Typ)
		}
	} else if len(prop.Types) == 0 {
		for _, Val := range newProp.Values {
			newProp.Types = append(newProp.Types, f.typ(f.getType(Val)))
		}
	} else {
		for x, Type := range prop.Types {
			switch Type.(type) {
			case FuncType:
				newProp.Types = append(newProp.Types, f.typ(newProp.Values[x].(FuncExpr).Type))
				continue
			}
			newProp.Types = append(newProp.Types, f.typ(Type))
		}
	}

	for i, Ident := range prop.Identifiers {
		switch newProp.Types[i].(type) {
		case FuncType:
			if !newProp.Types[i].(FuncType).Mut {
				newProp.Identifiers = append(newProp.Identifiers, f.NameSp.getStrctMethodName(Ident, Name))
				continue
			}
		}
		newProp.Identifiers = append(newProp.Identifiers, f.NameSp.getPropName(Ident))
	}

	return newProp
}

func (f *Formatter) strct(Typ StructType, Name Token) StructType {
	f.pushScope()
	strct := StructType{}
	strct.Props = []Declaration{}

	for _, prop := range Typ.Props {
		strct.Props = append(strct.Props, f.strctProp(prop, Name))
	}
	for _, superStruct := range Typ.SuperStructs {
		for _, prop := range f.getType(superStruct).(Typedef).Type.(StructType).Props {
			strct.Props = append(strct.Props, f.superStrctProp(BasicType{Expr: superStruct}, prop, Name))
		}
	}
	f.popScope()
	return strct
}

func (f *Formatter) expr(expr Expression) Expression {
	expr2 := expr

	switch expr.(type) {
	case IdentExpr:
		for _, buff := range Globals {
			if bytes.Compare([]byte(buff), expr.(IdentExpr).Value.Buff) == 0 {
				return expr
			}
		}
		expr2 = IdentExpr{Value: f.NameSp.getNewVarName(expr.(IdentExpr).Value)}
	case UnaryExpr:
		expr2 = UnaryExpr{Op: expr.(UnaryExpr).Op, Expr: f.expr(expr.(UnaryExpr).Expr)}
	case BinaryExpr:
		expr2 = BinaryExpr{Left: f.expr(expr.(BinaryExpr).Left), Op: expr.(BinaryExpr).Op, Right: f.expr(expr.(BinaryExpr).Right)}
	case PostfixUnaryExpr:
		expr2 = PostfixUnaryExpr{Op: expr.(PostfixUnaryExpr).Op, Expr: f.expr(expr.(PostfixUnaryExpr).Expr)}
	case TernaryExpr:
		expr2 = TernaryExpr{Cond: f.expr(expr.(TernaryExpr).Cond), Left: f.expr(expr.(TernaryExpr).Left), Right: f.expr(expr.(TernaryExpr).Right)}
	case ArrayLiteral:
		expr2 = ArrayLiteral{Exprs: f.exprArray(expr.(ArrayLiteral).Exprs)}
	case CallExpr:
		expr2 = f.callExpr(expr.(CallExpr))
	case TypeCast:
		expr2 = f.typeCast(expr.(TypeCast))
	case ArrayMemberExpr:
		expr2 = f.arrayMemberExpr(expr.(ArrayMemberExpr))
	case MemberExpr:
		expr2 = f.memberExpr(expr.(MemberExpr))
	case LenExpr:
		expr2 = f.lenExpr(expr.(LenExpr))
	case SizeExpr:
		expr2 = f.sizeExpr(expr.(SizeExpr))
	case CompoundLiteral:
		expr2 = f.compoundLiteral(expr.(CompoundLiteral))
	case FuncExpr:
		f.pushScope()
		expr2 = FuncExpr{Type: f.typ(expr.(FuncExpr).Type).(FuncType), Block: f.block(expr.(FuncExpr).Block)}
		f.popScope()
	case HeapAlloc:
		expr2 = HeapAlloc{Type: f.typ(expr.(HeapAlloc).Type), Val: f.expr(expr.(HeapAlloc).Val)}
	}
	return expr2
}

func (f *Formatter) callExpr(expr CallExpr) CallExpr {
	Args := f.exprArray(expr.Args)
	Function := f.expr(expr.Function)

	isPointer := false
	Typ := f.getRootType(f.getType(expr.Function))

	switch Typ.(type) {
	case PointerType:
		Typ = f.getRootType(Typ.(PointerType).BaseType)
		isPointer = true
	}

	switch expr.Function.(type) {
	case MemberExpr:
		Arg1 := Typ.(FuncType).ArgTypes[0]
		Base := expr.Function.(MemberExpr).Base

		oka := false
		switch Base.(type) {
		case IdentExpr:
			_, ok := f.Prefixes[string(Base.(IdentExpr).Value.Buff)]
			oka = ok
		}
		if !oka {
			BaseType := f.getRootType(f.getType(Base))

			switch Arg1.(type) {
			case PointerType:
				switch BaseType.(type) {
				case PointerType:
					Args = append([]Expression{f.expr(Base)}, Args...)
				default:
					Args = append([]Expression{f.expr(UnaryExpr{Expr: Base, Op: Token{Buff: []byte("&"), PrimaryType: AirthmaticOperator, SecondaryType: Mul}})}, Args...)
				}
			default:
				switch BaseType.(type) {
				case PointerType:
					Args = append([]Expression{f.expr(UnaryExpr{Expr: Base, Op: Token{Buff: []byte("*"), PrimaryType: AirthmaticOperator, SecondaryType: Mul}})}, Args...)
				default:
					Args = append([]Expression{f.expr(Base)}, Args...)
				}
			}
		}
	}

	if isPointer {
		return CallExpr{Function: UnaryExpr{Op: Token{Buff: []byte("*"), PrimaryType: AirthmaticOperator, SecondaryType: Mul}, Expr: Function}, Args: Args}
	}
	return CallExpr{Function: Function, Args: Args}
}

func (f *Formatter) typ(typ Type) Type {
	switch typ.(type) {
	case BasicType:
		return BasicType{Expr: f.expr(typ.(BasicType).Expr)}
	case NumberType:
		return f.typ(I32Type.Type)
	case PointerType:
		return PointerType{BaseType: f.typ(typ.(PointerType).BaseType)}
	case VecType:
		return VecType{BaseType: f.typ(typ.(VecType).BaseType)}
	case PromiseType:
		return PromiseType{BaseType: f.typ(typ.(PromiseType).BaseType)}
	case ConstType:
		return ConstType{BaseType: f.typ(typ.(ConstType).BaseType)}
	case CaptureType:
		return CaptureType{BaseType: f.typ(typ.(CaptureType).BaseType)}
	case StaticType:
		return StaticType{BaseType: f.typ(typ.(StaticType).BaseType)}
	case ImplictArrayType:
		return ImplictArrayType{BaseType: f.typ(typ.(ImplictArrayType).BaseType)}
	case ArrayType:
		return ArrayType{Size: typ.(ArrayType).Size, BaseType: f.typ(typ.(ArrayType).BaseType)}
	case FuncType:
		ArgNames := typ.(FuncType).ArgNames
		ArgTypes := typ.(FuncType).ArgTypes

		NewArgNames := make([]Token, len(ArgNames))
		NewArgTypes := make([]Type, len(ArgTypes))

		for i, name := range ArgNames {
			NewArgNames[i] = f.NameSp.getNewVarName(name)
		}
		for i, Typ := range ArgTypes {
			NewArgTypes[i] = f.typ(Typ)
		}
		return FuncType{Type: typ.(FuncType).Type, ArgTypes: NewArgTypes, ArgNames: NewArgNames, ReturnTypes: f.typeArray(typ.(FuncType).ReturnTypes)}
	case TupleType:
		return f.tupl(typ.(TupleType))
	case UnionType:
		f.pushScope()
		union := f.union(typ.(UnionType))
		f.popScope()
		return union
	}

	return typ
}

func (f *Formatter) typeArray(types []Type) []Type {
	typs := make([]Type, len(types))
	for i, typ := range types {
		typs[i] = f.typ(typ)
	}
	return typs
}

func (f *Formatter) typeCast(expr TypeCast) TypeCast {
	return TypeCast{Type: f.typ(expr.Type), Expr: f.expr(expr.Expr)}
	/*
		switch expr.Type.(type) {
		case BasicType:
			break
		default:
			return TypeCast{Type: f.typ(expr.Type), Expr: f.expr(expr.Expr)}
		}

		Typ := f.getType(expr.Type.(BasicType).Expr)

		switch Typ.(type) {
		case StructType:
			break
		default:
			return TypeCast{Type: f.typ(expr.Type), Expr: f.expr(expr.Expr)}
		}

		return TypeCast{
			Type: expr.Type,
			Expr: UnaryExpr{
				Op: Token{Buff: []byte("*"), PrimaryType: AirthmaticOperator, SecondaryType: Mul},
				Expr: TypeCast{
					Type: PointerType{BaseType: expr.Type},
					Expr: BinaryExpr{
						Left: UnaryExpr{Op: Token{Buff: []byte("&"), PrimaryType: BitwiseOperator, SecondaryType: And}, Expr: f.expr(expr)},
						Op:   Token{Buff: []byte("+"), PrimaryType: AirthmaticOperator, SecondaryType: Add},
						Right: CallExpr{
							Function: IdentExpr{Value: Token{Buff: []byte("offsetof"), PrimaryType: Identifier}},
							Args:     []Expression{expr.Type.(BasicType).Expr, Typ.(StructType).Props[0].Types[0]},
						},
					},
				},
			},
		}
	*/
}

func (f *Formatter) compoundLiteral(expr CompoundLiteral) CompoundLiteral {
	Name := f.typ(expr.Name)
	Typ := expr.Name

	prefix := f.NameSp.Base

	switch Typ.(type) {
	case BasicType:
		t := Typ.(BasicType)
		switch t.Expr.(type) {
		case MemberExpr:
			Ident := t.Expr.(MemberExpr).Base.(IdentExpr).Value
			if t, ok := f.Prefixes[string(f.NameSp.getActualName(Ident).Buff)]; ok {
				prefix = string(t[1:])
			}
		}
	default:
		return CompoundLiteral{Name: Name, Data: f.compoundLiteralData(expr.Data)}
	}

	Typ = f.getType(Typ.(BasicType).Expr)
	var StrctName Token

	switch Typ.(type) {
	case Typedef:
		StrctName = Typ.(Typedef).Name
	default:
		return CompoundLiteral{Name: Name, Data: f.compoundLiteralData(expr.Data)}
	}

	Typ = f.getRootType(Typ)

	switch Typ.(type) {
	case StructType:
		break
	default:
		return CompoundLiteral{Name: Name, Data: f.compoundLiteralData(expr.Data)}
	}

	strct := Typ.(StructType)
	data := f.compoundLiteralData(expr.Data)

	if len(data.Fields) == 0 && len(data.Values) > 0 {
		x := 0
		l := len(data.Values)

		for _, prop := range strct.Props {
			for j, Ident := range prop.Identifiers {
				t := prop.Types[j]
				switch t.(type) {
				case FuncType:
					if !t.(FuncType).Mut {
						continue
					}
				}
				data.Fields = append(data.Fields, f.NameSp.getPropName(Ident))
				x++
				if x <= l {
					continue
				}
				data.Values = append(data.Values, MemberExpr{
					Base: IdentExpr{Value: f.NameSp.getStrctDefaultNameFromPrefix(prefix, StrctName)},
					Prop: f.NameSp.getPropName(Ident),
				})
			}
		}
		for _, superSt := range strct.SuperStructs {
			superSt := f.getRootType(f.getType(superSt)).(StructType)

			for _, prop := range superSt.Props {
				for j, Ident := range prop.Identifiers {
					t := prop.Types[j]
					switch t.(type) {
					case FuncType:
						if !t.(FuncType).Mut {
							continue
						}
					}
					data.Fields = append(data.Fields, f.NameSp.getPropName(Ident))
					x++
					if x <= l {
						continue
					}
					data.Values = append(data.Values, MemberExpr{
						Base: IdentExpr{Value: f.NameSp.getStrctDefaultNameFromPrefix(prefix, StrctName)},
						Prop: f.NameSp.getPropName(Ident),
					})
				}
			}
		}
	} else {
		for _, prop := range strct.Props {
			for j, Ident := range prop.Identifiers {

				if hasField(data.Fields, Ident) {
					continue
				}
				t := prop.Types[j]

				switch t.(type) {
				case FuncType:
					if !t.(FuncType).Mut {
						continue
					}
				}

				data.Fields = append(data.Fields, f.NameSp.getPropName(Ident))
				data.Values = append(data.Values, MemberExpr{
					Base: IdentExpr{Value: f.NameSp.getStrctDefaultNameFromPrefix(prefix, StrctName)},
					Prop: f.NameSp.getPropName(Ident),
				})
			}
		}
		for _, superSt := range strct.SuperStructs {
			superSt := f.getRootType(f.getType(superSt)).(StructType)

			for _, prop := range superSt.Props {
				for j, Ident := range prop.Identifiers {
					t := prop.Types[j]
					switch t.(type) {
					case FuncType:
						if !t.(FuncType).Mut {
							continue
						}
					}

					data.Fields = append(data.Fields, f.NameSp.getPropName(Ident))
					data.Values = append(data.Values, MemberExpr{
						Base: IdentExpr{Value: f.NameSp.getStrctDefaultNameFromPrefix(prefix, StrctName)},
						Prop: f.NameSp.getPropName(Ident),
					})
				}
			}
		}
	}
	return CompoundLiteral{Name: Name, Data: data}
}

func hasField(fields []Token, field Token) bool {
	for _, tok := range fields {
		if bytes.Compare(tok.Buff, field.Buff) == 0 {
			return true
		}
	}
	return false
}

func (f *Formatter) compoundLiteralData(data CompoundLiteralData) CompoundLiteralData {
	newData := CompoundLiteralData{Values: make([]Expression, len(data.Values)), Fields: make([]Token, len(data.Fields))}
	for i, Val := range data.Values {
		newData.Values[i] = f.expr(Val)
	}
	for i, Field := range data.Fields {
		newData.Fields[i] = f.NameSp.getNewVarName(Field)
	}
	return newData
}

func (f *Formatter) arrayMemberExpr(expr ArrayMemberExpr) Expression {
	Typ := f.getRootType(f.getType(expr.Parent))

	switch Typ.(type) {
	case VecType:
		return ArrayMemberExpr{
			Parent: PointerMemberExpr{
				Base: f.expr(expr.Parent),
				Prop: Token{Buff: []byte("mem"), PrimaryType: Identifier},
			},
			Index: f.expr(expr.Index),
		}
	case TupleType:
		return MemberExpr{
			Base: f.expr(expr.Parent),
			Prop: Token{Buff: []byte("_" + string(expr.Index.(BasicLit).Value.Buff)), PrimaryType: Identifier},
		}
	default:
		return ArrayMemberExpr{Parent: f.expr(f.expr(expr.Parent)), Index: f.expr(expr.Index)}
	}
}

func (f *Formatter) lenExpr(expr LenExpr) Expression {
	// var Expr Expression
	Typ := expr.Type

	switch Typ.(type) {
	case BasicType:
		Typ = f.getType(Typ.(BasicType).Expr)
		// Expr = f.expr(Typ.(BasicType).Expr)
	}

	Typ = f.getRootType(Typ)
	/*
		switch Typ.(type) {
		case ArrayType:
			return CallExpr{Function: IdentExpr{Value: Token{Buff: []byte("len2")}}, Args: []Expression{Typ, Typ.(ArrayType).BaseType}}
		case DynamicType:
			switch Typ.(DynamicType).BaseType.(type) {
			case ImplictArrayType:
				return CallExpr{Function: IdentExpr{Value: Token{Buff: []byte("len")}}, Args: []Expression{Expr, Typ.(DynamicType).BaseType.(ImplictArrayType).BaseType}}
			case ArrayType:
				return CallExpr{Function: IdentExpr{Value: Token{Buff: []byte("len2")}}, Args: []Expression{Typ.(DynamicType).BaseType, Typ.(DynamicType).BaseType.(ArrayType).BaseType}}
			}
		}
	*/
	return CallExpr{Function: IdentExpr{Value: Token{Buff: []byte("len3")}}, Args: []Expression{Typ}}
}

func (f *Formatter) sizeExpr(expr SizeExpr) CallExpr {
	Expr := expr.Expr

	switch Expr.(type) {
	case Type:
		Expr = f.typ(Expr.(Type))
	default:
		Expr = f.expr(Expr)
	}

	return CallExpr{Function: IdentExpr{Value: Token{Buff: []byte("sizeof")}}, Args: []Expression{Expr}}
}

func (f *Formatter) memberExpr(expr MemberExpr) Expression {
	Typ := f.getType(expr.Base)

	if Typ == nil {
		switch expr.Base.(type) {
		case IdentExpr:
			val, ok := f.Prefixes[string(f.NameSp.getActualName(expr.Base.(IdentExpr).Value).Buff)]
			if ok {
				return IdentExpr{Value: f.NameSp.joinName(val, f.NameSp.getActualName(expr.Prop))}
			}
		}
	}

	isPointer := false

	switch Typ.(type) {
	case PointerType:
		Typ = Typ.(PointerType).BaseType
		isPointer = true
	}

	prefix := f.NameSp.Base
	isImported := false
	var table *SymbolTable = nil
	p2 := Expression(nil)

	switch Typ.(type) {
	case BasicType:
		switch Typ.(BasicType).Expr.(type) {
		case MemberExpr:
			Ident := Typ.(BasicType).Expr.(MemberExpr).Base.(IdentExpr).Value
			key := string(f.NameSp.getActualName(Ident).Buff)
			if t, ok := f.Prefixes[key]; ok {
				prefix = string(t[1:])
				isImported = true
				table = f.Imports[key]
				p2 = Typ.(BasicType).Expr.(MemberExpr).Base
			}
		}
		Typ = f.getType(Typ.(BasicType).Expr)
	case Typedef:
		if Typ.(Typedef).NameSpace.Buff != nil {
			isImported = true
			key := string(Typ.(Typedef).NameSpace.Buff)
			prefix = string(f.Prefixes[key][1:])
			table = f.Imports[key]
		}
	}

	switch Typ.(type) {
	case Typedef:
		switch Typ.(Typedef).Type.(type) {
		case EnumType:
			if isImported {
				return IdentExpr{Value: f.NameSp.getEnumPropFromPrefix(prefix, Typ.(Typedef).Name.Buff, expr.Prop)}
			}
			return IdentExpr{Value: f.NameSp.getEnumProp(Typ.(Typedef).Name.Buff, expr.Prop)}
		}
	}

	Typ99 := f.getRootType(Typ)
	var Typ2 StructType

	switch Typ99.(type) {
	case StructType:
		Typ2 = Typ99.(StructType)
	case VecType:
		return f.getVecProp(expr)
	case PromiseType:
		return f.getPromiseProp(expr)
	default:
		return MemberExpr{Base: f.expr(expr.Base), Prop: f.NameSp.getPropName(expr.Prop)}
	}

	ident := expr.Prop
	var Typ11 Type

	if isImported {
		Typ9 := StructType{}
		Typ9.Props = Typ2.Props
		Typ9.SuperStructs = make([]Expression, len(Typ2.SuperStructs))

		for i, e := range Typ2.SuperStructs {
			switch e.(type) {
			case IdentExpr:
				Typ9.SuperStructs[i] = MemberExpr{Base: p2, Prop: e.(IdentExpr).Value}
			default:
				Typ9.SuperStructs[i] = e
			}
		}

		Typ11 = f.ofNamespace(f.getPropType(ident, Typ9), p2, table)
	} else {
		Typ11 = f.getPropType(ident, Typ2)
	}

	switch Typ11.(type) {
	case FuncType:
		return IdentExpr{Value: f.NameSp.getStrctMethodNameFromPrefix(prefix, ident, Typ.(Typedef).Name)}
	}

	if isPointer {
		return PointerMemberExpr{Base: f.expr(expr.Base), Prop: f.NameSp.getPropName(expr.Prop)}
	}
	return MemberExpr{Base: f.expr(expr.Base), Prop: f.NameSp.getPropName(expr.Prop)}
}

func (f *Formatter) getVecProp(expr MemberExpr) Expression {
	switch string(expr.Prop.Buff) {
	case "length":
		return PointerMemberExpr{
			Base: f.expr(expr.Base),
			Prop: expr.Prop,
		}
	case "capacity":
		return PointerMemberExpr{
			Base: f.expr(expr.Base),
			Prop: expr.Prop,
		}
	case "push":
		return IdentExpr{Value: Token{Buff: []byte("VECTOR_PUSH")}}
	case "pop":
		return IdentExpr{Value: Token{Buff: []byte("VECTOR_POP")}}
	case "concat":
		return IdentExpr{Value: Token{Buff: []byte("VECTOR_CONCAT")}}
	case "free":
		return IdentExpr{Value: Token{Buff: []byte("VECTOR_FREE")}}
	case "clone":
		return IdentExpr{Value: Token{Buff: []byte("VECTOR_CLONE")}}
	}
	return nil
}

func (f *Formatter) getPromiseProp(expr MemberExpr) Expression {
	switch string(expr.Prop.Buff) {
	case "pending":
		return UnaryExpr{
			Op:   Token{Buff: []byte("!"), PrimaryType: LogicalOperator, SecondaryType: Not},
			Expr: PointerMemberExpr{Base: f.expr(expr.Base), Prop: Token{Buff: []byte("state"), PrimaryType: Identifier}},
		}
	case "resolved":
		return PointerMemberExpr{Base: f.expr(expr.Base), Prop: Token{Buff: []byte("state"), PrimaryType: Identifier}}
	case "resolve":
		return IdentExpr{Value: Token{Buff: []byte("PROMISE_RESOLVE")}}
	case "then":
		return IdentExpr{Value: Token{Buff: []byte("PROMISE_THEN")}}
	}
	return nil
}

func (f *Formatter) exprArray(array []Expression) []Expression {
	Exprs := []Expression{}
	for _, Expr := range array {
		Exprs = append(Exprs, f.expr(Expr))
	}
	return Exprs
}

func (f *Formatter) getType(expr Expression) Type {
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
		sym, ok := f.getSymbol(Ident, false)
		if ok {
			return sym.Type
		}
	case BinaryExpr:
		lt := f.getType(expr.(BinaryExpr).Left)

		switch lt.(type) {
		case NumberType:
		case InternalType:
			break
		default:
			return lt
		}
		return f.getType(expr.(BinaryExpr).Right)
	case TernaryExpr:
		lt := f.getType(expr.(TernaryExpr).Left)

		switch lt.(type) {
		case NumberType:
		case InternalType:
			break
		default:
			return lt
		}
		return f.getType(expr.(TernaryExpr).Right)
	case TypeCast:
		return expr.(TypeCast).Type
	case UnaryExpr:
		if expr.(UnaryExpr).Op.SecondaryType == Mul {
			Typ := f.getType(expr.(UnaryExpr).Expr)
			switch Typ.(type) {
			case PointerType:
				return Typ.(PointerType).BaseType
			}
		} else if expr.(UnaryExpr).Op.SecondaryType == And {
			return PointerType{BaseType: f.getType(expr.(UnaryExpr).Expr)}
		} else {
			return f.getType(expr.(UnaryExpr).Expr)
		}
	case PostfixUnaryExpr:
		return f.getType(expr.(PostfixUnaryExpr).Expr)
	case CallExpr:
		Typ := f.getType(expr.(CallExpr).Function)

		switch Typ.(type) {
		case InternalType:
			return InternalType{}
		}
		return Typ.(FuncType).ReturnTypes[0]
	case ArrayMemberExpr:
		Typ := f.getType(expr.(ArrayMemberExpr).Parent)

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
	case MemberExpr:
		Typ := f.getType(expr.(MemberExpr).Base)

		if Typ == nil {
			switch expr.(MemberExpr).Base.(type) {
			case IdentExpr:
				Ident := expr.(MemberExpr).Base.(IdentExpr).Value
				if t, ok := f.Imports[string(Ident.Buff)]; ok {
					sym, _ := t.Find(expr.(MemberExpr).Prop)
					return f.ofNamespace(sym.Type, expr.(MemberExpr).Base, t)
				}
			}
		}

		isImported := false
		var base Expression
		var table *SymbolTable

		switch Typ.(type) {
		case PointerType:
			Typ = Typ.(PointerType).BaseType
		case Typedef:
			return Typ.(Typedef).Type
		}

		Typ7 := f.getRootType(Typ)

		switch Typ7.(type) {
		case BasicType:
			switch Typ7.(BasicType).Expr.(type) {
			case MemberExpr:
				t, ok := f.Imports[string(Typ.(BasicType).Expr.(MemberExpr).Base.(IdentExpr).Value.Buff)]
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
				return f.getPropType(expr.(MemberExpr).Prop, Typ7.(StructType))
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
			return f.ofNamespace(f.getPropType(expr.(MemberExpr).Prop, Typ9), base, table)
		case UnionType:
			for x, prop := range Typ7.(UnionType).Identifiers {
				if bytes.Compare(prop.Buff, expr.(MemberExpr).Prop.Buff) == 0 {
					return Typ7.(UnionType).Types[x]
				}
			}
		case VecType:
			return f.getVectorPropType(Typ.(VecType), expr.(MemberExpr).Prop)
		case PromiseType:
			return f.getPromisePropType(Typ.(PromiseType), expr.(MemberExpr).Prop)
		}
	}

	return nil
}

func (f *Formatter) getPromisePropType(prom PromiseType, prop Token) Type {
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
	return nil
}

func (f *Formatter) getVectorPropType(vec VecType, prop Token) Type {
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
	return nil
}

func (f *Formatter) getPropType(Prop Token, strct StructType) Type {
	for _, prop := range strct.Props {
		for x, ident := range prop.Identifiers {
			if bytes.Compare(ident.Buff, Prop.Buff) != 0 {
				continue
			}
			if len(prop.Types) == 0 {
				return f.getType(prop.Values[x])
			} else if len(prop.Types) == 1 {
				return prop.Types[0]
			} else {
				return prop.Types[x]
			}
		}
	}
	for _, superSt := range strct.SuperStructs {
		return f.getPropType(Prop, f.getRootType(f.getType(superSt)).(StructType))
	}
	return nil
}

func (f *Formatter) getRootType(typ Type) Type {
	Typ := typ

	switch typ.(type) {
	case BasicType:
		Typ = f.getType(typ.(BasicType).Expr)
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

	return f.getRootType(Typ.(Typedef).Type)
}

func (f *Formatter) compareTypes(Type1 Type, Type2 Type) bool {
	switch Type2.(type) {
	case InternalType:
		return true
	case NumberType:
		root := f.getRootType(Type1)

		switch root.(type) {
		case NumberType:
			return true
		case InternalType:
			return true
		case BasicType:
			switch root.(BasicType).Expr.(type) {
			case IdentExpr:
				for _, buff := range GlobalTypes {
					if bytes.Compare([]byte(buff), f.NameSp.getActualName(root.(BasicType).Expr.(IdentExpr).Value).Buff) == 0 {
						return true
					}
				}
			}
		}
		return false
	case CaptureType:
		switch Type1.(type) {
		case CaptureType:
			return f.compareTypes(Type1.(CaptureType).BaseType, Type2.(CaptureType).BaseType)
		}
		return f.compareTypes(Type1, Type2.(CaptureType).BaseType)
	case StaticType:
		switch Type1.(type) {
		case StaticType:
			return f.compareTypes(Type1.(StaticType).BaseType, Type2.(StaticType).BaseType)
		}
		return f.compareTypes(Type1, Type2.(StaticType).BaseType)
	case ConstType:
		switch Type1.(type) {
		case ConstType:
			break
		default:
			return f.compareTypes(Type1, Type2.(ConstType).BaseType)
		}
		return f.compareTypes(Type1.(ConstType).BaseType, Type2.(ConstType).BaseType)
	}

	switch Type1.(type) {
	case InternalType:
		return true
	case NumberType:
		root := f.getRootType(Type2)

		switch root.(type) {
		case NumberType:
			return true
		case InternalType:
			return true
		case BasicType:
			switch root.(BasicType).Expr.(type) {
			case IdentExpr:
				for _, buff := range GlobalTypes {
					if bytes.Compare([]byte(buff), f.NameSp.getActualName(root.(BasicType).Expr.(IdentExpr).Value).Buff) == 0 {
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
			return bytes.Compare(f.NameSp.getActualName(Type1.(BasicType).Expr.(IdentExpr).Value).Buff, f.NameSp.getActualName(Type2.(BasicType).Expr.(IdentExpr).Value).Buff) == 0
		}
		switch Type2.(BasicType).Expr.(type) {
		case MemberExpr:
			switch Type1.(BasicType).Expr.(type) {
			case MemberExpr:
				break
			default:
				return false
			}
		}
		return f.compareTypes(BasicType{Expr: Type2.(BasicType).Expr.(MemberExpr).Base}, BasicType{Expr: Type1.(BasicType).Expr.(MemberExpr).Base}) && bytes.Compare(Type2.(BasicType).Expr.(MemberExpr).Prop.Buff, Type1.(BasicType).Expr.(MemberExpr).Prop.Buff) == 0
	case PointerType:
		switch Type2.(type) {
		case PointerType:
			return f.compareTypes(Type1.(PointerType).BaseType, Type2.(PointerType).BaseType)
		default:
			return false
		}
	case VecType:
		switch Type2.(type) {
		case VecType:
			return f.compareTypes(Type1.(VecType).BaseType, Type2.(VecType).BaseType)
		default:
			return false
		}
	case PromiseType:
		switch Type2.(type) {
		case PromiseType:
			return f.compareTypes(Type1.(PromiseType).BaseType, Type2.(PromiseType).BaseType)
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
		return f.compareTypes(Type1.(ArrayType).BaseType, Type2.(ArrayType).BaseType)
	case ConstType:
		switch Type2.(type) {
		case ConstType:
			break
		default:
			return f.compareTypes(Type1.(ConstType).BaseType, Type2)
		}
		return f.compareTypes(Type1.(ConstType).BaseType, Type2.(ConstType).BaseType)
	case FuncType:
		switch Type2.(type) {
		case FuncType:
			break
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
			if !f.compareTypes(arg, type2.ArgTypes[x]) {
				return false
			}
		}
		return f.compareTypes(type1.ReturnTypes[0], type2.ReturnTypes[0])
	case CaptureType:
		switch Type2.(type) {
		case CaptureType:
			return f.compareTypes(Type1.(CaptureType).BaseType, Type2.(CaptureType).BaseType)
		}
		return f.compareTypes(Type1.(CaptureType).BaseType, Type2)
	case StaticType:
		switch Type2.(type) {
		case StaticType:
			return f.compareTypes(Type1.(StaticType).BaseType, Type2.(StaticType).BaseType)
		}
		return f.compareTypes(Type1.(StaticType).BaseType, Type2)
	}

	return false
}

func (f *Formatter) appendBase(expr Expression, base Expression) Expression {
	switch expr.(type) {
	case IdentExpr:
		return MemberExpr{Base: base, Prop: expr.(IdentExpr).Value}
	case MemberExpr:
		return MemberExpr{Base: f.appendBase(expr.(MemberExpr).Base.(MemberExpr), base), Prop: expr.(MemberExpr).Prop}
	}
	return nil
}

func (f *Formatter) ofNamespace(typ Type, name Expression, t *SymbolTable) Type {
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
			return BasicType{Expr: f.appendBase(typ.(BasicType).Expr, name), Line: typ.LineM(), Column: typ.ColumnM()}
		}
	case PointerType:
		return PointerType{BaseType: f.ofNamespace(typ.(PointerType).BaseType, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case VecType:
		return VecType{BaseType: f.ofNamespace(typ.(VecType).BaseType, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case PromiseType:
		return VecType{BaseType: f.ofNamespace(typ.(PromiseType).BaseType, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case ArrayType:
		return ArrayType{BaseType: f.ofNamespace(typ.(ArrayType).BaseType, name, t), Size: typ.(ArrayType).Size, Line: typ.LineM(), Column: typ.ColumnM()}
	case ImplictArrayType:
		return ImplictArrayType{BaseType: f.ofNamespace(typ.(ImplictArrayType).BaseType, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case ConstType:
		return ConstType{BaseType: f.ofNamespace(typ.(ConstType).BaseType, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case FuncType:
		fnc := typ.(FuncType)
		return FuncType{Type: fnc.Type, ArgNames: fnc.ArgNames, ArgTypes: f.ofNamespaceArray(fnc.ArgTypes, name, t), ReturnTypes: f.ofNamespaceArray(fnc.ReturnTypes, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case TupleType:
		return TupleType{Types: f.ofNamespaceArray(typ.(TupleType).Types, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	case StructType:
		newProps := make([]Declaration, len(typ.(StructType).Props))
		newSuperStructs := make([]Expression, len(typ.(StructType).SuperStructs))

		for i, prop := range typ.(StructType).Props {
			newProps[i] = Declaration{Types: f.ofNamespaceArray(prop.Types, name, t), Identifiers: prop.Identifiers, Values: prop.Values}
		}
		for i, superStruct := range typ.(StructType).SuperStructs {
			newSuperStructs[i] = f.ofNamespace(BasicType{Expr: superStruct}, name, t).(BasicType).Expr
		}
		return StructType{Props: newProps, SuperStructs: newSuperStructs, Name: typ.(StructType).Name}
	case Typedef:
		for _, buff := range Globals {
			if bytes.Compare([]byte(buff), typ.(Typedef).Name.Buff) == 0 {
				return typ
			}
		}
		return Typedef{Name: typ.(Typedef).Name, DefaultName: typ.(Typedef).DefaultName, Type: f.ofNamespace(typ.(Typedef).Type, name, t), NameSpace: name.(IdentExpr).Value}
	case UnionType:
		return UnionType{Identifiers: typ.(UnionType).Identifiers, Types: f.ofNamespaceArray(typ.(UnionType).Types, name, t), Line: typ.LineM(), Column: typ.ColumnM()}
	}
	return typ
}

func (f *Formatter) ofNamespaceArray(types []Type, name Expression, t *SymbolTable) []Type {
	newTypes := make([]Type, len(types))
	for i, typ := range types {
		newTypes[i] = f.ofNamespace(typ, name, t)
	}
	return newTypes
}
