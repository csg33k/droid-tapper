package main

import (
	"log"
	"time"

	ebiten "github.com/hajimehoshi/ebiten/v2"

	"github.com/csg33k/droid-tapper/assets"
	"github.com/csg33k/droid-tapper/internal/adapters/audio"
	ebitenadapter "github.com/csg33k/droid-tapper/internal/adapters/ebiten"
	"github.com/csg33k/droid-tapper/internal/adapters/rng"
	"github.com/csg33k/droid-tapper/internal/adapters/storage"
	"github.com/csg33k/droid-tapper/internal/application"
	"github.com/csg33k/droid-tapper/internal/domain/game"
)

// assets embedded via assets package

func main() {
	store := storage.NewDefault()
	hs, _ := store.LoadHighScore()

	r := rng.New(time.Now().UnixNano())
	cfg := game.DefaultConfig()
	g := game.NewGame(r, cfg, hs)

	svc := application.NewService(g, ebitenadapter.Input{}, audio.Noop{}, store)
	ebg := ebitenadapter.New(svc)
	ebg.Renderer.Assets = ebitenadapter.LoadAssetsFS(assets.FS, "sprites")

	screenW, screenH := ebiten.ScreenSizeInFullscreen()
	// choose scale to fit within 80% of screen, preserving 4:3 aspect
	var winW, winH int
	if float64(screenW)*0.8/1600.0 < float64(screenH)*0.8/1200.0 {
		winW = int(float64(screenW) * 0.8)
		winH = int(float64(winW) * 1200.0 / 1600.0)
	} else {
		winH = int(float64(screenH) * 0.8)
		winW = int(float64(winH) * 1600.0 / 1200.0)
	}
	ebiten.SetWindowSize(winW, winH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Droid Tapper")
	if err := ebiten.RunGame(ebg); err != nil {
		log.Fatal(err)
	}
}
