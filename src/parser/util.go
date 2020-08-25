package parser

//IsChar checks if `b` is an english alphabet `(a-z|A-Z)`
func IsChar(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

//IsNumBi checks if `b` is a binary digit `(0|1)`
func IsNumBi(b byte) bool {
	return b == '0' || b == '1'
}

//IsNumOct checks if `b` is an octal digit `(0-7)`
func IsNumOct(b byte) bool {
	return (b >= '0' && b <= '7')
}

//IsNumDec checks if `b` is a decimal digit `(0-9)`
func IsNumDec(b byte) bool {
	return (b >= '0' && b <= '9')
}

//IsNumHex checks if `b` is a hexadecimal digit `(0-9|a-f|A-F)`
func IsNumHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'F') || (b >= 'a' && b <= 'f')
}

//IsIdentifierPart checks if `b` is a valid character for being a part of identifier `(a-z|A-Z|0-9|_)`
func IsIdentifierPart(b byte) bool {
	return IsChar(b) || IsNumDec(b) || b == '_' || b == '$'
}

//IsIdentifierBegining checks if `b` is valid for being the first character of an identifier `(a-z|A-Z|_)`
func IsIdentifierBegining(b byte) bool {
	return IsChar(b) || b == '_' || b == '$'
}

//IsStringDelimiter checks if `b` is a string delimiter `"`
func IsStringDelimiter(b byte) bool {
	return b == '"'
}

//IsCharDelimiter checks if `b` is a char delimiter `'`
func IsCharDelimiter(b byte) bool {
	return b == '\''
}

//IsSpace checks if `b` is a blank space (' '|'\n'|'\t|'\r')
func IsSpace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t'
}

//GetWordType retuns the right PrimaryTokenType for identifiers or keywords
func GetWordType(str string) PrimaryTokenType {
	if keywordType, ok := Keywords[str]; ok {
		return keywordType
	}
	return Identifier
}

func Pow(a, b int) int {
	r := 1
	for i := 0; i < b; i++ {
		r = r * a
	}
	return r
}

func HexToInt(b byte) int {
	if IsNumDec(b) {
		return int(b - 48)
	} else {
		return int(b - 65)
	}
}
