# Sapphire

[![Quality](https://github.com/dbut2/sapphire/actions/workflows/quality.yaml/badge.svg)](https://github.com/dbut2/sapphire/actions/workflows/quality.yaml)
[![Build](https://github.com/dbut2/sapphire/actions/workflows/build.yaml/badge.svg)](https://github.com/dbut2/sapphire/actions/workflows/build.yaml)

Sapphire is an emulator for the Game Boy Advance (GBA) system written in Go.

## Features

- Emulates GBA hardware interactions
- Allows ROM execution with BIOS emulation
- Supports DMA (Direct Memory Access) transfers
- Implements timers and display drawing
- Provides a windowed application for playing games
- Extensible through debugger hooks for custom behavior analysis

## Installation

To install Sapphire, you can clone the repository and build from source:

```bash
git clone https://github.com/dbut2/sapphire.git
cd sapphire
make build
```

This will compile the emulator and create an executable in the `build/` directory.

## Usage

To run a GBA game, you must have the game ROM file. Once you have it, you can start Sapphire and load the game as follows:

```bash
./Sapphire.app/Contents/MacOS/sapphire --game /path/to/game.gba
```

If the path to the ROM file is not provided using the `--game` flag, a file dialog will prompt you to select the game ROM file to load.

Ensure that the ROM file is a `.gba` file representing a Game Boy Advance game.

## Development and Debugging

Sapphire is a work in progress. Contributions are welcome. Visit the project's issues page to report any bugs or feature requests and to view the list of known issues.

For development purposes:

- `make build` - Build the project for macOS
- `make run` - Build and execute the project for macOS
- `make run-debug` - Execute the debugger for the project
- `make package` - Package the built binary for distribution
- `make test` - Run tests
- `make lint` - Check the code with a linter

In debug mode, custom hooks can be injected into the emulator's loop to analyze behavior or modify the emulator's operations.

## License

Sapphire is open-source software licensed under the [MIT License](https://opensource.org/licenses/MIT).

![Sapphire](sapphire.png)
```