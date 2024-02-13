# Sapphire

[![Quality](https://github.com/dbut2/sapphire/actions/workflows/quality.yaml/badge.svg)](https://github.com/dbut2/sapphire/actions/workflows/quality.yaml)
[![Build](https://github.com/dbut2/sapphire/actions/workflows/build.yaml/badge.svg)](https://github.com/dbut2/sapphire/actions/workflows/build.yaml)

### Big WIP energy

GBA Emulator written in Go

## Getting Started

To start using Sapphire, you will need to have Go installed on your machine.

### Prerequisites

- Go 1.21 or higher

### Running Sapphire

You can build and run Sapphire using the Makefile provided in the repository:

```bash
make run
```

Alternatively, you can manually build and run the emulator:

```bash
go build -o build/sapphire ./cmd/sapphire
./build/sapphire
```

### Controls

The emulator maps keyboard inputs to GBA controls:

- Arrow keys for directional input
- Z for A button
- X for B button
- A for L button
- S for R button
- Enter for Start
- Space for Select

### ROMs

Sapphire requires a GBA ROM to run. You can load a ROM using the `-g` or `--game` flag followed by the path to the ROM file:

```bash
./build/sapphire -g path/to/rom.gba
```

## Features

- ARM and Thumb instruction sets
- Memory-mapped I/O and interrupt handling
- Audio, graphics, timers, serial communication and more

## Contributing

Contributions are welcome. Please feel free to fork the project, make changes, and submit pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Thanks to the [TONC](http://www.coranac.com/tonc/text/) guide for the insightful GBA programming resources.
- Emulator development is guided by the info found on [GBATEK](https://problemkaputt.de/gbatek.htm).
