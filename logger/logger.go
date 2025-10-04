// Copyright 2025 Pano Operations Ltd
// This file is part of the Pano Client
//
// Pano is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Pano is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Pano. If not, see <http://www.gnu.org/licenses/>.

package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ethereum/go-ethereum/log"
)

// init with defaults.
func init() {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, true)))
}

// SetLevel sets level filter on log root handler.
// So it should be called last.
func SetLevel(l string) {
	lvl, err := levelFromString(l)
	if err != nil {
		panic(err)
	}

	log.SetDefault(
		log.NewLogger(
			log.NewTerminalHandlerWithLevel(os.Stderr, lvl, true),
		))
}

func levelFromString(lvlString string) (slog.Level, error) {
	switch lvlString {
	case "debug", "dbug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error", "eror":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("unknown level: %v", lvlString)
	}
}

//go:generate mockgen -source=logger.go -destination=logger_mock.go -package=logger

// Logger defined as an alias for log.Logger to allow mocking in tests.
type Logger interface {
	log.Logger
}
