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

type LLHead struct {
	First, Last *LLNode
}

type LLNode struct {
	Next *LLNode
	El *Node
}

func (h *LLHead) Append(el *Node) {
	n := &LLNode{
		Next: nil,
		El: el,
	}
	if h.First == nil {
		h.First = n
	} else {
		h.Last.Next = n
	}
	h.Last = n
}

func (h LLHead) String() string {
	s := "Form("
	if h.First == nil {
		s += "nil"
	} else {
		s += h.First.String()
	}
	s += ")"
	return s
}

func (n LLNode) String() (s string) {
	if n.El == nil {
		s = "nil"
	} else {
		s = n.El.String()
		if n.Next != nil {
			s += ", "
			s += n.Next.String()
		}
	}
	return
}

type Node struct {
	Type FormType
	Atom Atom  // TypeAtom
	Text Text  // TypeText
	Form *LLHead // TypeForm
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

type Atom string

func (a Atom) String() string {
	return fmt.Sprintf("Atom(%s)", string(a))
}

type Text string

func (t Text) String() string {
	return fmt.Sprintf("Text(%s)", string(t))
}

func Lex(tokens []tok.Token) *LLHead {
	root := &LLHead{}
	root.Append(&Node{
		Type: TypeAtom,
		Atom: "root",
	})
	forms := []*LLHead{root}
	for _, t := range tokens {
		top := forms[len(forms)-1]
		switch t.Type {
		case tok.TypeFormStart:
			head := &LLHead{}
			form := &Node{
				Type: TypeForm,
				Form: head,
			}
			top.Append(form)
			forms = append(forms, head)
		case tok.TypeAtom:
			atom := &Node{
				Type: TypeAtom,
				Atom: Atom(t.Text),
			}
			top.Append(atom)
		case tok.TypeText:
			text := &Node{
				Type: TypeText,
				Text: Text(t.Text),
			}
			top.Append(text)
		case tok.TypeFormEnd:
			forms = forms[:len(forms)-1]
		default:
			panic("invalid token")
		}
	}
	return root
}
