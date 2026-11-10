package application

import (
	"github.com/csg33k/droid-tapper/internal/domain/game"
	"github.com/csg33k/droid-tapper/internal/ports"
)

type Service struct {
	Game          *game.Game
	Input         ports.Input
	Audio         ports.AudioPlayer
	Store         ports.ScoreStore
	SettingsStore ports.SettingsStore

	Settings ports.Settings

	wasGameOver  bool
	leftTapTimer float64
	leftTapped   bool

	settingsSelected int
	settingsEditing  bool
	settingsEditKey  ports.Key
	settingsReturn   game.Status
}

func NewService(g *game.Game, in ports.Input, audio ports.AudioPlayer, store ports.ScoreStore, settingsStore ports.SettingsStore) *Service {
	s := &Service{Game: g, Input: in, Audio: audio, Store: store, SettingsStore: settingsStore}
	if settingsStore != nil {
		if st, err := settingsStore.LoadSettings(); err == nil {
			s.Settings = st
		}
	}
	if s.Settings.Bindings == nil {
		s.Settings.Bindings = map[ports.Key][]int{}
	}
	if s.Settings.PowerUps == nil {
		s.Settings.PowerUps = map[int]bool{0: true, 1: true, 2: true, 3: true}
	}
	if len(s.Settings.Bindings) == 0 {
		s.Settings.SoundEnabled = true
	}
	// sync power-up enables to game config
	for k := 0; k < len(s.Game.Config.EnabledPowerUps); k++ {
		s.Game.Config.EnabledPowerUps[k] = s.Settings.PowerUps[k]
	}
	if kb, ok := in.(ports.KeyBinder); ok {
		if len(s.Settings.Bindings) > 0 {
			kb.SetBindings(s.Settings.Bindings)
		}
		s.Settings.Bindings = kb.Bindings()
	}
	return s
}

func (s *Service) Update() error {
	g := s.Game

	// Global settings hotkey
	if g.Status != game.StatusSettings && s.Input.JustPressed(ports.KeySettings) {
		s.settingsReturn = g.Status
		g.Status = game.StatusSettings
	}

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

	// Settings
	if g.Status == game.StatusSettings {
		if s.settingsEditing {
			if kb, ok := s.Input.(ports.KeyBinder); ok {
				if code := kb.LastKeyPressed(); code >= 0 {
					key := s.settingsEditKey
					s.Settings.Bindings[key] = []int{code}
					if kb2, ok2 := s.Input.(ports.KeyBinder); ok2 {
						kb2.SetBindings(s.Settings.Bindings)
					}
					s.settingsEditing = false
					s.saveSettings()
				}
			}
			return nil
		}
		if s.Input.JustPressed(ports.KeyBack) || s.Input.JustPressed(ports.KeyQuit) {
			s.saveSettings()
			if s.settingsReturn != game.StatusSettings {
				g.Status = s.settingsReturn
			} else {
				g.Status = game.StatusMenu
			}
			s.settingsReturn = game.StatusMenu
			return nil
		}
		if s.Input.JustPressed(ports.KeyUp) {
			if s.settingsSelected > 0 {
				s.settingsSelected--
			}
		}
		if s.Input.JustPressed(ports.KeyDown) {
			// 1 (sound) + number of remappable keys
			// sound + 8 remappable keys + power-ups; last index = 8 + number of power-ups
			maxIdx := 8 + len(s.Game.Config.EnabledPowerUps)
			if s.settingsSelected < maxIdx {
				s.settingsSelected++
			}
		}
		if s.settingsSelected == 0 {
			if s.Input.JustPressed(ports.KeyLeft) || s.Input.JustPressed(ports.KeyRight) || s.Input.JustPressed(ports.KeyServe) {
				s.Settings.SoundEnabled = !s.Settings.SoundEnabled
				s.saveSettings()
			}
		} else if s.settingsSelected <= 8 {
			if s.Input.JustPressed(ports.KeyServe) || s.Input.JustPressed(ports.KeyStart) {
				s.settingsEditing = true
				keys := []ports.Key{ports.KeyLeft, ports.KeyRight, ports.KeyUp, ports.KeyDown, ports.KeyServe, ports.KeyPause, ports.KeyStart, ports.KeySettings}
				s.settingsEditKey = keys[s.settingsSelected-1]
			}
		} else {
			if s.Input.JustPressed(ports.KeyLeft) || s.Input.JustPressed(ports.KeyRight) || s.Input.JustPressed(ports.KeyServe) || s.Input.JustPressed(ports.KeyStart) {
				idx := s.settingsSelected - 9
				if idx >= 0 && idx < len(s.Game.Config.EnabledPowerUps) {
					s.Settings.PowerUps[idx] = !s.Settings.PowerUps[idx]
					s.Game.Config.EnabledPowerUps[idx] = s.Settings.PowerUps[idx]
					s.saveSettings()
				}
			}
		}
		return nil
	}

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
	if s.Settings.SoundEnabled {
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
			case game.EventPowerUp:
				s.Audio.Play(ports.SoundPowerUp)
			}
		}
	}
	g.Events = nil

	return nil
}

func (s *Service) SettingsSelected() int { return s.settingsSelected }
func (s *Service) SettingsEditing() bool { return s.settingsEditing }

func (s *Service) saveSettings() {
	if kb, ok := s.Input.(ports.KeyBinder); ok {
		kb.SetBindings(s.Settings.Bindings)
		s.Settings.Bindings = kb.Bindings()
	}
	if s.SettingsStore != nil {
		_ = s.SettingsStore.SaveSettings(s.Settings)
	}
}
