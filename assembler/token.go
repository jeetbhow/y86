package assembler

type TokenType uint8

const (
	halt TokenType = iota
	nop
	rrmovq
	irmovq
	rmmovq
	mrmovq
	addq
	subq
	andq
	xorq
	mulq
	divq
	modq
	jmp
	jle
	jl
	je
	jne
	jge
	jg
	call
	ret
	pushq
	popq
	reg
	colon
	comma
	lparen
	rparen
	pos
	quad
	number
	identifier
	invalid
	dot
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
func NewToken(tokenType TokenType, lex string, line uint, col uint) Token {
	return Token{
		tokenType,
		lex,
		line,
		col,
	}
}

func (t Token) String() string {
	return t.lex
}
