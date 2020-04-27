package eqdraw

import (
	"image"

	"golang.org/x/image/math/fixed"
)

var (
	runMargin = layoutResult{
		Height: fixed.Int26_6(2 << 6),
		Width:  fixed.Int26_6(4 << 6),
	}
)

// Run represents a horizontal series of terms.
type Run struct {
	layout      *layoutResult
	tallestTerm fixed.Int26_6

	Terms []node
}

// Bounds returns the width and height of the rendered term, as computed by
// the last layout pass. If no layout pass has occurred, the returned value
// will be nil.
func (r *Run) Bounds() *layoutResult {
	return r.layout
}

// Layout is called during the layout pass to compute the rendered size of this node.
func (r *Run) Layout(dc *drawContext) error {
	sz := runMargin

	var tallestTerm fixed.Int26_6
	for _, t := range r.Terms {
		if err := t.Layout(dc); err != nil {
			return err
		}
		b := t.Bounds()
		sz.Width += b.Width
		if b.Height > tallestTerm {
			tallestTerm = b.Height
		}
	}
	r.tallestTerm = tallestTerm

	sz.Height += r.tallestTerm
	r.layout = &sz
	return nil
}

// Draw is called to render the series of terms.
func (r *Run) Draw(dc *drawContext, pos fixed.Point26_6, clip image.Rectangle) error {
	pos.X += runMargin.Width / 2
	pos.Y += runMargin.Height / 2

	for _, t := range r.Terms {
		sz := t.Bounds()
		adjustY := (r.tallestTerm - sz.Height) / 2
		pos.Y += adjustY
		if err := t.Draw(dc, pos, clip); err != nil {
			return err
		}
		pos.Y -= adjustY
		pos.X += sz.Width
	}

	return nil
}
