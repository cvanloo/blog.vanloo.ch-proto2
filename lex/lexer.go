package lex

import (
	"fmt"
	"be/tok"
)

// Input:
// (title Hello World)

// Tokenizer:
// TokenFormStart{"("}
// TokenAtom{"title"}
// TokenText{"Hello World"}
// TokenFormEnd{")"}

// Lexer:
// Form {
//   Atom {
//     "title"
//   },
//   Form {
//     Atom {
//       "text"
//     },
//     Text {
//       "Hello World"
//     }
//   }
// }

type FormType int
const (
	TypeForm FormType = iota
	TypeAtom
	TypeText
)

type Node struct {
	Type FormType
	Atom Atom
	Text Text
	Form *Form
}

func (n Node) String() string {
	switch n.Type {
	case TypeForm:
		return n.Form.String()
	case TypeAtom:
		return n.Atom.String()
	case TypeText:
		return n.Text.String()
	default:
		panic("invalid type")
	}
}

type Form struct {
	Next *Form
	Element *Node
}

const NilForm = Form{
	Next: NilForm,
	Element: NilForm,
}

func (f Form) String() string {
	s := "Form("
	if f.Element == nil {
		s += "nil"
	} else {
		s += f.Element.String()
	}
	for c := f.Next; c != nil; c = c.Next {
		s += ", "
		if c.Element == nil {
			s += "nil"
		} else {
			s += c.Element.String()
		}
	}
	s += ")"
	return s
}

type Atom string

func (a Atom) String() string {
	return fmt.Sprintf("Atom(%s)", string(a))
}

type Text string

func (t Text) String() string {
	return fmt.Sprintf("Text(%s)", string(t))
}

func Lex(tokens []tok.Token) *Node {
	root := &Form{
		Element: &Node{
			Type: TypeAtom,
			Atom: Atom("root"),
		},
	}
	rootNode := &Node{
		Type: TypeForm,
		Form: root,
	}
	forms := []*Form{}
	forms = append(forms, root)
	for i, l := 0, len(tokens); i < l; i++ {
		t := tokens[i]
		switch t.Type {
		case tok.TypeFormStart:
			form := &Node{
				Type: TypeForm,
				Form: &Form{},
			}
			top := forms[len(forms)-1]
			top.Next = &Form{
				Element: form,
			}
			forms = append(forms, form.Form)
		case tok.TypeAtom:
			atom := &Node{
				Type: TypeAtom,
				Atom: Atom(t.Text),
			}
			top := forms[len(forms)-1]
			top.Next = &Form{
				Element: atom,
			}
		case tok.TypeText:
			text := &Node{
				Type: TypeText,
				Text: Text(t.Text),
			}
			top := forms[len(forms)-1]
			top.Next = &Form{
				Element: text,
			}
		case tok.TypeFormEnd:
			forms = forms[:len(forms)-1]
		default:
			panic("invalid token")
		}
	}
	return rootNode
}
