package motion

import (
	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"testing"
	"time"
)

func feed(e *Engine, ch int, mm int, at time.Time) []Event {
	var out []Event
	for i := 0; i < 3; i++ {
		out = append(out, e.Distance(ch, mm, at.Add(time.Duration(i)*10*time.Millisecond))...)
	}
	return out
}
func release(e *Engine, ch int, mm int, at time.Time) {
	feed(e, ch, mm, at)
	feed(e, ch, mm, at.Add(200*time.Millisecond))
}

func TestArmModeMovesAndActivates(t *testing.T) {
	c := config.Default()
	c.DwellMs = 500
	c.KneeLeftRaises = 2
	e := New(c)
	now := time.Unix(0, 0)
	feed(e, 0, 500, now)
	feed(e, 1, 500, now)
	if got := feed(e, 0, 400, now.Add(time.Second)); len(got) != 0 {
		t.Fatalf("first left %v", got)
	}
	release(e, 0, 500, now.Add(1100*time.Millisecond))
	got := feed(e, 0, 400, now.Add(1400*time.Millisecond))
	if len(got) != 1 || got[0].Kind != Show {
		t.Fatalf("arm %v", got)
	}
	got = feed(e, 1, 400, now.Add(1500*time.Millisecond))
	if len(got) != 1 || got[0].Kind != Move || got[0].Delta != 1 {
		t.Fatalf("move %v", got)
	}
	got = e.Tick(now.Add(2100 * time.Millisecond))
	if len(got) != 1 || got[0].Kind != Activate {
		t.Fatalf("activate %v", got)
	}
}

func TestConfirmRequiresLeftSequence(t *testing.T) {
	c := config.Default()
	c.KneeMode = "confirm"
	c.KneeLeftRaises = 1
	c.KneeRightDirection = -1
	e := New(c)
	now := time.Unix(0, 0)
	feed(e, 0, 500, now)
	feed(e, 1, 500, now)
	got := feed(e, 1, 400, now.Add(time.Second))
	if len(got) != 1 || got[0].Kind != Move || got[0].Delta != -1 {
		t.Fatalf("right %v", got)
	}
	if got = e.Tick(now.Add(4 * time.Second)); len(got) != 0 {
		t.Fatalf("right committed %v", got)
	}
	got = feed(e, 0, 400, now.Add(5*time.Second))
	if len(got) != 1 || got[0].Kind != Activate {
		t.Fatalf("left %v", got)
	}
}

func TestSideORPreventsDuplicateAndHandoff(t *testing.T) {
	c := config.Default()
	c.KneeLeftRaises = 1
	c.KneeChannels[2] = config.KneeChannel{Role: "left", ThresholdMM: 75}
	e := New(c)
	now := time.Unix(0, 0)
	feed(e, 0, 500, now)
	feed(e, 2, 500, now)
	if got := feed(e, 0, 400, now.Add(time.Second)); len(got) != 1 {
		t.Fatalf("first %v", got)
	}
	if got := feed(e, 2, 400, now.Add(1100*time.Millisecond)); len(got) != 0 {
		t.Fatalf("overlap %v", got)
	}
	release(e, 0, 500, now.Add(1200*time.Millisecond))
	if got := e.Connected(0, false); len(got) != 0 {
		t.Fatalf("handoff %v", got)
	}
}

func TestPartialSequenceExpires(t *testing.T) {
	c := config.Default()
	c.DwellMs = 250
	c.KneeLeftRaises = 2
	e := New(c)
	now := time.Unix(0, 0)
	feed(e, 0, 500, now)
	feed(e, 0, 400, now.Add(time.Second))
	release(e, 0, 500, now.Add(1200*time.Millisecond))
	e.Tick(now.Add(2 * time.Second))
	if got := feed(e, 0, 400, now.Add(3*time.Second)); len(got) != 0 {
		t.Fatalf("expired sequence %v", got)
	}
}

func TestDeskDominantDirectionAndRelease(t *testing.T) {
	c := config.Default()
	c.DeskEnabled = true
	c.DeskSensitivityMg = 350
	e := New(c)
	now := time.Unix(0, 0)
	e.Accel(0, 0, 1000, now)
	got := e.Accel(-500, 100, 1000, now.Add(10*time.Millisecond))
	if len(got) != 1 || got[0].Kind != Tile {
		t.Fatalf("left %v", got)
	}
	if got = e.Accel(600, 0, 1000, now.Add(20*time.Millisecond)); len(got) != 0 {
		t.Fatalf("rebound %v", got)
	}
	e.Accel(0, 0, 1000, now.Add(100*time.Millisecond))
	e.Accel(0, 0, 1000, now.Add(700*time.Millisecond))
	got = e.Accel(600, 0, 1000, now.Add(time.Second))
	if len(got) != 1 || got[0].Kind != Stack {
		t.Fatalf("right %v", got)
	}
}
