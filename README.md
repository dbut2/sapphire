# Sapphire

[![Quality](https://github.com/dbut2/sapphire/actions/workflows/quality.yaml/badge.svg)](https://github.com/dbut2/sapphire/actions/workflows/quality.yaml)
[![Build](https://github.com/dbut2/sapphire/actions/workflows/build.yaml/badge.svg)](https://github.com/dbut2/sapphire/actions/workflows/build.yaml)

## Overview

Sapphire is a Game Boy Advance (GBA) emulator written in Go. It aims to provide an accurate representation of GBA hardware allowing you to enjoy your favorite GBA games.

### Features

- High-level emulation of GBA hardware components
- Cross-platform support
- User-friendly graphical interface

### How to Run

To run a GBA game ROM using Sapphire:

1. Install the emulator using the provided installation files.
2. Run the emulator, passing the path to your game ROM file:

```shell
sapphire --game /path/to/your/game.rom
```

Alternatively, you can open the emulator and load the ROM file through the graphical user interface.

### Building from Source

To build Sapphire from source, ensure you have Go installed and run the following command:

```shell
make build
```

This will compile the source code and produce an executable in the `build` directory. You can then start the emulator with:

```shell
./Sapphire.app/Contents/MacOS/sapphire
```

### Testing

Run the test suite to ensure the emulator is working as expected:

```shell
make test
```

### Linting

To lint the code and ensure it follows Go best practices, run:

```shell
make lint
```

### Packaging

To package the build into a distributable format, use:

```shell
make package
```

This will create a `.zip` and `.dmg` file containing the emulator application in the `build` directory.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request or open an Issue on the [GitHub repository](https://github.com/dbut2/sapphire).

## License

Sapphire is released under the MIT License. See the LICENSE file for more details.

### Big WIP energy

GBA Emulator written in Go

```
