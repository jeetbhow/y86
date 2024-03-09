package model

type TokenType uint8

const (
	instruction TokenType = iota
	reg
	lparen
	rparen
	pos
	quad
	num
	label
	dir
	colon
	comma
	eof
)

// A lexical unit in the y86 assembly language.
type Token struct {
	tokenType TokenType
	lex       string
	line      uint
	col       uint
}

// Create a new token
func NewToken(tokType TokenType, lex string, line uint, col uint) Token {
	return Token{
		tokType,
		lex,
		line,
		col,
	}
}

func (t Token) String() string {
	return t.lex
}
