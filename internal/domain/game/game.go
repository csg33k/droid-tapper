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
	EventPowerUp
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
	EnabledPowerUps   [4]bool
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
		EnabledPowerUps:   [4]bool{true, true, true, true},
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

type PowerUpKind int

const (
	PowerSlowBartender PowerUpKind = iota
	PowerSpeedBartender
	PowerSlowCustomers
	PowerNuke
)

type PowerUp struct {
	ID    int
	BarID int
	X     float64
	Kind  PowerUpKind
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
	PowerUps  []PowerUp

	SpawnedCount      int
	TargetCount       int
	SpawnTimer        float64
	PowerUpTimer      float64
	LevelClearPending bool
	LevelClearTimer   float64
	LevelStartTimer   float64

	BartenderBoostTimer float64
	BartenderBoostMult  float64
	CustomerSlowTimer   float64

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
	g.PowerUps = nil
	g.Bartender = Bartender{BarID: 0, X: g.Config.BarTapX, Intent: 0}
	g.BartenderBoostTimer = 0
	g.BartenderBoostMult = 1
	g.CustomerSlowTimer = 0
	g.PowerUpTimer = 0
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
	if g.BartenderBoostTimer > 0 {
		g.BartenderBoostTimer -= dt
		if g.BartenderBoostTimer <= 0 {
			g.BartenderBoostTimer = 0
			g.BartenderBoostMult = 1
		}
	}
	if g.CustomerSlowTimer > 0 {
		g.CustomerSlowTimer -= dt
		if g.CustomerSlowTimer < 0 {
			g.CustomerSlowTimer = 0
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
	bartenderMult := 1.0
	if g.BartenderBoostTimer > 0 {
		bartenderMult = g.BartenderBoostMult
	}
	if b.Intent != 0 {
		b.VX = float64(b.Intent) * g.Config.BartenderSpeed * bartenderMult
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
	custSpeed := (g.Config.CustomerSpeed + float64(g.Level-1)*4)
	if g.CustomerSlowTimer > 0 {
		custSpeed *= 0.5
	}
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

	// Power-ups
	var alivePU []PowerUp
	for _, pu := range g.PowerUps {
		pu.X -= 90 * dt
		picked := false
		if pu.BarID == g.Bartender.BarID && math.Abs(g.Bartender.X-pu.X) <= 32 {
			picked = true
			switch pu.Kind {
			case PowerSlowBartender:
				if g.BartenderBoostTimer > 0 && g.BartenderBoostMult > 1 {
					// Crab cancels JS: net zero
					g.BartenderBoostTimer = 0
					g.BartenderBoostMult = 1
				} else {
					g.BartenderBoostMult = 0.5
					g.BartenderBoostTimer = 30
				}
			case PowerSpeedBartender:
				if g.BartenderBoostTimer > 0 && g.BartenderBoostMult < 1 {
					// JS cancels Crab: net zero
					g.BartenderBoostTimer = 0
					g.BartenderBoostMult = 1
				} else {
					g.BartenderBoostMult = 1.8
					g.BartenderBoostTimer = 30
				}
			case PowerSlowCustomers:
				g.CustomerSlowTimer = 30
			case PowerNuke:
				g.Mugs = nil
				for i := range g.Customers {
					g.Customers[i].State = CustomerApproaching
					g.Customers[i].Speed = 0
					g.Customers[i].Excited = true
				}
			}
			g.Events = append(g.Events, EventPowerUp)
		}
		if picked {
			continue
		}
		if pu.X < -20 {
			continue
		}
		alivePU = append(alivePU, pu)
	}
	g.PowerUps = alivePU

	g.PowerUpTimer += dt
	if g.PowerUpTimer >= 14 {
		g.PowerUpTimer = 0
		// choose from enabled power-ups
		var kinds []PowerUpKind
		for k := 0; k < len(g.Config.EnabledPowerUps); k++ {
			if g.Config.EnabledPowerUps[k] {
				kinds = append(kinds, PowerUpKind(k))
			}
		}
		if len(kinds) == 0 {
			g.PowerUpTimer = 0
		} else {
			barID := g.RNG.Intn(g.Config.BarCount)
			kind := kinds[g.RNG.Intn(len(kinds))]
			g.nextID++
			g.PowerUps = append(g.PowerUps, PowerUp{ID: g.nextID, BarID: barID, X: g.Config.BarLength, Kind: kind})
		}
	}

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
