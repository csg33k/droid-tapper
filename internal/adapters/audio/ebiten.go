package audio

import (
	"github.com/hajimehoshi/ebiten/v2/audio"

	"github.com/csg33k/droid-tapper/internal/ports"
)

type EbitenAudio struct {
	ctx     *audio.Context
	players map[ports.Sound]*audio.Player
}

func New() *EbitenAudio {
	ctx := audio.NewContext(sampleRate)
	a := &EbitenAudio{ctx: ctx, players: map[ports.Sound]*audio.Player{}}

	sounds := map[ports.Sound][]byte{
		ports.SoundBreak:   makeWAV(shatterSamples()),
		ports.SoundExcited: makeWAV(excitedSamples()),
		ports.SoundLevelUp: makeWAV(successSamples()),
	}
	// reuse shatter for life lost maybe
	sounds[ports.SoundLifeLost] = sounds[ports.SoundBreak]
	// simple click for serve/catch
	sounds[ports.SoundServe] = makeWAV(serveSamples())
	sounds[ports.SoundCatch] = makeWAV(catchSamples())
	sounds[ports.SoundGameOver] = makeWAV(shatterSamples())

	for s, data := range sounds {
		p := ctx.NewPlayerFromBytes(data)
		if p != nil {
			a.players[s] = p
		}
	}
	return a
}

func (a *EbitenAudio) Play(s ports.Sound) {
	if p, ok := a.players[s]; ok {
		_ = p.Rewind()
		p.Play()
	}
}
