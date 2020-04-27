package eqdraw

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/golang/freetype/truetype"
)

func findFontPath(base string) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	for _, p := range []string{
		"",
		"liberationsans",
		"/usr/share/eqdraw/fonts",
		"/usr/share/fonts/truetype/liberation",
		"/usr/local/share/fonts/truetype/liberation",
		"~/.fonts",
	} {
		p = filepath.Join(strings.Replace(p, "~", u.HomeDir, -1), base)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", errors.New("could not find font: " + base)
}

// DefaultFont returns the default font to use.
func DefaultFont() (*truetype.Font, error) {
	return loadFont("LiberationSans-Regular.ttf")
}

func loadFont(f string) (*truetype.Font, error) {
	p, err := findFontPath(f)
	if err != nil {
		return nil, fmt.Errorf("finding font: %w", err)
	}
	d, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading font: %w", err)
	}
	return truetype.Parse(d)
}
