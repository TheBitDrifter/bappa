# Coldbrew

Coldbrew is the client and scene management system for the [Bappa Framework](https://dl43t3h5ccph3.cloudfront.net/). It serves as the top-level interface for game developers, providing essential functionality for game lifecycle, rendering, input processing, scene transitions, and resource management.

## Features

- **Game Loop Management**: Handles update-render cycles with configurable timing
- **Scene Management**: Create, activate, and transition between multiple scenes
- **Multi-Camera Support**: Flexible viewport system with split-screen capabilities
- **Input Abstraction**: Device-independent input system for keyboard, mouse, gamepad, and touch
- **Asset Management**: Automatic loading and caching of sprites, animations, and sounds
- **System Organization**: Structured approach to game logic and rendering systems

## Installation

```bash
go get github.com/TheBitDrifter/coldbrew
```

## Docs

The official docs can be found [here](https://dl43t3h5ccph3.cloudfront.net/)

## Examples

Examples can be found [here](https://dl43t3h5ccph3.cloudfront.net/examples)

## Technical Foundation

Coldbrew is built on the excellent [Ebiten](https://github.com/hajimehoshi/ebiten) framework for Go, which provides hardware-accelerated rendering and cross-platform support. Ebiten handles the low-level graphics rendering, input detection, and platform compatibility, while Coldbrew extends these capabilities with a comprehensive entity-component system and game management features.

Coldbrew integrates with other Bappa Framework components to provide a complete game development ecosystem.

## License

MIT License - see the [LICENSE](LICENSE) file for details.
