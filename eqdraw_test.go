package eqdraw

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"
)

func testContext(t *testing.T, sz image.Rectangle) *drawContext {
	t.Helper()
	f, err := DefaultFont()
	if err != nil {
		t.Fatal(err)
	}
	o := truetype.Options{
		Size: 24,
	}
	ff := truetype.NewFace(f, &o)

	return &drawContext{
		o:   o,
		f:   f,
		ff:  ff,
		out: image.NewRGBA(sz),
	}
}

func TestLayout(t *testing.T) {
	tcs := []struct {
		name    string
		node    node
		results layoutResult
	}{
		{
			"term",
			&Term{Content: []rune{'h', 'e', 'l', 'l', 'o'}},
			layoutResult{
				Width:  fixed.Int26_6(56<<6 + 44),
				Height: fixed.Int26_6(27<<6 + 0),
			},
		},
		{
			"empty_parentheses",
			&Parenthesis{},
			layoutResult{
				Width:  fixed.Int26_6(17<<6 + 0),
				Height: fixed.Int26_6(36<<6 + 0),
			},
		},
		{
			"text_in_parentheses",
			&Parenthesis{Term: &Term{Content: []rune{'h', 'e', 'l', 'l', 'o'}}},
			layoutResult{
				Width:  fixed.Int26_6(75<<6 + 42),
				Height: fixed.Int26_6(39<<6 + 0),
			},
		},
		{
			"run_in_parentheses",
			&Parenthesis{Term: &Run{}},
			layoutResult{
				Width:  fixed.Int26_6(21<<6 + 0),
				Height: fixed.Int26_6(36<<6 + 0),
			},
		},
		{
			"run",
			&Run{Terms: []node{&Term{Content: []rune{'h', 'e', 'l', 'l', 'o'}}}},
			layoutResult{
				Width:  fixed.Int26_6(60<<6 + 44),
				Height: fixed.Int26_6(29<<6 + 0),
			},
		},
		{
			"div",
			&Div{Numerator: &Term{Content: []rune{'1'}}, Denominator: &Term{Content: []rune{'2'}}},
			layoutResult{
				Width:  fixed.Int26_6(21<<6 + 22),
				Height: fixed.Int26_6(72<<6 + 0),
			},
		},
		{
			"root",
			&Root{Term: &Term{Content: []rune{'1'}}},
			layoutResult{
				Width:  fixed.Int26_6(50<<6 + 57),
				Height: fixed.Int26_6(29<<6 + 0),
			},
		},
	}
	dc := testContext(t, image.Rect(0, 0, 500, 200))

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.node.Layout(dc); err != nil {
				t.Fatalf("Layout() failed: %v", err)
			}
			if *tc.node.Bounds() != tc.results {
				t.Errorf("results = %v, want %v", tc.node.Bounds(), tc.results)
			}
		})
	}
}

func TestDraw(t *testing.T) {
	const writeToTmp = "root"

	tcs := []struct {
		name string
		node node
	}{
		{
			"term",
			&Term{Content: []rune{'h', 'e', 'l', 'l', 'o'}},
		},
		{
			"text_in_parentheses",
			&Parenthesis{Term: &Term{Content: []rune{'h', 'e', 'l', 'l', 'o'}}},
		},
		{
			"multi_text_in_parentheses",
			&Parenthesis{Term: &Run{
				Terms: []node{
					&Term{Content: []rune{'h', 'e', 'l', 'l', 'o'}},
					&Term{Content: []rune{'M', 'A', 'T', 'E'}},
				},
			}},
		},
		{
			"div",
			&Div{Numerator: &Term{Content: []rune{'1'}}, Denominator: &Term{Content: []rune{'√', '2', 'a'}}},
		},
		{
			"root",
			&Root{Term: &Term{Content: []rune{'1', 'a'}}},
		},
	}

	clipRegion := image.Rectangle{
		Min: image.Point{X: -1, Y: -1},
		Max: image.Point{X: 20000, Y: 20000},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			dc := testContext(t, image.Rect(0, 0, 500, 200))
			if err := tc.node.Layout(dc); err != nil {
				t.Fatalf("Layout() failed: %v", err)
			}

			// Make a smaller canvas, and color the background white.
			dc.out = image.NewRGBA(image.Rectangle{Max: image.Point{X: tc.node.Bounds().Width.Ceil(), Y: tc.node.Bounds().Height.Ceil()}})
			draw.Draw(dc.out, clipRegion, image.NewUniform(color.White), image.Point{}, draw.Over)

			err := tc.node.Draw(dc, fixed.Point26_6{}, clipRegion)
			if err != nil {
				t.Fatalf("Draw() failed: %v", err)
			}

			if writeToTmp == tc.name {
				f, err := os.OpenFile("/tmp/test_output.png", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				if err := png.Encode(f, dc.out); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
