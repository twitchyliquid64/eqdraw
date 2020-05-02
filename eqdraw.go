// Package eqdraw renders mathematical equations.
package eqdraw

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type layoutResult struct {
	Width, Height fixed.Int26_6
}

type node interface {
	// Layout walks the node tree, recursively computing the sizes of each node.
	Layout(*DrawContext) error
	// Draw draws the node using the provided information and drawContext.
	Draw(dc *DrawContext, pos fixed.Point26_6, clip image.Rectangle) error
	// Bounds returns the width and height of the rendered term, as computed by
	// the last layout pass. If no layout pass has occurred, the returned value
	// will be nil.
	Bounds() *layoutResult
}

// DrawContext represents a context that can be used for generating
// equation renders.
type DrawContext struct {
	o   truetype.Options
	ff  font.Face
	ffi font.Face // italic font face
	f   *truetype.Font
	out *image.RGBA
}

// NewContext creates a new drawing context.
func NewContext(o truetype.Options) (*DrawContext, error) {
	f, err := DefaultFontRegular()
	if err != nil {
		return nil, err
	}

	ff := truetype.NewFace(f, &o)

	fi, err := DefaultFontItalic()
	if err != nil {
		return nil, err
	}
	ffi := truetype.NewFace(fi, &o)

	return &DrawContext{
		o:   o,
		f:   f,
		ff:  ff,
		ffi: ffi,
	}, nil
}

// DrawRGBA generates a RGBA image by drawing the given node. If uniform
// is non-nil, it will be drawn over the entire image before rendering
// the equation.
func (dc *DrawContext) DrawRGBA(n node, uniform *image.Uniform) (*image.RGBA, error) {
	if err := n.Layout(dc); err != nil {
		return nil, fmt.Errorf("layout: %w", err)
	}
	bounds := image.Rectangle{Max: image.Point{X: n.Bounds().Width.Ceil(), Y: n.Bounds().Height.Ceil()}}

	canvas := image.NewRGBA(bounds)
	if uniform != nil {
		draw.Draw(canvas, bounds, uniform, image.Point{}, draw.Over)
	}
	dc.out = canvas
	err := n.Draw(dc, fixed.Point26_6{}, bounds)
	if err != nil {
		return nil, fmt.Errorf("draw: %w", err)
	}
	dc.out = nil
	return canvas, nil
}
