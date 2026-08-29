package applog

import (
	"io"
	"log"
	"os"

	"github.com/JeremyProffittOrg/cli-controller/internal/config"
)

func Open() (*log.Logger, *os.File, error) {
	p, err := config.LogPath()
	if err != nil {
		return log.Default(), nil, err
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return log.Default(), nil, err
	}
	l := log.New(io.MultiWriter(f, os.Stderr), "", log.LstdFlags|log.Lmicroseconds)
	return l, f, nil
}
