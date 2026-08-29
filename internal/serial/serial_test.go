package serial

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
)

type fakePort struct {
	mu   sync.Mutex
	r    []byte
	w    []byte
	off  int
	dead bool
}

func (f *fakePort) Read(p []byte) (int, error) {
	deadline := time.Now().Add(150 * time.Millisecond)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		if f.dead {
			f.mu.Unlock()
			return 0, io.EOF
		}
		if f.off < len(f.r) {
			n := copy(p, f.r[f.off:])
			f.off += n
			f.mu.Unlock()
			return n, nil
		}
		f.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	return 0, nil
}

func (f *fakePort) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.w = append(f.w, p...)
	return len(p), nil
}

func (f *fakePort) Close() error {
	f.mu.Lock()
	f.dead = true
	f.mu.Unlock()
	return nil
}

func (f *fakePort) SetReadTimeout(time.Duration) error { return nil }

func TestFindPrefersLastSerialHandshake(t *testing.T) {
	ports := []PortInfo{
		{Name: "COM1", IsUSB: false},
		{Name: "COM3", IsUSB: false},
		{Name: "COM9", IsUSB: true, VID: "04E8", PID: "6860"},
		{Name: "COM21", IsUSB: true, VID: "303A", PID: "1001", SerialNumber: "DEAD"},
		{Name: "COM10", IsUSB: true, VID: "303A", PID: "1001", SerialNumber: "B0:81:84:97:1E:54", Product: "USB Serial Device"},
	}
	opened := []string{}
	open := func(name string) (Port, error) {
		opened = append(opened, name)
		switch name {
		case "COM10":
			return &fakePort{r: []byte("CLI-DIAL/1\n{\"v\":1,\"t\":\"hello\",\"fw\":\"0.1.0\",\"dev\":\"cli-dial\"}\n")}, nil
		default:
			return &fakePort{r: []byte("junk\n")}, nil
		}
	}
	list := func() ([]PortInfo, error) { return ports, nil }
	cfg := config.Default()
	cfg.LastSerial = "B0:81:84:97:1E:54"
	p, info, msg, err := Find(list, open, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if info.Name != "COM10" || !msg.Hello {
		t.Fatalf("info %+v msg %+v", info, msg)
	}
	if opened[0] != "COM10" {
		t.Fatalf("order %v", opened)
	}
}

func TestFindSkipsSamsungAndBluetooth(t *testing.T) {
	ports := []PortInfo{
		{Name: "COM3", IsUSB: false},
		{Name: "COM9", IsUSB: true, VID: "04E8", PID: "6860"},
		{Name: "COM10", IsUSB: true, VID: "303A", PID: "1001", SerialNumber: "ABC"},
	}
	open := func(name string) (Port, error) {
		if name != "COM10" {
			t.Fatalf("probed %s", name)
		}
		return &fakePort{r: []byte("CLI-DIAL/1\n")}, nil
	}
	_, info, _, err := Find(func() ([]PortInfo, error) { return ports, nil }, open, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "COM10" {
		t.Fatalf("%+v", info)
	}
}

func TestManualPort(t *testing.T) {
	cfg := config.Default()
	cfg.PortMode = "manual"
	cfg.Port = "COM10"
	open := func(name string) (Port, error) {
		if name != "COM10" {
			t.Fatalf("opened %s", name)
		}
		return &fakePort{r: []byte(`{"v":1,"t":"hello","dev":"cli-dial"}` + "\n")}, nil
	}
	_, info, msg, err := Find(func() ([]PortInfo, error) {
		return []PortInfo{{Name: "COM10", IsUSB: true, VID: "303A", PID: "1001"}}, nil
	}, open, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "COM10" || !msg.Hello {
		t.Fatalf("%+v %+v", info, msg)
	}
}

func TestNormalizeID(t *testing.T) {
	if normalizeID("303a") != "303A" || normalizeID("0x1001") != "1001" {
		t.Fatal(normalizeID("303a"), normalizeID("0x1001"))
	}
}
