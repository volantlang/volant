package parser

type FunctionType byte

// Never change these numbers, they are very sepcific
const (
	OrdFunction   FunctionType = 1
	AsyncFunction FunctionType = 2
	WorkFunction  FunctionType = 3
)

type LoopType byte

const (
	InitLoop LoopType = 1
	CondLoop LoopType = 2
	LoopLoop LoopType = 4
	NoneLoop LoopType = 8
)

type SwitchType byte

const (
	InitCondSwitch SwitchType = 1
	CondSwitch     SwitchType = 2
	NoneSwtch      SwitchType = 3
)

type CaseStruct struct {
	Condition Expression
	Block     Block
	Line      int
	Column    int
}

func (c CaseStruct) LineM() int {
	return c.Line
}
func (c CaseStruct) ColumnM() int {
	return c.Column
}

type Statement interface {
	isStatement()
	LineM() int
	ColumnM() int
}

type Expression interface {
	isExpression()
	isStatement()
	LineM() int
	ColumnM() int
}

type Type interface {
	isType()
	isExpression()
	isStatement()
	LineM() int
	ColumnM() int
}

type File struct {
	Path       string
	Statements []Statement
}

type (
	Block struct {
		Statements []Statement
		Line       int
		Column     int
	}
	Declaration struct {
		Identifiers []Token
		Types       []Type
		Values      []Expression
		Line        int
		Column      int
	}
	Import struct {
		Paths  []Token
		Line   int
		Column int
	}
	Loop struct {
		Type          LoopType
		InitStatement Statement
		Condition     Expression
		LoopStatement Statement
		Block         Block
		Line          int
		Column        int
	}
	Switch struct {
		Type           SwitchType
		InitStatement  Statement
		Expr           Expression
		Cases          []CaseStruct
		HasDefaultCase bool
		DefaultCase    Block
		Line           int
		Column         int
	}
	IfElseBlock struct {
		HasInitStmt   bool
		InitStatement Statement
		Conditions    []Expression
		Blocks        []Block
		ElseBlock     Block
		Line          int
		Column        int
	}
	Return struct {
		Values []Expression
		Line   int
		Column int
	}
	Assignment struct {
		Variables []Expression
		Op        Token
		Values    []Expression
		Line      int
		Column    int
	}
	Defer struct {
		Stmt   Statement
		Line   int
		Column int
	}
	Delete struct {
		Exprs  []Expression
		Line   int
		Column int
	}
	Typedef struct {
		Name        Token
		DefaultName Token
		Type        Type
		NameSpace   Token
		Line        int
		Column      int
	}

	Break struct {
		Line   int
		Column int
	}
	Continue struct {
		Line   int
		Column int
	}
	NullStatement struct {
		Line   int
		Column int
	}

	ExportStatement struct {
		Stmt   Statement
		Line   int
		Column int
	}
)

type (
	BasicLit struct {
		Value  Token
		Line   int
		Column int
	}

	BinaryExpr struct {
		Left   Expression
		Op     Token
		Right  Expression
		Line   int
		Column int
	}

	UnaryExpr struct {
		Op     Token
		Expr   Expression
		Line   int
		Column int
	}

	PostfixUnaryExpr struct {
		Op     Token
		Expr   Expression
		Line   int
		Column int
	}

	TernaryExpr struct {
		Cond   Expression
		Left   Expression
		Right  Expression
		Line   int
		Column int
	}

	FuncExpr struct {
		Type   FuncType
		Block  Block
		Line   int
		Column int
	}

	CallExpr struct {
		Function Expression
		Args     []Expression
		Line     int
		Column   int
	}

	TypeCast struct {
		Type   Type
		Expr   Expression
		Line   int
		Column int
	}

	IdentExpr struct {
		Value  Token
		Line   int
		Column int
	}

	MemberExpr struct {
		Base   Expression
		Prop   Token
		Line   int
		Column int
	}

	PointerMemberExpr struct {
		Base   Expression
		Prop   Token
		Line   int
		Column int
	}

	ArrayMemberExpr struct {
		Parent Expression
		Index  Expression
		Line   int
		Column int
	}

	CompoundLiteral struct {
		Name   Type
		Data   CompoundLiteralData
		Line   int
		Column int
	}

	CompoundLiteralData struct {
		Fields []Token
		Values []Expression
		Line   int
		Column int
	}

	ArrayLiteral struct {
		Exprs  []Expression
		Line   int
		Column int
	}

	HeapAlloc struct {
		Type   Type
		Val    Expression
		Line   int
		Column int
	}

	LenExpr struct {
		Type   Type
		Line   int
		Column int
	}

	SizeExpr struct {
		Expr   Expression
		Line   int
		Column int
	}
)

type (
	FuncType struct {
		Type        FunctionType
		ArgTypes    []Type
		ArgNames    []Token
		ReturnTypes []Type
		Mut         bool
		Line        int
		Column      int
	}

	StructType struct {
		Name         Token
		Props        []Declaration
		SuperStructs []Expression
		Line         int
		Column       int
	}

	TupleType struct {
		Types  []Type
		Line   int
		Column int
	}

	EnumType struct {
		Identifiers []Token
		Values      []Expression
		Line        int
		Column      int
	}

	UnionType struct {
		Identifiers []Token
		Types       []Type
		Line        int
		Column      int
	}

	BasicType struct {
		Expr   Expression
		Line   int
		Column int
	}

	PointerType struct {
		BaseType Type
		Line     int
		Column   int
	}

	VecType struct {
		BaseType Type
		Line     int
		Column   int
	}

	ConstType struct {
		BaseType Type
		Line     int
		Column   int
	}

	CaptureType struct {
		BaseType Type
		Line     int
		Column   int
	}

	PromiseType struct {
		BaseType Type
		Line     int
		Column   int
	}

	ImplictArrayType struct {
		BaseType Type
		Line     int
		Column   int
	}

	ArrayType struct {
		Size     Token
		BaseType Type
		Line     int
		Column   int
	}

	StaticType struct {
		BaseType Type
		Line     int
		Column   int
	}

	InternalType struct{}
	NumberType   struct{}
)

func (Block) isStatement()           {}
func (Declaration) isStatement()     {}
func (Import) isStatement()          {}
func (Loop) isStatement()            {}
func (Switch) isStatement()          {}
func (IfElseBlock) isStatement()     {}
func (Return) isStatement()          {}
func (Assignment) isStatement()      {}
func (NullStatement) isStatement()   {}
func (Break) isStatement()           {}
func (Continue) isStatement()        {}
func (Defer) isStatement()           {}
func (Delete) isStatement()          {}
func (ExportStatement) isStatement() {}

func (BasicLit) isExpression()            {}
func (BinaryExpr) isExpression()          {}
func (UnaryExpr) isExpression()           {}
func (CallExpr) isExpression()            {}
func (FuncExpr) isExpression()            {}
func (TernaryExpr) isExpression()         {}
func (PostfixUnaryExpr) isExpression()    {}
func (TypeCast) isExpression()            {}
func (IdentExpr) isExpression()           {}
func (MemberExpr) isExpression()          {}
func (ArrayMemberExpr) isExpression()     {}
func (CompoundLiteral) isExpression()     {}
func (CompoundLiteralData) isExpression() {}
func (HeapAlloc) isExpression()           {}
func (ArrayLiteral) isExpression()        {}
func (LenExpr) isExpression()             {}
func (SizeExpr) isExpression()            {}
func (PointerMemberExpr) isExpression()   {}

func (BasicLit) isStatement()            {}
func (BinaryExpr) isStatement()          {}
func (UnaryExpr) isStatement()           {}
func (CallExpr) isStatement()            {}
func (FuncExpr) isStatement()            {}
func (TernaryExpr) isStatement()         {}
func (PostfixUnaryExpr) isStatement()    {}
func (TypeCast) isStatement()            {}
func (IdentExpr) isStatement()           {}
func (MemberExpr) isStatement()          {}
func (ArrayMemberExpr) isStatement()     {}
func (CompoundLiteral) isStatement()     {}
func (CompoundLiteralData) isStatement() {}
func (HeapAlloc) isStatement()           {}
func (ArrayLiteral) isStatement()        {}
func (LenExpr) isStatement()             {}
func (SizeExpr) isStatement()            {}
func (PointerMemberExpr) isStatement()   {}

func (BasicType) isType()        {}
func (StructType) isType()       {}
func (EnumType) isType()         {}
func (TupleType) isType()        {}
func (UnionType) isType()        {}
func (FuncType) isType()         {}
func (ConstType) isType()        {}
func (PointerType) isType()      {}
func (ArrayType) isType()        {}
func (VecType) isType()          {}
func (ImplictArrayType) isType() {}
func (Typedef) isType()          {}
func (InternalType) isType()     {}
func (NumberType) isType()       {}
func (CaptureType) isType()      {}
func (StaticType) isType()       {}
func (PromiseType) isType()      {}

func (BasicType) isExpression()        {}
func (StructType) isExpression()       {}
func (EnumType) isExpression()         {}
func (TupleType) isExpression()        {}
func (UnionType) isExpression()        {}
func (FuncType) isExpression()         {}
func (ConstType) isExpression()        {}
func (PointerType) isExpression()      {}
func (ArrayType) isExpression()        {}
func (VecType) isExpression()          {}
func (ImplictArrayType) isExpression() {}
func (Typedef) isExpression()          {}
func (InternalType) isExpression()     {}
func (NumberType) isExpression()       {}
func (CaptureType) isExpression()      {}
func (StaticType) isExpression()       {}
func (PromiseType) isExpression()      {}

func (BasicType) isStatement()        {}
func (StructType) isStatement()       {}
func (EnumType) isStatement()         {}
func (TupleType) isStatement()        {}
func (UnionType) isStatement()        {}
func (FuncType) isStatement()         {}
func (ConstType) isStatement()        {}
func (PointerType) isStatement()      {}
func (ArrayType) isStatement()        {}
func (VecType) isStatement()          {}
func (ImplictArrayType) isStatement() {}
func (Typedef) isStatement()          {}
func (InternalType) isStatement()     {}
func (NumberType) isStatement()       {}
func (CaptureType) isStatement()      {}
func (StaticType) isStatement()       {}
func (PromiseType) isStatement()      {}

func (s Block) LineM() int {
	return s.Line
}
func (s Declaration) LineM() int {
	return s.Line
}
func (s Import) LineM() int {
	return s.Line
}
func (s Loop) LineM() int {
	return s.Line
}
func (s Switch) LineM() int {
	return s.Line
}
func (s IfElseBlock) LineM() int {
	return s.Line
}
func (s Return) LineM() int {
	return s.Line
}
func (s Assignment) LineM() int {
	return s.Line
}
func (s NullStatement) LineM() int {
	return s.Line
}
func (s Break) LineM() int {
	return s.Line
}
func (s Continue) LineM() int {
	return s.Line
}
func (s Defer) LineM() int {
	return s.Line
}
func (s Delete) LineM() int {
	return s.Line
}
func (s ExportStatement) LineM() int {
	return s.Line
}

func (s Block) ColumnM() int {
	return s.Column
}
func (s Declaration) ColumnM() int {
	return s.Column
}
func (s Import) ColumnM() int {
	return s.Column
}
func (s Loop) ColumnM() int {
	return s.Column
}
func (s Switch) ColumnM() int {
	return s.Column
}
func (s IfElseBlock) ColumnM() int {
	return s.Column
}
func (s Return) ColumnM() int {
	return s.Column
}
func (s Assignment) ColumnM() int {
	return s.Column
}
func (s NullStatement) ColumnM() int {
	return s.Column
}
func (s Break) ColumnM() int {
	return s.Column
}
func (s Continue) ColumnM() int {
	return s.Column
}
func (s Defer) ColumnM() int {
	return s.Column
}
func (s Delete) ColumnM() int {
	return s.Column
}
func (s ExportStatement) ColumnM() int {
	return s.Column
}

func (e BasicLit) LineM() int {
	return e.Line
}
func (e BinaryExpr) LineM() int {
	return e.Line
}
func (e UnaryExpr) LineM() int {
	return e.Line
}
func (e CallExpr) LineM() int {
	return e.Line
}
func (e FuncExpr) LineM() int {
	return e.Line
}
func (e TernaryExpr) LineM() int {
	return e.Line
}
func (e PostfixUnaryExpr) LineM() int {
	return e.Line
}
func (e TypeCast) LineM() int {
	return e.Line
}
func (e IdentExpr) LineM() int {
	return e.Line
}
func (e MemberExpr) LineM() int {
	return e.Line
}
func (e ArrayMemberExpr) LineM() int {
	return e.Line
}
func (e CompoundLiteral) LineM() int {
	return e.Line
}
func (e CompoundLiteralData) LineM() int {
	return e.Line
}
func (e HeapAlloc) LineM() int {
	return e.Line
}
func (e ArrayLiteral) LineM() int {
	return e.Line
}
func (e LenExpr) LineM() int {
	return e.Line
}
func (e SizeExpr) LineM() int {
	return e.Line
}
func (e PointerMemberExpr) LineM() int {
	return e.Line
}
func (e BasicLit) ColumnM() int {
	return e.Column
}
func (e BinaryExpr) ColumnM() int {
	return e.Column
}
func (e UnaryExpr) ColumnM() int {
	return e.Column
}
func (e CallExpr) ColumnM() int {
	return e.Column
}
func (e FuncExpr) ColumnM() int {
	return e.Column
}
func (e TernaryExpr) ColumnM() int {
	return e.Column
}
func (e PostfixUnaryExpr) ColumnM() int {
	return e.Column
}
func (e TypeCast) ColumnM() int {
	return e.Column
}
func (e IdentExpr) ColumnM() int {
	return e.Column
}
func (e MemberExpr) ColumnM() int {
	return e.Column
}
func (e ArrayMemberExpr) ColumnM() int {
	return e.Column
}
func (e CompoundLiteral) ColumnM() int {
	return e.Column
}
func (e CompoundLiteralData) ColumnM() int {
	return e.Column
}
func (e HeapAlloc) ColumnM() int {
	return e.Column
}
func (e ArrayLiteral) ColumnM() int {
	return e.Column
}
func (e LenExpr) ColumnM() int {
	return e.Column
}
func (e SizeExpr) ColumnM() int {
	return e.Column
}
func (e PointerMemberExpr) ColumnM() int {
	return e.Column
}

func (t BasicType) LineM() int {
	return t.Line
}
func (t StructType) LineM() int {
	return t.Line
}
func (t EnumType) LineM() int {
	return t.Line
}
func (t TupleType) LineM() int {
	return t.Line
}
func (t UnionType) LineM() int {
	return t.Line
}
func (t FuncType) LineM() int {
	return t.Line
}
func (t ConstType) LineM() int {
	return t.Line
}
func (t PointerType) LineM() int {
	return t.Line
}
func (t ArrayType) LineM() int {
	return t.Line
}
func (t VecType) LineM() int {
	return t.Line
}
func (t ImplictArrayType) LineM() int {
	return t.Line
}
func (t Typedef) LineM() int {
	return t.Line
}
func (t InternalType) LineM() int {
	return -1
}
func (t NumberType) LineM() int {
	return -1
}
func (t CaptureType) LineM() int {
	return t.Line
}
func (t StaticType) LineM() int {
	return t.Line
}
func (t PromiseType) LineM() int {
	return t.Line
}

func (t BasicType) ColumnM() int {
	return t.Column
}
func (t StructType) ColumnM() int {
	return t.Column
}
func (t EnumType) ColumnM() int {
	return t.Column
}
func (t TupleType) ColumnM() int {
	return t.Column
}
func (t UnionType) ColumnM() int {
	return t.Column
}
func (t FuncType) ColumnM() int {
	return t.Column
}
func (t ConstType) ColumnM() int {
	return t.Column
}
func (t PointerType) ColumnM() int {
	return t.Column
}
func (t ArrayType) ColumnM() int {
	return t.Column
}
func (t VecType) ColumnM() int {
	return t.Column
}
func (t ImplictArrayType) ColumnM() int {
	return t.Column
}
func (t Typedef) ColumnM() int {
	return t.Column
}
func (t InternalType) ColumnM() int {
	return -1
}
func (t NumberType) ColumnM() int {
	return -1
}
func (t CaptureType) ColumnM() int {
	return t.Column
}
func (t StaticType) ColumnM() int {
	return t.Column
}
func (t PromiseType) ColumnM() int {
	return t.Column
}

var VoidToken = Token{Buff: []byte("void"), PrimaryType: Identifier}
var VoidType = Typedef{Name: VoidToken, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$void"), PrimaryType: Identifier}}}}

var BoolToken = Token{Buff: []byte("bool"), PrimaryType: Identifier}
var BoolType = Typedef{Name: BoolToken, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$bool"), PrimaryType: Identifier}}}}

var U8Token = Token{Buff: []byte("u8"), PrimaryType: Identifier}
var U8Type = Typedef{Name: U8Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$u8"), PrimaryType: Identifier}}}}

var U16Token = Token{Buff: []byte("u16"), PrimaryType: Identifier}
var U16Type = Typedef{Name: U16Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$u16"), PrimaryType: Identifier}}}}

var U32Token = Token{Buff: []byte("u32"), PrimaryType: Identifier}
var U32Type = Typedef{Name: U32Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$u32"), PrimaryType: Identifier}}}}

var U64Token = Token{Buff: []byte("u64"), PrimaryType: Identifier}
var U64Type = Typedef{Name: U64Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$u64"), PrimaryType: Identifier}}}}

var I8Token = Token{Buff: []byte("i8"), PrimaryType: Identifier}
var I8Type = Typedef{Name: I8Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$i8"), PrimaryType: Identifier}}}}

var I16Token = Token{Buff: []byte("i16"), PrimaryType: Identifier}
var I16Type = Typedef{Name: I16Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$i16"), PrimaryType: Identifier}}}}

var I32Token = Token{Buff: []byte("i32"), PrimaryType: Identifier}
var I32Type = Typedef{Name: I32Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$i32"), PrimaryType: Identifier}}}}

var I64Token = Token{Buff: []byte("i64"), PrimaryType: Identifier}
var I64Type = Typedef{Name: I64Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$i64"), PrimaryType: Identifier}}}}

var F32Token = Token{Buff: []byte("f32"), PrimaryType: Identifier}
var F32Type = Typedef{Name: F32Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$f32"), PrimaryType: Identifier}}}}

var F64Token = Token{Buff: []byte("i64"), PrimaryType: Identifier}
var F64Type = Typedef{Name: F64Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$f64"), PrimaryType: Identifier}}}}

var True = IdentExpr{Value: Token{Buff: []byte("true"), PrimaryType: Identifier}}
var False = IdentExpr{Value: Token{Buff: []byte("false"), PrimaryType: Identifier}}

var SizeTToken = Token{Buff: []byte("size_t"), PrimaryType: Identifier}
var SizeTType = Typedef{Name: I64Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$size_t"), PrimaryType: Identifier}}}}

var UptrToken = Token{Buff: []byte("uptr"), PrimaryType: Identifier}
var UptrType = Typedef{Name: I64Token, Type: BasicType{Expr: IdentExpr{Value: Token{Buff: []byte("$uptr"), PrimaryType: Identifier}}}}

var Null = IdentExpr{Value: Token{Buff: []byte("null"), PrimaryType: Identifier}}

var Globals = []string{"u8", "u16", "u32", "u64", "i8", "i16", "i32", "i64", "uptr", "f32", "f64", "void", "bool", "size_t", "true", "false", "null"}
var GlobalTypes = []string{"u8", "u16", "u32", "u64", "i8", "i16", "i32", "i64", "uptr", "f32", "f64", "void", "bool", "size_t"}
