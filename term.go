package eqdraw

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/math/fixed"
)

const (
	termRenderBlocks = false
)

var (
	termMargin = layoutResult{
		Height: fixed.Int26_6(3 << 6),
		Width:  fixed.Int26_6(6 << 6),
	}
)

// Term represents a run of text to be rendered.
type Term struct {
	layout  *layoutResult
	Content []rune
}

// Bounds returns the width and height of the rendered term, as computed by
// the last layout pass. If no layout pass has occurred, the returned value
// will be nil.
func (t *Term) Bounds() *layoutResult {
	return t.layout
}

// Layout is called during the layout pass to compute the rendered size of this node.
func (t *Term) Layout(dc *DrawContext) error {
	var (
		prevC = rune(-1)
		w     = fixed.Int26_6(0)
	)
	for i := 0; i < len(t.Content); i++ {
		c := t.Content[i]
		ff := dc.ff
		if c >= 'a' && c <= 'z' {
			ff = dc.ffi
		}

		var kern fixed.Int26_6
		if prevC >= 0 {
			kern = ff.Kern(prevC, c)
		}
		a, ok := ff.GlyphAdvance(c)
		if !ok {
			continue
		}
		prevC = c
		w += a + kern
	}

	t.layout = &layoutResult{
		Height: dc.ff.Metrics().Height + termMargin.Height,
		Width:  w + termMargin.Width,
	}

	return nil
}

// Draw is called to render the term.
func (t *Term) Draw(dc *DrawContext, pos fixed.Point26_6, clip image.Rectangle) error {
	pos.X += termMargin.Width / 2
	pos.Y += dc.ff.Metrics().Ascent + termMargin.Height/2

	prevC := rune(-1)
	for i := 0; i < len(t.Content); i++ {
		c := t.Content[i]
		ff := dc.ff
		if c >= 'a' && c <= 'z' {
			ff = dc.ffi
		}

		if prevC >= 0 {
			pos.X += ff.Kern(prevC, c)
		}
		dr, mask, maskp, advance, ok := ff.Glyph(pos, c)
		if !ok {
			continue
		}
		if termRenderBlocks {
			mask = image.NewUniform(color.RGBA{A: 100})
		}

		draw.DrawMask(dc.out, dr.Intersect(clip), dc.fg, image.Point{}, mask, maskp, draw.Over)
		pos.X += advance
		prevC = c
	}

	return nil
}
