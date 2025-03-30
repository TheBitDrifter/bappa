<div align="center">
  <img src="https://github.com/user-attachments/assets/ba2ca552-dcb4-4379-8a8d-2e9a460fa452" width="120" height="120" alt="Bappa Framework Logo">
  <h1>The Bappa Framework</h1>
</div>

A code-first game engine for Go, providing an Entity Component System (ECS) architecture to enable developers to build captivating 2D games.

## Installation

Bappa is a cohesive collection of packages, with the main entrypoint being coldbrew:

```bash
go get github.com/TheBitDrifter/bappa/coldbrew@latest
```

## Examples

The best way to learn how to use Bappa is to read the examples and docs, or use a starter template!

<table>
  <tr>
    <td align="center"><img src="https://github.com/user-attachments/assets/35bc153f-fd00-4833-9970-a51108ada8e8" width="100%"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/2b8962d6-6315-4e8e-84ac-31e50f713977" width="100%"></td>
  </tr>
  <tr>
    <td align="center">LDTK Split Screen Platformer Template</td>
    <td align="center">Topdown Split Screen Template</td>
  </tr>
</table>

[Homepage](https://dl43t3h5ccph3.cloudfront.net) | [Examples](https://dl43t3h5ccph3.cloudfront.net/examples) | [Docs](https://dl43t3h5ccph3.cloudfront.net/docs)

## Framework Components

The Bappa Framework is organized as a monorepo, combining all core packages in one repository. This makes it easy to follow all development activity and ensures consistency across the framework.

### Coldbrew

- **Import Path**: `github.com/TheBitDrifter/bappa/coldbrew`
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/bappa/coldbrew)
- **Description**: Main package for client-side game operations, handling rendering, input, scene management and cameras

### Blueprint

- **Import Path**: `github.com/TheBitDrifter/bappa/blueprint`
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/bappa/blueprint)
- **Description**: Core component definitions and scene planning functionality

### Warehouse

- **Import Path**: `github.com/TheBitDrifter/bappa/warehouse`
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/bappa/warehouse)
- **Description**: Entity storage, archetype management, and query system for the ECS architecture

### Tteokbokki

- **Import Path**: `github.com/TheBitDrifter/bappa/tteokbokki`
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/bappa/tteokbokki)
- **Description**: Physics and collision detection systems

### Table

- **Import Path**: `github.com/TheBitDrifter/bappa/table`
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/bappa/table)
- **Description**: Efficient data storage optimized for game objects

## License

The Bappa Framework is licensed under the [MIT License](LICENSE)

### Acknowledgments

Bappa is built on top of [Ebiten](https://ebiten.org/), a dead simple 2D game library for Go. Ebiten is also licensed under the [Apache License 2.0](https://github.com/hajimehoshi/ebiten/blob/main/LICENSE).

This project wouldn't be possible without the incredible work of Hajime Hoshi and all the Ebiten contributors.
