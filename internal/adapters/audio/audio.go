package audio

import "github.com/csg33k/droid-tapper/internal/ports"

type Noop struct{}

func (Noop) Play(ports.Sound) {}
