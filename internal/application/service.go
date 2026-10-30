package application

import (
	"github.com/csg33k/droid-tapper/internal/domain/game"
	"github.com/csg33k/droid-tapper/internal/ports"
)

type Service struct {
	Game  *game.Game
	Input ports.Input
	Audio ports.AudioPlayer
	Store ports.ScoreStore

	wasGameOver  bool
	leftTapTimer float64
	leftTapped   bool
}

func NewService(g *game.Game, in ports.Input, audio ports.AudioPlayer, store ports.ScoreStore) *Service {
	return &Service{Game: g, Input: in, Audio: audio, Store: store}
}

func (s *Service) Update() error {
	g := s.Game

	// Menu
	if g.Status == game.StatusMenu {
		if s.Input.JustPressed(ports.KeyStart) || s.Input.JustPressed(ports.KeyServe) {
			g.Start()
		}
		return nil
	}

	// Game over -> allow restart
	if g.Status == game.StatusGameOver {
		s.leftTapped = false
		if s.Input.JustPressed(ports.KeyStart) || s.Input.JustPressed(ports.KeyServe) {
			g.Start()
		}
		// persist high score once
		if !s.wasGameOver {
			s.wasGameOver = true
			if g.HighScore > 0 {
				_ = s.Store.SaveHighScore(g.HighScore)
			}
		}
		return nil
	}
	s.wasGameOver = false

	// Paused
	if g.Status == game.StatusPaused {
		if s.Input.JustPressed(ports.KeyPause) {
			g.TogglePause()
		}
		return nil
	}

	// Playing input
	if s.Input.JustPressed(ports.KeyPause) {
		g.TogglePause()
		s.leftTapped = false
		return nil
	}
	if s.Input.JustPressed(ports.KeyUp) {
		g.MoveVertical(-1)
	}
	if s.Input.JustPressed(ports.KeyDown) {
		g.MoveVertical(1)
	}
	if s.Input.JustPressed(ports.KeyServe) {
		g.Serve()
	}
	if s.Input.JustPressed(ports.KeyLeft) {
		if s.leftTapped && s.leftTapTimer < 0.35 {
			g.ReturnToTap()
			s.leftTapped = false
			s.leftTapTimer = 0
		} else {
			s.leftTapped = true
			s.leftTapTimer = 0
		}
	}
	if s.leftTapped {
		s.leftTapTimer += 1.0 / 60.0
		if s.leftTapTimer > 0.35 {
			s.leftTapped = false
		}
	}

	dir := 0
	if s.Input.IsPressed(ports.KeyLeft) {
		dir -= 1
	}
	if s.Input.IsPressed(ports.KeyRight) {
		dir += 1
	}
	g.SetHorizontalIntent(dir)

	// Tick
	g.Tick(1.0 / 60.0)

	// Audio + events
	for _, e := range g.Events {
		switch e {
		case game.EventServe:
			s.Audio.Play(ports.SoundServe)
		case game.EventCatch:
			s.Audio.Play(ports.SoundCatch)
		case game.EventBreak:
			s.Audio.Play(ports.SoundBreak)
		case game.EventLifeLost:
			s.Audio.Play(ports.SoundLifeLost)
		case game.EventLevelUp:
			s.Audio.Play(ports.SoundLevelUp)
		case game.EventGameOver:
			s.Audio.Play(ports.SoundGameOver)
		case game.EventExcited:
			s.Audio.Play(ports.SoundExcited)
		}
	}
	g.Events = nil

	return nil
}
