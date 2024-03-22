// Package tok implements a tokenizer for the be markup language.
//
// Special symbols are
//   - '~'   (nbsp)
//   - '...' (ellipsis)
//   - '('   (start form)
//   - ')'   (end form)
// ...these need to be escaped if their literal interpretation is intended.
//
// Escape symbols are
//   - '\~'   (literal tilde)
//   - '\...' (literal three dots)
//   - '\('   (literal opening parenthesis)
//   - '\)'   (literal closing parenthesis)
//   - '\+'   (must come in pairs: starts (and ends) a raw text form)
//
// To separate multiple texts, the 't' or 'text' function is needed:
// ```
// <= This is all one text This still belongs to the same text
// => Text{0: `This is all one text This still belongs to the same text`}
// ```
//
// vs.
//
// ```
// <= (t This is all one text) This is a different text
// => Text{4: `This is all one text`}
// => Text{26: `This is a different text`}
// ```
//
// Text is also split by two newlines.
// Lines split by a single newline remain part of the same Text block, and are
// joined together (newline replaced by space).
// Multiple spaces are removed, so that only a single space remains.
// The only exception are raw strings (of the form '\+ ... \+').
package tok

import (
	"fmt"
	"log"
)

type TokenType int
const (
	TypeFormStart TokenType = iota
	TypeAtom
	TypeText
	TypeFormEnd
)

type (
	Token struct {
		Type TokenType
		Text string
		Pos int
	}

	tokFunc func() tokFunc

	Tokenizer struct {
		bs []rune
		l int
		pos int
		tokens []Token
		state tokFunc
		err error
	}

	TokenError struct {
		Msg string
		Pos int
		FileName string
	}
)

func NewTokenizer(bs []rune) *Tokenizer {
	return &Tokenizer{
		bs: bs,
		l: len(bs),
	}
}

func (t *Tokenizer) Tokenize() ([]Token, error) {
	t.state = t.tokTextOrForm // initial state [:init:]
	for t.state != nil {
		t.skipWhitespace()
		if t.pos >= t.l {
			t.state = t.tokEOF
		}
		t.state = t.state()
	}
	return t.tokens, t.err
}

func (t *Tokenizer) tokError(err error) tokFunc {
	t.err = err
	return nil
}

func (t *Tokenizer) tokTextOrForm() tokFunc { // initial state [:init:]
	if t.bs[t.pos] == '(' {
		return t.tokForm
	}
	return t.tokText
}

func (t *Tokenizer) tokText() tokFunc { // parse text
	// @fixme: text composition????
	var (
		textEnd = t.pos
		lastPos = textEnd
		quoted = false
		parsedText = ""
	)
	for textEnd < t.l && ((t.bs[textEnd] != ')' && t.bs[textEnd] != '(') || quoted) {
		if !quoted {
			if t.bs[textEnd] == ' ' { // merge excessive white space
				parsedText += string(t.bs[lastPos:textEnd])
				lastPos = textEnd + 1 // past space
				textEnd = lastPos

				for textEnd < t.l && t.bs[textEnd] == ' ' {
					textEnd++
				}
				lastPos = textEnd

				if textEnd < t.l && t.bs[textEnd] != '\n' && t.bs[textEnd] != '(' {
					parsedText += " "
					lastPos = textEnd
				}
			} else if t.bs[textEnd] == '\n' { // two newlines separate text blocks, lines divided by a single newline are joined
				if textEnd+1 < t.l {
					if t.bs[textEnd+1] == '\n' || t.bs[textEnd+1] == ')' {
						break // this text block is finished
						// @note: any further newlines are skipped in .Tokenize() by the call to .skipWhitespace()
					} else {
						// merge with next text block
						parsedText += string(t.bs[lastPos:textEnd])
						parsedText += " " // join with space
						lastPos = textEnd + 1 // past \n
						textEnd = lastPos
					}
				} else {
					parsedText += string(t.bs[lastPos:textEnd])
					lastPos = textEnd + 1
					textEnd = lastPos
				}
			} else if t.bs[textEnd] == '\\' {
				if textEnd+1 < t.l {
					esc := t.bs[textEnd+1]
					switch esc {
						case '(': fallthrough
						case ')': fallthrough
					case '\\':
						parsedText += string(t.bs[lastPos:textEnd])
						lastPos = textEnd + 1 // past backslash
						textEnd += 2          // past escaped char
					case '+':
						parsedText += string(t.bs[lastPos:textEnd])
						lastPos = textEnd + 2 // past escaped char
						textEnd += 2          // past escaped char
						quoted = !quoted
					default:
						return t.tokError(t.NewTokenError(fmt.Sprintf("invalid escape character: `%s`", string(esc))))
					}
				} else {
					return t.tokError(t.NewTokenError("unfinished escape character (did you mean `\\`?)"))
				}
			} else if t.bs[textEnd] == '~' {
				parsedText += string(t.bs[lastPos:textEnd])
				parsedText += "\u00A0" // no-break space
				lastPos = textEnd + 1  // past ~
				textEnd = lastPos
			} else if textEnd+2 < t.l && string(t.bs[textEnd:textEnd+3]) == "..." {
				parsedText += string(t.bs[lastPos:textEnd])
				parsedText += "\u2026" // horizontal ellipsis
				lastPos = textEnd + 3  // past ...
				textEnd = lastPos
			} else {
				textEnd++
			}
		} else {
			if t.bs[textEnd] == '\\' && textEnd+1 < t.l && t.bs[textEnd+1] == '+' {
				parsedText += string(t.bs[lastPos:textEnd])
				lastPos = textEnd + 2
				textEnd = lastPos
				quoted = false
			} else {
				textEnd++
			}
		}
	}
	parsedText += string(t.bs[lastPos:textEnd])
	t.tokens = append(t.tokens, Token{
		Type: TypeText,
		Text: parsedText,
		Pos: t.pos,
	})
	t.pos = textEnd

	return t.tokNilOrTextOrForm
}

func (t *Tokenizer) tokForm() tokFunc { // parse form start
	t.tokens = append(t.tokens, Token{
		Type: TypeFormStart,
		Text: "(",
		Pos: t.pos,
	})
	t.pos++

	return t.tokNilOrAtom
}

func (t *Tokenizer) tokNilOrAtom() tokFunc {
	r := t.bs[t.pos]
	if r == '(' {
		return t.tokError(t.NewTokenError("cannot start form / expected atom or nil"))
	}
	if r == ')' {
		return t.tokNil
	}
	if isAtomChar(r) {
		return t.tokAtom
	}
	return t.tokError(t.NewTokenError(fmt.Sprintf("invalid character: `%s` / expected nil or atom", string(r))))
}

func (t *Tokenizer) tokNil() tokFunc { // parse form end
	t.tokens = append(t.tokens, Token{
		Type: TypeFormEnd,
		Text: ")",
		Pos: t.pos,
	})
	t.pos++

	return t.tokNilOrTextOrForm
}

func (t *Tokenizer) tokAtom() tokFunc { // parse atom
	atomEnd := t.pos
	for atomEnd < t.l && isAtomChar(t.bs[atomEnd]) {
		atomEnd++
	}
	t.tokens = append(t.tokens, Token{
		Type: TypeAtom,
		Text: string(t.bs[t.pos:atomEnd]),
		Pos: t.pos,
	})
	t.pos = atomEnd

	return t.tokNilOrTextOrForm
}

func (t *Tokenizer) tokNilOrTextOrForm() tokFunc {
	r := t.bs[t.pos]
	if r == ')' {
		return t.tokNil
	}
	if r == '(' {
		return t.tokForm
	}
	return t.tokText
}

func (t *Tokenizer) tokEOF() tokFunc {
	t.tokens = append(
		t.tokens,
		Token{
			Type: TypeFormStart,
			Text: "(",
			Pos: t.pos,
		},
		Token{
			Type: TypeAtom,
			Text: "eof",
			Pos: t.pos,
		},
		Token{
			Type: TypeFormEnd,
			Text: ")",
			Pos: t.pos,
		},
	)

	return nil
}

func (t *Tokenizer) skipWhitespace() {
	for t.pos < t.l && isWhitespace(t.bs[t.pos]) {
		// @todo: count line / column
		t.pos++
	}
}

func isWhitespace(r rune) bool {
	ws := []rune{' ', '\n', '\r', '\t', '\v', '\f'}
	for _, w := range ws {
		if r == w {
			return true
		}
	}
	return false
}

func isAlphaLower(r rune) bool {
	return r >= 'a' && r <= 'z' || r == '-' || r == '@'
}

func isNum(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAtomChar(r rune) bool {
	return isAlphaLower(r) || isNum(r)
}

func (t *Tokenizer) NewTokenError(msg string) TokenError {
	return TokenError{
		Msg: msg,
		Pos: t.pos,
		FileName: "@todo: implement",
	}
}

func (e TokenError) Error() string {
	return fmt.Sprintf("%s[%d]: %s", e.FileName, e.Pos, e.Msg)
}

func (t Token) String() string {
	switch (t.Type) {
	case TypeFormStart:
		return fmt.Sprintf("FormStart{%d: `%s`}", t.Pos, VisibleString(t.Text))
	case TypeAtom:
		return fmt.Sprintf("Atom{%d: `%s`}", t.Pos, VisibleString(t.Text))
	case TypeText:
		return fmt.Sprintf("Text{%d: `%s`}", t.Pos, VisibleString(t.Text))
	case TypeFormEnd:
		return fmt.Sprintf("FormEnd{%d: `%s`}", t.Pos, VisibleString(t.Text))
	}
	log.Fatalf("invalid token type: %v", t.Type)
	return fmt.Sprintf("Invalid[%d]{%d: `%s`}", t.Type, t.Pos, VisibleString(t.Text))
}

func VisibleString(s string) string {
	asciiSpecialLookup := [...]string{
		"<NUL>",
		"<SOH>",
		"<STX>",
		"<ETX>",
		"<EOT>",
		"<ENQ>",
		"<ACK>",
		"\\a",
		"\\b",
		"\\t",
		"\\n",
		"\\v",
		"\\f",
		"\\r",
		"<SO>",
		"<SI>",
		"<DLE>",
		"<DC1>",
		"<DC2>",
		"<DC3>",
		"<DC4>",
		"<NAK>",
		"<SYN>",
		"<ETB>",
		"<CAN>",
		"<EM>",
		"<SUB>",
		"<ESC>",
		"<FS>",
		"<GS>",
		"<RS>",
		"<US>",
	}
	v := ""
	for _, r := range s {
		if r >= 32 && r <= 126 {
			v += string(r)
		} else if r == 127 {
			v += "<DEL>"
		} else if r >= 0 && r <= 31 {
			v += asciiSpecialLookup[r]
		} else /* is unicode (probably) */{
			v += fmt.Sprintf("<%U>", r)
		}
	}
	return v
}

