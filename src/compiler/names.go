package compiler

import (
	"bytes"
	. "parser"
	"strconv"
)

var num int = 0

type Namespace struct {
	Base     string
	BasePath string
	Num      int
}

func (n *Namespace) Init(nm int) {
	n.Base = strconv.Itoa(nm) + "_"
	n.Num = nm
}

func (n *Namespace) getEnumProp(enumName []byte, prop Token) Token {
	return Token{Buff: []byte("e" + n.Base + string(enumName) + "_" + string(prop.Buff)), PrimaryType: Identifier, SecondaryType: SecondaryNullType, Line: prop.Line, Column: prop.Column, Flags: 3}
}

func (n *Namespace) getNewVarName(token Token) Token {
	if token.Flags != 0 {
		return token
	}
	if isGlobal(token.Buff) {
		return token
	}
	if isInternal(token) {
		return Token{Buff: token.Buff[1:], PrimaryType: Identifier, SecondaryType: SecondaryNullType, Line: token.Line, Column: token.Column, Flags: 2}
	}
	return Token{Buff: []byte("v" + n.Base + string(token.Buff)), PrimaryType: Identifier, SecondaryType: SecondaryNullType, Line: token.Line, Column: token.Column, Flags: 1}
}

func isGlobal(buf []byte) bool {
	for _, buff := range []string{"u8", "u16", "u32", "u64", "i8", "i16", "132", "i64", "float", "double", "void", "true", "false"} {
		if bytes.Compare([]byte(buff), buf) == 0 {
			return true
		}
	}
	return false
}

func getLastImportPrefix() string {
	return "v" + strconv.Itoa(num-1) + "_"
}

func getPropName(prop Token) Token {
	return Token{Buff: []byte("p_" + string(prop.Buff)), PrimaryType: prop.PrimaryType, SecondaryType: prop.SecondaryType, Line: prop.Line, Column: prop.Column, Flags: 7}
}

func (n *Namespace) getVecPropName(prop Token) Token {
	newToken := Token{Line: prop.Line, Column: prop.Column, PrimaryType: Identifier, SecondaryType: SecondaryNullType, Flags: 10}

	switch string(prop.Buff) {
	case "push":
		newToken.Buff = []byte("VECTOR_PUSH")
	case "pop":
		newToken.Buff = []byte("VECTOR_POP")
	default:
		newToken.Buff = prop.Buff
	}
	return newToken
}

func (n *Namespace) getActualName(token Token) Token {
	if token.Flags == 0 {
		return token
	}
	newToken := Token{Line: token.Line, Column: token.Column, PrimaryType: token.PrimaryType, SecondaryType: token.SecondaryType, Flags: 0}
	newToken.Buff = make([]byte, len(token.Buff)+len(n.Base)+1)

	switch token.Flags {
	case 2:
		copy(newToken.Buff, []byte("$"+string(token.Buff)))
	case 7:
		copy(newToken.Buff, token.Buff[2:])
	case 8:
		copy(newToken.Buff, token.Buff[1:])
	default:
		copy(newToken.Buff, token.Buff[1+len(n.Base):])
	}
	return newToken
}

func (n *Namespace) getStrctDefaultName(strct Token) Token {
	if strct.Flags != 0 {
		return strct
	}
	return Token{Line: strct.Line, Column: strct.Column, PrimaryType: strct.PrimaryType, SecondaryType: strct.SecondaryType, Flags: 4, Buff: []byte("d" + n.Base + string(strct.Buff))}
}

func (n *Namespace) getStrctMethodName(name Token, strct Token) Token {
	return Token{Line: strct.Line, Column: strct.Column, PrimaryType: strct.PrimaryType, SecondaryType: strct.SecondaryType, Flags: 5, Buff: []byte("m" + n.Base + string(name.Buff) + "_" + string(strct.Buff))}
}

func (n *Namespace) joinName(prefix []byte, name Token) Token {
	return Token{Line: name.Line, Column: name.Column, PrimaryType: name.PrimaryType, SecondaryType: name.SecondaryType, Flags: 6, Buff: []byte(string(prefix) + string(name.Buff))}
}

func isInternal(tok Token) bool {
	buff := tok.Buff
	return len(buff) > 1 && bytes.Compare(buff[:1], []byte("$")) == 0
}

func (n *Namespace) getStrctDefaultNameFromPrefix(prefix string, strct Token) Token {
	if strct.Flags != 0 {
		return strct
	}
	return Token{Line: strct.Line, Column: strct.Column, PrimaryType: strct.PrimaryType, SecondaryType: strct.SecondaryType, Flags: 4, Buff: []byte("d" + prefix + string(strct.Buff))}
}

func (n *Namespace) getPropName(token Token) Token {
	if token.Flags != 0 {
		return token
	}
	return Token{Buff: []byte("p_" + string(token.Buff)), PrimaryType: Identifier, SecondaryType: SecondaryNullType, Line: token.Line, Column: token.Column, Flags: 8}
}

func (n *Namespace) getStrctMethodNameFromPrefix(prefix string, name Token, strct Token) Token {
	return Token{Line: strct.Line, Column: strct.Column, PrimaryType: strct.PrimaryType, SecondaryType: strct.SecondaryType, Flags: 5, Buff: []byte("m" + prefix + string(name.Buff) + "_" + string(strct.Buff))}
}

func (n *Namespace) getEnumPropFromPrefix(prefix string, enumName []byte, prop Token) Token {
	return Token{Buff: []byte("e" + prefix + string(enumName) + "_" + string(prop.Buff)), PrimaryType: Identifier, SecondaryType: SecondaryNullType, Line: prop.Line, Column: prop.Column, Flags: 3}
}
