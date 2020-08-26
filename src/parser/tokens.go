package parser

import "strconv"

// PrimaryTokenType enum
type PrimaryTokenType byte

//SecondaryTokenType enum
type SecondaryTokenType byte

// The numbers here are just to keep values different, they dont mean anything specific
// Primary token type

const (
	PrimaryNullType PrimaryTokenType = 0

	Identifier PrimaryTokenType = 1

	NumberLiteral PrimaryTokenType = 2
	StringLiteral PrimaryTokenType = 3
	CharLiteral   PrimaryTokenType = 4

	StringDelimiter PrimaryTokenType = 6
	CharDelimiter   PrimaryTokenType = 7

	LeftParen       PrimaryTokenType = 8
	RightParen      PrimaryTokenType = 9
	LeftBrace       PrimaryTokenType = 10
	RightBrace      PrimaryTokenType = 11
	LeftCurlyBrace  PrimaryTokenType = 12
	RightCurlyBrace PrimaryTokenType = 13
	SemiColon       PrimaryTokenType = 14
	Comma           PrimaryTokenType = 15

	// Operators
	AirthmaticOperator PrimaryTokenType = 51
	AssignmentOperator PrimaryTokenType = 52
	RelationalOperator PrimaryTokenType = 53
	LogicalOperator    PrimaryTokenType = 54
	BitwiseOperator    PrimaryTokenType = 55
	SpecialOperator    PrimaryTokenType = 56

	// Keywords
	ForKeyword      PrimaryTokenType = 101
	SwitchKeyword   PrimaryTokenType = 102
	IfKeyword       PrimaryTokenType = 103
	ElseKeyword     PrimaryTokenType = 104
	FunctionKeyword PrimaryTokenType = 105
	StructKeyword   PrimaryTokenType = 106
	TupleKeyword    PrimaryTokenType = 107
	EnumKeyword     PrimaryTokenType = 108
	CaseKeyword     PrimaryTokenType = 109
	AsyncKeyword    PrimaryTokenType = 110
	WorkKeyword     PrimaryTokenType = 111
	ImportKeyword   PrimaryTokenType = 113
	DeferKeyword    PrimaryTokenType = 114
	ReturnKeyword   PrimaryTokenType = 115
	DefaultKeyword  PrimaryTokenType = 116
	BreakKeyword    PrimaryTokenType = 117
	ContinueKeyword PrimaryTokenType = 118
	NewKeyword      PrimaryTokenType = 119
	ConstKeyword    PrimaryTokenType = 120
	VecKeyword      PrimaryTokenType = 121
	DeleteKeyword   PrimaryTokenType = 122
	TypedefKeyword  PrimaryTokenType = 123
	CastKeyword     PrimaryTokenType = 124
	LenKeyword      PrimaryTokenType = 125
	SizeKeyword     PrimaryTokenType = 126
	ExportKeyword   PrimaryTokenType = 127
	UnionKeyword    PrimaryTokenType = 128
	StaticKeyword   PrimaryTokenType = 129
	CaptureKeyword  PrimaryTokenType = 130
	PromiseKeyword  PrimaryTokenType = 131

	// the parser stops parsing when it receives either of these types and shows the correct error message
	EOF        PrimaryTokenType = 254
	ErrorToken PrimaryTokenType = 255
)

// Secondary token type
const (
	SecondaryNullType SecondaryTokenType = 0

	// Radix for number literals
	DecimalRadix     SecondaryTokenType = 1
	BinaryRadix      SecondaryTokenType = 2
	OctalRadix       SecondaryTokenType = 3
	HexadecimalRadix SecondaryTokenType = 4

	// Airthmatic Operators
	Add     SecondaryTokenType = 11
	Sub     SecondaryTokenType = 12
	Mul     SecondaryTokenType = 13
	Div     SecondaryTokenType = 14
	Modulus SecondaryTokenType = 15

	// Assignment Operators
	AddEqual     SecondaryTokenType = 21
	SubEqual     SecondaryTokenType = 22
	MulEqual     SecondaryTokenType = 23
	DivEqual     SecondaryTokenType = 24
	ModulusEqual SecondaryTokenType = 25
	Equal        SecondaryTokenType = 26
	AddAdd       SecondaryTokenType = 27
	SubSub       SecondaryTokenType = 28

	// Realtional Operators
	EqualEqual   SecondaryTokenType = 31
	NotEqual     SecondaryTokenType = 32
	Greater      SecondaryTokenType = 33
	Less         SecondaryTokenType = 34
	LessEqual    SecondaryTokenType = 35
	GreaterEqual SecondaryTokenType = 36

	// Logical Operators
	AndAnd SecondaryTokenType = 41
	OrOr   SecondaryTokenType = 42
	Not    SecondaryTokenType = 43

	// Bitwise Opeartors
	LeftShift   SecondaryTokenType = 51
	RightShift  SecondaryTokenType = 52
	Or          SecondaryTokenType = 53
	And         SecondaryTokenType = 54
	ExclusiveOr SecondaryTokenType = 55
	BitwiseNot  SecondaryTokenType = 56

	// Special Operators
	Colon    SecondaryTokenType = 61
	QuesMark SecondaryTokenType = 62
	Dot      SecondaryTokenType = 63
	DotDot   SecondaryTokenType = 65

	// Encoding for char literals
	Byte1Char SecondaryTokenType = 71
	Byte2Char SecondaryTokenType = 72
	Byte3Char SecondaryTokenType = 73
	Byte4Char SecondaryTokenType = 74

	// For use with ErrorToken
	UnexpectedEOF SecondaryTokenType = 101
	NotFound      SecondaryTokenType = 102
	UnknownChar   SecondaryTokenType = 103
)

// Keywords urgh idk what to write
var Keywords = map[string]PrimaryTokenType{
	"if":       IfKeyword,
	"else":     ElseKeyword,
	"for":      ForKeyword,
	"switch":   SwitchKeyword,
	"case":     CaseKeyword,
	"enum":     EnumKeyword,
	"struct":   StructKeyword,
	"async":    AsyncKeyword,
	"work":     WorkKeyword,
	"import":   ImportKeyword,
	"defer":    DeferKeyword,
	"func":     FunctionKeyword,
	"return":   ReturnKeyword,
	"default":  DefaultKeyword,
	"break":    BreakKeyword,
	"continue": ContinueKeyword,
	"tuple":    TupleKeyword,
	"new":      NewKeyword,
	"const":    ConstKeyword,
	"vec":      VecKeyword,
	"delete":   DeleteKeyword,
	"typedef":  TypedefKeyword,
	"cast":     CastKeyword,
	// "len":      LenKeyword,
	"sizeof":  SizeKeyword,
	"export":  ExportKeyword,
	"union":   UnionKeyword,
	"static":  StaticKeyword,
	"capture": CaptureKeyword,
	"promise": PromiseKeyword,
	// more stuff
}

// Token ...
type Token struct {
	PrimaryType   PrimaryTokenType
	SecondaryType SecondaryTokenType
	Buff          []byte
	Line          int
	Column        int
	Flags         int
}

// --------------------------------------------
// --------------------------------------------
// ---------------- Debug Info ----------------
// --------------------------------------------
// --------------------------------------------

// PrimaryTypes map with debug info for PrimaryTokenType
var PrimaryTypes map[PrimaryTokenType]string = map[PrimaryTokenType]string{
	PrimaryNullType: "Null",

	Identifier: "Identifier",

	NumberLiteral: "NumberLiteral",
	StringLiteral: "StringLiteral",
	CharLiteral:   "CharLiteral",

	StringDelimiter: "StringDelimiter",
	CharDelimiter:   "CharDelimiter",

	LeftParen:       "LeftParen",
	RightParen:      "RightParen",
	LeftBrace:       "LeftBrace",
	RightBrace:      "RightBrace",
	LeftCurlyBrace:  "LeftCurlyBrace",
	RightCurlyBrace: "RightCurlyBrace",
	SemiColon:       "SemiColon",
	Comma:           "Comma",

	AirthmaticOperator: "AirthmaticOperator",
	AssignmentOperator: "AssignmentOperator",
	RelationalOperator: "RelationalOperator",
	LogicalOperator:    "LogicalOperator",
	BitwiseOperator:    "BitwiseOperator",
	SpecialOperator:    "SpecialOperator",

	ForKeyword:      "ForKeyword",
	SwitchKeyword:   "SwitchKeyword",
	IfKeyword:       "IfKeyword",
	ElseKeyword:     "ElseKeyword",
	FunctionKeyword: "FunctionKeyword",
	StructKeyword:   "StructKeyword",
	TupleKeyword:    "TupleKeyword",
	EnumKeyword:     "EnumKeyword",
	CaseKeyword:     "CaseKeyword",
	AsyncKeyword:    "AsyncKeyword",
	WorkKeyword:     "WorkKeyword",
	ImportKeyword:   "ImportKeyowrd",

	EOF:        "EOF",
	ErrorToken: "ErrorToken",
}

// SecondaryTypes map with debug info for SecondaryTokenType
var SecondaryTypes map[SecondaryTokenType]string = map[SecondaryTokenType]string{
	SecondaryNullType: "Null",

	DecimalRadix:     "DecimalRadix",
	BinaryRadix:      "BinaryRadix",
	OctalRadix:       "OctalRadix",
	HexadecimalRadix: "HexadecimalRadix",

	Add:     "Add",
	Sub:     "Sub",
	Mul:     "Mul",
	Div:     "Div",
	Modulus: "Modulus",

	AddEqual:     "AddEqual",
	SubEqual:     "SubEqual",
	MulEqual:     "MulEqual",
	DivEqual:     "DivEqual",
	ModulusEqual: "ModulusEqual",
	Equal:        "Equal",

	AddAdd: "AddAdd",
	SubSub: "SubSub",

	EqualEqual:   "EqualEqual",
	NotEqual:     "NotEqual",
	Greater:      "Greater",
	Less:         "Less",
	LessEqual:    "LessEqual",
	GreaterEqual: "GreaterEqual",

	AndAnd: "AndAnd",
	OrOr:   "OrOr",
	Not:    "Not",

	LeftShift:   "LeftShift",
	RightShift:  "RightShift",
	Or:          "Or",
	And:         "And",
	ExclusiveOr: "ExclusiveOr",

	Colon:    "Colon",
	QuesMark: "QuesMark",
	Dot:      "Dot",
	DotDot:   "DotDot",

	Byte1Char: "Byte1Char",
	Byte2Char: "Byte2Char",
	Byte3Char: "Byte3Char",
	Byte4Char: "Byte4Char",

	UnexpectedEOF: "UnexpectedEOF",
	NotFound:      "NotFound",
	UnknownChar:   "UnknownChar",
}

// Serialize serializes a token
func (token Token) Serialize() string {
	return "{\n\tPrimaryType:\t" + PrimaryTypes[token.PrimaryType] + ",\n\tSecondaryType:\t" + SecondaryTypes[token.SecondaryType] + ",\n\tValue:\t" + string(token.Buff) + "\n\tLine:\t" + strconv.Itoa(token.Line) + "\n\tColumn:\t" + strconv.Itoa(token.Column) + "\n},"
}
