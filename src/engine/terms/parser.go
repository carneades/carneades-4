// Copyright © 2016 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Terms Parser

package terms

import (
	"fmt"
	"os"
	"strings"
	sc "text/scanner"
	// "go/scanner"
	// "go/token"
)

func ReadString(src string) (result Term, ok bool) {
	// src is the input that we want to tokenize.
	var s sc.Scanner
	// var s *sc.Scanner
	// Initialize the scanner.
	s.Init(strings.NewReader(src))
	s.Error = err

	/* var tok rune
	for tok != sc.EOF {
		tok = s.Scan()
		fmt.Println("At position", s.Pos(), ":", s.TokenText())
	}
	*/

	result, _, ok = readBIConstraint(&s)
	return
}
func err(s *sc.Scanner, str string) {
	fmt.Fprintln(os.Stderr, "*** Parse Error before[", s.Pos(), "]:", str)
}

func readBIConstraint(s *sc.Scanner) (t Term, tok rune, ok bool) {
	fmt.Printf("readBIConstraint : \n")
	t, tok, ok = expression(s, s.Scan())
	if tok == sc.EOF {
		return
	}
	if tok == ',' {
		t1 := List{t}
		for tok == ',' {
			t, tok, ok = expression(s, s.Scan())
			fmt.Printf("<<< expression: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
			if !ok {
				return t1, tok, false
			}
			t1 = append(t1, t)
		}
		t = t1
	}
	if tok == '.' {
		tok = s.Scan()
	}
	if tok != sc.EOF {
		err(s, fmt.Sprintf("',' or EOF exspected, not '%c'", tok))
	}
	return
}

func expression(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	fmt.Printf("expression : '%s'\n", f(tok1))
	t, tok, ok = simple_expression(s, tok1)
	fmt.Printf("<<< simple_expression: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
	op := ""
	// named operator
	if tok <= 0 {
		if tok == sc.Ident && s.TokenText() == "in" {
			op = "in"
		} else {
			return
		}
	} else {
		// sign operator
		tok2 := s.Peek()
		switch tok {
		case '=':
			op = "="
		case '<':
			switch tok2 {
			case '>':
				op = "<>"
				tok = s.Scan()
			case '=':
				op = "<="
				tok = s.Scan()
			}
		case '!':
			if tok2 != '=' {
				err(s, "missing '=' after '!' in expression")
			} else {
				tok = s.Scan()
			}
			op = "!="
		case '>':
			if tok2 == '=' {
				op = ">="
				tok = s.Scan()
			} else {
				op = ">"
			}

		}
	}
	if op == "" {
		return
	}
	// simple expression with op
	ex := Compound{Functor: op}
	tex := []Term{t}
	t, tok, ok = simple_expression(s, s.Scan())
	fmt.Printf("<<< simple_expression: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
	if !ok {
		return
	}
	tex = append(tex, t)
	ex.Args = tex
	return ex, tok, ok
}

func simple_expression(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	fmt.Printf("simple_expression : '%s'\n", f(tok1))
	// sign
	monop := ""
	if tok1 == '+' {
		tok1 = s.Scan()
	} else {
		if tok1 == '-' {
			monop = "-"
			tok1 = s.Scan()
		}
	}
	t, tok, ok = sterm(s, tok1)
	fmt.Printf("<<< sterm: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)

	op := ""
	if tok <= 0 {
		if tok == sc.Ident && s.TokenText() == "or" {
			op = "or"
		} else {
			return
		}
	} else {
		tok2 := s.Peek()
		switch tok {
		case '-':
			op = "-"
			tok = s.Scan()
		case '+':
			op = "+"
			tok = s.Scan()
		case '|':
			if tok2 != '|' {
				err(s, "missing '|' after '|' in expression")
			} else {
				tok = s.Scan()
			}
			tok = s.Scan()
			op = "||"

		}
	}

	if op != "" {
		ex := Compound{Functor: op}
		tex := []Term{t}
		t, tok, ok = sterm(s, tok)
		fmt.Printf("<<< sterm: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
		if !ok {
			return
		}
		tex = append(tex, t)
		ex.Args = tex
		t = ex
	}

	if monop != "" {
		t = Compound{Functor: monop, Args: []Term{t}}
	}
	return
}

func sterm(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	fmt.Printf("sterm : '%s'\n", f(tok1))
	t, tok, ok = factor(s, tok1)
	fmt.Printf("<<< factor: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
	op := ""
	// named operator
	if tok <= 0 {
		if tok == sc.Ident {
			switch s.TokenText() {
			case "div":
				op = "div"
			case "mod":
				op = "mod"
			case "and":
				op = "and"
			default:
				return
			}
		} else {
			return
		}
	} else {
		// sign operator
		// tok2 := s.Peek()
		switch tok {
		case '*':
			op = "*"
		case '/':
			op = "/"
		case '%':
			op = "%"
		}
	}
	if op == "" {
		return
	}
	// factor with op
	ex := Compound{Functor: op}
	tex := []Term{t}
	t, tok, ok = sterm(s, s.Scan())
	fmt.Printf("<<< sterm: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
	if !ok {
		return
	}
	tex = append(tex, t)
	ex.Args = tex
	return ex, tok, ok

}

func factor(s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	fmt.Printf("factor : '%s'\n", f(tok1))
	tok = tok1
	ok = true

	switch tok1 {
	case '[': //
		list := List{}
		if s.Peek() == ']' {
			s.Scan()
			return list, s.Scan(), true
		}
		tok = ','
		pos := s.Pos()
		for tok == ',' {
			t, tok, ok = expression(s, s.Scan())
			fmt.Printf("<<< expression in [ factor: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
			if !ok {
				t = list
				return
			}
			list = append(list, t)
		}
		t = list
		// to do |
		/*if tok == '|' {
			// [ expr | variable ]
		}*/
		if tok != ']' {
			err(s, fmt.Sprintf("missing closed ']' for the open '[' at position %s", pos))
			return t, tok, false
		}
		return t, s.Scan(), true
	case '(':
		pos := s.Pos()
		t, tok, ok = expression(s, s.Scan())
		fmt.Printf("<<< expression in ( factor: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
		if !ok {
			return
		}
		if tok != ')' {
			err(s, fmt.Sprintf("missing closed ')' for the open '(' at position %s", pos))
			return t, tok, false
		}
		tok = s.Scan()
	case sc.Ident:
		t, tok, ok = factor_name(s.TokenText(), s, s.Scan())
		fmt.Printf("<<< factor_name: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
	case sc.Int:
		t, tok, ok = sInt(s)
		fmt.Printf("<<< sInt: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
	case sc.Float:
		t, tok, ok = sFloat(s)
		fmt.Printf("<<< sFloat: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
	case sc.Char:
		t, tok, ok = sChar(s)
		fmt.Printf("<<< sChar: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
	case sc.String:
		t, tok, ok = sString(s)
		fmt.Printf("<<< sString: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
	case sc.RawString:
		t, tok, ok = sString(s)
		fmt.Printf("<<< sRawString: term: %s tok: '%s' ok: %v \n", t.String(), f(tok), ok)
	// case sc.Comment:
	case sc.EOF:
		err(s, "EOF missing term")
		return nil, tok, false
	default:
		err(s, fmt.Sprintf("unexpected character '%c', expect <variable>, <name>, <constant>, '(' or '['", tok))
		t = nil
	}
	return t, tok, true
}

func factor_name(name string, s *sc.Scanner, tok1 rune) (t Term, tok rune, ok bool) {
	fmt.Printf("factor_name : %s '%s'\n", name, f(tok1))
	t = nil
	ok = true
	tok = tok1
	if tok != '(' {
		return bi_0(name, tok)
	}
	// to do: distinguish between CHR- and Built In-Constraint
	args := []Term{}
	if s.Peek() == ')' {
		s.Scan()

		return Compound{Functor: name, Args: []Term{}}, s.Scan(), true
	}
	tok = ','
	pos := s.Pos()
	for tok == ',' {
		t, tok, ok = expression(s, s.Scan())
		if !ok {
			t = Compound{Functor: name, Args: args}
			return
		}
		args = append(args, t)
	}
	t = Compound{Functor: name, Args: args}

	if tok != ')' {
		err(s, fmt.Sprintf("missing closed ')' for the open '(' at position %s", pos))
		return t, tok, false
	}
	return t, s.Scan(), true
}

func sInt(s *sc.Scanner) (Term, rune, bool) {
	fmt.Printf("sInt : '%s'\n", s.TokenText())
	var (
		i   int
		err error
	)
	_, err = fmt.Sscan(s.TokenText(), &i)
	if err == nil {
		return Int(i), s.Scan(), true
	}
	return Int(i), s.Scan(), false
}

func sFloat(s *sc.Scanner) (Term, rune, bool) {
	fmt.Printf("sFloat : '%s'\n", s.TokenText())
	var (
		f   float64
		err error
	)
	_, err = fmt.Sscan(s.TokenText(), &f)
	if err == nil {
		return Float(f), s.Scan(), true
	}
	return Float(f), s.Scan(), false
}

func sString(s *sc.Scanner) (Term, rune, bool) {
	fmt.Printf("sString : '%s'\n", s.TokenText())
	var (
		str string
		err error
	)
	_, err = fmt.Sscan(s.TokenText(), &str)
	if err == nil {
		return String(str), s.Scan(), true
	}
	return String(str), s.Scan(), false
}

func sChar(s *sc.Scanner) (Term, rune, bool) {
	fmt.Printf("sChar : '%s'\n", s.TokenText())
	var (
		c   rune
		err error
	)
	_, err = fmt.Sscan(s.TokenText(), &c)
	if err == nil {
		return String(c), s.Scan(), true
	}
	return String(c), s.Scan(), false
}

func f(tok rune) string {
	if tok > 0 {
		return string(tok)
	}

	switch tok {
	case sc.Ident:
		return "Ident"
	case sc.Int:
		return "Int"
	case sc.Float:
		return "Float"
	case sc.Char:
		return "Char"
	case sc.String:
		return "String"
	case sc.RawString:
		return "RawString"
	case sc.Comment:
		return "Comment"
	case sc.EOF:
		return "EOF"
	}
	return "??"
}

func bi_0(n string, tk rune) (t Term, tok rune, ok bool) {
	fmt.Printf("bi_0 : '%s'\n", n)
	tok = tk
	ok = true
	switch n {
	case "true":
		t = Bool(true)
		return
	case "false":
		t = Bool(false)
		return
	}

	c := n[0]
	if c >= 'A' && c <= 'Z' {
		t = Variable{Name: n}
		return
	} else {
		t = Atom(n)
		return
	}
}
