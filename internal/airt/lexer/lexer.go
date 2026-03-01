package lexer

import "unicode"

type Token struct{ Type, Lit string }

type Lexer struct {
	src []rune
	pos int
}

var kw = map[string]string{"let": "LET", "func": "FUNC", "async": "ASYNC", "return": "RETURN", "if": "IF", "else": "ELSE", "while": "WHILE", "true": "TRUE", "false": "FALSE", "null": "NULL", "import": "IMPORT", "as": "AS", "throw": "THROW", "try": "TRY", "catch": "CATCH", "component": "COMPONENT", "await": "AWAIT", "and": "AND", "or": "OR", "not": "NOT", "end": "END"}

func New(s string) *Lexer  { return &Lexer{src: []rune(s)} }
func (l *Lexer) eof() bool { return l.pos >= len(l.src) }
func (l *Lexer) peek() rune {
	if l.eof() {
		return 0
	}
	return l.src[l.pos]
}
func (l *Lexer) next() rune { r := l.peek(); l.pos++; return r }

func (l *Lexer) Tokens() []Token {
	out := []Token{}
	for !l.eof() {
		c := l.peek()
		switch {
		case c == ' ' || c == '\t' || c == '\r':
			l.next()
		case c == '\n':
			l.next()
			out = append(out, Token{"NEWLINE", "\n"})
		case c == '#':
			for !l.eof() && l.peek() != '\n' {
				l.next()
			}
		case unicode.IsDigit(c):
			out = append(out, l.number())
		case unicode.IsLetter(c) || c == '_':
			out = append(out, l.ident())
		case c == '"' || c == '\'':
			out = append(out, l.str())
		default:
			out = append(out, l.sym())
		}
	}
	return append(out, Token{"EOF", ""})
}

func (l *Lexer) number() Token {
	s := l.pos
	for !l.eof() && (unicode.IsDigit(l.peek()) || l.peek() == '.') {
		l.next()
	}
	return Token{"NUMBER", string(l.src[s:l.pos])}
}
func (l *Lexer) ident() Token {
	s := l.pos
	for !l.eof() && (unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_') {
		l.next()
	}
	t := string(l.src[s:l.pos])
	if k, ok := kw[t]; ok {
		return Token{k, t}
	}
	return Token{"IDENT", t}
}
func (l *Lexer) str() Token {
	q := l.next()
	s := []rune{}
	for !l.eof() && l.peek() != q {
		s = append(s, l.next())
	}
	if !l.eof() {
		l.next()
	}
	return Token{"STRING", string(s)}
}
func (l *Lexer) sym() Token {
	two := ""
	if l.pos+1 < len(l.src) {
		two = string(l.src[l.pos : l.pos+2])
	}
	for _, p := range []string{"==", "!=", "<=", ">="} {
		if two == p {
			l.pos += 2
			return Token{p, p}
		}
	}
	c := string(l.next())
	return Token{c, c}
}
