package ebiten

import (
	"image"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Assets holds loaded sprites.
type Assets struct {
	Bartender    *ebiten.Image
	Customers    []*ebiten.Image
	Mug          *ebiten.Image
	EmptyMug     *ebiten.Image
	Exit         *ebiten.Image
	BarTexture   *ebiten.Image
	FloorTexture *ebiten.Image
	ExitDrape    *ebiten.Image
}

// LoadAssets loads sprites from dir. Missing files are silently skipped.
func LoadAssets(dir string) *Assets {
	a := &Assets{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return a
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
			continue
		}
		img, _, err := ebitenutil.NewImageFromFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		switch {
		case base == "bartender":
			a.Bartender = img
		case strings.HasPrefix(base, "customer"):
			a.Customers = append(a.Customers, img)
		case base == "mug_empty":
			a.EmptyMug = img
		case base == "mug":
			a.Mug = img
		case base == "exit":
			a.Exit = img
		case base == "bar_texture":
			a.BarTexture = img
		}
	}
	a.Customers = append(a.Customers, generatePenguins()...)
	return a
}

// LoadAssetsFS loads sprites from an embedded filesystem.
func LoadAssetsFS(fsys fs.FS, dir string) *Assets {
	a := &Assets{}
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return a
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
			continue
		}
		f, err := fsys.Open(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			continue
		}
		ei := ebiten.NewImageFromImage(img)
		base := strings.TrimSuffix(name, filepath.Ext(name))
		switch {
		case base == "bartender":
			a.Bartender = ei
		case strings.HasPrefix(base, "customer"):
			a.Customers = append(a.Customers, ei)
		case base == "mug_empty":
			a.EmptyMug = ei
		case base == "mug":
			a.Mug = ei
		case base == "exit":
			a.Exit = ei
		case base == "bar_texture":
			a.BarTexture = ei
		}
	}
	a.Customers = append(a.Customers, generatePenguins()...)
	return a
}

func (a *Assets) customerFor(id int) *ebiten.Image {
	if a == nil || len(a.Customers) == 0 {
		return nil
	}
	return a.Customers[id%len(a.Customers)]
}

func drawSprite(screen *ebiten.Image, img *ebiten.Image, cx, cy, size float64) {
	drawSpriteTinted(screen, img, cx, cy, size, color.White)
}

func drawSpriteTinted(screen *ebiten.Image, img *ebiten.Image, cx, cy, size float64, tint color.Color) {
	if img == nil {
		return
	}
	w, h := img.Size()
	scale := size / float64(w)
	if float64(h)*scale > size {
		scale = size / float64(h)
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(cx-size/2, cy-size/2)
	op.ColorScale.ScaleWithColor(tint)
	screen.DrawImage(img, op)
}

func generatePenguins() []*ebiten.Image {
	var out []*ebiten.Image
	for i := 0; i < 3; i++ {
		img := ebiten.NewImage(64, 64)
		// body
		ebitenutil.DrawRect(img, 16, 8, 32, 48, color.Black)
		// belly
		ebitenutil.DrawRect(img, 24, 24, 16, 24, color.White)
		// eyes
		vector.DrawFilledCircle(img, 24, 18, 3, color.White, false)
		vector.DrawFilledCircle(img, 40, 18, 3, color.White, false)
		// beak
		ebitenutil.DrawRect(img, 28, 24, 8, 4, color.RGBA{R: 255, G: 165, B: 0, A: 255})
		// variants
		if i == 1 {
			ebitenutil.DrawRect(img, 16, 8, 32, 6, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		} else if i == 2 {
			ebitenutil.DrawRect(img, 18, 4, 28, 6, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		}
		out = append(out, img)
	}
	return out
}
