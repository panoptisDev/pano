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
	"sort"

	"github.com/panoptisDev/pano/config/flags"
	"github.com/panoptisDev/pano/debug"
	"github.com/panoptisDev/pano/version"
	"gopkg.in/urfave/cli.v1"
)

func Run() error {
	return RunWithArgs(os.Args)
}

func RunWithArgs(args []string) error {
	app := cli.NewApp()
	app.Name = "panotool"
	app.Usage = "the Pano management tool"
	app.Version = version.StringWithCommit()
	app.Flags = []cli.Flag{
		flags.DataDirFlag,
		flags.CacheFlag,
		flags.LiveDbCacheFlag,
		flags.ArchiveCacheFlag,
		flags.StateDbCacheCapacityFlag,
	}
	app.Flags = append(app.Flags, debug.Flags...)

	app.Commands = []cli.Command{
		{
			Name:  "genesis",
			Usage: "Initialize the database from a genesis file",
			Description: `
    panotool --datadir=<datadir> genesis genesis-file.g

Requires a first argument of the genesis file to import.
Initialize the database using data from the genesis file.
`,

			ArgsUsage: "<filename>",
			Action:    gfileGenesisImport,
			Flags: []cli.Flag{
				ExperimentalFlag,
				ModeFlag,
			},

			Subcommands: []cli.Command{
				{
					Name:      "json",
					Usage:     "Initialize the database from a testing JSON genesis file",
					ArgsUsage: "<filename>",
					Action:    jsonGenesisImport,
					Flags: []cli.Flag{
						ExperimentalFlag,
						ModeFlag,
					},
					Description: `
    panotool --datadir=<datadir> genesis json --experimental genesis-file.json

Requires a first argument of the JSON genesis file to import.
Initialize the database using data from the experimental genesis file.
`,
				},
				{
					Name:      "fake",
					Usage:     "Initialize the database for a fakenet testing network",
					ArgsUsage: "<validators>",
					Action:    fakeGenesisImport,
					Flags: []cli.Flag{
						ModeFlag,
						FakeUpgrades,
					},
					Description: `
    panotool --datadir=<datadir> genesis fake <N> [--mode=validator] [--upgrades=upgrades]

Requires the number of validators in the fake network as the first argument.
Initialize the database for a testing fakenet.
--upgrades can be used to define the network features, default is pano hardfork feature set.
`,
				},
				{
					Name:      "export",
					Usage:     "Export current state into a genesis file",
					ArgsUsage: "<filename> [--mode=validator]",
					Action:    exportGenesis,
					Flags: []cli.Flag{
						ModeFlag,
					},
					Description: `
Export current state into a genesis file.
Requires a first argument of the file to write to.
Use --mode=validator to generate a genesis without an archive section.
`,
				},
				{
					Name:      "sign",
					Usage:     "Sign genesis file",
					ArgsUsage: "<filename>",
					Action:    signGenesis,
					Description: `
Add signature into an exported genesis file.
`,
				},
			},
		},

		{
			Name:        "check",
			Usage:       "Check EVM database consistency",
			Description: "Verifies the consistency of the EVM state database.",
			Subcommands: []cli.Command{
				{
					Name:   "live",
					Usage:  "Check EVM live state database",
					Action: checkLive,
					Description: `
    panotool --datadir=<datadir> check live

Verifies the consistency of the EVM state database.
The live state is used for blocks processing.
`,
				},
				{
					Name:   "archive",
					Usage:  "Check EVM archive states database",
					Action: checkArchive,
					Description: `
    panotool --datadir=<datadir> check archive

Verifies the consistency of the EVM state database.
The archive state is used for RPC - allows to handle state-related RPC queries.
`,
				},
			},
		},

		{
			Name:        "heal",
			Usage:       "Fix database in dirty state",
			Action:      heal,
			Description: "Tries to recover database corrupted by incorrect termination of the client.",
		},

		{
			Name:        "compact",
			Usage:       "Compact all pebble databases",
			Action:      compactDbs,
			Description: "Compacts (optimize) all the Pebble databases in the data directory.",
		},

		{
			Name:      "cli",
			Usage:     "Start an interactive JavaScript environment, attach to a node",
			ArgsUsage: "[endpoint]",
			Action:    remoteConsole,
			Flags: []cli.Flag{
				JSpathFlag,
				PreloadJSFlag,
				ExecFlag,
			},
			Description: `
The Pano console is an interactive shell for the JavaScript runtime environment
which exposes a node admin interface as well as the Dapp JavaScript API.
See https://github.com/ethereum/go-ethereum/wiki/JavaScript-Console.
This command allows to open a console attached to a running Pano node.`,
		},

		{
			Name:     "events",
			Usage:    "Export/import blockchain events",
			Category: "MISCELLANEOUS COMMANDS",

			Subcommands: []cli.Command{
				{
					Name:      "export",
					Usage:     "Export blockchain events",
					ArgsUsage: "<filename> [<epochFrom> <epochTo>]",
					Action:    exportEvents,
					Description: `
Requires a first argument of the file to write to.
Optional second and third arguments control the first and
last epoch to write. If the file ends with .gz, the output will
be gzipped.
`,
				},
				{
					Action:    importEvents,
					Name:      "import",
					Usage:     "Import blockchain events file",
					ArgsUsage: "<filename> (<filename 2> ... <filename N>)",
					Flags: []cli.Flag{
						ModeFlag,
						flags.SuppressFramePanicFlag,
					},
					Description: `
    panotool --datadir=<datadir> events import <filenames> [--mode=validator]

The import command imports events from RLP-encoded files.
Events are fully verified.`,
				},
			},
		},

		{
			Action:      checkConfig,
			Name:        "checkconfig",
			Usage:       "Checks configuration file",
			ArgsUsage:   "<filename>",
			Category:    "MISCELLANEOUS COMMANDS",
			Description: `The checkconfig checks configuration file.`,
		},
		{
			Action:      dumpConfig,
			Name:        "dumpconfig",
			Usage:       "Show configuration values",
			ArgsUsage:   "<filename>",
			Category:    "MISCELLANEOUS COMMANDS",
			Description: `The dumpconfig command shows configuration values.`,
		},

		{
			Name:     "account",
			Usage:    "Manage accounts",
			Category: "ACCOUNT COMMANDS",
			Description: `

Manage accounts, list all existing accounts, import a private key into a new
account, create a new account or update an existing account.

It supports interactive mode, when you are prompted for password as well as
non-interactive mode where passwords are supplied via a given password file.
Non-interactive mode is only meant for scripted use on test networks or known
safe environments.

Make sure you remember the password you gave when creating a new account (with
either new or import). Without it you are not able to unlock your account.

Note that exporting your key in unencrypted format is NOT supported.

Keys are stored under <DATADIR>/keystore.
It is safe to transfer the entire directory or the individual keys therein
between ethereum nodes by simply copying.

Make sure you backup your keys regularly.`,
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "Print summary of existing accounts",
					Action: accountList,
					Flags: []cli.Flag{
						flags.DataDirFlag,
						flags.KeyStoreDirFlag,
					},
					Description: `
Print a short summary of all accounts`,
				},
				{
					Name:   "new",
					Usage:  "Create a new account",
					Action: accountCreate,
					Flags: []cli.Flag{
						flags.DataDirFlag,
						flags.KeyStoreDirFlag,
						flags.PasswordFileFlag,
						flags.LightKDFFlag,
					},
					Description: `
Creates a new account and prints the address.

The account is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your account in the future.

For non-interactive use the passphrase can be specified with the --password flag:

    panotool account new --password=file

Note, this is meant to be used for testing only, it is a bad idea to save your
password to file or expose in any other way.
`,
				},
				{
					Name:      "update",
					Usage:     "Update an existing account",
					Action:    accountUpdate,
					ArgsUsage: "<address>",
					Flags: []cli.Flag{
						flags.DataDirFlag,
						flags.KeyStoreDirFlag,
						flags.LightKDFFlag,
					},
					Description: `
Update an existing account.

The account is saved in the newest version in encrypted format, you are prompted
for a passphrase to unlock the account and another to save the updated file.

This same command can therefore be used to migrate an account of a deprecated
format to the newest format or change the password for an account.

For non-interactive use the passphrase can be specified with the --password flag:

    panotool account update --password=file <address>

Since only one password can be given, only format update can be performed,
changing your password is only possible interactively.
`,
				},
				{
					Name:   "import",
					Usage:  "Import a private key into a new account",
					Action: accountImport,
					Flags: []cli.Flag{
						flags.DataDirFlag,
						flags.KeyStoreDirFlag,
						flags.PasswordFileFlag,
						flags.LightKDFFlag,
					},
					ArgsUsage: "<keyFile>",
					Description: `
    panotool account import <keyfile>

Imports an unencrypted private key from <keyfile> and creates a new account.
Prints the address.

The keyfile is assumed to contain an unencrypted private key in hexadecimal format.

The account is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your account in the future.

For non-interactive use the passphrase can be specified with the --password flag:

    panotool account import --password=file <keyfile>

Note:
As you can directly copy your encrypted accounts to another Pano instance,
this import mechanism is not needed when you transfer an account between
nodes.
`,
				},
			},
		},

		{
			Name:     "validator",
			Usage:    "Manage validators",
			Category: "VALIDATOR COMMANDS",
			Description: `
Create a new validator private key.

It supports interactive mode, when you are prompted for password as well as
non-interactive mode where passwords are supplied via a given password file.
Non-interactive mode is only meant for scripted use on test networks or known
safe environments.

Make sure you remember the password you gave when creating a new validator key.
Without it you are not able to unlock your validator key.

Note that exporting your key in unencrypted format is NOT supported.

Keys are stored under <DATADIR>/keystore/validator.
It is safe to transfer the entire directory or the individual keys therein
between Pano nodes by simply copying.

Make sure you backup your keys regularly.
`,
			Subcommands: []cli.Command{
				{
					Name:   "new",
					Usage:  "Create a new validator key",
					Action: validatorKeyCreate,
					Flags: []cli.Flag{
						flags.DataDirFlag,
						flags.KeyStoreDirFlag,
						flags.PasswordFileFlag,
					},
					Description: `
Creates a new validator private key and prints the public key.

The key is saved in encrypted format, you are prompted for a passphrase.

You must remember this passphrase to unlock your key in the future.

For non-interactive use the passphrase can be specified with the --validator.password flag:

Note, this is meant to be used for testing only, it is a bad idea to save your
password to file or expose in any other way.
`,
				},
				{
					Name:   "convert",
					Usage:  "Convert an account key to a validator key",
					Action: validatorKeyConvert,
					Flags: []cli.Flag{
						flags.DataDirFlag,
						flags.KeyStoreDirFlag,
					},
					ArgsUsage: "<account address> <validator pubkey>",
					Description: `
Converts an account private key to a validator private key and saves in the validator keystore.
`,
				},
			},
		},
	}

	app.Before = func(ctx *cli.Context) error {
		return debug.Setup(ctx)
	}
	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		return nil
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	return app.Run(args)
}
