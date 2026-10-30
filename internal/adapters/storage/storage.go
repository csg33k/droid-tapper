package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type JSONStore struct {
	Path string
}

func NewDefault() *JSONStore {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return &JSONStore{Path: filepath.Join(dir, "droid-tapper", "highscore.json")}
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
