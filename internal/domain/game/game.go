package game

import "math"

// RNG is the randomness port used by the domain.
type RNG interface {
	Intn(n int) int
	Float64() float64
}

// Status of the game.
type Status int

const (
	StatusMenu Status = iota
	StatusPlaying
	StatusPaused
	StatusGameOver
	StatusSettings
)

// Event emitted by the domain for adapters (audio, etc.).
type Event int

const (
	EventNone Event = iota
	EventServe
	EventCatch
	EventBreak
	EventLifeLost
	EventLevelUp
	EventGameOver
	EventExcited
)

type CustomerState int

const (
	CustomerApproaching CustomerState = iota
	CustomerSatisfied
)

type Config struct {
	ScreenW, ScreenH  float64
	BarCount          int
	BarLength         float64
	BarTapX           float64
	BartenderSpeed    float64
	CustomerSpeed     float64 // base approach speed
	SatisfiedSpeed    float64
	MugSpeed          float64
	EmptyMugSpeed     float64
	SpawnIntervalBase float64
	TargetCountBase   int
	TargetCountPerLvl int
	Lives             int
	ServeScore        int
	CatchScore        int
	LevelClearDelay   float64
	CatchRadius       float64
	CustomerReachX    float64
	VerticalCooldown  float64
}

func DefaultConfig() Config {
	return Config{
		ScreenW:           1600,
		ScreenH:           1200,
		BarCount:          4,
		BarLength:         1520,
		BarTapX:           160,
		BartenderSpeed:    440,
		CustomerSpeed:     100,
		SatisfiedSpeed:    180,
		MugSpeed:          360,
		EmptyMugSpeed:     300,
		SpawnIntervalBase: 3.0,
		TargetCountBase:   12,
		TargetCountPerLvl: 6,
		Lives:             5,
		ServeScore:        100,
		CatchScore:        50,
		LevelClearDelay:   3.0,
		CatchRadius:       96,
		CustomerReachX:    180,
		VerticalCooldown:  0.12,
	}
}

type Bar struct {
	ID int
	Y  float64
}

type Bartender struct {
	BarID        int
	X            float64
	VX           float64
	Intent       int
	MoveCooldown float64
}

type Customer struct {
	ID      int
	BarID   int
	X       float64
	VX      float64
	State   CustomerState
	Speed   float64
	Excited bool
}

type Mug struct {
	ID     int
	BarID  int
	X      float64
	VX     float64
	Filled bool
}

type Game struct {
	Status    Status
	Level     int
	Score     int
	Lives     int
	HighScore int

	Bars      []Bar
	Bartender Bartender
	Customers []Customer
	Mugs      []Mug

	SpawnedCount      int
	TargetCount       int
	SpawnTimer        float64
	LevelClearPending bool
	LevelClearTimer   float64
	LevelStartTimer   float64

	RNG    RNG
	Config Config

	Events []Event
	nextID int
}

func NewGame(rng RNG, cfg Config, highScore int) *Game {
	g := &Game{
		Status:    StatusMenu,
		Level:     1,
		Lives:     cfg.Lives,
		HighScore: highScore,
		RNG:       rng,
		Config:    cfg,
	}
	g.Bars = make([]Bar, cfg.BarCount)
	barSpacing := cfg.ScreenH / float64(cfg.BarCount+1)
	for i := 0; i < cfg.BarCount; i++ {
		g.Bars[i] = Bar{ID: i, Y: barSpacing * float64(i+1)}
	}
	g.resetLevelState()
	return g
}

func (g *Game) resetLevelState() {
	g.SpawnedCount = 0
	g.TargetCount = g.Config.TargetCountBase + (g.Level-1)*g.Config.TargetCountPerLvl
	g.SpawnTimer = 0
	g.LevelClearPending = false
	g.LevelClearTimer = 0
	g.LevelStartTimer = 2.0
	g.Customers = nil
	g.Mugs = nil
	g.Bartender = Bartender{BarID: 0, X: g.Config.BarTapX, Intent: 0}
}

func (g *Game) Start() {
	g.Status = StatusPlaying
	g.Level = 1
	g.Score = 0
	g.Lives = g.Config.Lives
	g.resetLevelState()
}

func (g *Game) TogglePause() {
	if g.Status == StatusPlaying {
		g.Status = StatusPaused
	} else if g.Status == StatusPaused {
		g.Status = StatusPlaying
	}
}

func (g *Game) MoveVertical(delta int) {
	if g.Status != StatusPlaying {
		return
	}
	b := &g.Bartender
	if b.MoveCooldown > 0 {
		return
	}
	nb := b.BarID + delta
	if nb < 0 {
		nb = g.Config.BarCount - 1
	}
	if nb >= g.Config.BarCount {
		nb = 0
	}
	b.BarID = nb
	b.MoveCooldown = g.Config.VerticalCooldown
}

func (g *Game) ReturnToTap() {
	if g.Status != StatusPlaying {
		return
	}
	g.Bartender.X = g.Config.BarTapX
	g.Bartender.Intent = 0
	g.Bartender.VX = 0
}

func (g *Game) SetHorizontalIntent(dir int) {
	if dir < -1 {
		dir = -1
	}
	if dir > 1 {
		dir = 1
	}
	g.Bartender.Intent = dir
}

func (g *Game) Serve() {
	if g.Status != StatusPlaying {
		return
	}
	b := &g.Bartender
	if b.X > g.Config.BarTapX+20 {
		return
	}
	g.nextID++
	m := Mug{
		ID:     g.nextID,
		BarID:  b.BarID,
		X:      b.X + 10,
		VX:     g.Config.MugSpeed,
		Filled: true,
	}
	g.Mugs = append(g.Mugs, m)
	g.Events = append(g.Events, EventServe)
}

func (g *Game) Tick(dt float64) {
	if g.Status != StatusPlaying {
		return
	}
	if dt <= 0 {
		return
	}
	if g.LevelStartTimer > 0 {
		g.LevelStartTimer -= dt
		if g.LevelStartTimer < 0 {
			g.LevelStartTimer = 0
		}
	}

	// Bartender movement
	b := &g.Bartender
	if b.MoveCooldown > 0 {
		b.MoveCooldown -= dt
		if b.MoveCooldown < 0 {
			b.MoveCooldown = 0
		}
	}
	if b.Intent != 0 {
		b.VX = float64(b.Intent) * g.Config.BartenderSpeed
	} else {
		b.VX = 0
	}
	b.X += b.VX * dt
	if b.X < g.Config.BarTapX-10 {
		b.X = g.Config.BarTapX - 10
	}
	if b.X > g.Config.BarLength-20 {
		b.X = g.Config.BarLength - 20
	}

	// Mugs
	for i := range g.Mugs {
		m := &g.Mugs[i]
		speed := m.VX
		if speed == 0 {
			if m.Filled {
				speed = g.Config.MugSpeed
			} else {
				speed = -g.Config.EmptyMugSpeed
			}
		}
		m.X += speed * dt
	}

	// Customers
	custSpeed := g.Config.CustomerSpeed + float64(g.Level-1)*4
	for i := range g.Customers {
		c := &g.Customers[i]
		if c.State == CustomerApproaching {
			if c.Speed == 0 {
				c.Speed = custSpeed
			}
			c.X -= c.Speed * dt
			if !c.Excited && c.X <= 200 {
				c.Excited = true
				g.Events = append(g.Events, EventExcited)
			}
		} else {
			if c.Speed == 0 {
				c.Speed = g.Config.SatisfiedSpeed
			}
			c.X += c.Speed * dt
		}
	}

	// Collisions: filled mug -> customer
	satisfied := map[int]bool{}
	consumedMug := map[int]bool{}
	for mi := range g.Mugs {
		m := &g.Mugs[mi]
		if !m.Filled || consumedMug[m.ID] {
			continue
		}
		for ci := range g.Customers {
			c := &g.Customers[ci]
			if c.BarID != m.BarID || c.State != CustomerApproaching {
				continue
			}
			if satisfied[c.ID] {
				continue
			}
			if m.X >= c.X-10 {
				consumedMug[m.ID] = true
				satisfied[c.ID] = true
				c.State = CustomerSatisfied
				c.Speed = g.Config.SatisfiedSpeed
				g.nextID++
				g.Mugs = append(g.Mugs, Mug{
					ID:     g.nextID,
					BarID:  c.BarID,
					X:      c.X,
					VX:     -g.Config.EmptyMugSpeed,
					Filled: false,
				})
				g.Score += g.Config.ServeScore
				g.Events = append(g.Events, EventServe)
			}
		}
	}

	// Catch empty mugs by bartender / breaks
	var aliveMugs []Mug
	for _, m := range g.Mugs {
		if consumedMug[m.ID] {
			continue
		}
		if !m.Filled {
			// catch check
			if m.BarID == b.BarID && math.Abs(b.X-m.X) <= g.Config.CatchRadius {
				g.Score += g.Config.CatchScore
				g.Events = append(g.Events, EventCatch)
				continue
			}
			if m.X <= 0 {
				g.loseLife()
				g.Events = append(g.Events, EventBreak)
				continue
			}
		} else {
			if m.X >= g.Config.BarLength {
				g.loseLife()
				g.Events = append(g.Events, EventBreak)
				continue
			}
		}
		aliveMugs = append(aliveMugs, m)
	}
	g.Mugs = aliveMugs

	// Customers reaching bartender (life loss) + remove satisfied off-screen
	var aliveCust []Customer
	for _, c := range g.Customers {
		if c.State == CustomerApproaching && c.X <= g.Config.CustomerReachX && c.BarID == g.Bartender.BarID {
			// only lose life if on same bar as bartender? Customer reaches bar end regardless.
		}
		if c.State == CustomerApproaching && c.X <= g.Config.BarTapX {
			g.loseLife()
			g.Events = append(g.Events, EventLifeLost)
			continue
		}
		if c.State == CustomerSatisfied && c.X >= g.Config.BarLength+20 {
			continue
		}
		aliveCust = append(aliveCust, c)
	}
	g.Customers = aliveCust

	// Spawner
	if g.SpawnedCount < g.TargetCount {
		g.SpawnTimer += dt
		interval := g.spawnInterval()
		if g.SpawnTimer >= interval {
			g.SpawnTimer = 0
			barID := 0
			if g.Config.BarCount > 0 {
				barID = g.RNG.Intn(g.Config.BarCount)
			}
			g.nextID++
			g.Customers = append(g.Customers, Customer{
				ID:    g.nextID,
				BarID: barID,
				X:     g.Config.BarLength,
				VX:    -custSpeed,
				State: CustomerApproaching,
				Speed: custSpeed,
			})
			g.SpawnedCount++
		}
	}

	// Level completion
	if g.SpawnedCount >= g.TargetCount && len(g.Customers) == 0 && len(g.Mugs) == 0 {
		if !g.LevelClearPending {
			g.LevelClearPending = true
			g.LevelClearTimer = g.Config.LevelClearDelay
		}
	}
	if g.LevelClearPending {
		g.LevelClearTimer -= dt
		if g.LevelClearTimer <= 0 {
			g.Level++
			g.Events = append(g.Events, EventLevelUp)
			g.resetLevelState()
		}
	}

	if g.Lives <= 0 && g.Status == StatusPlaying {
		g.Status = StatusGameOver
		if g.Score > g.HighScore {
			g.HighScore = g.Score
		}
		g.Events = append(g.Events, EventGameOver)
	}
}

func (g *Game) spawnInterval() float64 {
	v := g.Config.SpawnIntervalBase / (1 + float64(g.Level-1)*0.08)
	if v < 0.7 {
		v = 0.7
	}
	return v
}

func (g *Game) loseLife() {
	g.Lives--
	if g.Lives < 0 {
		g.Lives = 0
	}
}
