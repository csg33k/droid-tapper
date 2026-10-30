package ebiten

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/csg33k/droid-tapper/internal/application"
)

type Game struct {
	Service  *application.Service
	Renderer Renderer
	Width    int
	Height   int
}

func New(svc *application.Service) *Game {
	return &Game{Service: svc, Renderer: Renderer{}, Width: 1600, Height: 1200}
}

func (g *Game) Update() error {
	return g.Service.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Renderer.Draw(screen, g.Service.Game)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.Width, g.Height
}
