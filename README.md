# CogMem Go Library

[![Go Report Card](https://goreportcard.com/badge/github.com/lexlapax/cogmem)](https://goreportcard.com/report/github.com/lexlapax/cogmem)
[![Build Status](https://github.com/lexlapax/cogmem/actions/workflows/ci.yml/badge.svg)](https://github.com/lexlapax/cogmem/actions)

CogMem is a cognitively-inspired memory library for Go, providing persistent, structured, dynamic, and context-rich long-term memory (episodic, semantic, procedural) functionality for AI agent systems. This repository contains the Go module scaffold under `cogmem-go/`, currently in Phase 1 of implementation.

---

## Features

CogMem implements:

* Clean/Hexagonal Architecture (Domain, Application, Infrastructure)
* PostgreSQL persistence with the pgvector extension
* Partitioned Episodic Memory storage and retrieval (vector similarity & recency filters)
* Modular scripting via Lua engine
* Valence scoring and metadata support
* Test-Driven Development (TDD) and Docker-first workflow


---

## Current Status

See [Implementation Plan](implementation-plan.md) for the full roadmap and upcoming tasks.

## Getting Started


# CogMem

<!-- Short, engaging description of your project. What does it do? Who is it for? -->

## Table of Contents

*   [Getting Started](#getting-started)
    *   [Prerequisites](#prerequisites)
    *   [Installation](#installation)
    *   [Running the Application](#running-the-application)
*   [Key Documentation](#key-documentation)
*   [Usage](#usage)
*   [Project Structure](#project-structure)
*   [Running Tests](#running-tests)
*   [Deployment](#deployment)
*   [Contributing](#contributing)
*   [License](#license)

## Getting Started

<!-- Instructions on how to get the project set up and running locally. -->

### Prerequisites

* Go 1.24+
* Docker & Docker Compose
* PostgreSQL 15+ (for production/testing) or use Docker Compose for local development

### Clone and Install

```bash
git clone https://github.com/lexlapax/cogmem.git
cd cogmem-go
make tidy    # download Go module dependencies
```

### Docker Compose (Local PostgreSQL)

```bash
make docker-up   # spin up Postgres+pgvector
```
Then run library checks and tests:
```bash
make all         # fmt, vet, and test
```

*Note:* A `docker-compose.yml` providing PostgreSQL with pgvector support will be added as part of the infrastructure setup (Phase 1).

## Key Documentation

Understand the project goals, design, and plan:

## Configuration

CogMem uses Viper for configuration, loading in order:
1. `config.yaml` in the working directory
2. `.env` file
3. Environment variables

Example `config.yaml` placed alongside your application:
```yaml
database_url: postgres://postgres:password@localhost:5432/cogmem?sslmode=disable
embedding_dim: 1536
decay_base_rate: 0.01
decay_valence_weight: 0.5
decay_interval: 1h
# Additional settings (e.g., Lua sandbox) follow here
```
*   **Implementation Plan:** [./implementation-plan.md](./implementation-plan.md)
*   **Structure Philosophy:** [./project-structure.md](./project-structure.md)

## Project Structure

High-level layout under `cogmem-go/`:

```text
cogmem-go/
├── cmd/                      # CLI or example consumers
├── internal/
│   ├── domain/               # Entities, repository & service interfaces
│   ├── application/          # Use-case services and ports
│   ├── infrastructure/       # Persistence adapters, engines, config, logging
│   └── port/                 # Shared interface definitions
├── pkg/                      # Public client interface and implementation
├── migrations/               # Database migration scripts
├── scripts/                  # Default/example Lua scripts
├── test/integration/         # Integration tests and fixtures
├── go.mod
└── go.sum
```

## Documentation


## Testing

Run all unit and integration tests:
```bash
cd cogmem-go
go test ./...
```
Integration tests will spin up PostgreSQL with pgvector via Testcontainers.
## Other Considerations
### Deployment
<!-- Briefly describe the deployment process or link to more detailed documentation. -->
<!-- Mention CI/CD pipelines if applicable. -->
## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

TBD