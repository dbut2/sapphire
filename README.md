# Sapphire

[![Quality](https://github.com/dbut2/sapphire/actions/workflows/quality.yaml/badge.svg)](https://github.com/dbut2/sapphire/actions/workflows/quality.yaml)

Sapphire is an emulator for the Game Boy Advance (GBA) system written in Go.

## Requirements

To run Sapphire, you will need a Game Boy Advance BIOS file (`bios.gba`). This file should be placed in the `gba/` directory before building, or you can use the `make decrypt` command if you have the encrypted version and the necessary key.

## Installation

To install Sapphire, you can clone the repository and build it from source:

```bash
git clone https://github.com/dbut2/sapphire.git
cd sapphire
make build
```

This will compile the emulator and create an executable in the `build/` directory, and also package it into `Sapphire.app`.

## Usage

To run a GBA game, you will need the game ROM file. Once you have it, you can run Sapphire and load the game as follows:

```bash
./Sapphire.app/Contents/MacOS/sapphire --game /path/to/game.gba
```

Alternatively, if you have not provided the path to the ROM file using the `--game` flag, a file dialog will prompt you to select the game ROM file to load.

Ensure the ROM file is a `.gba` file that represents a Game Boy Advance game.

### Controls

| GBA Button | Keyboard Key |
|------------|--------------|
| A          | Z            |
| B          | X            |
| L          | A            |
| R          | S            |
| Start      | Enter        |
| Select     | Backspace    |
| D-Pad      | Arrow Keys   |
| Fast Forward| Space       |

## Development

Sapphire is a work in progress, and contributions are welcome. Visit the project's issues page to report any bugs or feature requests and to see the list of known issues.

For development, besides the standard Go tools, Sapphire includes a `Makefile` that simplifies common tasks:

- `make clean` - Remove any build artifacts.
- `make build` - Build the project for macOS.
- `make wasm` - Build the project for WebAssembly.
- `make run` - Build and execute the project for macOS.
- `make package` - Package the built binary for distribution.
- `make test` - Run tests.
- `make lint` - Run linter to check the code.
- `make decrypt` - Decrypt BIOS and default gamepak (requires `SAPPHIRE_KEY`).

## License

Sapphire is licensed under the [MIT License](https://opensource.org/licenses/MIT).
