package avatars

import (
	"fmt"
	gim "github.com/ozankasikci/go-image-merge"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

type PresenceImage struct {
	ImageName string
	Present   bool
}

func CreateAvatar(presence []PresenceImage) (string, error) {

	var grids []*gim.Grid

	for _, p := range presence {
		var c color.Color = color.White
		if p.Present == false {
			c = color.Gray{Y: 100}
		}
		grids = append(grids, &gim.Grid{
			ImageFilePath:   "./avatars/" + p.ImageName,
			BackgroundColor: c,
		})
	}
	x := len(presence)
	y := 1
	if len(presence) >= 4 {
		x = 4
	}
	if len(presence) > 4 {
		y = int(math.Ceil(float64(len(presence)) / 4))
	}
	rgba, err := gim.New(grids, x, y).Merge()

	if err != nil {
		return "", fmt.Errorf("cannot create image %w", err)
	}
	file, err := os.Create("./presence.png")
	if err != nil {
		return "", fmt.Errorf("cannot save image %w", err)
	}
	err = png.Encode(file, rgba)
	if err != nil {
		return "", fmt.Errorf("cannot encode image %w", err)
	}
	name := file.Name()
	path, err := filepath.Abs(filepath.Dir(file.Name()))
	if err != nil {
		return "", fmt.Errorf("cannot encode image %w", err)
	}
	return filepath.Join(path, name), nil
}
