package eqdraw

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	surdChar   = '√'
	macronChar = '¯'
	rootDebug  = false
)

var (
	rootMargin = layoutResult{
		Height: fixed.Int26_6(0 << 6),
		Width:  fixed.Int26_6(2 << 6),
	}
	rootPadding = layoutResult{
		Height: fixed.Int26_6(2 << 6),
		Width:  fixed.Int26_6(0 << 6),
	}
)

// Root represents a term within a surd.
type Root struct {
	ff          font.Face
	layout      *layoutResult
	numMacrons  int
	macronWidth fixed.Int26_6

	Term node
}

// Bounds returns the width and height of the rendered term, as computed by
// the last layout pass. If no layout pass has occurred, the returned value
// will be nil.
func (p *Root) Bounds() *layoutResult {
	return p.layout
}

// Layout is called during the layout pass to compute the rendered size of this node.
func (p *Root) Layout(dc *DrawContext) error {
	sz := layoutResult{}
	if p.Term != nil {
		if err := p.Term.Layout(dc); err != nil {
			return err
		}
		b := p.Term.Bounds()
		sz.Width += b.Width + rootPadding.Width
		sz.Height += b.Height + rootPadding.Height
	}

	// Determine the appropriate font size so the surd is taller than the term.
	for fz := dc.o.Size; fz < 144; fz++ {
		opts := dc.o
		opts.Size = fz
		p.ff = truetype.NewFace(dc.f, &opts)
		if h := p.ff.Metrics().Height; h >= sz.Height {
			sz.Height = h
			break
		}
	}

	// Determine how many macron characters are needed for the top bar.
	mw, _, _ := p.ff.GlyphBounds(macronChar)
	p.numMacrons = int(math.Ceil(float64(sz.Width.Ceil()) / float64((mw.Max.X - mw.Min.X).Ceil())))
	p.macronWidth = fixed.Int26_6(p.numMacrons) * (mw.Max.X - mw.Min.X)
	// If the macrons are slightly larger, update the width.
	if p.macronWidth > sz.Width {
		sz.Width = p.macronWidth
	}

	// Add the widths for the surd.
	a, _ := p.ff.GlyphAdvance(surdChar)
	sz.Width += a

	sz.Height += rootMargin.Height
	sz.Width += rootMargin.Width
	p.layout = &sz
	return nil
}

// computeYAdjustment returns the vertical distance the macron needs to be moved,
// to line up with the surd glyph.
func (p *Root) computeYAdjustment() fixed.Int26_6 {
	sb, _, _ := p.ff.GlyphBounds(surdChar)
	mb, _, _ := p.ff.GlyphBounds(macronChar)
	return sb.Min.Y - mb.Min.Y - (mb.Max.Y-mb.Min.Y)/3
}

// Draw is called to render the parentheses and its contained terms.
func (p *Root) Draw(dc *DrawContext, pos fixed.Point26_6, clip image.Rectangle) error {
	m := p.ff.Metrics()
	pos.X += rootMargin.Width / 2
	pos.Y += rootMargin.Height / 2

	if rootDebug {
		// baseline (red).
		for i := 0; i < 22; i++ {
			dc.out.Set(pos.X.Floor()+i, (pos.Y + m.Ascent - m.Descent).Round(), color.RGBA{255, 0, 0, 255})
		}
		// ascent (blue).
		for i := 0; i < 22; i++ {
			dc.out.Set(pos.X.Floor()+i, (pos.Y + m.Descent).Round(), color.RGBA{0, 0, 255, 255})
		}
		// descent (green).
		for i := 0; i < 22; i++ {
			dc.out.Set(pos.X.Floor()+i, (pos.Y + m.Ascent).Round(), color.RGBA{0, 255, 0, 255})
		}
	}

	pos.Y += m.Ascent
	dr, mask, maskp, advance, _ := p.ff.Glyph(pos, surdChar)
	if rootDebug {
		draw.DrawMask(dc.out, dr.Intersect(clip), dc.fg, image.Point{}, image.NewUniform(color.RGBA{A: 120}), maskp, draw.Over)
	}
	draw.DrawMask(dc.out, dr.Intersect(clip), dc.fg, image.Point{}, mask, maskp, draw.Over)

	pos.X += advance
	p2 := pos
	p2.Y += p.computeYAdjustment()
	p2.X -= 32
	for x := 0; x < p.numMacrons; x++ {
		dr, mask, maskp, advance, _ := p.ff.Glyph(p2, macronChar)
		draw.DrawMask(dc.out, dr.Intersect(clip), dc.fg, image.Point{}, mask, maskp, draw.Over)
		p2.X += advance
	}

	pos.X += (p.macronWidth - p.Term.Bounds().Width) / 3
	pos.Y += rootPadding.Height - m.Ascent
	if p.Term != nil {
		if err := p.Term.Draw(dc, pos, clip); err != nil {
			return err
		}
		pos.X += p.Term.Bounds().Width
		pos.Y += m.Ascent
	}
	return nil
}
