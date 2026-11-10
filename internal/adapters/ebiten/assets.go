package ebiten

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
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
	PowerJS      *ebiten.Image
	PowerTS      *ebiten.Image
	PowerCrab    *ebiten.Image
	PowerNuke    *ebiten.Image
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
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".svg" {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		var img *ebiten.Image
		if ext == ".svg" {
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			img = rasterizeSVG(data, 1254)
			if img == nil {
				continue
			}
		} else {
			loaded, _, err := ebitenutil.NewImageFromFile(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			img = loaded
		}
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
		case strings.HasPrefix(base, "penguin"):
			a.Customers = append(a.Customers, img)
		}
	}
	a.PowerJS, a.PowerTS, a.PowerCrab, a.PowerNuke = generatePowerIcons()
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
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".svg" {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		var ei *ebiten.Image
		if ext == ".svg" {
			f, err := fsys.Open(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			data, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				continue
			}
			img := rasterizeSVG(data, 1254)
			if img == nil {
				continue
			}
			ei = img
		} else {
			f, err := fsys.Open(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			img, _, err := image.Decode(f)
			f.Close()
			if err != nil {
				continue
			}
			ei = ebiten.NewImageFromImage(img)
		}
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
		case strings.HasPrefix(base, "penguin"):
			a.Customers = append(a.Customers, ei)
		}
	}
	a.PowerJS, a.PowerTS, a.PowerCrab, a.PowerNuke = generatePowerIcons()
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

func rasterizeSVG(data []byte, size int) *ebiten.Image {
	icon, err := oksvg.ReadIconStream(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
	if w <= 0 || h <= 0 {
		w, h = size, size
	}
	if size > 0 {
		// preserve aspect ratio within size x size
		if w >= h {
			h = int(float64(h) * float64(size) / float64(w))
			w = size
		} else {
			w = int(float64(w) * float64(size) / float64(h))
			h = size
		}
		if w <= 0 {
			w = size
		}
		if h <= 0 {
			h = size
		}
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	icon.SetTarget(0, 0, float64(w), float64(h))
	icon.Draw(raster, 1.0)
	return ebiten.NewImageFromImage(img)
}

func iconFace() font.Face {
	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: 28, DPI: 72})
	if err != nil {
		return nil
	}
	return face
}

func generatePowerIcons() (js, ts, crab, nuke *ebiten.Image) {
	yellow := color.RGBA{R: 247, G: 223, B: 30, A: 255}
	blue := color.RGBA{R: 49, G: 120, B: 198, A: 255}
	red := color.RGBA{R: 222, G: 60, B: 50, A: 255}

	js = ebiten.NewImage(64, 64)
	ebitenutil.DrawRect(js, 8, 8, 48, 48, yellow)
	vector.StrokeRect(js, 8, 8, 48, 48, 2, color.Black, false)
	if face := iconFace(); face != nil {
		text.Draw(js, "JS", face, 14, 44, color.Black)
	}

	ts = ebiten.NewImage(64, 64)
	ebitenutil.DrawRect(ts, 8, 8, 48, 48, blue)
	vector.StrokeRect(ts, 8, 8, 48, 48, 2, color.White, false)
	if face := iconFace(); face != nil {
		text.Draw(ts, "TS", face, 14, 44, color.White)
	}

	crab = ebiten.NewImage(64, 64)
	// body
	vector.DrawFilledCircle(crab, 32, 36, 16, red, false)
	// claws
	vector.DrawFilledCircle(crab, 11, 30, 7, red, false)
	vector.DrawFilledCircle(crab, 53, 30, 7, red, false)
	// eyes
	vector.DrawFilledCircle(crab, 24, 23, 5, red, false)
	vector.DrawFilledCircle(crab, 40, 23, 5, red, false)
	vector.DrawFilledCircle(crab, 24, 23, 2, color.Black, false)
	vector.DrawFilledCircle(crab, 40, 23, 2, color.Black, false)
	// legs
	for i := 0; i < 3; i++ {
		y := float32(34 + i*7)
		vector.StrokeLine(crab, 18, y, 8, y+6, 3, red, false)
		vector.StrokeLine(crab, 46, y, 56, y+6, 3, red, false)
	}

	nuke = ebiten.NewImage(64, 64)
	// yellow radiation trefoil on a dark disc
	vector.DrawFilledCircle(nuke, 32, 32, 26, color.RGBA{R: 250, G: 210, B: 40, A: 255}, false)
	black := color.RGBA{R: 20, G: 20, B: 20, A: 255}
	for i := 0; i < 3; i++ {
		ang := float64(i)*2*math.Pi/3 - math.Pi/2
		cx := float32(32 + 13*math.Cos(ang))
		cy := float32(32 + 13*math.Sin(ang))
		vector.DrawFilledCircle(nuke, cx, cy, 9, black, false)
	}
	vector.DrawFilledCircle(nuke, 32, 32, 6, black, false)

	return js, ts, crab, nuke
}
