<div align="center">
  
<h2>
<div><img src="https://github.com/user-attachments/assets/ba2ca552-dcb4-4379-8a8d-2e9a460fa452" width="120" height="120" alt="Bappa Framework Logo"></div>
  Bappa
  <p>(www.bappa.net)</p>
</h2>
</div>

A code-first 2D game engine/framework for Go, providing an Entity Component System (ECS) architecture to enable developers to build captivating projects.

> ⚠️ **Project Status: (Shelved)**
>
> **Bappa** is a game engine, no longer in active development.
>
> It was a self-directed project to build a foundational framework for 2D games in Go ([Ebiten](https://ebitengine.org/)), and it was used to create the prototype ["Concrete Echos."](https://github.com/TheBitDrifter/concrete_echos)
>
> This project was a massive learning experience and contains many early design decisions that I would architect differently today. The lessons learned while building this engine are the foundation for any future, more robust version.
>
> I've left the repository public as a "read-only" example of this design and iteration process.

## Getting Started

The best way to get started with Bappa is through examples and documentation, or use a `bappacreate` starter template for hands on experience!

[Getting Started](https://bappa.net/docs/getting-started) | [Examples](https://bappa.net/examples) | [Docs](https://bappa.net/docs)

<table>
  <tr>
    <td align="center"><img src="https://github.com/user-attachments/assets/35bc153f-fd00-4833-9970-a51108ada8e8" width="100%"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/2b8962d6-6315-4e8e-84ac-31e50f713977" width="100%"></td>
    <td align="center"><img src="https://github.com/user-attachments/assets/fab72558-50c0-4a24-b01e-77d5547ae905" width="100%"></td>
  </tr>
  <tr>
    <td align="center">LDTK Split Screen Platformer Template</td>
    <td align="center">Topdown Split Screen Template</td>
    <td align="center">LDTK Netcode POC Template</td>
  </tr>
</table>



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

### Drip

- **Import Path**: `github.com/TheBitDrifter/bappa/drip`
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/bappa/drip)
- **Description**: Basic TCP server to run core Bappa systems, receive input, and broadcast state

## License

The Bappa Framework is licensed under the [MIT License](LICENSE)

### Acknowledgments

Bappa is built on top of [Ebiten](https://ebiten.org/), a dead simple 2D game library for Go. Ebiten is also licensed under the [Apache License 2.0](https://github.com/hajimehoshi/ebiten/blob/main/LICENSE).

This project wouldn't be possible without the incredible work of Hajime Hoshi and all the Ebiten contributors.
