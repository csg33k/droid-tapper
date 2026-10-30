package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/csg33k/droid-tapper/internal/ports"
)

type JSONStore struct {
	Path         string
	SettingsPath string
}

func NewDefault() *JSONStore {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return &JSONStore{
		Path:         filepath.Join(dir, "droid-tapper", "highscore.json"),
		SettingsPath: filepath.Join(dir, "droid-tapper", "settings.json"),
	}
}

func (s *JSONStore) LoadHighScore() (int, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	var v struct {
		HighScore int `json:"high_score"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return 0, err
	}
	return v.HighScore, nil
}

func (s *JSONStore) SaveHighScore(score int) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(struct {
		HighScore int `json:"high_score"`
	}{HighScore: score}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0o644)
}

func (s *JSONStore) LoadSettings() (ports.Settings, error) {
	data, err := os.ReadFile(s.SettingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ports.Settings{SoundEnabled: true, Bindings: map[ports.Key][]int{}}, nil
		}
		return ports.Settings{}, err
	}
	var v struct {
		SoundEnabled bool          `json:"sound_enabled"`
		Bindings     map[int][]int `json:"bindings"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return ports.Settings{}, err
	}
	bindings := map[ports.Key][]int{}
	for k, codes := range v.Bindings {
		bindings[ports.Key(k)] = codes
	}
	return ports.Settings{SoundEnabled: v.SoundEnabled, Bindings: bindings}, nil
}

func (s *JSONStore) SaveSettings(st ports.Settings) error {
	if err := os.MkdirAll(filepath.Dir(s.SettingsPath), 0o755); err != nil {
		return err
	}
	bindings := map[int][]int{}
	for k, codes := range st.Bindings {
		bindings[int(k)] = codes
	}
	data, err := json.MarshalIndent(struct {
		SoundEnabled bool          `json:"sound_enabled"`
		Bindings     map[int][]int `json:"bindings"`
	}{SoundEnabled: st.SoundEnabled, Bindings: bindings}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.SettingsPath, data, 0o644)
}
