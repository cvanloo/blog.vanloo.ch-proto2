package be

type Type int
const (
	TypeText
	TypeAtom
	TypeForm
)

type LL struct {
	Next *LL
	Type Type
	Text string
	Atom string
	Form *LL
}

type (
	Context struct {
		scope Scope
	}
	Scope map[string]beFun
	beFun struct {
		isMacro bool
		func([]Renderable) Renderable
	}
)

func (s Scope) Resolve(name string) (beFun, error) {
	if f, ok := s[name]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("function not in scope: %s", name)
}

func (c *Context) Eval2(l *LL) (Renderable, error) {
	switch l.Type {
	case TypeText:
		return Text(l.Text), nil
	case TypeAtom:
		fun, ok := c.scope.Resolve(l.Atom)
		if fun.isMacro {
			// @todo: pass unevaluated args
			return Apply2(fun, args)
		} else {
			args := []Renderable{}
			for arg := l.Next; arg != nil; arg = arg.Next {
				r, err := Eval2(arg)
				if err != nil {
					return nil, err
				}
				args = append(args, r)
			}
			return Apply2(fun, args)
		}
	case TypeForm:
		return c.Eval2(l.Form)
	}
}

func (c *Context) Apply2(fun beFun, args []Renderable) (Renderable, error) {
	c.scope.Push(Scope{})
	defer c.scope.Pop()
	return fun(arg)
}

func TestEval2() {
	// (author (name cvl) (email cvl@vanloo.ch))
	code1 := &LL{
		Type: TypeAtom,
		Atom: "author",
		Next: &LL{
			Type: TypeForm,
			Form: &LL{
				Type: TypeAtom,
				Atom: "name",
				Next: &LL{
					Type: TypeText,
					Text: cvl,
				},
			},
			Next: &LL{
				Type: TypeForm,
				Form: &LL{
					Type: TypeAtom,
					Atom: "email",
					Next: &LL{
						Type: TypeText,
						Text: "cvl@vanloo.ch"
					},
				},
			},
		},
	}

	// ^-- author is a macro?
	fmt.Printf("%+v\n", Eval2(code1))

	// (section "Section Heading" "Some content followed by" (sidenote "a sidenote" "Look at me, I'm a sidenote."))
	code2 := &LL{
		Type: TypeAtom,
		Atom: "section",
		Next: &LL{
			Type: TypeText,
			Text: "Section Heading",
			Next: &LL{
				Type: TypeText,
				Text: "Some content followed by",
				Next: &LL{
					Type: TypeForm,
					Form: &LL{
						Type: TypeAtom,
						Atom: "sidenote",
						Next: &LL{
							Type: TypeText,
							Text: "a sidenote.",
							Next: &LL{
								Type: TypeText,
								Text: "Look at me, I'm a sidenote.",
							}
						},
					},
				},
			},
		},
	}
	fmt.Printf("%+v\n", Eval2(code2))
}
