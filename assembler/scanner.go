package assembler

import (
	"errors"
	"unicode"
)

// not a token
type NaT struct{}

func (nat *NaT) Error() string {
	return "invalid token"
}

// Maps strings to TokenTypes.
var tokenMap = map[string]TokenType{
	"halt":   halt,
	"nop":    nop,
	"rrmovq": rrmovq,
	"irmovq": irmovq,
	"rmmovq": rrmovq,
	"mrmovq": mrmovq,
	"addq":   addq,
	"subq":   subq,
	"andq":   andq,
	"xorq":   xorq,
	"mulq":   mulq,
	"divq":   divq,
	"modq":   modq,
	"jmp":    jmp,
	"jle":    jle,
	"jl":     jl,
	"je":     je,
	"jne":    jne,
	"jge":    jge,
	"jg":     jg,
	"call":   call,
	"ret":    ret,
	"pushq":  pushq,
	"popq":   popq,
}

var regSet = map[string]bool{
	"%rax": true,
	"%rcx": true,
	"%rdx": true,
	"%rbx": true,
	"%rsp": true,
	"%rbp": true,
	"%rsi": true,
	"%rdi": true,
	"%r10": true,
	"%r11": true,
	"%r12": true,
	"%r13": true,
	"%r14": true,
	"%r15": true,
}

// Scans a source string and generates a list of tokens.
type Scanner struct {
	src    string  // the source code
	curr   int     // points at the current unprocessed character
	start  int     // the start of the sliding window
	line   uint    // the current line
	col    uint    // the current col
	tokens []Token // a list of tokens
}

// Create a new scanner and set its source string.
func NewScanner(src string) *Scanner {
	return &Scanner{
		src,
		0,
		0,
		1,
		1,
		[]Token{},
	}
}

// Scans the source file and returns a list of tokens.
func (s *Scanner) Scan() ([]Token, error) {
	for !s.isAtEnd() {
		err := s.next()
		if err != nil {
			return nil, err
		}
	}
	s.addTokenLiteral(eof, "")
	return s.tokens, nil
}

// Return the current character and advance the scanner.
func (s *Scanner) advance() rune {
	r := s.src[s.curr]
	s.curr++
	s.col++
	return rune(r)
}

// Return the current character that the scanner is pointing at without advancing it.
func (s *Scanner) peek() rune {
	return rune(s.src[s.curr])
}

// Returns true if the scanner is at the end of the file and false if it is not
func (s *Scanner) isAtEnd() bool {
	return s.curr >= len(s.src)
}

// Add a token literal to the token list.
func (s *Scanner) addTokenLiteral(tokenType TokenType, literal string) {
	s.tokens = append(s.tokens, NewToken(tokenType, literal, s.line, s.col))
}

// Add a token to the token list.
func (s *Scanner) addToken(tokenType TokenType) {
	lex := s.src[s.start:s.curr]
	s.tokens = append(s.tokens, NewToken(tokenType, lex, s.line, s.col))
}

// Match a sequence of numbers in the source string. Returns an invalid token error if
// the scanner encounters a non-numerical character.
func (s *Scanner) matchNumber(r rune) bool {
	for !s.isAtEnd() && unicode.IsNumber(r) {
		r = s.advance()
	}

	if !s.isAtEnd() {
		s.curr--
	}

	if r != ',' && !unicode.IsSpace(r) && !unicode.IsNumber(r) {
		return false
	}
	s.addToken(number)
	return true
}

// Match a sequence of alphanumeric characters in the source string.
func (s *Scanner) matchIdentifier(r rune) {
	for !s.isAtEnd() && r != ':' && !unicode.IsSpace(r) {
		r = s.advance()
	}

	if !s.isAtEnd() || r != ':' {
		s.curr--
	}

	lex := s.src[s.start:s.curr]
	keyword, ok := tokenMap[lex]
	if ok {
		s.addToken(keyword)
	} else {
		s.addToken(identifier)
	}
}

// Match a register in the source string.
func (s *Scanner) matchReg() bool {
	r := s.advance()
	switch r {
	case '8':
		s.addTokenLiteral(reg, "%r8")
	case '9':
		s.addTokenLiteral(reg, "%r9")
	case '1':
		r = s.advance()
		switch r {
		case '0':
			s.addTokenLiteral(reg, "%r10")
		case '1':
			s.addTokenLiteral(reg, "%r11")
		case '2':
			s.addTokenLiteral(reg, "%r12")
		case '3':
			s.addTokenLiteral(reg, "%r13")
		case '4':
			s.addTokenLiteral(reg, "%r14")
		case '5':
			s.addTokenLiteral(reg, "%r15")
		default:
			return false
		}
	default:
		return false
	}
	return true
}

// Return the next token from the source file.
func (s *Scanner) next() error {
	s.start = s.curr
	r := s.advance()

	switch {
	case r == '\n':
		s.line++
		s.col = 1
	case r == '(':
		s.addTokenLiteral(lparen, "(")
	case r == ')':
		s.addTokenLiteral(lparen, ")")
	case r == ':':
		s.addTokenLiteral(lparen, ":")
	case r == ',':
		s.addTokenLiteral(lparen, ",")
	case r == '.':
		s.addTokenLiteral(dot, ".")
	case r == '0':
		if r = s.peek(); r == 'x' {
			s.advance()
			s.matchNumber(r)
		} else {
			s.matchIdentifier(r)
		}
	case r == '%':
		if r = s.advance(); r == 'r' {
			if !s.matchReg() {
				return errors.New("invalid token")
			}
		} else {
			s.matchIdentifier(r)
		}
	case unicode.IsNumber(r):
		s.matchNumber(r)
	case unicode.IsLetter(r) || unicode.IsSymbol(r):
		s.matchIdentifier(r)
	}

	return nil
}
