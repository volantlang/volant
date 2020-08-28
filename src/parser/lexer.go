package parser

import (
	"error"
	"math"
	"strconv"
)

//Lexer does stuff
type Lexer struct {
	Buffer   []byte // code read from some file or something
	Position int    // position of char to process next in Buffer

	// for error messages and all
	Line   int // line count of the current char in code
	Column int // column count of the current char in the current line
}

func (lexer *Lexer) readToBuffer() byte {
	// read few bytes from a file or command line or whatever
	// lexer.Buffer = append(lexer.Buffer, 'a') // just for now
	return 0 // eof
}

func (lexer *Lexer) nextChar() (byte, bool) {
	character, ok := lexer.peek()
	if !ok {
		return 0, false
	}
	lexer.eatLastByte()
	return character, true
}

// get the next character without incrementing the position counter of lexer
func (lexer *Lexer) peek() (byte, bool) {
	// if no more characters to read in the Buffer
	if lexer.Position == len(lexer.Buffer) {
		// read few bytes from the source file and push it to Buffer
		// return 0 if eof reached
		if lexer.readToBuffer() == 0 {
			return 0, false // marks eof (tells the Lexer to stop lexing)
		}
	}

	// reached only when the Buffer has more bytes to read
	return lexer.Buffer[lexer.Position], true
}

// increment the position counter of Lexer by 1
func (lexer *Lexer) eatLastByte() {
	lexer.Position++
	lexer.Column++
}

// increment the line count and zero the column count when a newline is reached
func (lexer *Lexer) shiftLine() {
	lexer.Position++
	lexer.Line++
	lexer.Column = 0
}

func (lexer *Lexer) skipSpaces() {
	// no need to check for eof
	for next, _ := lexer.peek(); ; next, _ = lexer.peek() {
		if next == '\n' {
			lexer.shiftLine()
		} else if next == ' ' || next == '\t' || next == '\r' {
			lexer.eatLastByte()
		} else { // eof handled here
			break
		}
	}
}

func (lexer *Lexer) skipUntilNewline() {
	for next, ok := lexer.peek(); ok && next != '\n'; next, _ = lexer.peek() {
		lexer.eatLastByte()
	}
}

func (lexer *Lexer) multilineComment() {
	for {
		lexer.skipSpaces()

		if next, ok := lexer.peek(); !ok {
			error.New("Expected end of multiline comment, got eof.", lexer.Line, lexer.Column)
		} else if next != '*' {
			lexer.eatLastByte()
			continue
		}
		lexer.eatLastByte()

		if next, ok := lexer.peek(); !ok {
			error.New("Expected end of multiline comment, got eof.", lexer.Line, lexer.Column)
		} else if next != '/' {
			continue
		}

		lexer.eatLastByte()
		break
	}
}

func (lexer *Lexer) skipComments() {
	if next, _ := lexer.peek(); next != byte('/') {
		return
	}
	lexer.eatLastByte()
	if next, _ := lexer.peek(); next == byte('/') {
		lexer.skipUntilNewline()
	} else if next == '*' {
		lexer.eatLastByte()
		lexer.multilineComment()
	} else {
		lexer.Position--
		lexer.Column--
	}
}

func (lexer *Lexer) PrepNext() {
	for {
		if next, _ := lexer.peek(); IsSpace(next) {
			lexer.skipSpaces() // skiping spaces/tabs/newlines
		}
		if next, _ := lexer.peek(); next == byte('/') {
			lexer.eatLastByte()
		} else {
			break
		}
		if next, _ := lexer.peek(); next == '/' {
			lexer.skipUntilNewline()
		} else if next == '*' {
			lexer.multilineComment()
		} else {
			lexer.Position--
			lexer.Column--
			break
		}
	}
}

// NextToken returns next token
func (lexer *Lexer) NextToken() Token {
	lexer.PrepNext()
	character, ok := lexer.peek()
	if !ok {
		// return eof token (tell the parser to stop further parsing)
		return Token{PrimaryType: EOF, SecondaryType: SecondaryNullType, Buff: nil, Line: lexer.Line, Column: lexer.Column}
	} else if IsIdentifierBegining(character) {
		return lexer.lexWord() // identifier or keyword or macros
	} else if IsNumDec(character) {
		return lexer.lexNumber() // number, hex starting with 0x, octals with 0o, decimals either directly or with 0d, and binary with 0b
	} else if IsStringDelimiter(character) { // string delimiter is "
		lexer.eatLastByte()
		return lexer.lexString()
	} else if IsCharDelimiter(character) { // char delimiter is '
		lexer.eatLastByte()
		return lexer.lexChar() // just a single byte
	} else {
		if op := lexer.lexOperator(); op.SecondaryType != NotFound {
			return op
		} else if op = lexer.lexDelimiter(); op.SecondaryType != NotFound {
			return op
		}
	}
	error.New("Unknown character.", lexer.Line, lexer.Column)
	return Token{PrimaryType: ErrorToken, SecondaryType: UnknownChar, Buff: nil}
}

func (lexer *Lexer) lexNumber() Token {
	var radix SecondaryTokenType = DecimalRadix
	num := []byte{}

	line := lexer.Line
	column := lexer.Column

	// already checked for eof in NextToken()
	if character, _ := lexer.nextChar(); character == '0' {

		// check either the number has a base specified or we have to implictly assume it
		// no need to check for eof
		if next, _ := lexer.peek(); next == 'd' { // decimal
			radix = DecimalRadix
			lexer.eatLastByte()
			lexer.eatLastByte()

			if n3, _ := lexer.peek(); n3 == '0' || n3 == '_' {
				lexer.eatLastByte()
				for n, _ := lexer.peek(); n == '0' || n == '_'; n, _ = lexer.peek() {
					lexer.eatLastByte()
				}
			} else if !IsNumDec(n3) {
				num = append(num, '0')
			}
		} else if next == 'b' { // binary
			radix = BinaryRadix
			lexer.eatLastByte()
			num = append(num, '0')
			num = append(num, next)
		} else if next == 'o' { // octal
			radix = OctalRadix
			lexer.eatLastByte()
			num = append(num, '0')
		} else if next == 'x' { // hexadecimal
			radix = HexadecimalRadix
			lexer.eatLastByte()
			num = append(num, '0')
			num = append(num, next)
		} else if next == '0' {
			lexer.eatLastByte()
			if n3, _ := lexer.peek(); n3 == '0' || n3 == '_' {
				lexer.eatLastByte()
				for n, _ := lexer.peek(); n == '0' || n == '_'; n, _ = lexer.peek() {
					lexer.eatLastByte()
				}
			} else if !IsNumDec(n3) {
				num = append(num, '0')
			}
		} else {
			num = append(num, '0')
		}
	} else {
		num = append(num, character)
	}

	for m, _ := lexer.peek(); m == '_'; m, _ = lexer.peek() {
		lexer.eatLastByte()
	}

	switch radix {
	case DecimalRadix:
		for n, _ := lexer.peek(); IsNumDec(n) || n == '.'; n, _ = lexer.peek() {
			num = append(num, n)
			lexer.eatLastByte()
			for n2, _ := lexer.peek(); n2 == '_'; n2, _ = lexer.peek() {
				lexer.eatLastByte()
			}
		}
	case BinaryRadix:
		for n, _ := lexer.peek(); IsNumBi(n) || n == '.'; n, _ = lexer.peek() {
			num = append(num, n)
			lexer.eatLastByte()
			for n2, _ := lexer.peek(); n2 == '_'; n2, _ = lexer.peek() {
				lexer.eatLastByte()
			}
		}
	case OctalRadix:
		for n, _ := lexer.peek(); IsNumOct(n) || n == '.'; n, _ = lexer.peek() {
			num = append(num, n)
			lexer.eatLastByte()
			for n2, _ := lexer.peek(); n2 == '_'; n2, _ = lexer.peek() {
				lexer.eatLastByte()
			}
		}
	case HexadecimalRadix:
		for n, _ := lexer.peek(); IsNumHex(n) || n == '.'; n, _ = lexer.peek() {
			num = append(num, n)
			lexer.eatLastByte()
			for n2, _ := lexer.peek(); n2 == '_'; n2, _ = lexer.peek() {
				lexer.eatLastByte()
			}
		}
	}

	return Token{Buff: num, PrimaryType: NumberLiteral, SecondaryType: radix, Line: line, Column: column}
}

func (lexer *Lexer) lexChar() Token {

	// stores the utf32 code of char
	num := 0

	var encoding SecondaryTokenType

	line := lexer.Line
	column := lexer.Column

	character, ok := lexer.peek()

	if !ok {
		// Error: Expected char, got eof
		return Token{PrimaryType: ErrorToken, SecondaryType: UnexpectedEOF, Buff: nil, Line: lexer.Line, Column: lexer.Column}
	} else if character == '\n' {
		// Error: Expected char, got end of line
		return Token{PrimaryType: ErrorToken, SecondaryType: UnexpectedEOF, Buff: nil, Line: lexer.Line, Column: lexer.Column}
	}

	lexer.eatLastByte()

	if character == '\\' {
		next, ok := lexer.peek()

		if !ok { // Error: Expected char, got eof
			return Token{PrimaryType: ErrorToken, SecondaryType: UnexpectedEOF, Buff: nil, Line: lexer.Line, Column: lexer.Column}
		} else if character == '\n' {
			// Error: Expected char, got end of line
			return Token{PrimaryType: ErrorToken, SecondaryType: UnexpectedEOF, Buff: nil, Line: lexer.Line, Column: lexer.Column}
		}

		lexer.eatLastByte()

		switch next {
		case 't':
			num = '\t'
		case 'n':
			num = '\n'
		case 'r':
			num = '\r'
		case '\'':
			num = '\''
		case '\\':
			num = '\\'
		case 'u': // 2 byte char
			for i := 0; i < 4; i++ {
				chr, ok := lexer.peek()

				if !ok { // Error: Expected escape sequence, got eof
					error.New("exprected escape sequence, got eof.", line, column)
				} else if IsNumHex(chr) {
					num += HexToInt(chr) * Pow(16, (3-i))
					lexer.eatLastByte()
				} else { // Error: Invalid character in escape sequence, expected (0-9|A-F|a-f)
					error.New("invalid character in escape sequence.", line, column)
				}
			}

			encoding = Byte2Char
		case 'U': // 4 byte char
			for i := 0; i < 8; i++ {
				chr, ok := lexer.peek()

				if !ok { // Error: Expected escape sequence, got eof
					error.New("exprected escape sequence, got eof.", line, column)
				} else if IsNumHex(chr) {
					num += HexToInt(chr) * Pow(16, (7-i))
					lexer.eatLastByte()
				} else { // Error: Invalid character in escape sequence, expected (0-9|A-F|a-f)
					error.New("invalid character in escape sequence.", line, column)
				}
			}
			encoding = Byte4Char
		default:
			// idk we can prob throw an error here saying that its an invalid escape sequence?
			num = int(next)
		}
		if nextChar, _ := lexer.peek(); nextChar == '\'' {
			lexer.eatLastByte() // increament the positon as `nextChar` was `'` as expected
			return Token{PrimaryType: CharLiteral, SecondaryType: encoding, Buff: []byte(strconv.Itoa(num)), Line: line, Column: column}
		}
		error.New("exprected ', got eof.", line, column)
	} else if character>>7 == 0 { // 1 byte char
		encoding = Byte1Char
	} else if character>>5 == 0b110 { // 2 byte char
		encoding = Byte2Char
	} else if character>>4 == 0b1110 { // 3 byte char
		encoding = Byte3Char
	} else if character>>3 == 0b11110 { // 4 byte char
		encoding = Byte4Char
	} else {
		// Error: Unsupported/Invalid character encoding
		return Token{PrimaryType: ErrorToken, SecondaryType: UnknownChar, Buff: nil, Line: lexer.Line, Column: lexer.Column}
	}

	num = int(character)
	len := int(encoding) - int(Byte1Char) + 1

	for i := 1; i < len; i++ {
		if chr, ok := lexer.peek(); !ok {
			// Error: expected char, got eof
			return Token{PrimaryType: ErrorToken, SecondaryType: UnexpectedEOF, Buff: nil, Line: lexer.Line, Column: lexer.Column}
		} else if chr>>6 == 0b10 {
			lexer.eatLastByte()
			num += int(math.Pow(float64(chr), 16*float64(i)))
		} else {
			// Error: Invalid char
			return Token{PrimaryType: ErrorToken, SecondaryType: SecondaryNullType, Buff: nil, Line: lexer.Line, Column: lexer.Column}
		}
	}

	if nextChar, ok := lexer.peek(); !ok {
		// Error: Expected char demlimiter, got eof
		return Token{PrimaryType: ErrorToken, SecondaryType: UnexpectedEOF, Buff: nil, Line: lexer.Line, Column: lexer.Column}
	} else if nextChar == '\'' {
		lexer.eatLastByte() // increament the positon as `nextChar` was `'` as expected
		return Token{PrimaryType: CharLiteral, SecondaryType: encoding, Buff: []byte(strconv.Itoa(num)), Line: line, Column: column}
	}

	// Error: Invalid character {nedxtChar} at {Lexer.line}:{Lexer.column}. Expected `'`.
	// print error and all (maybe use the Buff property for storing error message?)
	return Token{PrimaryType: ErrorToken, SecondaryType: SecondaryNullType, Buff: nil, Line: lexer.Line, Column: lexer.Column}
}

func (lexer *Lexer) lexString() Token {
	str := []byte{'"'}

	line := lexer.Line
	column := lexer.Column
	size := 1

	for character, ok := lexer.peek(); !IsStringDelimiter(character); character, ok = lexer.peek() {

		if !ok {
			// Error: Expected end of string literal, got eof
			error.New("expected \", got eof.", lexer.Line, lexer.Column)
		} else if character == '\n' {
			// Error: Expected end of string literal, got end of line
			error.New("expected \", got end of line.", lexer.Line, lexer.Column)
		}

		str = append(str, character)
		lexer.eatLastByte()

		if character == '\\' {
			next, ok := lexer.peek()

			if !ok {
				// Error: Expected end of string literal, got eof
				return Token{PrimaryType: ErrorToken, SecondaryType: UnexpectedEOF, Buff: nil, Line: lexer.Line, Column: lexer.Column}
			} else if character == '\n' {
				// Error: Expected end of string literal, got end of line
				return Token{PrimaryType: ErrorToken, SecondaryType: UnexpectedEOF, Buff: nil, Line: lexer.Line, Column: lexer.Column}
			}

			if chr, _ := lexer.peek(); chr == 'u' {
				size += 3
			} else if chr == 'U' {
				size += 7
			}

			str = append(str, next)
			character = next
			lexer.eatLastByte()
		}
		size++
	}

	lexer.eatLastByte() // eat '"'
	str = append(str, '"')
	return Token{PrimaryType: StringLiteral, SecondaryType: SecondaryNullType, Flags: size, Buff: str, Line: line, Column: column}
}

func (lexer *Lexer) lexWord() Token {
	word := []byte{}

	column := lexer.Column
	line := lexer.Line

	character, _ := lexer.nextChar()
	word = append(word, character)

	for w, _ := lexer.peek(); IsIdentifierPart(w); w, _ = lexer.peek() {
		word = append(word, w)
		lexer.eatLastByte()
	}

	return Token{PrimaryType: GetWordType(string(word)), SecondaryType: SecondaryNullType, Buff: word, Line: line, Column: column}
}

func (lexer *Lexer) lexOperator() Token {
	line := lexer.Line
	column := lexer.Column

	switch character, _ := lexer.peek(); character {
	case '*':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '=':
			lexer.eatLastByte()
			return Token{PrimaryType: AssignmentOperator, SecondaryType: MulEqual, Buff: []byte("*="), Line: line, Column: column}
		default:
			return Token{PrimaryType: AirthmaticOperator, SecondaryType: Mul, Buff: []byte("*"), Line: line, Column: column}
		}
	case '/':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '=':
			lexer.eatLastByte()
			return Token{PrimaryType: AssignmentOperator, SecondaryType: DivEqual, Buff: []byte("/="), Line: line, Column: column}
		default:
			return Token{PrimaryType: AirthmaticOperator, SecondaryType: Div, Buff: []byte("/"), Line: line, Column: column}
		}
	case '%':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '=':
			lexer.eatLastByte()
			return Token{PrimaryType: AssignmentOperator, SecondaryType: ModulusEqual, Buff: []byte("%="), Line: line, Column: column}
		default:
			return Token{PrimaryType: AirthmaticOperator, SecondaryType: Modulus, Buff: []byte("%"), Line: line, Column: column}
		}
	case '+':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '=':
			lexer.eatLastByte()
			return Token{PrimaryType: AssignmentOperator, SecondaryType: AddEqual, Buff: []byte("+="), Line: line, Column: column}
		case '+':
			lexer.eatLastByte()
			return Token{PrimaryType: AssignmentOperator, SecondaryType: AddAdd, Buff: []byte("++"), Line: line, Column: column}
		default:
			return Token{PrimaryType: AirthmaticOperator, SecondaryType: Add, Buff: []byte("+"), Line: line, Column: column}
		}
	case '-':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '=':
			lexer.eatLastByte()
			return Token{PrimaryType: AssignmentOperator, SecondaryType: SubEqual, Buff: []byte("-="), Line: line, Column: column}
		case '-':
			lexer.eatLastByte()
			return Token{PrimaryType: AssignmentOperator, SecondaryType: SubSub, Buff: []byte("--"), Line: line, Column: column}
		default:
			return Token{PrimaryType: AirthmaticOperator, SecondaryType: Sub, Buff: []byte("-"), Line: line, Column: column}
		}
	case '=':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '=':
			lexer.eatLastByte()
			return Token{PrimaryType: RelationalOperator, SecondaryType: EqualEqual, Buff: []byte("=="), Line: line, Column: column}
		default:
			return Token{PrimaryType: AssignmentOperator, SecondaryType: Equal, Buff: []byte("="), Line: line, Column: column}
		}
	case '!':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '=':
			lexer.eatLastByte()
			return Token{PrimaryType: RelationalOperator, SecondaryType: NotEqual, Buff: []byte("!="), Line: line, Column: column}
		default:
			return Token{PrimaryType: LogicalOperator, SecondaryType: Not, Buff: []byte("!"), Line: line, Column: column}
		}
	case '>':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '=':
			lexer.eatLastByte()
			return Token{PrimaryType: RelationalOperator, SecondaryType: GreaterEqual, Buff: []byte(">="), Line: line, Column: column}
		case '>':
			lexer.eatLastByte()
			return Token{PrimaryType: BitwiseOperator, SecondaryType: RightShift, Buff: []byte(">>"), Line: line, Column: column}
		default:
			return Token{PrimaryType: RelationalOperator, SecondaryType: Greater, Buff: []byte(">"), Line: line, Column: column}
		}
	case '<':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '=':
			lexer.eatLastByte()
			return Token{PrimaryType: RelationalOperator, SecondaryType: LessEqual, Buff: []byte("<="), Line: line, Column: column}
		case '<':
			lexer.eatLastByte()
			return Token{PrimaryType: BitwiseOperator, SecondaryType: LeftShift, Buff: []byte("<<"), Line: line, Column: column}
		default:
			return Token{PrimaryType: RelationalOperator, SecondaryType: Less, Buff: []byte("<"), Line: line, Column: column}
		}
	case '&':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '&':
			lexer.eatLastByte()
			return Token{PrimaryType: LogicalOperator, SecondaryType: AndAnd, Buff: []byte("&&"), Line: line, Column: column}
		default:
			return Token{PrimaryType: BitwiseOperator, SecondaryType: And, Buff: []byte("&"), Line: line, Column: column}
		}
	case '|':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '|':
			lexer.eatLastByte()
			return Token{PrimaryType: LogicalOperator, SecondaryType: OrOr, Buff: []byte("||"), Line: line, Column: column}
		default:
			return Token{PrimaryType: BitwiseOperator, SecondaryType: Or, Buff: []byte("|"), Line: line, Column: column}
		}
	case '.':
		lexer.eatLastByte()
		switch next, _ := lexer.peek(); next {
		case '.':
			lexer.eatLastByte()
			return Token{PrimaryType: SpecialOperator, SecondaryType: DotDot, Buff: []byte(".."), Line: line, Column: column}
		default:
			return Token{PrimaryType: SpecialOperator, SecondaryType: Dot, Buff: []byte("."), Line: line, Column: column}
		}
	case '^':
		lexer.eatLastByte()
		return Token{PrimaryType: BitwiseOperator, SecondaryType: ExclusiveOr, Buff: []byte("^"), Line: line, Column: column}
	case '~':
		lexer.eatLastByte()
		return Token{PrimaryType: BitwiseOperator, SecondaryType: BitwiseNot, Buff: []byte("~"), Line: line, Column: column}
	case ':':
		lexer.eatLastByte()
		return Token{PrimaryType: SpecialOperator, SecondaryType: Colon, Buff: []byte(":"), Line: line, Column: column}
	case '?':
		lexer.eatLastByte()
		return Token{PrimaryType: SpecialOperator, SecondaryType: QuesMark, Buff: []byte("?"), Line: line, Column: column}
	}

	return Token{PrimaryType: ErrorToken, SecondaryType: NotFound, Buff: nil, Line: line, Column: column}
}

func (lexer *Lexer) lexDelimiter() Token {
	line := lexer.Line
	column := lexer.Column

	switch character, _ := lexer.peek(); character {
	case '(':
		lexer.eatLastByte()
		return Token{Buff: []byte("("), PrimaryType: LeftParen, SecondaryType: SecondaryNullType, Line: line, Column: column}
	case ')':
		lexer.eatLastByte()
		return Token{Buff: []byte(")"), PrimaryType: RightParen, SecondaryType: SecondaryNullType, Line: line, Column: column}
	case '{':
		lexer.eatLastByte()
		return Token{Buff: []byte("{"), PrimaryType: LeftCurlyBrace, SecondaryType: SecondaryNullType, Line: line, Column: column}
	case '}':
		lexer.eatLastByte()
		return Token{Buff: []byte("}"), PrimaryType: RightCurlyBrace, SecondaryType: SecondaryNullType, Line: line, Column: column}
	case '[':
		lexer.eatLastByte()
		return Token{Buff: []byte("["), PrimaryType: LeftBrace, SecondaryType: SecondaryNullType, Line: line, Column: column}
	case ']':
		lexer.eatLastByte()
		return Token{Buff: []byte("]"), PrimaryType: RightBrace, SecondaryType: SecondaryNullType, Line: line, Column: column}
	case ';':
		lexer.eatLastByte()
		return Token{Buff: []byte(";"), PrimaryType: SemiColon, SecondaryType: SecondaryNullType, Line: line, Column: column}
	case ',':
		lexer.eatLastByte()
		return Token{Buff: []byte(","), PrimaryType: Comma, SecondaryType: SecondaryNullType, Line: line, Column: column}
	}

	return Token{PrimaryType: ErrorToken, SecondaryType: NotFound, Buff: nil, Line: line, Column: column}
}
