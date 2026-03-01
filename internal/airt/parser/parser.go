package parser

import (
	"airt/internal/airt/ast"
	"airt/internal/airt/lexer"
	"fmt"
	"strconv"
)

type Parser struct {
	t []lexer.Token
	p int
}

func New(t []lexer.Token) *Parser      { return &Parser{t: t} }
func (p *Parser) cur() lexer.Token     { return p.t[p.p] }
func (p *Parser) adv() lexer.Token     { x := p.cur(); p.p++; return x }
func (p *Parser) check(tt string) bool { return p.cur().Type == tt }
func (p *Parser) match(tt string) bool {
	if p.check(tt) {
		p.adv()
		return true
	}
	return false
}
func (p *Parser) eat(tt, msg string) lexer.Token {
	if !p.check(tt) {
		panic(msg)
	}
	return p.adv()
}
func (p *Parser) nl() {
	for p.match("NEWLINE") {
	}
}

func (p *Parser) Program() *ast.Program {
	body := []ast.Stmt{}
	p.nl()
	for !p.check("EOF") {
		body = append(body, p.Stmt())
		p.nl()
	}
	return &ast.Program{Body: body}
}
func (p *Parser) Stmt() ast.Stmt {
	switch p.cur().Type {
	case "IMPORT":
		p.adv()
		path := p.eat("STRING", "expected import path").Lit
		alias := ""
		if p.match("AS") {
			alias = p.eat("IDENT", "expected alias").Lit
		}
		return &ast.ImportStmt{Path: path, Alias: alias}
	case "LET":
		p.adv()
		n := p.eat("IDENT", "expected name").Lit
		p.eat("=", "expected =")
		return &ast.LetStmt{Name: n, Value: p.Expr(0)}
	case "ASYNC":
		p.adv()
		return p.funcStmt(true, false)
	case "FUNC":
		return p.funcStmt(false, false)
	case "COMPONENT":
		p.adv()
		return p.componentStmt()
	case "IF":
		p.adv()
		test := p.Expr(0)
		p.nl()
		th := p.block([]string{"ELSE", "END"})
		el := []ast.Stmt{}
		if p.match("ELSE") {
			p.nl()
			el = p.block([]string{"END"})
		}
		p.eat("END", "expected end")
		return &ast.IfStmt{Test: test, Then: th, Else: el}
	case "WHILE":
		p.adv()
		test := p.Expr(0)
		p.nl()
		b := p.block([]string{"END"})
		p.eat("END", "expected end")
		return &ast.WhileStmt{Test: test, Body: b}
	case "RETURN":
		p.adv()
		if p.check("NEWLINE") || p.check("END") || p.check("EOF") {
			return &ast.ReturnStmt{}
		}
		return &ast.ReturnStmt{Value: p.Expr(0)}
	case "THROW":
		p.adv()
		return &ast.ThrowStmt{Value: p.Expr(0)}
	case "TRY":
		p.adv()
		p.nl()
		tb := p.block([]string{"CATCH"})
		p.eat("CATCH", "expected catch")
		n := p.eat("IDENT", "expected catch name").Lit
		p.nl()
		cb := p.block([]string{"END"})
		p.eat("END", "expected end")
		return &ast.TryCatchStmt{Try: tb, Name: n, Catch: cb}
	default:
		return &ast.ExprStmt{Value: p.Expr(0)}
	}
}
func (p *Parser) funcStmt(async, comp bool) ast.Stmt {
	p.eat("FUNC", "expected func")
	n := p.eat("IDENT", "expected name").Lit
	p.eat("(", "expected (")
	ps := []string{}
	for !p.check(")") {
		ps = append(ps, p.eat("IDENT", "expected param").Lit)
		if !p.match(",") {
			break
		}
	}
	p.eat(")", "expected )")
	p.nl()
	b := p.block([]string{"END"})
	p.eat("END", "expected end")
	return &ast.FuncStmt{Name: n, Params: ps, Body: b, Async: async, Component: comp}
}

func (p *Parser) componentStmt() ast.Stmt {
	n := p.eat("IDENT", "expected component name").Lit
	p.eat("(", "expected (")
	ps := []string{}
	for !p.check(")") {
		ps = append(ps, p.eat("IDENT", "expected param").Lit)
		if !p.match(",") {
			break
		}
	}
	p.eat(")", "expected )")
	p.nl()
	b := p.block([]string{"END"})
	p.eat("END", "expected end")
	return &ast.FuncStmt{Name: n, Params: ps, Body: b, Component: true}
}

func (p *Parser) block(stop []string) []ast.Stmt {
	b := []ast.Stmt{}
	p.nl()
	for !p.check("EOF") {
		s := false
		for _, x := range stop {
			if p.check(x) {
				s = true
			}
		}
		if s {
			break
		}
		b = append(b, p.Stmt())
		p.nl()
	}
	return b
}

func (p *Parser) Expr(pr int) ast.Expr {
	l := p.prefix()
	for prec(p.cur().Type) > pr {
		op := p.adv().Type
		r := p.Expr(prec(op))
		l = &ast.BinaryExpr{Op: op, L: l, R: r}
	}
	return l
}
func (p *Parser) prefix() ast.Expr {
	if p.match("NUMBER") {
		v, _ := strconv.ParseFloat(p.t[p.p-1].Lit, 64)
		return &ast.Literal{Value: v}
	}
	if p.match("STRING") {
		return &ast.Literal{Value: p.t[p.p-1].Lit}
	}
	if p.match("TRUE") {
		return &ast.Literal{Value: true}
	}
	if p.match("FALSE") {
		return &ast.Literal{Value: false}
	}
	if p.match("NULL") {
		return &ast.Literal{Value: nil}
	}
	if p.match("NOT") || p.match("-") {
		op := p.t[p.p-1].Type
		return &ast.UnaryExpr{Op: op, X: p.Expr(7)}
	}
	if p.match("AWAIT") {
		return &ast.AwaitExpr{X: p.Expr(7)}
	}
	if p.match("(") {
		x := p.Expr(0)
		p.eat(")", "expected )")
		return p.post(x)
	}
	if p.match("[") {
		a := []ast.Expr{}
		for !p.check("]") {
			a = append(a, p.Expr(0))
			if !p.match(",") {
				break
			}
		}
		p.eat("]", "expected ]")
		return &ast.ArrayLit{Items: a}
	}
	if p.match("{") {
		f := []ast.ObjectField{}
		for !p.check("}") {
			k := p.eat("IDENT", "expected key").Lit
			p.eat(":", "expected :")
			f = append(f, ast.ObjectField{Key: k, Value: p.Expr(0)})
			if !p.match(",") {
				break
			}
		}
		p.eat("}", "expected }")
		return &ast.ObjectLit{Fields: f}
	}
	if p.match("IDENT") {
		return p.post(&ast.Ident{Name: p.t[p.p-1].Lit})
	}
	panic(fmt.Sprintf("unexpected token %s", p.cur().Type))
}
func (p *Parser) post(x ast.Expr) ast.Expr {
	for {
		if p.match("(") {
			a := []ast.Expr{}
			for !p.check(")") {
				a = append(a, p.Expr(0))
				if !p.match(",") {
					break
				}
			}
			p.eat(")", "expected )")
			x = &ast.CallExpr{Callee: x, Args: a}
			continue
		}
		if p.match(".") {
			x = &ast.MemberExpr{Obj: x, Prop: p.eat("IDENT", "expected prop").Lit}
			continue
		}
		if p.match("=") {
			x = &ast.AssignExpr{Left: x, Right: p.Expr(0)}
		}
		break
	}
	return x
}
func prec(op string) int {
	m := map[string]int{"OR": 1, "AND": 2, "==": 3, "!=": 3, "<": 4, ">": 4, "<=": 4, ">=": 4, "+": 5, "-": 5, "*": 6, "/": 6, "%": 6}
	return m[op]
}
