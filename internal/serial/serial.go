package serial

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
	"github.com/JeremyProffittOrg/cli-controller/internal/protocol"
	goserial "go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

const (
	EspressifVID = "303A"
	EspressifPID = "1001"
	SamsungVID   = "04E8"
	Baud         = 115200
	HelloWait    = 3 * time.Second
)

type PortInfo struct {
	Name         string
	IsUSB        bool
	VID          string
	PID          string
	SerialNumber string
	Product      string
}

type Port interface {
	io.ReadWriteCloser
	SetReadTimeout(t time.Duration) error
}

type Lister func() ([]PortInfo, error)
type Opener func(name string) (Port, error)

type realPort struct {
	goserial.Port
}

func (p realPort) SetReadTimeout(t time.Duration) error {
	return p.Port.SetReadTimeout(t)
}

func DefaultLister() ([]PortInfo, error) {
	raw, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}
	out := make([]PortInfo, 0, len(raw))
	for _, p := range raw {
		if p == nil || p.Name == "" {
			continue
		}
		out = append(out, PortInfo{
			Name:         p.Name,
			IsUSB:        p.IsUSB,
			VID:          normalizeID(p.VID),
			PID:          normalizeID(p.PID),
			SerialNumber: strings.ToUpper(strings.TrimSpace(p.SerialNumber)),
			Product:      p.Product,
		})
	}
	return out, nil
}

func DefaultOpener(name string) (Port, error) {
	p, err := goserial.Open(name, &goserial.Mode{BaudRate: Baud})
	if err != nil {
		return nil, err
	}
	return realPort{p}, nil
}

func normalizeID(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "0X")
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return "0"
	}
	if len(s) < 4 {
		s = strings.Repeat("0", 4-len(s)) + s
	}
	return s
}

func skipAuto(p PortInfo) bool {
	if !p.IsUSB {
		return true
	}
	if p.VID == SamsungVID {
		return true
	}
	return false
}

func isDialUSB(p PortInfo) bool {
	return p.IsUSB && p.VID == EspressifVID && p.PID == EspressifPID
}

func probeOrder(ports []PortInfo, lastSerial string) []PortInfo {
	lastSerial = strings.ToUpper(strings.TrimSpace(lastSerial))
	var first, second []PortInfo
	seen := map[string]bool{}
	if lastSerial != "" {
		for _, p := range ports {
			if isDialUSB(p) && p.SerialNumber == lastSerial {
				first = append(first, p)
				seen[p.Name] = true
			}
		}
	}
	for _, p := range ports {
		if seen[p.Name] {
			continue
		}
		if isDialUSB(p) {
			second = append(second, p)
			seen[p.Name] = true
		}
	}
	return append(first, second...)
}

func waitHello(p Port, wait time.Duration) (protocol.DeviceMsg, error) {
	_ = p.SetReadTimeout(200 * time.Millisecond)
	deadline := time.Now().Add(wait)
	var buf bytes.Buffer
	tmp := make([]byte, 256)
	for time.Now().Before(deadline) {
		n, err := p.Read(tmp)
		if n > 0 {
			for _, line := range protocol.SplitLines(&buf, tmp[:n]) {
				m, perr := protocol.ParseDeviceLine(line)
				if perr == nil && m.Hello {
					return m, nil
				}
			}
		}
		if err != nil && err != io.EOF {
			// timeout is expected
		}
	}
	return protocol.DeviceMsg{}, fmt.Errorf("no hello")
}

func OpenAndHello(open Opener, name string) (Port, protocol.DeviceMsg, error) {
	p, err := open(name)
	if err != nil {
		return nil, protocol.DeviceMsg{}, err
	}
	m, err := waitHello(p, HelloWait)
	if err != nil {
		_ = p.Close()
		return nil, protocol.DeviceMsg{}, err
	}
	b, err := protocol.HelloHost()
	if err != nil {
		_ = p.Close()
		return nil, protocol.DeviceMsg{}, err
	}
	if _, err := p.Write(b); err != nil {
		_ = p.Close()
		return nil, protocol.DeviceMsg{}, err
	}
	_ = p.SetReadTimeout(200 * time.Millisecond)
	return p, m, nil
}

func Find(list Lister, open Opener, cfg config.Config) (Port, PortInfo, protocol.DeviceMsg, error) {
	ports, err := list()
	if err != nil {
		return nil, PortInfo{}, protocol.DeviceMsg{}, err
	}
	if cfg.PortMode == "manual" && cfg.Port != "" {
		info := PortInfo{Name: cfg.Port}
		for _, p := range ports {
			if p.Name == cfg.Port {
				info = p
				break
			}
		}
		port, msg, err := OpenAndHello(open, cfg.Port)
		return port, info, msg, err
	}
	candidates := probeOrder(ports, cfg.LastSerial)
	var last error
	for _, c := range candidates {
		if skipAuto(c) && !isDialUSB(c) {
			continue
		}
		port, msg, err := OpenAndHello(open, c.Name)
		if err == nil {
			return port, c, msg, nil
		}
		last = err
	}
	if last == nil {
		last = fmt.Errorf("no dial USB port")
	}
	return nil, PortInfo{}, protocol.DeviceMsg{}, last
}

func ListPresent(list Lister) ([]PortInfo, error) {
	if list == nil {
		list = DefaultLister
	}
	return list()
}

func PortLabel(p PortInfo) string {
	label := p.Name
	if p.Product != "" {
		label += " — " + p.Product
	} else if p.IsUSB {
		label += " — USB " + p.VID + ":" + p.PID
	}
	return label
}

type EventHandler func(protocol.DeviceMsg)
type ConnHandler func(connected bool, port PortInfo)

type Manager struct {
	List   Lister
	Open   Opener
	OnMsg  EventHandler
	OnConn ConnHandler
	Write  func([]byte) error

	cfg  func() config.Config
	stop chan struct{}
	cur  Port
}

func NewManager(cfg func() config.Config) *Manager {
	return &Manager{
		List: DefaultLister,
		Open: DefaultOpener,
		cfg:  cfg,
		stop: make(chan struct{}),
	}
}

func (m *Manager) Stop() {
	select {
	case <-m.stop:
	default:
		close(m.stop)
	}
	if m.cur != nil {
		_ = m.cur.Close()
	}
}

func (m *Manager) Send(b []byte) error {
	if m.cur == nil {
		return fmt.Errorf("not connected")
	}
	_, err := m.cur.Write(b)
	return err
}

func (m *Manager) Run() {
	delay := time.Second
	for {
		select {
		case <-m.stop:
			return
		default:
		}
		err := m.session()
		if err == nil {
			delay = time.Second
		}
		select {
		case <-m.stop:
			return
		case <-time.After(delay):
		}
		if delay < 5*time.Second {
			delay *= 2
		}
		if delay > 5*time.Second {
			delay = 5 * time.Second
		}
	}
}

func (m *Manager) session() error {
	list := m.List
	open := m.Open
	if list == nil {
		list = DefaultLister
	}
	if open == nil {
		open = DefaultOpener
	}
	cfg := m.cfg()
	p, info, _, err := Find(list, open, cfg)
	if err != nil {
		return err
	}
	m.cur = p
	if m.OnConn != nil {
		m.OnConn(true, info)
	}
	defer func() {
		_ = p.Close()
		m.cur = nil
		if m.OnConn != nil {
			m.OnConn(false, PortInfo{})
		}
	}()
	_ = p.SetReadTimeout(200 * time.Millisecond)
	var buf bytes.Buffer
	tmp := make([]byte, 256)
	lastBeat := time.Now()
	for {
		select {
		case <-m.stop:
			return nil
		default:
		}
		if time.Since(lastBeat) >= time.Second {
			if b, err := protocol.Ping(); err == nil {
				_, _ = p.Write(b)
			}
			lastBeat = time.Now()
		}
		n, err := p.Read(tmp)
		if n > 0 {
			for _, line := range protocol.SplitLines(&buf, tmp[:n]) {
				msg, perr := protocol.ParseDeviceLine(line)
				if perr != nil {
					continue
				}
				if m.OnMsg != nil {
					m.OnMsg(msg)
				}
			}
		}
		if err != nil {
			if te, ok := err.(interface{ Timeout() bool }); ok && te.Timeout() {
				continue
			}
			return err
		}
	}
}
