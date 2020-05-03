package eqdraw

import (
	"image"
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

	layout *layoutResult

	Term node
}

// Bounds returns the width and height of the rendered term, as computed by
// the last layout pass. If no layout pass has occurred, the returned value
// will be nil.
func (p *Parenthesis) Bounds() *layoutResult {
	return p.layout
}

// Layout is called during the layout pass to compute the rendered size of this node.
func (p *Parenthesis) Layout(dc *DrawContext) error {
	sz := layoutResult{}
	if p.Term != nil {
		if err := p.Term.Layout(dc); err != nil {
			return err
		}
		b := p.Term.Bounds()
		sz.Width += b.Width
		sz.Height += b.Height
	}

	// Determine the appropriate font size so the parentheses wraps the term.
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
func (p *Parenthesis) Draw(dc *DrawContext, pos fixed.Point26_6, clip image.Rectangle) error {
	asc := p.ff.Metrics().Ascent
	pos.X += paraMargin.Width / 2
	pos.Y += asc + paraMargin.Height/4

	dr, mask, maskp, advance, _ := p.ff.Glyph(pos, '(')
	draw.DrawMask(dc.out, dr.Intersect(clip), dc.fg, image.Point{}, mask, maskp, draw.Over)
	pos.X += advance

	if p.Term != nil {
		pos.Y -= asc
		if err := p.Term.Draw(dc, pos, clip); err != nil {
			return err
		}
		pos.X += p.Term.Bounds().Width
		pos.Y += asc
	}

	dr, mask, maskp, advance, _ = p.ff.Glyph(pos, ')')
	draw.DrawMask(dc.out, dr.Intersect(clip), dc.fg, image.Point{}, mask, maskp, draw.Over)
	return nil
}
