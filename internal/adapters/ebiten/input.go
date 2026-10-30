package ebiten

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/csg33k/droid-tapper/internal/ports"
)

type Input struct {
	bindings map[ports.Key][]ebiten.Key
}

func defaultBindings() map[ports.Key][]ebiten.Key {
	return map[ports.Key][]ebiten.Key{
		ports.KeyLeft:     {ebiten.KeyA, ebiten.KeyArrowLeft},
		ports.KeyRight:    {ebiten.KeyD, ebiten.KeyArrowRight},
		ports.KeyUp:       {ebiten.KeyW, ebiten.KeyArrowUp},
		ports.KeyDown:     {ebiten.KeyS, ebiten.KeyArrowDown},
		ports.KeyServe:    {ebiten.KeySpace},
		ports.KeyPause:    {ebiten.KeyP},
		ports.KeyStart:    {ebiten.KeyEnter},
		ports.KeyQuit:     {ebiten.KeyEscape},
		ports.KeySettings: {ebiten.KeyO},
		ports.KeyBack:     {ebiten.KeyBackspace, ebiten.KeyEscape},
	}
}

func (i *Input) ensureBindings() {
	if i.bindings == nil {
		i.bindings = defaultBindings()
	}
}

func (i *Input) IsPressed(k ports.Key) bool {
	i.ensureBindings()
	for _, ek := range i.bindings[k] {
		if ebiten.IsKeyPressed(ek) {
			return true
		}
	}
	return false
}

func (i *Input) JustPressed(k ports.Key) bool {
	i.ensureBindings()
	for _, ek := range i.bindings[k] {
		if inpututil.IsKeyJustPressed(ek) {
			return true
		}
	}
	return false
}

func (i *Input) SetBindings(b map[ports.Key][]int) {
	i.ensureBindings()
	defaults := defaultBindings()
	for k, codes := range b {
		// start from defaults, then add custom keys (dedup)
		seen := map[ebiten.Key]bool{}
		var keys []ebiten.Key
		for _, dk := range defaults[k] {
			if !seen[dk] {
				seen[dk] = true
				keys = append(keys, dk)
			}
		}
		for _, c := range codes {
			ek := ebiten.Key(c)
			if !seen[ek] {
				seen[ek] = true
				keys = append(keys, ek)
			}
		}
		i.bindings[k] = keys
	}
}

func (i *Input) Bindings() map[ports.Key][]int {
	i.ensureBindings()
	out := map[ports.Key][]int{}
	for k, keys := range i.bindings {
		for _, ek := range keys {
			out[k] = append(out[k], int(ek))
		}
	}
	return out
}

func (i *Input) LastKeyPressed() int {
	for k := 0; k <= int(ebiten.KeyMax); k++ {
		if inpututil.IsKeyJustPressed(ebiten.Key(k)) {
			return k
		}
	}
	return -1
}
