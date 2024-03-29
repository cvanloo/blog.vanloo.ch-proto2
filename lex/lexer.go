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
	TypeForm FormType = 1 << 0
	TypeAtom FormType = 1 << 1
	TypeText FormType = 1 << 2
)

func (t FormType) String() (str string) {
	if (t & TypeForm) != 0 {
		str += "Form"
	}
	if (t & TypeAtom) != 0 {
		if len(str) > 0 {
			str += " | "
		}
		str += "Atom"
	}
	if (t & TypeText) != 0 {
		if len(str) > 0 {
			str += " | "
		}
		str += "Text"
	}
	if len(str) > 0 {
		return str
	}
	return "None"
}

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
	return h.StringIndent(0)
}

func tabs(n int) (s string) {
	for i := 0; i < n; i++ {
		s += "  "
	}
	return
}

func (h LLHead) StringIndent(level int) (s string) {
	s += tabs(level); s += "Form("
	if h.First == nil {
		s += tabs(level+1); s += "nil"
	} else {
		s += "\n"
		s += h.First.StringIndent(level+1)
		s += "\n"
	}
	s += tabs(level); s += ")"
	return s
}

func (n LLNode) String() (s string) {
	return n.StringIndent(0)
}

func (n LLNode) StringIndent(level int) (s string) {
	if n.El == nil {
		s += tabs(level); s += "nil"
	} else {
		s += n.El.StringIndent(level)
		if n.Next != nil {
			s += ",\n"
			s += n.Next.StringIndent(level)
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
	return n.StringIndent(0)
}

func (n Node) StringIndent(level int) string {
	switch n.Type {
	case TypeForm:
		return n.Form.StringIndent(level)
	case TypeAtom:
		return n.Atom.StringIndent(level)
	case TypeText:
		return n.Text.StringIndent(level)
	default:
		panic("invalid type")
	}
}

type Atom string

func (a Atom) String() string {
	return a.StringIndent(0)
}

func (a Atom) StringIndent(level int) string {
	return tabs(level) + fmt.Sprintf("Atom(%s)", tok.VisibleString(string(a)))
}

type Text string

func (t Text) String() string {
	return t.StringIndent(0)
}

func (t Text) StringIndent(level int) string {
	return tabs(level) + fmt.Sprintf("Text(%s)", tok.VisibleString(string(t)))
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
