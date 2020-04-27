package eqdraw

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var (
	paraMargin = layoutResult{
		Height: fixed.Int26_6(12 << 6),
		Width:  fixed.Int26_6(1 << 6),
	}
)

// Parenthesis represents terms contained within parentheses.
type Parenthesis struct {
	ff font.Face

	layout      *layoutResult
	tallestTerm fixed.Int26_6

	Terms []node
}

// Bounds returns the width and height of the rendered term, as computed by
// the last layout pass. If no layout pass has occurred, the returned value
// will be nil.
func (p *Parenthesis) Bounds() *layoutResult {
	return p.layout
}

// Layout is called during the layout pass to compute the rendered size of this node.
func (p *Parenthesis) Layout(dc *drawContext) error {
	sz := layoutResult{}

	var tallestTerm fixed.Int26_6
	for _, t := range p.Terms {
		if err := t.Layout(dc); err != nil {
			return err
		}
		b := t.Bounds()
		sz.Width += b.Width
		sz.Height += b.Height
		if b.Height > tallestTerm {
			tallestTerm = b.Height
		}
	}
	p.tallestTerm = tallestTerm

	// Determine the appropriate font size so the parentheses wrap all terms vertically.
	for fz := dc.o.Size; fz < 144; fz++ {
		opts := dc.o
		opts.Size = fz
		p.ff = truetype.NewFace(dc.f, &opts)
		if h := p.ff.Metrics().Height; h >= sz.Height {
			sz.Height = h
			break
		}
	}

	// Add the widths for the two parentheses.
	a, _ := p.ff.GlyphAdvance('(')
	sz.Width += a
	a, _ = p.ff.GlyphAdvance(')')
	sz.Width += a

	sz.Height += paraMargin.Height
	sz.Width += paraMargin.Width
	p.layout = &sz
	return nil
}

// Draw is called to render the parentheses and its contained terms.
func (p *Parenthesis) Draw(dc *drawContext, pos fixed.Point26_6, clip image.Rectangle) error {
	src := image.NewUniform(color.Black)
	asc := p.ff.Metrics().Ascent
	pos.X += paraMargin.Width / 2
	pos.Y += asc + paraMargin.Height/4

	dr, mask, maskp, advance, _ := p.ff.Glyph(pos, '(')
	draw.DrawMask(dc.out, dr.Intersect(clip), src, image.Point{}, mask, maskp, draw.Over)
	pos.X += advance

	pos.Y -= asc
	for _, t := range p.Terms {
		sz := t.Bounds()
		adjustY := (p.tallestTerm - sz.Height) / 2
		pos.Y += adjustY
		if err := t.Draw(dc, pos, clip); err != nil {
			return err
		}
		pos.Y -= adjustY
		pos.X += sz.Width
	}
	pos.Y += asc

	dr, mask, maskp, advance, _ = p.ff.Glyph(pos, ')')
	draw.DrawMask(dc.out, dr.Intersect(clip), src, image.Point{}, mask, maskp, draw.Over)
	return nil
}
