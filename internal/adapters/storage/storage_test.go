package storage

import (
	"os"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	dir, _ := os.MkdirTemp("", "hs")
	defer os.RemoveAll(dir)
	s := &JSONStore{Path: dir + "/highscore.json"}
	if err := s.SaveHighScore(42); err != nil {
		t.Fatal(err)
	}
	v, err := s.LoadHighScore()
	if err != nil || v != 42 {
		t.Fatalf("got %d err %v", v, err)
	}
}
