package eqdraw

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAsciiEquation(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		expected node
		err      error
	}{
		{
			name:  "basic nospace",
			input: "1+2",
			expected: &Run{Terms: []node{
				&Term{Content: []rune{'1'}},
				&Term{Content: []rune{'+'}},
				&Term{Content: []rune{'2'}},
			}},
		},
		{
			name:  "basic space",
			input: "1 + 2",
			expected: &Run{Terms: []node{
				&Term{Content: []rune{'1'}},
				&Term{Content: []rune{'+'}},
				&Term{Content: []rune{'2'}},
			}},
		},
		{
			name:  "basic parenthesis nospace",
			input: "2(b+1)",
			expected: &Run{Terms: []node{
				&Term{Content: []rune{'2'}},
				&Parenthesis{Term: &Run{
					Terms: []node{
						&Term{Content: []rune{'b'}},
						&Term{Content: []rune{'+'}},
						&Term{Content: []rune{'1'}},
					},
				}},
			}},
		},
		{
			name:  "basic parenthesis",
			input: "2(b + 1 )",
			expected: &Run{Terms: []node{
				&Term{Content: []rune{'2'}},
				&Parenthesis{Term: &Run{
					Terms: []node{
						&Term{Content: []rune{'b'}},
						&Term{Content: []rune{'+'}},
						&Term{Content: []rune{'1'}},
					},
				}},
			}},
		},
		{
			name:  "parenthesis superfluous",
			input: "2(1)",
			expected: &Run{Terms: []node{
				&Term{Content: []rune{'2'}},
				&Parenthesis{
					Term: &Term{Content: []rune{'1'}},
				},
			}},
		},
		{
			name:  "line",
			input: "y = mx + b",
			expected: &Run{Terms: []node{
				&Term{Content: []rune{'y'}},
				&Term{Content: []rune{'='}},
				&Term{Content: []rune{'m', 'x'}},
				&Term{Content: []rune{'+'}},
				&Term{Content: []rune{'b'}},
			}},
		},
		{
			name:  "sqrt",
			input: "sqrt(12 - a)",
			expected: &Root{Term: &Run{Terms: []node{
				&Term{Content: []rune{'1', '2'}},
				&Term{Content: []rune{'-'}},
				&Term{Content: []rune{'a'}},
			}}},
		},
		{
			name:  "div",
			input: "1/2",
			expected: &Div{
				Numerator:   &Term{Content: []rune{'1'}},
				Denominator: &Term{Content: []rune{'2'}},
			},
		},
		{
			name:  "div unwrap",
			input: "(1 + 2)/(2+1)",
			expected: &Div{
				Numerator:   &Run{Terms: []node{&Term{Content: []rune{'1'}}, &Term{Content: []rune{'+'}}, &Term{Content: []rune{'2'}}}},
				Denominator: &Run{Terms: []node{&Term{Content: []rune{'2'}}, &Term{Content: []rune{'+'}}, &Term{Content: []rune{'1'}}}},
			},
		},
		{
			name:  "eq with div",
			input: "2x = 1/2",
			expected: &Run{Terms: []node{&Term{Content: []rune{'2', 'x'}}, &Term{Content: []rune{'='}}, &Div{
				Numerator:   &Term{Content: []rune{'1'}},
				Denominator: &Term{Content: []rune{'2'}},
			}}},
		},
		{
			name:  "eq with div stack overflow",
			input: "y = mx + b/2",
			expected: &Run{Terms: []node{&Term{Content: []rune{'y'}}, &Term{Content: []rune{'='}}, &Div{
				Numerator:   &Run{Terms: []node{&Term{Content: []rune{'m', 'x'}}, &Term{Content: []rune{'+'}}, &Term{Content: []rune{'b'}}}},
				Denominator: &Term{Content: []rune{'2'}},
			}}},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out, err := ParseASCIIEquation(tc.input)
			if err != tc.err {
				t.Errorf("err = %v, want %v", err, tc.err)
			}
			if diff := cmp.Diff(out, tc.expected,
				cmp.AllowUnexported(Run{}), cmp.AllowUnexported(Term{}), cmp.AllowUnexported(Parenthesis{}), cmp.AllowUnexported(Root{}), cmp.AllowUnexported(Div{})); diff != "" {
				t.Errorf("output differed:\n%s", diff)
			}
		})
	}
}
