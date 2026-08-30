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
}

type Engine struct {
	cfg                config.Config
	channels           [4]channelState
	sideRaised         map[string]bool
	leftCount          int
	lastLeft           time.Time
	armed              bool
	deadline           time.Time
	gravityX, gravityY float64
	haveGravity        bool
	deskLocked         bool
	deskReleaseSince   time.Time
}

func New(cfg config.Config) *Engine { e := &Engine{}; e.Configure(cfg); return e }

func (e *Engine) Configure(cfg config.Config) {
	cfg.Normalize()
	e.cfg = cfg
	if e.sideRaised == nil {
		e.sideRaised = map[string]bool{"left": false, "right": false}
	}
}

func (e *Engine) Connected(ch int, ok bool) []Event {
	if ch < 0 || ch >= len(e.channels) {
		return nil
	}
	s := &e.channels[ch]
	s.connected = ok
	if !ok {
		s.samples = nil
		s.baseline = 0
		s.raised = false
		s.releaseSince = time.Time{}
	}
	return e.updateSides(time.Now())
}

func (e *Engine) Distance(ch, mm int, now time.Time) []Event {
	if ch < 0 || ch >= len(e.channels) || mm <= 0 || mm > 4000 {
		return nil
	}
	s := &e.channels[ch]
	s.connected = true
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
	threshold := e.cfg.KneeChannels[ch].ThresholdMM
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
		s.baseline += (float64(med) - s.baseline) * 0.02
	}
	return e.updateSides(now)
}

func median3(v []int) int { a := append([]int(nil), v...); sort.Ints(a); return a[len(a)/2] }

func (e *Engine) updateSides(now time.Time) []Event {
	var out []Event
	for _, side := range []string{"left", "right"} {
		raised := false
		for i := range e.channels {
			if e.cfg.KneeChannels[i].Role == side && e.channels[i].connected && e.channels[i].raised {
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

func (e *Engine) Accel(x, y, _ int, now time.Time) []Event {
	if !e.cfg.DeskEnabled {
		e.haveGravity = false
		e.deskLocked = false
		return nil
	}
	if !e.haveGravity {
		e.gravityX = float64(x)
		e.gravityY = float64(y)
		e.haveGravity = true
		return nil
	}
	dx, dy := float64(x)-e.gravityX, float64(y)-e.gravityY
	e.gravityX += dx * 0.08
	e.gravityY += dy * 0.08
	rx, ry := rotate(dx, dy, e.cfg.DeskOrientation)
	absX, absY := abs(rx), abs(ry)
	peak := absX
	if absY > peak {
		peak = absY
	}
	release := float64(e.cfg.DeskSensitivityMg) * 0.4
	if e.deskLocked {
		if peak < release {
			if e.deskReleaseSince.IsZero() {
				e.deskReleaseSince = now
			}
			if now.Sub(e.deskReleaseSince) >= 500*time.Millisecond {
				e.deskLocked = false
				e.deskReleaseSince = time.Time{}
			}
		} else {
			e.deskReleaseSince = time.Time{}
		}
		return nil
	}
	if peak < float64(e.cfg.DeskSensitivityMg) {
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
	e.deskLocked = true
	action := map[string]string{"left": e.cfg.DeskLeft, "right": e.cfg.DeskRight, "forward": e.cfg.DeskForward, "back": e.cfg.DeskBack}[dir]
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
