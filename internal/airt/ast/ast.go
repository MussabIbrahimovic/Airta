package ast

type Node interface{ node() }

type Program struct{ Body []Stmt }

func (*Program) node() {}

type Stmt interface {
	Node
	stmt()
}
type Expr interface {
	Node
	expr()
}

type ImportStmt struct{ Path, Alias string }

func (*ImportStmt) node() {}
func (*ImportStmt) stmt() {}

type LetStmt struct {
	Name  string
	Value Expr
}

func (*LetStmt) node() {}
func (*LetStmt) stmt() {}

type FuncStmt struct {
	Name      string
	Params    []string
	Body      []Stmt
	Async     bool
	Component bool
}

func (*FuncStmt) node() {}
func (*FuncStmt) stmt() {}

type IfStmt struct {
	Test       Expr
	Then, Else []Stmt
}

func (*IfStmt) node() {}
func (*IfStmt) stmt() {}

type WhileStmt struct {
	Test Expr
	Body []Stmt
}

func (*WhileStmt) node() {}
func (*WhileStmt) stmt() {}

type ReturnStmt struct{ Value Expr }

func (*ReturnStmt) node() {}
func (*ReturnStmt) stmt() {}

type ThrowStmt struct{ Value Expr }

func (*ThrowStmt) node() {}
func (*ThrowStmt) stmt() {}

type TryCatchStmt struct {
	Try   []Stmt
	Name  string
	Catch []Stmt
}

func (*TryCatchStmt) node() {}
func (*TryCatchStmt) stmt() {}

type ExprStmt struct{ Value Expr }

func (*ExprStmt) node() {}
func (*ExprStmt) stmt() {}

type Ident struct{ Name string }

func (*Ident) node() {}
func (*Ident) expr() {}

type Literal struct{ Value any }

func (*Literal) node() {}
func (*Literal) expr() {}

type ArrayLit struct{ Items []Expr }

func (*ArrayLit) node() {}
func (*ArrayLit) expr() {}

type ObjectField struct {
	Key   string
	Value Expr
}
type ObjectLit struct{ Fields []ObjectField }

func (*ObjectLit) node() {}
func (*ObjectLit) expr() {}

type UnaryExpr struct {
	Op string
	X  Expr
}

func (*UnaryExpr) node() {}
func (*UnaryExpr) expr() {}

type BinaryExpr struct {
	Op   string
	L, R Expr
}

func (*BinaryExpr) node() {}
func (*BinaryExpr) expr() {}

type CallExpr struct {
	Callee Expr
	Args   []Expr
}

func (*CallExpr) node() {}
func (*CallExpr) expr() {}

type MemberExpr struct {
	Obj  Expr
	Prop string
}

func (*MemberExpr) node() {}
func (*MemberExpr) expr() {}

type AssignExpr struct{ Left, Right Expr }

func (*AssignExpr) node() {}
func (*AssignExpr) expr() {}

type AwaitExpr struct{ X Expr }

func (*AwaitExpr) node() {}
func (*AwaitExpr) expr() {}
