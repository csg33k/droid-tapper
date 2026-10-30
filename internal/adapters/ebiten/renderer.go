package ebiten

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/csg33k/droid-tapper/internal/domain/game"
)

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
	Assets *Assets
}

func (r Renderer) Draw(screen *ebiten.Image, g *game.Game) {
	screen.Fill(colorBg)
	cfg := g.Config

	// HUD
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d   Level: %d   Lives: %d   High: %d", g.Score, g.Level, g.Lives, g.HighScore), 20, 20)

	// Bars
	for _, bar := range g.Bars {
		ebitenutil.DrawRect(screen, 40, bar.Y-16, cfg.BarLength, 32, colorBar)
		// tap marker
		ebitenutil.DrawRect(screen, cfg.BarTapX-12, bar.Y-28, 24, 56, colorTap)
	}

	// Customers
	for _, c := range g.Customers {
		barY := g.Bars[c.BarID].Y
		if spr := r.Assets.customerFor(c.ID); spr != nil {
			drawSprite(screen, spr, c.X, barY, 80)
		} else {
			col := colorCust
			if c.State == game.CustomerSatisfied {
				col = color.RGBA{R: 120, G: 220, B: 120, A: 255}
			}
			ebitenutil.DrawRect(screen, c.X-24, barY-28, 48, 56, col)
		}
	}

	// Mugs
	for _, m := range g.Mugs {
		barY := g.Bars[m.BarID].Y
		if m.Filled && r.Assets != nil && r.Assets.Mug != nil {
			drawSprite(screen, r.Assets.Mug, m.X, barY, 48)
		} else if !m.Filled && r.Assets != nil && r.Assets.EmptyMug != nil {
			drawSprite(screen, r.Assets.EmptyMug, m.X, barY, 48)
		} else {
			col := colorMug
			if !m.Filled {
				col = colorEmpty
			}
			ebitenutil.DrawRect(screen, m.X-12, barY-12, 24, 24, col)
		}
	}

	// Bartender
	b := g.Bartender
	barY := g.Bars[b.BarID].Y
	if r.Assets != nil && r.Assets.Bartender != nil {
		drawSprite(screen, r.Assets.Bartender, b.X, barY, 96)
	} else {
		ebitenutil.DrawRect(screen, b.X-28, barY-36, 56, 72, colorBartend)
	}

	// Overlays
	switch g.Status {
	case game.StatusMenu:
		ebitenutil.DebugPrintAt(screen, "DROID TAPPER", 720, 500)
		ebitenutil.DebugPrintAt(screen, "Press ENTER or SPACE to start", 650, 580)
		ebitenutil.DebugPrintAt(screen, "L=up N=left R=down T=right  Space=serve  P=pause", 540, 660)
	case game.StatusPaused:
		ebitenutil.DebugPrintAt(screen, "PAUSED (P to resume)", 690, 580)
	case game.StatusGameOver:
		ebitenutil.DebugPrintAt(screen, "GAME OVER", 730, 500)
		ebitenutil.DebugPrintAt(screen, "Press ENTER or SPACE to restart", 650, 580)
	}
}
