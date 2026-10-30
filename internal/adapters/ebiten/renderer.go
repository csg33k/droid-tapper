package ebiten

import (
	"fmt"
	"image/color"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"

	"github.com/csg33k/droid-tapper/internal/domain/game"
)

var (
	fontCache  = map[float64]font.Face{}
	fontMu     sync.Mutex
	parsedFont *opentype.Font
	parseErr   error
	parseOnce  sync.Once
)

func getFont(scale float64) font.Face {
	size := 12.0 * scale
	if size < 12 {
		size = 12
	}
	parseOnce.Do(func() {
		parsedFont, parseErr = opentype.Parse(goregular.TTF)
	})
	if parseErr != nil {
		return nil
	}
	fontMu.Lock()
	defer fontMu.Unlock()
	if f, ok := fontCache[size]; ok {
		return f
	}
	face, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{Size: size, DPI: 72})
	if err != nil {
		return nil
	}
	fontCache[size] = face
	return face
}

func drawText(screen *ebiten.Image, msg string, x, y int, scale float64) {
	drawTextColored(screen, msg, x, y, scale, color.White)
}

func drawTextColored(screen *ebiten.Image, msg string, x, y int, scale float64, col color.Color) {
	face := getFont(scale)
	if face == nil {
		// fallback
		img := ebiten.NewImage(4000, 128)
		ebitenutil.DebugPrintAt(img, msg, 0, 0)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(x), float64(y))
		op.ColorScale.ScaleWithColor(col)
		screen.DrawImage(img, op)
		return
	}
	text.Draw(screen, msg, face, x, y+int(12*scale), col)
}

var (
	colorBg      = color.RGBA{R: 20, G: 16, B: 24, A: 255}
	colorBar     = color.RGBA{R: 60, G: 40, B: 30, A: 255}
	colorTap     = color.RGBA{R: 120, G: 90, B: 60, A: 255}
	colorBartend = color.RGBA{R: 80, G: 180, B: 220, A: 255}
	colorCust    = color.RGBA{R: 220, G: 120, B: 80, A: 255}
	colorMug     = color.RGBA{R: 240, G: 230, B: 180, A: 255}
	colorEmpty   = color.RGBA{R: 200, G: 200, B: 200, A: 255}
	colorText    = color.RGBA{R: 230, G: 230, B: 230, A: 255}
)

type Renderer struct {
	Assets     *Assets
	floorImage *ebiten.Image
}

func (r *Renderer) Draw(screen *ebiten.Image, g *game.Game) {
	screen.Fill(colorBg)
	cfg := g.Config

	// Floor texture: subtle binary pattern (cached)
	if r.floorImage == nil {
		r.floorImage = ebiten.NewImage(int(cfg.ScreenW), int(cfg.ScreenH))
		phrase := "golang rocks!!"
		var hexes []string
		for _, b := range []byte(phrase) {
			hexes = append(hexes, fmt.Sprintf("%02x", b))
		}
		hexLine := strings.Join(hexes, " ")
		fullLine := strings.Repeat(hexLine+"  ", 12)
		for y := 0; y < int(cfg.ScreenH); y += 20 {
			drawTextColored(r.floorImage, fullLine, 0, y, 1.4, color.RGBA{R: 30, G: 80, B: 90, A: 50})
		}
	}
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(r.floorImage, op)

	// Exit drape on the right, top to bottom
	drapeW := 140
	drapeX := int(cfg.ScreenW) - drapeW
	ebitenutil.DrawRect(screen, float64(drapeX), 0, float64(drapeW), cfg.ScreenH, color.RGBA{R: 60, G: 20, B: 30, A: 200})
	// folds
	for i := 0; i < 4; i++ {
		ebitenutil.DrawRect(screen, float64(drapeX+i*35), 0, 2, cfg.ScreenH, color.RGBA{R: 40, G: 10, B: 20, A: 180})
	}
	drawTextColored(screen, "EXIT", drapeX+20, int(cfg.ScreenH)/2-40, 4, color.RGBA{R: 80, G: 255, B: 120, A: 255})

	// HUD
	drawText(screen, fmt.Sprintf("Score: %d   Level: %d   Lives: %d   High: %d", g.Score, g.Level, g.Lives, g.HighScore), 20, 20, 3)

	// Bars
	for _, bar := range g.Bars {
		if r.Assets != nil && r.Assets.BarTexture != nil {
			op := &ebiten.DrawImageOptions{}
			w, h := r.Assets.BarTexture.Size()
			op.GeoM.Scale(cfg.BarLength/float64(w), 32.0/float64(h))
			op.GeoM.Translate(40, bar.Y-16)
			screen.DrawImage(r.Assets.BarTexture, op)
		} else {
			ebitenutil.DrawRect(screen, 40, bar.Y-16, cfg.BarLength, 32, colorBar)
		}
		// tap marker
		ebitenutil.DrawRect(screen, cfg.BarTapX-12, bar.Y-28, 24, 56, colorTap)
		// exit handled by full-height drape on the right
	}

	// Customers
	for _, c := range g.Customers {
		barY := g.Bars[c.BarID].Y
		if spr := r.Assets.customerFor(c.ID); spr != nil {
			tints := []color.Color{
				color.White,
				color.RGBA{R: 255, G: 200, B: 200, A: 255},
				color.RGBA{R: 200, G: 255, B: 200, A: 255},
				color.RGBA{R: 200, G: 200, B: 255, A: 255},
				color.RGBA{R: 255, G: 255, B: 200, A: 255},
				color.RGBA{R: 255, G: 200, B: 255, A: 255},
			}
			drawSpriteTinted(screen, spr, c.X, barY, 112, tints[c.ID%len(tints)])
		} else {
			col := colorCust
			if c.State == game.CustomerSatisfied {
				col = color.RGBA{R: 120, G: 220, B: 120, A: 255}
			}
			ebitenutil.DrawRect(screen, c.X-32, barY-40, 64, 80, col)
		}
	}

	// Mugs
	for _, m := range g.Mugs {
		barY := g.Bars[m.BarID].Y
		if m.Filled && r.Assets != nil && r.Assets.Mug != nil {
			drawSprite(screen, r.Assets.Mug, m.X, barY, 64)
		} else if !m.Filled && r.Assets != nil && r.Assets.EmptyMug != nil {
			drawSprite(screen, r.Assets.EmptyMug, m.X, barY, 64)
		} else {
			col := colorMug
			if !m.Filled {
				col = colorEmpty
			}
			ebitenutil.DrawRect(screen, m.X-16, barY-16, 32, 32, col)
		}
	}

	// Bartender
	b := g.Bartender
	barY := g.Bars[b.BarID].Y
	if r.Assets != nil && r.Assets.Bartender != nil {
		drawSprite(screen, r.Assets.Bartender, b.X, barY, 128)
	} else {
		ebitenutil.DrawRect(screen, b.X-40, barY-48, 80, 96, colorBartend)
	}

	// Level start overlay
	if g.LevelStartTimer > 0 {
		drawText(screen, fmt.Sprintf("LEVEL %d", g.Level), int(cfg.ScreenW)/2-160, 120, 6)
	}

	// Overlays
	switch g.Status {
	case game.StatusMenu:
		drawText(screen, "DROID TAPPER", 600, 420, 4)
		drawText(screen, "Press ENTER or SPACE to start", 560, 560, 3)
		drawText(screen, "L=up N=left R=down T=right  Space=serve  P=pause", 360, 680, 2)
	case game.StatusPaused:
		drawText(screen, "PAUSED (P to resume)", 600, 560, 3)
	case game.StatusGameOver:
		drawText(screen, "GAME OVER", 640, 420, 4)
		drawText(screen, "Press ENTER or SPACE to restart", 560, 560, 3)
	}
}
