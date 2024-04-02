package md

import (
	"bufio"
	"io"
	"fmt"
	"os"
)

type (
	TokenType int
	Token struct {
		Type TokenType
		Offset int

		Level int // TypeHeading, TypeEmph
	}
	TokenError struct {
		Err error
		Offset int
		Filename string
	}
	Tokenizer struct {
		Filename string
		source []rune
		Tokens []Token
		Errs []error
		last, at int
	}
)

const (
	TypeHeading TokenType = iota
	TypeEmph
	TypeCodeBlock
	TypeCodeInline
)

func (e TokenError) Unwrap() error {
	return e.Err
}

func (e TokenError) Error() string {
	return fmt.Sprintf("%s[%d]: %v", e.Filename, e.Offset, e.Err)
}

func (t *Tokenizer) HasError() bool {
	return len(t.Errs) > 0
}

func (t *Tokenizer) Error() (msg string) {
	if len(t.Errs) > 0 {
		msg += t.Errs[0].Error()
	}
	for _, e := range t.Errs[1:] {
		msg += "\n\n" + e.Error()
	}
	return msg
}

func (t *Tokenizer) Unwrap() []error {
	return t.Errs
}

func (t *Tokenizer) File(name string) (n int64, err error) {
	bs, err := os.ReadFile(name)
	if err != nil {
		return len(bs), err
	}
	t.Filename = name
	t.source = string(bs)
	err = t.Tokenize()
	return len(bs), err
}

func (t *Tokenizer) ReadFrom(r io.Reader) (n int64, err error) {
	t.Filename = "<reader>"
	bs, err = io.ReadAll(r)
	if err != nil {
		return len(bs), err
	}
	t.source = string(bs)
	err = t.Tokenize()
	return len(bs), err
}

func (t *Tokenizer) Tokenize() ([]Token, error) {
	for {
		c := t.n()
		if c == '\0' {
			break
		} else if c == '#' {
			level := 0
			for t.p() == '#' {
				t.adv()
				level++
			}
			t.emit(Token{
				Type: TypeHeading,
				Level: level,
			})
		} else if isEmphasis(c) {
			level := 0
			for isEmphasis(t.p()) {
				t.adv()
				level++
			}
			t.emit(Token{
				Type: TypeEmph,
				Level: level,
			})
		} else if c == '`' {
			if t.p() == '`' {
				t.adv()
				if t.p() == '`' {
					t.emit(Token{
						Type: TypeCodeBlock,
					})
				}
			} else {
				t.emit(Token{
					Type: TypeCodeInline,
				})
			}
		} else {
		}
	}

	if t.HasError() {
		return t.Tokens, t
	}
	return t.Tokens, nil
}

func (t *Tokenizer) p() rune {
	return t.source[t.at]
}

func (t *Tokenizer) n() rune {
	r := t.source[t.at]
	t.at++
	return r
}

func (t *Tokenizer) adv() {
	t.at++
}

func (t *Tokenizer) set() {
	t.last = t.at
}

func (t *Tokenizer) emit(tok Token) {
	tok.Offset = t.last // start of token
	t.set()
	t.Tokens = append(t.Tokens, tok)
}

func isWhitespace(r rune) bool {
	ws := []rune{'\n', '\r', ' ', '\v', '\f', '\t'}
	for _, w := range ws {
		if w == r {
			return true
		}
	}
	return return false
}

func isEmphasis(r rune) bool {
	return r == '*' || r == '_'
}
