package runtime

import (
	"airt/internal/airt/ast"
	"airt/internal/airt/lexer"
	"airt/internal/airt/parser"
	"airt/internal/airt/stdlib"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Env struct {
	p *Env
	v map[string]any
}

func NewEnv(p *Env) *Env             { return &Env{p: p, v: map[string]any{}} }
func (e *Env) Def(k string, val any) { e.v[k] = val }
func (e *Env) Get(k string) (any, bool) {
	if v, ok := e.v[k]; ok {
		return v, true
	}
	if e.p != nil {
		return e.p.Get(k)
	}
	return nil, false
}
func (e *Env) Set(k string, val any) bool {
	if _, ok := e.v[k]; ok {
		e.v[k] = val
		return true
	}
	if e.p != nil {
		return e.p.Set(k, val)
	}
	return false
}

type fn struct {
	n   *ast.FuncStmt
	c   *Ctx
	env *Env
}
type ret struct{ v any }

type Ctx struct {
	Root  *Env
	Args  []string
	cache map[string]map[string]any
}

func New(args []string) *Ctx {
	c := &Ctx{Root: NewEnv(nil), Args: args, cache: map[string]map[string]any{}}
	stdlib.Inject(func(k string, v any) { c.Root.Def(k, v) }, args)
	return c
}
func (c *Ctx) Parse(src string) (*ast.Program, error) {
	defer func() { recover() }()
	return parser.New(lexer.New(src).Tokens()).Program(), nil
}

func (c *Ctx) RunFile(file string) (map[string]any, error) { return c.module(file) }
func (c *Ctx) module(file string) (ex map[string]any, err error) {
	abs, _ := filepath.Abs(file)
	if m, ok := c.cache[abs]; ok {
		return m, nil
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("parse error: %v", r)
		}
	}()
	prg := parser.New(lexer.New(string(b)).Tokens()).Program()
	env := NewEnv(c.Root)
	ex = map[string]any{}
	env.Def("exports", ex)
	c.cache[abs] = ex
	if _, err = c.block(prg.Body, env, abs); err != nil {
		return nil, err
	}
	return ex, nil
}
func (c *Ctx) block(b []ast.Stmt, e *Env, file string) (any, error) {
	for _, s := range b {
		v, err := c.stmt(s, e, file)
		if err != nil {
			return nil, err
		}
		if _, ok := v.(ret); ok {
			return v, nil
		}
	}
	return nil, nil
}
func (c *Ctx) stmt(s ast.Stmt, e *Env, file string) (any, error) {
	switch n := s.(type) {
	case *ast.ImportStmt:
		p := filepath.Join(filepath.Dir(file), n.Path)
		m, er := c.module(p)
		if er != nil {
			return nil, er
		}
		if n.Alias != "" {
			e.Def(n.Alias, m)
		}
		return nil, nil
	case *ast.LetStmt:
		v, er := c.expr(n.Value, e, file)
		if er != nil {
			return nil, er
		}
		e.Def(n.Name, v)
		return nil, nil
	case *ast.FuncStmt:
		e.Def(n.Name, fn{n, c, e})
		return nil, nil
	case *ast.ExprStmt:
		return c.expr(n.Value, e, file)
	case *ast.IfStmt:
		t, er := c.expr(n.Test, e, file)
		if er != nil {
			return nil, er
		}
		if truthy(t) {
			return c.block(n.Then, NewEnv(e), file)
		}
		return c.block(n.Else, NewEnv(e), file)
	case *ast.WhileStmt:
		for {
			t, er := c.expr(n.Test, e, file)
			if er != nil {
				return nil, er
			}
			if !truthy(t) {
				break
			}
			v, er := c.block(n.Body, NewEnv(e), file)
			if er != nil {
				return nil, er
			}
			if _, ok := v.(ret); ok {
				return v, nil
			}
		}
		return nil, nil
	case *ast.ReturnStmt:
		if n.Value == nil {
			return ret{nil}, nil
		}
		v, er := c.expr(n.Value, e, file)
		return ret{v}, er
	case *ast.ThrowStmt:
		v, er := c.expr(n.Value, e, file)
		if er != nil {
			return nil, er
		}
		return nil, errors.New(fmt.Sprint(v))
	case *ast.TryCatchStmt:
		v, er := c.block(n.Try, NewEnv(e), file)
		if er == nil {
			return v, nil
		}
		ce := NewEnv(e)
		ce.Def(n.Name, er.Error())
		return c.block(n.Catch, ce, file)
	default:
		return nil, fmt.Errorf("unsupported stmt")
	}
}

func (c *Ctx) call(cal any, args []any, file string, e *Env) (any, error) {
	switch f := cal.(type) {
	case fn:
		sc := NewEnv(f.env)
		for i, p := range f.n.Params {
			if i < len(args) {
				sc.Def(p, args[i])
			} else {
				sc.Def(p, nil)
			}
		}
		v, er := f.c.block(f.n.Body, sc, file)
		if er != nil {
			return nil, er
		}
		if rr, ok := v.(ret); ok {
			return rr.v, nil
		}
		return nil, nil
	case func(...any) (any, error):
		return f(args...)
	default:
		return nil, fmt.Errorf("not callable")
	}
}

func (c *Ctx) expr(x ast.Expr, e *Env, file string) (any, error) {
	switch n := x.(type) {
	case *ast.Literal:
		return n.Value, nil
	case *ast.Ident:
		if v, ok := e.Get(n.Name); ok {
			return v, nil
		}
		return nil, fmt.Errorf("undefined %s", n.Name)
	case *ast.ArrayLit:
		a := []any{}
		for _, it := range n.Items {
			v, er := c.expr(it, e, file)
			if er != nil {
				return nil, er
			}
			a = append(a, v)
		}
		return a, nil
	case *ast.ObjectLit:
		m := map[string]any{}
		for _, f := range n.Fields {
			v, er := c.expr(f.Value, e, file)
			if er != nil {
				return nil, er
			}
			m[f.Key] = v
		}
		return m, nil
	case *ast.UnaryExpr:
		v, er := c.expr(n.X, e, file)
		if er != nil {
			return nil, er
		}
		if n.Op == "-" {
			return num(v) * -1, nil
		}
		return !truthy(v), nil
	case *ast.BinaryExpr:
		l, er := c.expr(n.L, e, file)
		if er != nil {
			return nil, er
		}
		if n.Op == "AND" {
			if truthy(l) {
				return c.expr(n.R, e, file)
			}
			return l, nil
		}
		if n.Op == "OR" {
			if truthy(l) {
				return l, nil
			}
			return c.expr(n.R, e, file)
		}
		r, er := c.expr(n.R, e, file)
		if er != nil {
			return nil, er
		}
		return bin(n.Op, l, r)
	case *ast.CallExpr:
		cal, er := c.expr(n.Callee, e, file)
		if er != nil {
			return nil, er
		}
		args := []any{}
		for _, a := range n.Args {
			v, er := c.expr(a, e, file)
			if er != nil {
				return nil, er
			}
			args = append(args, v)
		}
		return c.call(cal, args, file, e)
	case *ast.MemberExpr:
		o, er := c.expr(n.Obj, e, file)
		if er != nil {
			return nil, er
		}
		if m, ok := o.(map[string]any); ok {
			return m[n.Prop], nil
		}
		return nil, nil
	case *ast.AssignExpr:
		rv, er := c.expr(n.Right, e, file)
		if er != nil {
			return nil, er
		}
		switch l := n.Left.(type) {
		case *ast.Ident:
			if !e.Set(l.Name, rv) {
				return nil, fmt.Errorf("undefined %s", l.Name)
			}
			return rv, nil
		case *ast.MemberExpr:
			o, er := c.expr(l.Obj, e, file)
			if er != nil {
				return nil, er
			}
			if m, ok := o.(map[string]any); ok {
				m[l.Prop] = rv
				return rv, nil
			}
			return nil, fmt.Errorf("invalid assignment")
		}
	case *ast.AwaitExpr:
		return c.expr(n.X, e, file)
	}
	return nil, fmt.Errorf("unsupported expr")
}

func (c *Ctx) CallExport(file, name string, args ...any) error {
	m, err := c.RunFile(file)
	if err != nil {
		return err
	}
	v, ok := m[name]
	if !ok {
		return fmt.Errorf("missing export %s", name)
	}
	_, err = c.call(v, args, file, c.Root)
	return err
}
func truthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	case float64:
		return t != 0
	default:
		return true
	}
}
func num(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	default:
		return 0
	}
}
func bin(op string, l, r any) (any, error) {
	switch op {
	case "+":
		if ls, ok := l.(string); ok {
			return ls + fmt.Sprint(r), nil
		}
		if rs, ok := r.(string); ok {
			return fmt.Sprint(l) + rs, nil
		}
		return num(l) + num(r), nil
	case "-":
		return num(l) - num(r), nil
	case "*":
		return num(l) * num(r), nil
	case "/":
		return num(l) / num(r), nil
	case "%":
		return float64(int(num(l)) % int(num(r))), nil
	case "==":
		return fmt.Sprint(l) == fmt.Sprint(r), nil
	case "!=":
		return fmt.Sprint(l) != fmt.Sprint(r), nil
	case "<":
		return num(l) < num(r), nil
	case "<=":
		return num(l) <= num(r), nil
	case ">":
		return num(l) > num(r), nil
	case ">=":
		return num(l) >= num(r), nil
	}
	return nil, fmt.Errorf("unknown op %s", op)
}

func MustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

var _ = time.Now
