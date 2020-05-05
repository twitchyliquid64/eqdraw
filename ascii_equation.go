package eqdraw

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type specKind uint8

// Valid specKind values.
const (
	kindTerms = iota
	kindParenthesis
	kindRoot
)

type termType uint8

const (
	termNormal termType = iota
	termBinOp
)

type eqSpec struct {
	kind  specKind
	terms []node
}

func (s eqSpec) String() string {
	var b strings.Builder
	for _, t := range s.terms {
		switch n := t.(type) {
		case *Term:
			b.WriteString(string(n.Content))
		default:
			b.WriteString(fmt.Sprint(t))
		}
	}
	return b.String()
}

// postProcess splits a sequence of terms containing a '/' symbol, into the
// numerator + denominator under a new Div node.
func (s *eqSpec) postProcess() {
	idx := -1
	for i, n := range s.terms {
		if t, ok := n.(*Term); ok && string(t.Content) == "/" {
			idx = i
		}
	}

	if idx > 0 {
		// Don't capture anything delimited by an equals sign.
		var (
			num       = s.terms[:idx]
			remaining = []node{}
		)
		for i := len(num) - 1; i > 0; i-- {
			if t, isTerm := num[i].(*Term); isTerm && string(t.Content) == "=" {
				remaining = num[:i+1]
				num = num[i+1:]
				break
			}
		}

		var tmp eqSpec
		tmp.pushNode(eqSpec{terms: num})
		tmp.pushNode(eqSpec{terms: s.terms[idx+1:]})
		d := &Div{
			Numerator:   tmp.terms[0],
			Denominator: tmp.terms[1],
		}
		if paren, isParenth := d.Numerator.(*Parenthesis); isParenth {
			d.Numerator = paren.Term
		}
		if paren, isParenth := d.Denominator.(*Parenthesis); isParenth {
			d.Denominator = paren.Term
		}

		if len(remaining) > 0 {
			s.terms = append(remaining, d)
		} else {
			s.terms = []node{d}
		}
	}
}

func (s *eqSpec) push(term []rune, kind termType) {
	if len(term) == 0 {
		return
	}

	switch kind {
	case termNormal:
		s.terms = append(s.terms, &Term{Content: term})
	case termBinOp:
		s.terms = append(s.terms, &Term{Content: term})
	}
}

func (s *eqSpec) pushNode(in eqSpec) {
	in.postProcess()

	var out node
	switch len(in.terms) {
	case 0:
		return
	case 1:
		out = in.terms[0]
	default:
		out = &Run{Terms: in.terms}
	}

	switch in.kind {
	case kindRoot:
		s.terms = append(s.terms, &Root{Term: out})
	case kindParenthesis:
		s.terms = append(s.terms, &Parenthesis{Term: out})
	default:
		s.terms = append(s.terms, out)
	}
}

func binOp(in rune) bool {
	switch in {
	case '+', '-', '*', '/':
		return true
	}
	return false
}

// ParseASCIIEquation attempts to generate the node tree by parsing an
// ascii representation of the equation.
func ParseASCIIEquation(inp string) (node, error) {
	var (
		nextTerm    termType
		inQuotes    = false
		quoteChar   = '\''
		accumulator []rune
		out         eqSpec
		stack       []eqSpec
	)

	input := []byte(inp)
	pos := 0
	for len(input) > 0 {
		c, size := utf8.DecodeRune(input)
		input = input[size:]
		pos++

		switch {
		case inQuotes && c == quoteChar: // Terminating quote reached
			out.push(accumulator, nextTerm)
			accumulator = []rune{}
			nextTerm = termNormal
			inQuotes = false

		case inQuotes: // Still in a quoted term
			accumulator = append(accumulator, c)

		case !inQuotes && c == '\'': // New quoted term
			inQuotes = true
			quoteChar = '\''

		case !inQuotes && c == '(': // Start parenthesis
			switch {
			case string(accumulator) == "sqrt":
				stack = append(stack, out)
				out = eqSpec{kind: kindRoot}
			default:
				out.push(accumulator, nextTerm)
				stack = append(stack, out)
				out = eqSpec{kind: kindParenthesis}
			}
			accumulator = []rune{}
			nextTerm = termNormal

		case !inQuotes && c == ')': // End parenthesis
			out.push(accumulator, nextTerm)
			accumulator = []rune{}
			nextTerm = termNormal
			if len(stack) == 0 {
				return nil, fmt.Errorf("unmatched end parenthesis at position %d", pos)
			}
			tmp := out
			out = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			out.pushNode(tmp)

		case !inQuotes && (c == ',' || c == ' '): // End of term
			out.push(accumulator, nextTerm)
			accumulator = []rune{}
			nextTerm = termNormal

		case !inQuotes && binOp(c): // Split on binary op
			out.push(accumulator, nextTerm)
			accumulator = []rune{}
			out.push([]rune{c}, termBinOp)
			nextTerm = termNormal

		default:
			accumulator = append(accumulator, c)
			switch accumulator {

			}
		}
	}
	out.push(accumulator, nextTerm)

	if len(stack) != 0 {
		return nil, fmt.Errorf("unmatched start parenthesis")
	}

	out.postProcess()
	switch len(out.terms) {
	case 1:
		return out.terms[0], nil
	case 0:
		return nil, nil
	default:
		return &Run{Terms: out.terms}, nil
	}
}
