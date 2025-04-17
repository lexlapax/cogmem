## CogMem Library: Phased Prototyping Implementation Plan

**Vision:** Develop a robust, testable, and maintainable Go library implementing the core CogMem architecture (v2.0 ADD) suitable for integration into AI agent systems. The prototype will demonstrate key features incrementally across three phases.

**Guiding Principles:**

*   **Test-Driven Development (TDD):** Tests will be written before or alongside implementation code. Unit tests cover domain logic and application services (using mocks). Integration tests cover persistence, engine interactions, and client behavior (using test containers).
*   **Docker-First:** Docker Compose will manage the PostgreSQL+pgvector dependency for local development and integration testing, ensuring environment consistency.
*   **Feature-Based Slicing (Vertical Slices):** Each phase delivers demonstrable end-to-end functionality through the library's public interface (`CogMemClient`), cutting across internal layers.
*   **Layered Architecture (Onion/Clean/Hexagonal):** The Go codebase (`internal/` directory) will strictly separate Domain, Application, and Infrastructure concerns, using interfaces (`ports/` or within domain/application) to mediate dependencies.

**Proposed Go Project Structure (Illustrative):**

```
cogmem/
├── cmd/                   # Example consumer/CLI tool (Optional, for testing library)
│   └── main.go
├── internal/
│   ├── domain/            # Core business logic & entities (independent)
│   │   ├── entity/        # Core structs: EpisodicMemory, SemanticNode, Edge, Script, ValenceScore, PartitionContext
│   │   ├── repository/    # Repository interfaces: Episodic, SemanticNode, SemanticEdge, Script
│   │   ├── service/       # Domain service interfaces: ValenceCalculator, ScriptExecutor, DecayLogic
│   │   └── valueobject/   # Value objects if needed
│   ├── application/       # Use cases / Application services (depend on domain)
│   │   ├── service/       # MemoryService, ScriptingService, NodeService, EdgeService, DecayService
│   │   ├── dto/           # Data Transfer Objects used by application services internally
│   │   └── port/          # Application input ports (interfaces implemented by services)
│   ├── infrastructure/    # Implementations of ports & external concerns (depend on domain interfaces)
│   │   ├── persistence/   # Database adapters (implements domain.repository interfaces)
│   │   │   └── postgres/  # Postgres implementation using pgx
│   │   ├── engine/        # Engine implementations (implements domain.service interfaces)
│   │   │   ├── lua/       # Lua script executor (gopher-lua)
│   │   │   └── valence/   # Valence calculator implementation
│   │   ├── config/        # Configuration loading (Viper)
│   │   └── log/           # Logging setup
│   └── port/              # Interfaces defining how infrastructure communicates (driving side for DB, driven side for Engines) - Matches domain repo/service interfaces
├── pkg/                   # Public library code
│   ├── cogmem/            # Main package with CogMemClient interface, NewClient constructor, public DTOs
│   └── client/            # Internal implementation of CogMemClient interface
├── migrations/            # Database schema migrations (e.g., using goose or sql-migrate)
├── scripts/               # Default/example Lua scripts
├── test/                  # Integration tests, test helpers, fixtures
│   └── integration/
├── .github/               # CI/CD workflows
├── config.yaml            # Default configuration file
├── docker-compose.yml     # For development/testing database
├── Dockerfile             # For building example consumer or test environment
└── go.mod
└── go.sum
```
*(Note: `internal/port` usage can vary; interfaces might live closer to their users in Domain/Application layers instead)*

---

### Phase 1: Foundational Storage, Retrieval & Partitioning

**Goal:** Establish the core library structure, configuration, persistence layer for Episodic Memory, and the public client interface for basic storage and retrieval, strictly enforcing partitioning. Deliver the first testable end-to-end vertical slice.

**Key Vertical Slice:** As a consuming application, I can initialize the `CogMemClient`, store text (`EpisodicMemoryInput`) associated with a specific `PartitionContext` (user/entity), and later retrieve relevant memories using vector similarity search or recentness filters, receiving `EpisodicMemoryOutput` DTOs, with partitioning correctly applied.

**Tasks & Implementation Details:**

1.  **Project Setup & Tooling (Infrastructure):**
    *   Initialize Go project (`go mod init`), set up directory structure.
    *   Set up Git repository & basic branching strategy.
    *   Create `docker-compose.yml` with a `postgres` service using an image supporting `pgvector` (e.g., `pgvector/pgvector:pg15` or official postgres + init script). Expose port for testing.
    *   Implement configuration loading (`internal/infrastructure/config`) using Viper, defining `cogmem.Config` struct, loading from `config.yaml`, `.env`, and env vars.
    *   Set up basic logging (`internal/infrastructure/log`).
    *   Establish testing infrastructure: Add `stretchr/testify`, configure Go test runner. Set up `testcontainers-go` for PostgreSQL integration tests.
    *   Define initial CI pipeline (e.g., GitHub Actions) for linting, vetting, building, and running unit & integration tests.

2.  **Domain Layer (Phase 1):**
    *   Define core entities: `EpisodicMemory`, `PartitionContext` (`internal/domain/entity`). Focus on fields needed for Phase 1 (ID, UserID, EntityID, Content, Embedding, Timestamp, ShareScope, LastAccessed, AccessibilityScore - default 1.0).
    *   Define repository interface: `EpisodicRepository` (`internal/domain/repository`) with methods: `Save(ctx, mem)`, `FindByID(ctx, id, pCtx)`, `FindByVector(ctx, vector, limit, pCtx)`, `FindRecent(ctx, limit, pCtx)`. Methods must accept `PartitionContext`.
    *   *TDD:* Write unit tests for entity validation/creation logic (if any).

3.  **Infrastructure Layer (Persistence - Phase 1):**
    *   Implement `EpisodicRepository` using `pgxpool` (`internal/infrastructure/persistence/postgres`). Connect using `DatabaseURL` from config.
    *   Write DB migration script (`migrations/`) for `episodic_memory` table (SQL definition from ADD). Include `pgvector` extension creation. Use a migration tool like `goose`.
    *   Implement robust partitioning logic within repository methods: dynamically add `WHERE user_id = $1 AND (entity_id IS NULL OR entity_id = $2)` clauses based on `PartitionContext`. Handle `share_scope` filtering if needed for basic 'user' scope initially.
    *   *TDD:* Write integration tests (`test/integration/`) for `PostgresEpisodicRepository`. Use test containers to spin up Postgres. Verify:
        *   Saving and retrieving by ID works.
        *   Vector search (`FindByVector`) returns expected results using cosine similarity/distance.
        *   `FindRecent` returns correctly ordered results.
        *   **Crucially:** Test that queries only return data matching the provided `PartitionContext`, and fail/return empty for incorrect contexts.

4.  **Application Layer (Phase 1):**
    *   Define `MemoryService` interface (`internal/application/port`) and implementation (`internal/application/service`).
    *   Implement use cases: `StoreEpisodic(ctx, pCtx, dto)`, `RetrieveEpisodic(ctx, pCtx, queryDto)`.
    *   Inject `domain.repository.EpisodicRepository`.
    *   Handle mapping between public DTOs (`cogmem.EpisodicMemoryInput/Output`) and internal domain entities.
    *   Implement embedding generation logic: For Phase 1, this might be a *placeholder* or call out to a separate embedding service/library (dependency injected). Define embedding dimension based on config.
    *   *TDD:* Write unit tests for `MemoryService`, mocking the `EpisodicRepository` and embedding generation. Test input validation, DTO mapping, and correct repository method calls.

5.  **Public Library Interface (`pkg/cogmem`, `pkg/client` - Phase 1):**
    *   Define the public `CogMemClient` interface in `pkg/cogmem/client.go` with methods for Phase 1: `StoreEpisodic`, `RetrieveEpisodic`, `GetEpisodicByID`, `Close`, plus `NewClient` constructor signature.
    *   Define public DTOs (`EpisodicMemoryInput`, `EpisodicMemoryOutput`, `PartitionContext`, `RetrievalQuery` - limited fields) in `pkg/cogmem/dto.go` (or similar).
    *   Implement `NewClient` function in `pkg/cogmem/factory.go` (or similar). It should:
        *   Load configuration using Viper (`internal/infrastructure/config`).
        *   Initialize `pgxpool` connection pool.
        *   Initialize `PostgresEpisodicRepository`.
        *   Initialize `MemoryService` (with repo & embedding generator).
        *   Instantiate the internal client implementation (`pkg/client/client.go`).
        *   Return the `CogMemClient` interface.
    *   Implement the internal `client` struct (`pkg/client/client.go`) which holds references to application services and implements the `CogMemClient` interface methods by calling the appropriate service methods. Implement `Close()` to close the DB pool.
    *   *TDD:* Write integration tests for the `NewClient` function: verify config loading, successful initialization with valid config, error handling with invalid config. Write integration tests for the client methods implemented in this phase, ensuring they interact correctly with the underlying services and database (can use test containers).

6.  **Documentation & Examples:**
    *   Initialize `README.md` explaining library purpose, setup, and Phase 1 usage.
    *   Add GoDoc comments to public interfaces, structs, and functions.
    *   Provide a simple example in `cmd/` showing how to initialize and use the Phase 1 client methods.

**Phase 1 Outcome:** A consumable Go library (`cogmem`) installable via `go get`. A consuming application can initialize the client, connect to a Postgres DB (managed externally), store episodic memories with partitioning, and retrieve them via vector search or recentness. The library structure follows Clean Architecture principles and is supported by unit and integration tests.

---

### Phase 2: Integrating Core Cognitive Features (Valence & Scripting)

**Goal:** Introduce the Valence Engine for scoring memories and the Lua Scripting Engine for custom operations. Enable basic Semantic Node storage.

**Key Vertical Slices:**
*   As a consuming application, when I store episodic memory, it is automatically assigned a calculated `ValenceScore` (v1: polarity focus), and I can filter/sort retrieval results based on this valence.
*   As a consuming application, I can store predefined Lua scripts and execute them via the `CogMemClient`, passing parameters and receiving results, allowing custom memory queries (e.g., find negative memories for a user).
*   As a consuming application, I can store and retrieve basic `SemanticNode` data.

**Tasks & Implementation Details:**

1.  **Domain Layer (Phase 2):**
    *   Define `ValenceScore` entity (`internal/domain/entity`) with Polarity, GoalRelevance, Arousal.
    *   Define `ValenceCalculator` interface (`internal/domain/service`) with `Calculate(ctx, text, taskHint) (ValenceScore, error)`.
    *   Add `Valence` field (type `ValenceScore`, stored as JSONB) to `EpisodicMemory` entity.
    *   Define `Script` entity (`internal/domain/entity`) with Name, Language, Code, etc.
    *   Define `ScriptRepository` interface (`internal/domain/repository`) with `Save`, `FindByName`.
    *   Define `ScriptExecutor` interface (`internal/domain/service`) with `Execute(ctx, scriptCode, params) (result, logs, error)`.
    *   Define `SemanticNode` entity (`internal/domain/entity`).
    *   Define `SemanticNodeRepository` interface (`internal/domain/repository`) with `Save`, `FindByID`, `FindByNameAndType`.
    *   *TDD:* Unit tests for new entities/validation.

2.  **Infrastructure Layer (Engines - Phase 2):**
    *   Implement `ValenceCalculator` (`internal/infrastructure/engine/valence`) using the v1 algorithm (keyword lists, potentially `vader-go`). Make keywords/weights configurable via Viper.
    *   Implement `ScriptExecutor` (`internal/infrastructure/engine/lua`) using `gopher-lua`. Implement sandboxing (disable standard libs per config). Implement safe Go function injection for DB access (`luaSafeFindEpisodic`, etc.) which *must* accept and enforce `PartitionContext`. Handle execution timeouts from config. Capture `print` statements as logs.
    *   *TDD:* Unit tests for `ValenceCalculator` logic. Unit tests for `ScriptExecutor` sandboxing setup. Integration tests for executing Lua scripts that call the injected safe DB functions (using test containers, verifying partitioning is respected within Lua calls).

3.  **Infrastructure Layer (Persistence - Phase 2):**
    *   Update `episodic_memory` migration/schema to add `valence JSONB` column. Update `PostgresEpisodicRepository` to store/retrieve valence and potentially add `WHERE` clauses or `ORDER BY` clauses based on valence filters in `RetrievalQuery`.
    *   Implement `procedural_scripts` table migration/schema. Implement `PostgresScriptRepository`.
    *   Implement `semantic_nodes` table migration/schema. Implement `PostgresSemanticNodeRepository`.
    *   *TDD:* Update `EpisodicRepository` integration tests for valence storage/retrieval/filtering. Write integration tests for `ScriptRepository` and `SemanticNodeRepository`.

4.  **Application Layer (Phase 2):**
    *   Inject `ValenceCalculator` into `MemoryService`. Modify `StoreEpisodic` to calculate and store valence (unless `ExplicitValence` is provided in input). Modify `RetrieveEpisodic` to handle valence filters/sorting from `RetrievalQuery`.
    *   Create `ScriptingService` (`internal/application/service`). Implement `ExecuteScript(ctx, pCtx, scriptName, params)` use case. Inject `ScriptRepository` (to find script code) and `ScriptExecutor`. Ensure `PartitionContext` is passed correctly if needed by script execution context.
    *   Create `NodeService` (`internal/application/service`). Implement `StoreNode`, `GetNodeByID`, `FindNodes`. Inject `SemanticNodeRepository`. Handle DTO/entity mapping.
    *   *TDD:* Update `MemoryService` unit tests. Write unit tests for `ScriptingService` and `NodeService` (mocking repos/engines).

5.  **Public Library Interface (`pkg/cogmem`, `pkg/client` - Phase 2):**
    *   Add `ValenceScore` struct to public DTOs. Add `ExplicitValence` to `EpisodicMemoryInput`. Add valence filters/sorting options to `RetrievalQuery`. Add `Valence` to `EpisodicMemoryOutput`.
    *   Add `StoreScript`, `ExecuteScript` methods to `CogMemClient` interface. Add `ScriptExecutionInput/Output` DTOs.
    *   Add `StoreSemanticNode`, `GetSemanticNodeByID`, `FindSemanticNodes` methods to `CogMemClient`. Add `SemanticNodeInput/Output` DTOs.
    *   Implement the new methods in the internal client (`pkg/client/client.go`) by calling the corresponding application services.
    *   Update `NewClient` to initialize new repositories and services.
    *   *TDD:* Update integration tests for `Store/RetrieveEpisodic` to check valence handling. Write integration tests for `StoreScript`, `ExecuteScript`, `StoreSemanticNode`, `GetSemanticNodeByID`, `FindSemanticNodes` client methods.

6.  **Documentation & Examples:**
    *   Document Valence Engine v1 logic and configuration.
    *   Document Scripting Engine usage, sandboxing, available safe Lua functions, and error handling.
    *   Document Semantic Node operations.
    *   Update `README.md` and examples in `cmd/`.

**Phase 2 Outcome:** The library now calculates and stores valence for episodic memories, allowing valence-aware retrieval. Users can store and execute sandboxed Lua scripts for custom memory operations. Basic semantic graph node management is available.

---

### Phase 3: Advanced Memory Dynamics & Semantic Relations

**Goal:** Implement remaining core features: Memory Lifespan/Decay, basic Causal Layer support, Semantic Edges, and initial Persona concepts.

**Key Vertical Slices:**
*   As a consuming application, stored memories have an `accessibility_score` that decays over time based on valence, influencing retrieval order. I can start/stop this background decay process.
*   As a consuming application, I can store basic `causal_links` with episodic memories and retrieve them.
*   As a consuming application, I can store and retrieve relationships (`SemanticEdge`) between semantic nodes.
*   As a consuming application, I can provide a `Persona` hint during script execution to potentially influence which scripts are chosen or how they behave (basic implementation).

**Tasks & Implementation Details:**

1.  **Domain Layer (Phase 3):**
    *   Add `AccessibilityScore` and `LastAccessed` fields to relevant entities (`EpisodicMemory`, `SemanticNode`).
    *   Define `DecayLogic` interface (`internal/domain/service`) with `ApplyDecay(ctx, pCtx?)`. Context needed if decay applies per-partition.
    *   Add `CausalLinks` field (e.g., `map[string][]uuid.UUID`) to `EpisodicMemory` entity.
    *   Define `SemanticEdge` entity (`internal/domain/entity`).
    *   Define `SemanticEdgeRepository` interface (`internal/domain/repository`) with `Save`, `FindEdgesByNodeID`, `FindEdgesByRelationType`.
    *   Define `Persona` concept (e.g., a simple string type or struct) in domain.
    *   *TDD:* Unit tests for new/updated entities.

2.  **Infrastructure Layer (Persistence - Phase 3):**
    *   Update migrations/schemas for `episodic_memory` and `semantic_nodes` to include `accessibility_score` (default 1.0) and `last_accessed` (default now). Add `causal_links JSONB` to `episodic_memory`.
    *   Implement `semantic_edges` table migration/schema.
    *   Implement `PostgresSemanticEdgeRepository`.
    *   Update `PostgresEpisodicRepository` and `PostgresSemanticNodeRepository` retrieval methods to update `last_accessed` on read. Implement logic to use `accessibility_score` in sorting options (e.g., `ORDER BY accessibility_score DESC`).
    *   Implement `PostgresEpisodicRepository` methods to store/retrieve `causal_links`.
    *   *TDD:* Integration tests for `SemanticEdgeRepository`. Update episodic/node repository tests to verify `last_accessed` updates and `accessibility_score` sorting. Test causal link storage/retrieval.

3.  **Infrastructure Layer (Engines - Phase 3):**
    *   Implement `DecayLogic` (`internal/infrastructure/engine/decay`) applying the formula from ADD, using configured base rate and valence weight. It will need to fetch items, calculate new scores, and update them via repositories. Consider batching updates.
    *   Update `ScriptRepository` to potentially filter scripts based on `allowed_personas` if a persona hint is provided.
    *   Update `ScriptExecutor` (Lua) to potentially inject the current `Persona` hint into the Lua environment for scripts to use conditionally.
    *   *TDD:* Unit tests for `DecayLogic` calculation. Integration test for decay logic interacting with DB (can use manipulated timestamps). Update scripting tests for persona filtering/injection.

4.  **Application Layer (Phase 3):**
    *   Create `DecayService` (`internal/application/service`). Implement `RunDecayCycle(ctx)` which uses `DecayLogic`.
    *   Create `EdgeService` (`internal/application/service`). Implement `StoreEdge`, `FindEdges`. Inject `SemanticEdgeRepository`.
    *   Modify `MemoryService` and `NodeService` retrieval use cases to update `last_accessed`.
    *   Modify `MemoryService` to handle storing/retrieving `causal_links` DTO field.
    *   Modify `ScriptingService` `ExecuteScript` use case to accept an optional `Persona` hint and pass it down for filtering/injection.
    *   *TDD:* Unit tests for `DecayService`, `EdgeService`. Update other service tests for `last_accessed`, causal links, and persona handling.

5.  **Public Library Interface (`pkg/cogmem`, `pkg/client` - Phase 3):**
    *   Add `AccessibilityScore`, `LastAccessed` to relevant output DTOs (`EpisodicMemoryOutput`, `SemanticNodeOutput`).
    *   Add `CausalLinks` to `EpisodicMemoryInput/Output`.
    *   Add `StoreSemanticEdge`, `FindEdges` methods to `CogMemClient`. Add `SemanticEdgeInput/Output` DTOs.
    *   Add `StartBackgroundTasks() (stop func(ctx context.Context) error, err error)` method to `CogMemClient`.
    *   Modify `ExecuteScript` signature/input DTO to optionally accept a `Persona` string/struct.
    *   Implement the new/modified methods in the internal client (`pkg/client/client.go`). The `StartBackgroundTasks` method will:
        *   Initialize the `DecayService`.
        *   Launch a goroutine that runs `DecayService.RunDecayCycle` periodically based on `config.DecayInterval`.
        *   Return a `stop` function that signals the goroutine to terminate gracefully (e.g., using a context cancellation or channel).
    *   Update `NewClient` to initialize the `EdgeService` and `DecayService`.
    *   *TDD:* Integration tests for `StartBackgroundTasks` (verify it starts/stops, potentially mock decay logic for testing lifecycle). Integration tests for `StoreSemanticEdge`, `FindEdges`. Integration tests for storing/retrieving causal links via client. Test `ExecuteScript` with persona hints. Verify `accessibility_score` appears in relevant DTOs.

6.  **Refinement & Polish:**
    *   Implement Valence Aggregation for Semantic Nodes (potentially as part of `StoreEpisodic` or a background task).
    *   Refine error handling across the library (use custom error types).
    *   Add more comprehensive logging.
    *   Profile performance-critical paths (retrieval, decay) and optimize queries/code.
    *   Add more complex default Lua scripts (`scripts/`).

7.  **Documentation & Examples:**
    *   Document Lifespan/Decay mechanism, configuration, and background task management.
    *   Document Causal Layer usage.
    *   Document Semantic Edge operations.
    *   Document Persona hints.
    *   Provide comprehensive GoDoc for all public interfaces/structs/functions.
    *   Update `README.md` and examples in `cmd/` to cover all features.

**Phase 3 Outcome:** A feature-complete prototype of the CogMem library according to the ADD v2.0 vision. It includes advanced memory dynamics like decay, support for semantic relations (edges) and causality, and initial hooks for personas, all accessible via a well-tested Go client interface. The library is ready for integration into an agent system for further evaluation.