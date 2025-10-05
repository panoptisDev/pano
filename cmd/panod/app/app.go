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

package app

import (
	"os"

	"github.com/panoptisDev/pano/config/flags"
	"gopkg.in/urfave/cli.v1"
)

// Run starts panod with the regular command line arguments.
func Run() error {
	return RunWithArgs(os.Args, nil)
}

// AppControl is a struct of channels facilitating the interaction of a test
// harness with a panod application instance.
type AppControl struct {
	// Upon a successful start of the panod node, the node ID is sent to this
	// channel. The channel is closed when the process stops.
	NodeIdAnnouncement chan<- string
	// Upon a successful start of the panod node, the HTTP port used by the HTTP
	// server is sent to this channel. The channel is closed when the process
	HttpPortAnnouncement chan<- string
	// The process is stopped by sending a message through this channel, or by
	// closing it.
	Shutdown <-chan struct{}
}

// RunWithArgs starts panod with the given command line arguments.
// An optional httpPortAnnouncement channel can be provided to announce the HTTP
// port used by the HTTP server of the started panod node. The channel is
// closed when the process stops.
// Another optional stop channel can be provided. By sending a message through
// this channel, or closing it, the shutdown of the process is triggered.
func RunWithArgs(
	args []string,
	control *AppControl,
) error {
	app := initApp()
	initAppHelp()

	// If present, take ownership and inject the control struct into the action.
	if control != nil {
		// Disable txPool validation, only to be used in tests.
		app.Flags = append(app.Flags, &flags.TEST_ONLY_DisableTransactionPoolValidation)
		if control.NodeIdAnnouncement != nil {
			defer close(control.NodeIdAnnouncement)
		}
		if control.HttpPortAnnouncement != nil {
			defer close(control.HttpPortAnnouncement)
		}
		app.Action = func(ctx *cli.Context) error {
			return lachesisMainInternal(
				ctx,
				control,
			)
		}
	}

	return app.Run(args)
}
