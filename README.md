<div align="center">
  <img src="https://github.com/user-attachments/assets/ba2ca552-dcb4-4379-8a8d-2e9a460fa452" width="120" height="120" alt="Bappa Framework Logo">
  <h1>The Bappa Framework</h1>
</div>

A code-first game engine for Go, providing an Entity Component System (ECS) architecture to enable developers to build captivating 2D games.

## Installation

Bappa is a cohesive collection of packages, with the main entrypoint being coldbrew:
```bash
go get github.com/TheBitDrifter/coldbrew@latest
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


[Homepage](https://dl43t3h5ccph3.cloudfront.net) | [Examples](https://dl43t3h5ccph3.cloudfront.net/examples) |  [docs](https://dl43t3h5ccph3.cloudfront.net/docs)


## Core Repositories
These are the packages/libraries that form the Bappa Framework. With the exception of coldbrew, they can be used independently if desired.

### Coldbrew

- [GitHub Repository](https://github.com/TheBitDrifter/coldbrew)
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/coldbrew)
- **Description**: Main package for client-side game operations, handling rendering, input, scene management and cameras

### Blueprint

- [GitHub Repository](https://github.com/TheBitDrifter/blueprint)
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/blueprint)
- **Description**: Core component definitions and scene planning functionality


### Warehouse

- [GitHub Repository](https://github.com/TheBitDrifter/warehouse)
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/warehouse)
- **Description**: Entity storage, archetype management, and query system for the ECS architecture

### Tteokbokki

- [GitHub Repository](https://github.com/TheBitDrifter/tteokbokki)
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/tteokbokki)
- **Description**: Physics and collision detection systems

### Table

- [GitHub Repository](https://github.com/TheBitDrifter/table)
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/table)
- **Description**: Efficient data storage optimized for game objects

### Mask

- [GitHub Repository](https://github.com/TheBitDrifter/mask)
- [GoDoc](https://pkg.go.dev/github.com/TheBitDrifter/mask)
- **Description**: Bitmasking utilities for component filtering

## Tools & Utilities

### BappaCreate

- [GitHub Repository](https://github.com/TheBitDrifter/bappacreate)
- **Description**: Template generator tool for quickly creating new Bappa game projects
- **Templates**:
  - [Topdown](https://github.com/TheBitDrifter/bappacreate/tree/main/templates/topdown)
  - [Topdown Split-Screen](https://github.com/TheBitDrifter/bappacreate/tree/main/templates/topdown-split)
  - [Platformer](https://github.com/TheBitDrifter/bappacreate/tree/main/templates/platformer)
  - [Platformer Split-Screen](https://github.com/TheBitDrifter/bappacreate/tree/main/templates/platformer-split)
  - [Sandbox](https://github.com/TheBitDrifter/bappacreate/tree/main/templates/sandbox)
