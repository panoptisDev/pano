# Pano

EVM-compatible chain secured by the Lachesis consensus algorithm.

## Building the source

Building Pano requires both a Go (version 1.24 or later) and a C compiler. You can install
them using your favourite package manager. Once the dependencies are installed, run:

```sh
make all
```
The build outputs are ```build/panod``` and ```build/panotool``` executables.

## Initialization of the Pano Database

You will need a genesis file to join a network. See [lachesis_launch](https://github.com/panoptisDev/lachesis_launch) for details on obtaining one. Once you have a genesis file, initialize the DB:

```sh
panotool --datadir=<target DB path> genesis <path to the genesis file>
```

## Running `panod`

Going through all the possible command line flags is out of scope here,
but we've enumerated a few common parameter combos to get you up to speed quickly
on how you can run your own `panod` instance.

### Launching a network

To launch a `panod` read-only (non-validator) node for network specified by the genesis file:

```sh
panod --datadir=<DB path>
```

### Configuration

As an alternative to passing the numerous flags to the `panod` binary, you can also pass a
configuration file via:

```sh
panod --datadir=<DB path> --config /path/to/your/config.toml
```

To get an idea of what the file should look like you can use the `dumpconfig` subcommand to
export the default configuration:

```sh
panotool --datadir=<DB path> dumpconfig
```

### Validator

To create a new validator private key:

```sh
panotool --datadir=<DB path> validator new
```

To launch a validator, use the `--validator.id` and `--validator.pubkey` flags. See the [Pano Documentation](https://docs.panoptisDev.com/) for details on obtaining a validator ID and registering your initial stake.

```sh
panod --datadir=<DB path> --validator.id=YOUR_ID --validator.pubkey=0xYOUR_PUBKEY
```

`panod` will prompt for a password to decrypt your validator private key. Optionally, use `--validator.password` to specify a password file.

#### Participation in discovery

Optionally, specify your public IP to improve connectivity. Ensure your TCP/UDP p2p port (5050 by default) is open:

```sh
panod --datadir=<DB path> --nat=extip:1.2.3.4
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. Please also review our [Code of Conduct](CODE_OF_CONDUCT.md).
