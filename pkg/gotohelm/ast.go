package gotohelm

import (
	"fmt"
	"io"
	"strconv"
)

type Node interface {
	Write(io.Writer)
}

type Until struct {
	Expr Node
}

func (u *Until) Write(w io.Writer) {
	w.Write([]byte("until ("))
	u.Expr.Write(w)
	w.Write([]byte("|int)"))
}

type UntilStep struct {
	Start Node
	Stop  Node
	Step  Node
}

func (u *UntilStep) Write(w io.Writer) {
	w.Write([]byte("untilStep ("))
	u.Start.Write(w)
	w.Write([]byte("|int)"))
	w.Write([]byte(" ("))
	u.Stop.Write(w)
	w.Write([]byte("|int)"))
	w.Write([]byte(" ("))
	u.Step.Write(w)
	w.Write([]byte("|int)"))
}

type ParenExpr struct {
	Expr Node
}

func (s *ParenExpr) Write(w io.Writer) {
	w.Write([]byte("("))
	s.Expr.Write(w)
	w.Write([]byte(")"))
}

type Selector struct {
	Expr  Node
	Field string
	// Inlined indicates if `Field` is a JSON inlined (embedded) field or not.
	Inlined bool
}

func (s *Selector) Write(w io.Writer) {
	s.Expr.Write(w)
	// If this Selector is referencing an inlined field, don't emit it as
	// gotohelm's "object model" is the JSON representation of structs, not
	// go's representation.
	if !s.Inlined {
		fmt.Fprintf(w, ".%s", s.Field)
	}
}

type Nil struct{}

func (*Nil) Write(w io.Writer) {
	// nil is strange for some reason, in many cases it's acceptable to just
	// have `nil` but in others, you'll get `nil is not a command` errors.
	// {{ $_ := nil }} Doesn't work
	// {{ $_ := (eq nil nil) }} Works
	// It's too difficult to inspect all the cases and use nil in some but not
	// others, instead wrap nil in a function that just returns nil.
	// (fromJSON "null") doesn't work quite as expected but coalesce seems to
	// do the trick.
	w.Write([]byte(`(coalesce nil)`))
}

type Statement struct {
	NoCapture bool
	Expr      Node
}

func (s *Statement) Write(w io.Writer) {
	fmt.Fprintf(w, "{{- ")
	if !s.NoCapture {
		fmt.Fprintf(w, "$_ := ")
	}
	s.Expr.Write(w)
	fmt.Fprintf(w, " -}}\n")
}

type Binary struct {
	LHS Node
	Op  string
	RHS Node
}

func (b *Binary) Write(w io.Writer) {
	b.LHS.Write(w)
	fmt.Fprintf(w, " %s ", b.Op)
	b.RHS.Write(w)
}

type Ident struct {
	Name string
}

func (i *Ident) Write(w io.Writer) {
	fmt.Fprintf(w, "$%s", i.Name)
}

type BuiltInCall struct {
	FuncName  string
	Arguments []Node
}

func (c *BuiltInCall) Write(w io.Writer) {
	fmt.Fprintf(w, "(%s ", c.FuncName)
	for i, arg := range c.Arguments {
		if i > 0 {
			fmt.Fprintf(w, " ")
		}
		arg.Write(w)
	}
	fmt.Fprintf(w, ")")
}

type Cast struct {
	To string
	X  Node
}

func (c *Cast) Write(w io.Writer) {
	fmt.Fprintf(w, "(")
	c.X.Write(w)
	fmt.Fprintf(w, " | %s)", c.To)
}

type Call struct {
	FuncName  string
	Arguments []Node
}

func (c *Call) Write(w io.Writer) {
	args := &DictLiteral{
		KeysValues: []*KeyValue{
			{
				Key: `"a"`,
				Value: &BuiltInCall{
					FuncName:  "list",
					Arguments: c.Arguments,
				},
			},
		},
	}

	fmt.Fprintf(w, `(get (fromJson (include %q `, c.FuncName)
	args.Write(w)
	fmt.Fprintf(w, `)) %q)`, "r")
}

type Assignment struct {
	LHS Node
	New bool
	RHS Node
}

func (a *Assignment) Write(w io.Writer) {
	fmt.Fprintf(w, "{{- ")
	a.LHS.Write(w)
	fmt.Fprintf(w, " ")
	if a.New {
		fmt.Fprintf(w, ":")
	}
	fmt.Fprintf(w, "= ")
	a.RHS.Write(w)
	fmt.Fprintf(w, " -}}\n")
}

type DictLiteral struct {
	KeysValues []*KeyValue
}

func (d *DictLiteral) Write(w io.Writer) {
	fmt.Fprintf(w, "(dict ")
	for _, p := range d.KeysValues {
		p.Write(w)
		fmt.Fprintf(w, " ")
	}
	fmt.Fprintf(w, ")")
}

type KeyValue struct {
	Key   string
	Value Node
}

func (p *KeyValue) Write(w io.Writer) {
	fmt.Fprintf(w, "%s ", p.Key)
	p.Value.Write(w)
}

type File struct {
	Source string
	Name   string
	Header string
	Funcs  []*Func
}

func (f *File) Write(w io.Writer) {
	if f.Source != "" {
		fmt.Fprintf(w, "{{- /* Generated from %q */ -}}\n\n", f.Source)
	}
	w.Write([]byte(f.Header))
	for _, s := range f.Funcs {
		s.Write(w)
		w.Write([]byte{'\n'})
	}
}

type Func struct {
	Namespace  string
	Name       string
	Params     []Node
	Statements []Node
}

func (f *Func) Write(w io.Writer) {
	fmt.Fprintf(w, "{{- define %q -}}\n", f.Namespace+"."+f.Name)
	for i := range f.Params {
		fmt.Fprintf(w, "{{- ")
		f.Params[i].Write(w)
		fmt.Fprintf(w, " := (index .a %d) -}}\n", i)
	}
	fmt.Fprintf(w, "{{- range $_ := (list 1) -}}\n")
	for _, s := range f.Statements {
		s.Write(w)
	}
	fmt.Fprintf(w, "{{- end -}}\n")
	fmt.Fprintf(w, "{{- end -}}\n")
}

type Return struct {
	Expr Node
}

func (r *Return) Write(w io.Writer) {
	fmt.Fprintf(w, "{{- (dict %q ", "r")
	r.Expr.Write(w)
	fmt.Fprintf(w, ") | toJson -}}\n")
	fmt.Fprintf(w, "{{- break -}}\n")
}

type Literal struct {
	Value string
}

func NewLiteral(unquoted string) *Literal {
	return &Literal{Value: strconv.Quote(unquoted)}
}

func (l *Literal) Write(w io.Writer) {
	fmt.Fprintf(w, "%s", l.Value)
}

type Block struct {
	Statements []Node
}

func (b *Block) Write(w io.Writer) {
	for _, s := range b.Statements {
		s.Write(w)
	}
}

type Range struct {
	Key   Node
	Value Node
	Over  Node
	Body  Node
}

func (r *Range) Write(w io.Writer) {
	fmt.Fprintf(w, "{{- range ")
	if r.Key != nil {
		r.Key.Write(w)
	} else {
		w.Write([]byte("$_"))
	}
	fmt.Fprintf(w, ", ")
	if r.Value != nil {
		r.Value.Write(w)
	} else {
		w.Write([]byte("$_"))
	}
	fmt.Fprintf(w, " := ")
	r.Over.Write(w)
	fmt.Fprintf(w, " -}}\n")
	r.Body.Write(w)
	fmt.Fprintf(w, "{{- end -}}\n")
}

type IfStmt struct {
	Init Node
	Cond Node
	Body Node
	Else Node
}

func (i *IfStmt) Write(w io.Writer) {
	if i.Init != nil {
		i.Init.Write(w)
	}

	fmt.Fprintf(w, "{{- if ")
	i.Cond.Write(w)
	fmt.Fprintf(w, " -}}\n")

	if i.Body != nil {
		i.Body.Write(w)
	}

	if i.Else != nil {
		fmt.Fprintf(w, "{{- else -}}")
		if _, ok := i.Else.(*IfStmt); !ok {
			fmt.Fprintf(w, "\n")
		}
		i.Else.Write(w)
	}

	fmt.Fprintf(w, "{{- end -}}\n")
}
