// Package eqdraw renders mathematical equations.
package eqdraw

import (
	"image"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type layoutResult struct {
	Width, Height fixed.Int26_6
}

type node interface {
	// Layout walks the node tree, recursively computing the sizes of each node.
	Layout(*drawContext) error
	// Draw draws the node using the provided information and drawContext.
	Draw(dc *drawContext, pos fixed.Point26_6, clip image.Rectangle) error
	// Bounds returns the width and height of the rendered term, as computed by
	// the last layout pass. If no layout pass has occurred, the returned value
	// will be nil.
	Bounds() *layoutResult
}

type drawContext struct {
	ff  font.Face
	f   *truetype.Font
	out *image.RGBA
}
