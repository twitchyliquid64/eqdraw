package eqdraw

import (
	"image"
	"image/color"

	"golang.org/x/image/math/fixed"
)

var (
	divMargin = layoutResult{
		Height: fixed.Int26_6(8 << 6),
		Width:  fixed.Int26_6(2 << 6),
	}
	divLineThickness = 2
	divLineSpacing   = 4
)

// Div represents one term dividing another
type Div struct {
	layout *layoutResult

	Numerator   node
	Denominator node
}

// Bounds returns the width and height of the rendered term, as computed by
// the last layout pass. If no layout pass has occurred, the returned value
// will be nil.
func (d *Div) Bounds() *layoutResult {
	return d.layout
}

// Layout is called during the layout pass to compute the rendered size of this node.
func (d *Div) Layout(dc *drawContext) error {
	sz := divMargin
	sz.Height += fixed.I(divLineThickness) + fixed.I(divLineSpacing*2)

	if err := d.Numerator.Layout(dc); err != nil {
		return err
	}
	if err := d.Denominator.Layout(dc); err != nil {
		return err
	}
	nb := d.Numerator.Bounds()
	sz.Height += nb.Height
	db := d.Denominator.Bounds()
	sz.Height += db.Height

	if nb.Width > db.Width {
		sz.Width += nb.Width
	} else {
		sz.Width += db.Width
	}

	d.layout = &sz
	return nil
}

// Draw is called to render the parentheses and its contained terms.
func (d *Div) Draw(dc *drawContext, pos fixed.Point26_6, clip image.Rectangle) error {
	pos.Y += divMargin.Height / 2

	nb := d.Numerator.Bounds()
	adjX := (d.layout.Width - nb.Width + 1) / 2
	pos.X += adjX
	if err := d.Numerator.Draw(dc, pos, clip); err != nil {
		return err
	}
	pos.X -= adjX
	pos.Y += nb.Height + fixed.I(divLineSpacing)

	for x := 1; x < d.layout.Width.Ceil()-2; x++ {
		for y := 0; y < divLineThickness; y++ {
			dc.out.Set(x, pos.Y.Round()+y, color.Black)
		}
	}

	pos.Y += fixed.I(divLineThickness) + fixed.I(divLineSpacing)
	db := d.Denominator.Bounds()
	adjX = (d.layout.Width - db.Width + 1) / 2
	pos.X += adjX
	if err := d.Denominator.Draw(dc, pos, clip); err != nil {
		return err
	}

	return nil
}
