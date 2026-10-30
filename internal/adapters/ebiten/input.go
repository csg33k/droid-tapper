package ebiten

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/csg33k/droid-tapper/internal/ports"
)

type Input struct{}

var keyMap = map[ports.Key][]ebiten.Key{
	ports.KeyLeft:  {ebiten.KeyN, ebiten.KeyArrowLeft, ebiten.KeyA},
	ports.KeyRight: {ebiten.KeyT, ebiten.KeyArrowRight, ebiten.KeyD},
	ports.KeyUp:    {ebiten.KeyL, ebiten.KeyArrowUp, ebiten.KeyW},
	ports.KeyDown:  {ebiten.KeyR, ebiten.KeyArrowDown, ebiten.KeyS},
	ports.KeyServe: {ebiten.KeySpace},
	ports.KeyPause: {ebiten.KeyP},
	ports.KeyStart: {ebiten.KeyEnter},
	ports.KeyQuit:  {ebiten.KeyEscape},
}

func (Input) IsPressed(k ports.Key) bool {
	for _, ek := range keyMap[k] {
		if ebiten.IsKeyPressed(ek) {
			return true
		}
	}
	return false
}

func (Input) JustPressed(k ports.Key) bool {
	for _, ek := range keyMap[k] {
		if inpututil.IsKeyJustPressed(ek) {
			return true
		}
	}
	return false
}
