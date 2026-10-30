package ports

// Key is a logical input key.
type Key int

const (
	KeyLeft Key = iota
	KeyRight
	KeyUp
	KeyDown
	KeyServe
	KeyPause
	KeyStart
	KeyQuit
)

// Input is the input port.
type Input interface {
	IsPressed(k Key) bool
	JustPressed(k Key) bool
}

// Sound identifiers.
type Sound int

const (
	SoundServe Sound = iota
	SoundCatch
	SoundBreak
	SoundLifeLost
	SoundLevelUp
	SoundGameOver
)

// AudioPlayer is the audio port.
type AudioPlayer interface {
	Play(s Sound)
}

// ScoreStore is the persistence port.
type ScoreStore interface {
	LoadHighScore() (int, error)
	SaveHighScore(score int) error
}
