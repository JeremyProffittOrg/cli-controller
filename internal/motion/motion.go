package motion

import (
	"sort"
	"time"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
)

type EventKind int

const (
	Show EventKind = iota + 1
	Move
	Activate
	Tile
	Stack
)

type Event struct {
	Kind  EventKind
	Delta int
}

type channelState struct {
	samples      []int
	baseline     float64
	raised       bool
	releaseSince time.Time
	connected    bool
	lastAt       time.Time
}

type deskState struct {
	gravityX, gravityY float64
	haveGravity        bool
	locked             bool
	releaseSince       time.Time
}

type Engine struct {
	cfg        config.Config
	tof        map[string]*channelState
	desk       map[string]*deskState
	sideRaised map[string]bool
	leftCount  int
	lastLeft   time.Time
	armed      bool
	deadline   time.Time
}

// baselineTau is the resting-distance tracking time constant in seconds. The
// Dial streams each VL53L4CD at 20 Hz, so adaptation is set by wall-clock
// time rather than sample count and does not change when more sensors share
// the bus.
const baselineTau = 6.0

// baselineMaxAlpha caps one update after a long gap between samples.
const baselineMaxAlpha = 0.1

func New(cfg config.Config) *Engine { e := &Engine{}; e.Configure(cfg); return e }

func (e *Engine) Configure(cfg config.Config) {
	cfg.Normalize()
	e.cfg = cfg
	if e.tof == nil {
		e.tof = map[string]*channelState{}
	}
	if e.desk == nil {
		e.desk = map[string]*deskState{}
	}
	if e.sideRaised == nil {
		e.sideRaised = map[string]bool{"left": false, "right": false}
	}
}

func (e *Engine) tofState(id string) *channelState {
	s := e.tof[id]
	if s == nil {
		s = &channelState{}
		e.tof[id] = s
	}
	return s
}

func (e *Engine) deskState(id string) *deskState {
	s := e.desk[id]
	if s == nil {
		s = &deskState{}
		e.desk[id] = s
	}
	return s
}

func (e *Engine) Connected(id string, ok bool) []Event {
	if id == "" {
		return nil
	}
	s := e.tofState(id)
	s.connected = ok
	if !ok {
		s.samples = nil
		s.baseline = 0
		s.raised = false
		s.releaseSince = time.Time{}
		s.lastAt = time.Time{}
	}
	return e.updateSides(time.Now())
}

func (e *Engine) Distance(id string, mm int, now time.Time) []Event {
	if id == "" || mm <= 0 || mm > 4000 {
		return nil
	}
	s := e.tofState(id)
	s.connected = true
	prevAt := s.lastAt
	s.lastAt = now
	s.samples = append(s.samples, mm)
	if len(s.samples) > 3 {
		s.samples = s.samples[len(s.samples)-3:]
	}
	if len(s.samples) < 3 {
		return nil
	}
	med := median3(s.samples)
	if s.baseline == 0 {
		s.baseline = float64(med)
		return nil
	}
	ctrl, ok := e.cfg.FindSensor(id)
	if !ok || ctrl.Kind != "tof" || ctrl.Role == "off" {
		return nil
	}
	threshold := ctrl.ThresholdMM
	drop := s.baseline - float64(med)
	if !s.raised && drop >= float64(threshold) {
		s.raised = true
		s.releaseSince = time.Time{}
	} else if s.raised {
		if drop <= float64(threshold)/2 {
			if s.releaseSince.IsZero() {
				s.releaseSince = now
			}
			if now.Sub(s.releaseSince) >= 150*time.Millisecond {
				s.raised = false
				s.releaseSince = time.Time{}
			}
		} else {
			s.releaseSince = time.Time{}
		}
	} else {
		alpha := baselineMaxAlpha
		if !prevAt.IsZero() {
			alpha = now.Sub(prevAt).Seconds() / baselineTau
			if alpha > baselineMaxAlpha {
				alpha = baselineMaxAlpha
			}
			if alpha < 0 {
				alpha = 0
			}
		}
		s.baseline += (float64(med) - s.baseline) * alpha
	}
	return e.updateSides(now)
}

func median3(v []int) int { a := append([]int(nil), v...); sort.Ints(a); return a[len(a)/2] }

func (e *Engine) updateSides(now time.Time) []Event {
	var out []Event
	for _, side := range []string{"left", "right"} {
		raised := false
		for _, ctrl := range e.cfg.Sensors {
			if ctrl.Kind != "tof" || ctrl.Role != side {
				continue
			}
			s := e.tof[ctrl.ID]
			if s != nil && s.connected && s.raised {
				raised = true
				break
			}
		}
		if raised && !e.sideRaised[side] {
			out = append(out, e.sideRise(side, now)...)
		}
		e.sideRaised[side] = raised
	}
	return out
}

func (e *Engine) sideRise(side string, now time.Time) []Event {
	if side == "right" {
		if e.cfg.KneeMode == "confirm" {
			return []Event{{Kind: Move, Delta: e.cfg.KneeRightDirection}}
		}
		if e.armed {
			e.deadline = now.Add(time.Duration(e.cfg.DwellMs) * time.Millisecond)
			return []Event{{Kind: Move, Delta: e.cfg.KneeRightDirection}}
		}
		return nil
	}
	if !e.lastLeft.IsZero() && now.Sub(e.lastLeft) > time.Duration(e.cfg.DwellMs)*time.Millisecond {
		e.leftCount = 0
	}
	e.leftCount++
	e.lastLeft = now
	if e.leftCount < e.cfg.KneeLeftRaises {
		return nil
	}
	e.leftCount = 0
	if e.cfg.KneeMode == "confirm" {
		return []Event{{Kind: Activate}}
	}
	e.armed = true
	e.deadline = now.Add(time.Duration(e.cfg.DwellMs) * time.Millisecond)
	return []Event{{Kind: Show}}
}

func (e *Engine) Tick(now time.Time) []Event {
	if e.leftCount > 0 && !e.lastLeft.IsZero() && now.Sub(e.lastLeft) > time.Duration(e.cfg.DwellMs)*time.Millisecond {
		e.leftCount = 0
	}
	if e.armed && !e.deadline.IsZero() && !now.Before(e.deadline) {
		e.armed = false
		e.deadline = time.Time{}
		return []Event{{Kind: Activate}}
	}
	return nil
}

func (e *Engine) Accel(id string, x, y, _ int, now time.Time) []Event {
	ctrl, ok := e.cfg.FindSensor(id)
	if !ok || ctrl.Kind != "accel" || !ctrl.Enabled {
		return nil
	}
	s := e.deskState(id)
	if !s.haveGravity {
		s.gravityX = float64(x)
		s.gravityY = float64(y)
		s.haveGravity = true
		return nil
	}
	dx, dy := float64(x)-s.gravityX, float64(y)-s.gravityY
	s.gravityX += dx * 0.08
	s.gravityY += dy * 0.08
	rx, ry := rotate(dx, dy, ctrl.Orientation)
	absX, absY := abs(rx), abs(ry)
	peak := absX
	if absY > peak {
		peak = absY
	}
	release := float64(ctrl.SensitivityMg) * 0.4
	if s.locked {
		if peak < release {
			if s.releaseSince.IsZero() {
				s.releaseSince = now
			}
			if now.Sub(s.releaseSince) >= 500*time.Millisecond {
				s.locked = false
				s.releaseSince = time.Time{}
			}
		} else {
			s.releaseSince = time.Time{}
		}
		return nil
	}
	if peak < float64(ctrl.SensitivityMg) {
		return nil
	}
	dir := ""
	if absX >= absY {
		if rx < 0 {
			dir = "left"
		} else {
			dir = "right"
		}
	} else {
		if ry < 0 {
			dir = "forward"
		} else {
			dir = "back"
		}
	}
	s.locked = true
	action := map[string]string{"left": ctrl.Left, "right": ctrl.Right, "forward": ctrl.Forward, "back": ctrl.Back}[dir]
	if action == "tile" {
		return []Event{{Kind: Tile}}
	}
	if action == "stack" {
		return []Event{{Kind: Stack}}
	}
	return nil
}

func rotate(x, y float64, degrees int) (float64, float64) {
	switch degrees {
	case 90:
		return y, -x
	case 180:
		return -x, -y
	case 270:
		return -y, x
	default:
		return x, y
	}
}
func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
