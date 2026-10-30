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
	KeySettings
	KeyBack
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
	SoundExcited
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

// Settings holds user preferences.
type Settings struct {
	SoundEnabled bool
	Bindings     map[Key][]int
}

// SettingsStore is the settings persistence port.
type SettingsStore interface {
	LoadSettings() (Settings, error)
	SaveSettings(Settings) error
}

// KeyBinder allows rebinding keys at runtime.
type KeyBinder interface {
	SetBindings(map[Key][]int)
	Bindings() map[Key][]int
	LastKeyPressed() int
}
