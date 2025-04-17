## CogMem: Architecture Design Document (ADD) - Library Edition

**Version:** 2.0
**Date:** October 26, 2023
**Author:** AI Research Analyst (Conceptual)

**1. Overview**

This document details the technical architecture for the **CogMem Go Library**, a cognitively-inspired memory system designed to be integrated into AI agent frameworks or other applications requiring advanced memory capabilities for Large Language Models (LLMs). Instead of a standalone service, CogMem is provided as a Go library with a defined client interface.

CogMem aims to provide persistent, structured, dynamic, and contextually rich memory, addressing limitations in standard LLM context windows and simple RAG systems. Key features include distinct Long-Term Memory (LTM) components (Episodic, Semantic Graph, Procedural), Memory Valence scoring, fine-grained Memory Partitioning, a Lua-based Scripting Engine, valence-modulated Memory Lifespan modeling, a Causal Memory Layer, Valence Aggregation, and mechanisms to potentially export data for external visualization.

This ADD specifies implementation choices, library structure, client interfaces, data schemas, key algorithms, configuration management, and interaction patterns to guide the library's development and integration. The concept of Working Memory (WM) is considered external to this library, typically managed by the consuming agent system (e.g., as the LLM prompt context).

**2. Guiding Principles & Technology Choices**

*   **Library Integration:** Designed as a Go library (`cogmem`) with a clear entry point (`cogmem.NewClient`) and interface for use within other Go applications (agent systems).
*   **Modularity:** Internal components (Persistence, Engines) are loosely coupled via interfaces, promoting testability and maintainability within the library.
*   **Performance:** Golang is chosen for efficiency. Database interactions are optimized using `pgx` and appropriate indexing. Concurrency is leveraged where applicable (e.g., background tasks if managed).
*   **Configuration:** Flexible configuration via YAML files (`config.yaml`) overrideable by `.env` files and environment variables, managed using `viper`.
*   **Reliability:** Leverage mature technologies (Postgres, Go). Implement robust error handling and comprehensive testing (TDD).
*   **Consolidated Persistence:** Utilize PostgreSQL (>= v15 recommended) with `pgvector` and potentially other extensions as the primary persistence layer, managed by the consuming application's infrastructure.
*   **Extensibility:** The scripting engine and modular design allow for future enhancements.
*   **Clear Boundaries:** The library manages LTM operations. The consuming application manages the LLM interaction, WM/prompt construction, and the lifecycle of CogMem background tasks (like decay).

**Technology Choices:**

*   **Implementation Language:** **Golang (Go)**
    *   *Justification:* Ideal for building performant, concurrent, and maintainable libraries. Static typing enhances reliability. Excellent ecosystem.
*   **Database:** **PostgreSQL (>= version 15 recommended)** with **`pgvector`** extension.
    *   *Justification:* Consolidates diverse data needs (relational, vector, JSONB) into a single, robust system. Assumes the consuming application provides or manages the PostgreSQL instance. Simplifies the library's external dependencies. JSONB is used for flexible metadata (valence, causal links). Standard SQL tables used for Episodic, Semantic Nodes/Edges, and Scripts.
*   **Scripting Engine Language:** **Lua**
    *   *Justification:* Lightweight, fast, easily embeddable (`gopher-lua`), simple syntax, good sandboxing potential, good C/Go interop for controlled database access via library functions.
*   **Configuration Management:** **`viper`** library.
    *   *Justification:* Handles YAML, `.env`, and environment variable loading with defined precedence, simplifying configuration for library consumers.

**3. Library Architecture**

CogMem is structured as a Go library. The consuming application interacts with it through a `CogMemClient` interface. Internally, the library follows layered principles (Domain, Application, Ports, Infrastructure).

```mermaid
graph TD
    subgraph ConsumingAgentSystem["Consuming Agent System - Go Application"]
        AgentLogic["Agent Core Logic"]
        ConfigLoader["Configuration Loader (Viper)"]
        ClientInstance["CogMemClient Instance"]

        AgentLogic -->|Calls Methods| ClientInstance
        ConfigLoader -->|Provides Config| ClientInstance
        AgentLogic -->|Manages Lifecycle| ClientInstance
    end

    subgraph CogMemLibrary["CogMem Library - Internal Structure"]
        PublicClient["Public Interface: CogMemClient"]
        ClientImpl["Client Implementation"]

        subgraph ApplicationLayer["Application Layer"]
            MemoryService["Memory Service"]
            ScriptingService["Scripting Service"]
            NodeService["Semantic Node Service"]
            EdgeService["Semantic Edge Service"]
            DecayService["Decay Service"]
        end

        subgraph DomainLayer["Domain Layer"]
            Entities["Memory, Node, Edge, Script, Valence, PartitionCtx..."]
            Interfaces["Repositories, Calculators, Executors..."]
        end

        subgraph InfrastructureLayer["Infrastructure Layer"]
            subgraph Persistence["Persistence (Postgres/pgx)"]
                EpisodicRepo["Episodic Repo Impl"]
                SemanticRepo["Semantic Repo Impl"]
                ScriptRepo["Script Repo Impl"]
            end
            subgraph Engines["Engines"]
                LuaExecutor["Lua Script Executor"]
                ValenceCalc["Valence Calculator Impl"]
            end
            PartitionLogic["Partition Logic Impl"]
            ConfigModule["Internal Config Structs"]
        end

        ClientImpl --> MemoryService
        ClientImpl --> ScriptingService
        ClientImpl --> NodeService
        ClientImpl --> EdgeService
        ClientImpl --> DecayService

        MemoryService --> DomainLayer
        ScriptingService --> DomainLayer
        NodeService --> DomainLayer
        EdgeService --> DomainLayer
        DecayService --> DomainLayer

        DomainLayer --> Interfaces
        Interfaces -->|Implemented By| InfrastructureLayer

        Persistence --> DBPG["External PostgreSQL Instance"]
    end

    PublicClient -->|Implemented By| ClientImpl

    style DBPG fill:#lightgrey,stroke:#333,stroke-width:2px
    style CogMemLibrary fill:#ccf,stroke:#333,stroke-width:2px
    style ClientInstance fill:#f9f,stroke:#333,stroke-width:2px
```

**Key Points:**

*   The "API Layer" is replaced by the `CogMemClient` interface and its implementation within the library.
*   The consuming application is responsible for providing the database connection details and managing the PostgreSQL instance itself.
*   Background tasks (like Decay) are initiated and managed by the consuming application via methods on the `CogMemClient`.

**4. `CogMemClient` Interface (Primary Entrypoint)**

```go
package cogmem

import (
	"context"
	"time"

	"github.com/google/uuid"
	// Potentially import domain types if needed publicly, or use DTOs
	// "github.com/yourorg/cogmem/internal/domain/memory"
	// "github.com/yourorg/cogmem/internal/domain/valence"
)

// Config struct loaded via Viper
type Config struct {
	DatabaseURL      string        `mapstructure:"database_url"`
	EmbeddingDim     int           `mapstructure:"embedding_dim"`
	DecayBaseRate    float64       `mapstructure:"decay_base_rate"`
	DecayValenceWeight float64       `mapstructure:"decay_valence_weight"`
	DecayInterval    time.Duration `mapstructure:"decay_interval"`
	// Add other config fields (e.g., Lua sandbox settings)
}

// --- DTOs for Public Interface ---
// (Define simple structs for input/output to avoid exposing internal domain complex types directly)

type PartitionContext struct {
	UserID   uuid.UUID
	EntityID *uuid.UUID // Optional
	// Scope/access level info if needed beyond user/entity
}

type ValenceScore struct {
	Polarity      float64
	GoalRelevance float64
	Arousal       float64
}

type EpisodicMemoryInput struct {
	Content        string
	Source         string // e.g., "user", "agent"
	Timestamp      *time.Time // Optional, defaults to now
	SessionID      *uuid.UUID // Optional
	Embedding      []float32  // Optional: If provided by caller
	ExplicitValence *ValenceScore // Optional: Override calculated valence
	CausalLinks    map[string][]uuid.UUID // Optional: e.g., {"caused_by": [id1], "leads_to": [id2]}
	Metadata       map[string]interface{} // Optional
	ShareScope     string // Optional: Default 'user'
}

type EpisodicMemoryOutput struct {
	ID                uuid.UUID
	Timestamp         time.Time
	Content           string
	Source            string
	Valence           ValenceScore
	CausalLinks       map[string][]uuid.UUID
	Metadata          map[string]interface{}
	AccessibilityScore float64
	LastAccessed      time.Time
	// Embedding excluded by default for brevity
}

type SemanticNodeInput struct {
	NodeType   string
	Name       string
	Properties map[string]interface{}
	Summary    *string // Optional
	ShareScope string  // Optional: Default 'user'
}

type SemanticNodeOutput struct {
	ID                  uuid.UUID
	NodeType            string
	Name                string
	Properties          map[string]interface{}
	AggregatedValence   ValenceScore // Placeholder - aggregation logic needed
	Summary             *string
	AccessibilityScore  float64
	LastAccessed        time.Time
}

type SemanticEdgeInput struct {
	SourceNodeID  uuid.UUID
	TargetNodeID  uuid.UUID
	RelationType  string
	Properties    map[string]interface{}
	Timestamp     *time.Time // Optional
	ShareScope    string     // Optional: Default 'user'
}

type SemanticEdgeOutput struct {
	ID            uuid.UUID
	SourceNodeID  uuid.UUID
	TargetNodeID  uuid.UUID
	RelationType  string
	Properties    map[string]interface{}
	Timestamp     time.Time
}


type RetrievalQuery struct {
	QueryText        *string     // For similarity search
	Vector           []float32   // Alternative for similarity search
	Limit            int
	MinTimestamp     *time.Time  // Filter
	MaxTimestamp     *time.Time  // Filter
	MinValencePolarity *float64    // Filter
	SourceFilter     *string     // Filter
	RequiredMetadata map[string]interface{} // Filter
	SortBy           string      // e.g., "timestamp_desc", "relevance", "valence_abs_desc"
}

type ScriptExecutionInput struct {
	ScriptName string
	Parameters map[string]interface{} // Parameters passed to Lua script
}

type ScriptExecutionOutput struct {
	Success bool
	Result  interface{} // Lua script return value(s)
	Logs    string      // Captured output/logs from script
	Error   string      // Error message if execution failed
}


// CogMemClient defines the public interface for interacting with the CogMem library.
type CogMemClient interface {
	// --- Episodic Memory ---
	StoreEpisodic(ctx context.Context, pCtx PartitionContext, input EpisodicMemoryInput) (uuid.UUID, error)
	RetrieveEpisodic(ctx context.Context, pCtx PartitionContext, query RetrievalQuery) ([]EpisodicMemoryOutput, error)
	GetEpisodicByID(ctx context.Context, pCtx PartitionContext, id uuid.UUID) (*EpisodicMemoryOutput, error)
	UpdateEpisodic(ctx context.Context, pCtx PartitionContext, id uuid.UUID, updates map[string]interface{}) error // e.g., update metadata, causal links
	DeleteEpisodic(ctx context.Context, pCtx PartitionContext, id uuid.UUID) error

	// --- Semantic Memory ---
	StoreSemanticNode(ctx context.Context, pCtx PartitionContext, input SemanticNodeInput) (uuid.UUID, error)
	GetSemanticNodeByID(ctx context.Context, pCtx PartitionContext, id uuid.UUID) (*SemanticNodeOutput, error)
	FindSemanticNodes(ctx context.Context, pCtx PartitionContext, nodeType *string, name *string, properties map[string]interface{}) ([]SemanticNodeOutput, error)
	StoreSemanticEdge(ctx context.Context, pCtx PartitionContext, input SemanticEdgeInput) (uuid.UUID, error)
	FindEdges(ctx context.Context, pCtx PartitionContext, sourceNodeID *uuid.UUID, targetNodeID *uuid.UUID, relationType *string) ([]SemanticEdgeOutput, error)
	// ... other semantic graph operations (Update, Delete)

	// --- Procedural Memory (Scripts) ---
	StoreScript(ctx context.Context, pCtx PartitionContext, name, description, language, code string, triggerEvents []string, allowedPersonas []string, shareScope string, metadata map[string]interface{}) (uuid.UUID, error)
	ExecuteScript(ctx context.Context, pCtx PartitionContext, input ScriptExecutionInput) (ScriptExecutionOutput, error)
	// ... GetScriptByName, DeleteScript ...

	// --- Background Task Management ---
	// StartBackgroundTasks starts long-running goroutines (e.g., for decay).
	// Returns a function to gracefully stop the tasks.
	StartBackgroundTasks() (stop func(ctx context.Context) error, err error)

	// --- Direct Valence Calculation (Optional Utility) ---
	CalculateValence(ctx context.Context, text string, taskHint string) (ValenceScore, error)

	// --- Cleanup ---
	// Close releases any resources held by the client (e.g., database connection pool).
	Close() error
}

// NewClient is the constructor for the CogMemClient.
// It loads configuration using Viper (config.yaml, .env, env vars) and initializes dependencies.
func NewClient(configPath string) (CogMemClient, error) {
	// 1. Load config using Viper
	// 2. Initialize DB connection pool (pgxpool)
	// 3. Initialize internal services/repositories (ValenceCalc, Repos, ScriptExecutor...)
	// 4. Return the client implementation
	panic("Not implemented") // Placeholder
}

```


**5. Configuration (`config.yaml`, `.env`, Env Vars)**

*   **Loading:** Use `viper` within `NewClient`.
*   **Precedence:** Env Vars > `.env` file > `config.yaml` file > Defaults.
*   **Example `config.yaml`:**

```yaml
database_url: "postgres://user:password@localhost:5432/cogmemdb?sslmode=disable"
embedding_dim: 1536
decay_base_rate: 1.0e-7
decay_valence_weight: 0.8
decay_interval: "1h" # Parse using time.ParseDuration
lua:
  sandbox:
    disallowed_modules: ["os", "io", "debug", "dofile", "loadfile"] # Example
  execution_timeout: "5s"
# Add other config sections as needed
```

*   **Example `.env`:**
    ```dotenv
    COGMENT_DATABASE_URL="postgres://prod_user:prod_secret@prod_host:5432/cogmemdb_prod"
    ```
    (Viper needs `viper.AutomaticEnv()` and potentially `viper.SetEnvPrefix("COG MEM")` / `viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))`)

**6. Data Models & Schemas (PostgreSQL)**
Includes tables `episodic_memory`, `semantic_nodes`, `semantic_edges`, `procedural_scripts` with appropriate columns, types (including `VECTOR`, `JSONB`), and indexes. The consuming application is responsible for running migrations to create these tables in the provided database. Migrations should be bundled with the library (`migrations/`).

```sql
-- Required Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

-- Episodic Memory Table
CREATE TABLE episodic_memory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,          -- Partitioning key
    entity_id UUID,                 -- Partitioning key (optional, for orgs)
    session_id UUID,                -- Optional: Link related events
    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source TEXT,                    -- e.g., 'user_input', 'agent_output', 'system_event'
    content TEXT NOT NULL,          -- The raw text content of the memory
    content_embedding VECTOR(1536), -- Store embeddings (e.g., OpenAI Ada-002 dim)
    valence JSONB,                  -- Store valence score: {'polarity': float, 'goal_relevance': float, 'arousal': float}
    causal_links JSONB,             -- Store links: {'caused_by': [uuid], 'leads_to': [uuid]}
    metadata JSONB,                 -- Other arbitrary metadata
    share_scope TEXT NOT NULL DEFAULT 'user', -- 'user', 'entity', 'shared:<uuid>', 'public'
    accessibility_score FLOAT NOT NULL DEFAULT 1.0,
    last_accessed TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for Episodic Memory
CREATE INDEX idx_episodic_user_entity ON episodic_memory (user_id, entity_id);
CREATE INDEX idx_episodic_timestamp ON episodic_memory (timestamp);
CREATE INDEX idx_episodic_valence_polarity ON episodic_memory USING GIN ((valence -> 'polarity')); -- Example index on JSONB
CREATE INDEX idx_episodic_embedding ON episodic_memory USING HNSW (content_embedding vector_cosine_ops); -- Or IVFFlat


-- Semantic Nodes Table
CREATE TABLE semantic_nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,          -- Partitioning key
    entity_id UUID,                 -- Partitioning key
    node_type TEXT NOT NULL,        -- e.g., 'person', 'project', 'concept', 'client'
    name TEXT NOT NULL,             -- Primary label/identifier for the node
    properties JSONB,               -- Store node attributes
    aggregated_valence JSONB,       -- Store aggregated valence over time
    summary TEXT,                   -- Optional textual summary
    share_scope TEXT NOT NULL DEFAULT 'user',
    accessibility_score FLOAT NOT NULL DEFAULT 1.0,
    last_accessed TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for Semantic Nodes
CREATE INDEX idx_semantic_nodes_user_entity ON semantic_nodes (user_id, entity_id);
CREATE INDEX idx_semantic_nodes_type_name ON semantic_nodes (node_type, name);
CREATE INDEX idx_semantic_nodes_properties ON semantic_nodes USING GIN (properties);


-- Semantic Edges Table (Relationships)
CREATE TABLE semantic_edges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,          -- Partitioning key (inherited from nodes usually)
    entity_id UUID,                 -- Partitioning key
    source_node_id UUID NOT NULL REFERENCES semantic_nodes(id) ON DELETE CASCADE,
    target_node_id UUID NOT NULL REFERENCES semantic_nodes(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL,    -- e.g., 'works_on', 'related_to', 'caused_by', 'client_of'
    properties JSONB,               -- Store edge attributes (e.g., weight, timestamp)
    timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    share_scope TEXT NOT NULL DEFAULT 'user'
);

-- Indexes for Semantic Edges
CREATE INDEX idx_semantic_edges_user_entity ON semantic_edges (user_id, entity_id);
CREATE INDEX idx_semantic_edges_source_type ON semantic_edges (source_node_id, relation_type);
CREATE INDEX idx_semantic_edges_target_type ON semantic_edges (target_node_id, relation_type);


-- Procedural Scripts Table
CREATE TABLE procedural_scripts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,          -- Partitioning key (who owns/can manage script)
    entity_id UUID,                 -- Partitioning key
    name TEXT UNIQUE NOT NULL,      -- Unique identifier for the script
    description TEXT,
    script_language TEXT NOT NULL DEFAULT 'lua', -- 'lua' or 'python' (though ADD focuses on Lua)
    script_code TEXT NOT NULL,      -- The actual Lua code
    trigger_events TEXT[],          -- Events that can trigger this script (e.g., 'on_memory_retrieval', 'manual')
    allowed_personas TEXT[],        -- Personas allowed to use this script
    share_scope TEXT NOT NULL DEFAULT 'user',
    metadata JSONB                  -- e.g., version, author
);

-- Indexes for Procedural Scripts
CREATE INDEX idx_procedural_scripts_user_entity ON procedural_scripts (user_id, entity_id);
CREATE INDEX idx_procedural_scripts_trigger ON procedural_scripts USING GIN (trigger_events);
```

**6. Key Algorithms & Logic**

*   **6.1. Valence Engine v1 Algorithm (Go Module)**
    1.  **Input:** `text` (string), `context` (struct containing `task_type`, optional `user_goals`).
    2.  **Initialize Scores:** `polarity=0.0`, `goal_relevance=0.0`, `arousal=0.0`.
    3.  **Polarity Calculation:**
        *   Use predefined lists of positive/negative keywords (weighted). Sum weights of matched keywords.
        *   *Optionally:* Feed `text` into a simple Go sentiment analysis library (e.g., vader-go) to get a compound score.
        *   Combine keyword score and library score (e.g., average or prioritize one).
        *   Normalize final polarity score to range [-1.0, 1.0].
    4.  **Goal Relevance Calculation:**
        *   Check `context.task_type` or keywords in `text` against a predefined map: `{"client_complaint": 0.9, "contract_renewal": 0.8, "project_blocker": 0.9, "positive_feedback": 0.7, ...}`.
        *   If `context.user_goals` provided, check for keyword matches.
        *   Assign highest matching score. Default to 0.0 if no match.
    5.  **Arousal Calculation:**
        *   Base arousal on absolute polarity: `arousal = abs(polarity) * 0.5`.
        *   Check for intensity keywords ("urgent", "critical", "immediately", "amazing", "terrible"). Add bonus for each match (e.g., +0.2 per keyword).
        *   Check for high goal relevance: `if goal_relevance > 0.7: arousal += 0.2`.
        *   Clamp final arousal score to range [0.0, 1.0].
    6.  **Output:** Return `ValenceScore{Polarity: polarity, GoalRelevance: goal_relevance, Arousal: arousal}` (struct easily marshallable to JSONB).

*   **6.2. Scripting Engine Sandboxing (gopher-lua)**
    1.  **Initialize Lua State:** Create a new `lua.LState`.
    2.  **Disable Standard Libraries:** Use `LState.DoString` or module loading mechanisms to explicitly *remove* or *nil out* potentially dangerous standard Lua modules: `os`, `io` (partially, allow `print` maybe), `dofile`, `loadfile`, `debug`.
    3.  **Inject Safe Go Functions:** Define Go functions that perform specific, controlled actions (like accessing the DB via the LTM Access Layer, which enforces partitioning). Expose these functions to the Lua environment using `LState.SetGlobal` or by creating custom modules loadable via `require`.
        *   Example Go func: `luaSafeFindMemories(L *lua.LState) int`. Takes Lua table args, performs DB query *with partitioning*, pushes results back to Lua stack.
        *   Example Lua call: `local results = CogMemDB.findMemories({type='episodic', limit=10, query='client issue'})`
    4.  **Set Resource Limits:** Use context deadlines or custom mechanisms within `gopher-lua` (if available, or via interrupting goroutines) to limit script execution time. Monitor memory usage if possible (though harder with standard Lua).
    5.  **Execute Script:** Run the user script code within this sandboxed `LState`.

*   **6.3. Memory Decay Function Logic**
    1.  **Trigger:** Run periodically (e.g., every hour) or lazily when a memory item is accessed.
    2.  **Selection:** Identify memory items not recently accessed (e.g., `last_accessed < NOW() - interval`).
    3.  **Calculation:** For each selected item:
        *   `currentTime = NOW()`
        *   `lastAccessTime = item.last_accessed`
        *   `timeDelta = currentTime - lastAccessTime` (in seconds or appropriate units)
        *   `currentAccessibility = item.accessibility_score`
        *   `valencePolarity = item.valence -> 'polarity'` (extract from JSONB)
        *   `k = Config.DecayBaseRate` (e.g., 1e-7)
        *   `W = Config.DecayValenceWeight` (e.g., 0.8)
        *   `newAccessibility = currentAccessibility * exp(-k * (1.0 - W * abs(valencePolarity)) * timeDelta)`
    4.  **Update:** Update the `accessibility_score` in the database.
    5.  **Pruning (Optional):** Items below a certain `accessibility_score` threshold could be flagged for archiving or deletion in a separate process.


**7. Key Algorithms & Logic**
*   **Valence Engine v1 Algorithm:** Implemented as an internal Go component.
*   **Scripting Engine Sandboxing:** Implemented using `gopher-lua` with careful standard library restriction and injection of safe Go functions for DB access (which *must* enforce partitioning).
*   **Memory Decay Function Logic:** Implemented internally, but its execution loop is managed via `StartBackgroundTasks`/`stop` initiated by the consumer.


*   **7.1. Valence Engine v1 Algorithm (Go Module)**
    1.  **Input:** `text` (string), `context` (struct containing `task_type`, optional `user_goals`).
    2.  **Initialize Scores:** `polarity=0.0`, `goal_relevance=0.0`, `arousal=0.0`.
    3.  **Polarity Calculation:**
        *   Use predefined lists of positive/negative keywords (weighted). Sum weights of matched keywords.
        *   *Optionally:* Feed `text` into a simple Go sentiment analysis library (e.g., vader-go) to get a compound score.
        *   Combine keyword score and library score (e.g., average or prioritize one).
        *   Normalize final polarity score to range [-1.0, 1.0].
    4.  **Goal Relevance Calculation:**
        *   Check `context.task_type` or keywords in `text` against a predefined map: `{"client_complaint": 0.9, "contract_renewal": 0.8, "project_blocker": 0.9, "positive_feedback": 0.7, ...}`.
        *   If `context.user_goals` provided, check for keyword matches.
        *   Assign highest matching score. Default to 0.0 if no match.
    5.  **Arousal Calculation:**
        *   Base arousal on absolute polarity: `arousal = abs(polarity) * 0.5`.
        *   Check for intensity keywords ("urgent", "critical", "immediately", "amazing", "terrible"). Add bonus for each match (e.g., +0.2 per keyword).
        *   Check for high goal relevance: `if goal_relevance > 0.7: arousal += 0.2`.
        *   Clamp final arousal score to range [0.0, 1.0].
    6.  **Output:** Return `ValenceScore{Polarity: polarity, GoalRelevance: goal_relevance, Arousal: arousal}` (struct easily marshallable to JSONB).

*   **7.2. Scripting Engine Sandboxing (gopher-lua)**
    1.  **Initialize Lua State:** Create a new `lua.LState`.
    2.  **Disable Standard Libraries:** Use `LState.DoString` or module loading mechanisms to explicitly *remove* or *nil out* potentially dangerous standard Lua modules: `os`, `io` (partially, allow `print` maybe), `dofile`, `loadfile`, `debug`.
    3.  **Inject Safe Go Functions:** Define Go functions that perform specific, controlled actions (like accessing the DB via the LTM Access Layer, which enforces partitioning). Expose these functions to the Lua environment using `LState.SetGlobal` or by creating custom modules loadable via `require`.
        *   Example Go func: `luaSafeFindMemories(L *lua.LState) int`. Takes Lua table args, performs DB query *with partitioning*, pushes results back to Lua stack.
        *   Example Lua call: `local results = CogMemDB.findMemories({type='episodic', limit=10, query='client issue'})`
    4.  **Set Resource Limits:** Use context deadlines or custom mechanisms within `gopher-lua` (if available, or via interrupting goroutines) to limit script execution time. Monitor memory usage if possible (though harder with standard Lua).
    5.  **Execute Script:** Run the user script code within this sandboxed `LState`.

*   **7.3. Memory Decay Function Logic**
    1.  **Trigger:** Run periodically (e.g., every hour) or lazily when a memory item is accessed.
    2.  **Selection:** Identify memory items not recently accessed (e.g., `last_accessed < NOW() - interval`).
    3.  **Calculation:** For each selected item:
        *   `currentTime = NOW()`
        *   `lastAccessTime = item.last_accessed`
        *   `timeDelta = currentTime - lastAccessTime` (in seconds or appropriate units)
        *   `currentAccessibility = item.accessibility_score`
        *   `valencePolarity = item.valence -> 'polarity'` (extract from JSONB)
        *   `k = Config.DecayBaseRate` (e.g., 1e-7)
        *   `W = Config.DecayValenceWeight` (e.g., 0.8)
        *   `newAccessibility = currentAccessibility * exp(-k * (1.0 - W * abs(valencePolarity)) * timeDelta)`
    4.  **Update:** Update the `accessibility_score` in the database.
    5.  **Pruning (Optional):** Items below a certain `accessibility_score` threshold could be flagged for archiving or deletion in a separate process.



**8. Interaction Flows (Sequence Diagrams - Library Perspective)**

*   **8.1. Storing New Memory Item**

```mermaid
sequenceDiagram
    participant Agent as "Consuming Agent System"
    participant Client as "CogMemClient Instance"
    participant AppLayer as "Application Layer (Internal)"
    participant Infra as "Infrastructure Layer (Internal)"
    participant DBPG as "PostgreSQL"

    Agent->>Client: StoreEpisodic(ctx, pCtx, input)
    Client->>AppLayer: MemoryService.StoreEpisodic(...)
    AppLayer->>Infra: ValenceCalculator.Calculate(...)
    Infra-->>AppLayer: Return valenceScore
    AppLayer->>Infra: EpisodicRepo.Save(..., memoryWithValence)
    Infra->>DBPG: INSERT INTO episodic_memory...
    DBPG-->>Infra: Insert OK
    Infra-->>AppLayer: Success (memoryID)
    AppLayer-->>Client: Success (memoryID)
    Client-->>Agent: Return memoryID, nil
```

*   **8.2. Retrieving Memory**

```mermaid
sequenceDiagram
    participant Agent as "Consuming Agent System"
    participant Client as "CogMemClient Instance"
    participant AppLayer as "Application Layer (Internal)"
    participant Infra as "Infrastructure Layer (Internal)"
    participant DBPG as "PostgreSQL"

    Agent->>Client: RetrieveEpisodic(ctx, pCtx, query)
    Client->>AppLayer: MemoryService.RetrieveEpisodic(...)
    Note over AppLayer: Logic decides retrieval strategy (vector, keyword, filters, sorting based on query.SortBy)
    AppLayer->>Infra: EpisodicRepo.FindByVector / FindRecent / ... (queryCriteria, pCtx)
    Infra->>DBPG: SELECT ... FROM episodic_memory WHERE ... [Partition Filters] ... ORDER BY ...
    DBPG-->>Infra: Return memory rows
    Infra-->>AppLayer: Return internal memory structs
    AppLayer->>AppLayer: Map internal structs to EpisodicMemoryOutput DTOs
    AppLayer-->>Client: Return []EpisodicMemoryOutput
    Client-->>Agent: Return []EpisodicMemoryOutput, nil
```

*   **8.3. Starting Background Decay Task**

```mermaid
sequenceDiagram
    participant Agent as "Consuming Agent System"
    participant Client as "CogMemClient Instance"
    participant AppLayer as "Application Layer (Internal)"
    participant Infra as "Infrastructure Layer (Internal)"
    participant DBPG as "PostgreSQL"

    Agent->>Client: StartBackgroundTasks()
    Client->>AppLayer: DecayService.StartDecayLoop(config.DecayInterval)
    Note over AppLayer: Spawns a goroutine that periodically calls decay logic.
    AppLayer-->>Client: Return stop function, nil
    Client-->>Agent: Return stop function, nil

    loop "Decay Loop (Goroutine)"
        AppLayer->>AppLayer: Wait(DecayInterval)
        AppLayer->>Infra: DecayLogic.ApplyDecay(pCtx)\n// Uses partition context if applicable
        Infra->>DBPG: UPDATE episodic_memory SET accessibility_score = ... WHERE ...
    end

    Note right of Agent: Later...
    Agent->>Client: stop(ctx)\n// Call the returned stop function
    Client->>AppLayer: Signal Decay Loop Goroutine to stop
    AppLayer->>AppLayer: Goroutine exits cleanly
```

**9. Deployment & Integration**

*   **Library Usage:** The consuming Go application imports the `cogmem` library.
*   **Dependency Management:** The consuming application adds `cogmem` to its `go.mod`.
*   **Database Management:** The consuming application is responsible for provisioning, managing, and providing connection details for the PostgreSQL database instance (including installing `pgvector`). It must also run the database migrations provided by the `cogmem` library.
*   **Configuration:** The consuming application provides the path to the `config.yaml` file (or relies solely on env vars/defaults) when calling `cogmem.NewClient`. It manages `.env` files and environment variables.
*   **Lifecycle:** The consuming application controls the lifecycle of the `CogMemClient` (calling `NewClient`, `StartBackgroundTasks`, and `Close`).

**10. Security Considerations**

*   **Database Credentials:** The consuming application is responsible for securely managing and providing the `database_url`. The library should not log credentials.
*   **Scripting Engine Sandboxing:** Remains critical. The library *must* implement robust sandboxing for Lua to prevent scripts from accessing unauthorized resources within the consuming application's environment or performing unsafe database operations. Expose only necessary, partition-aware functions to Lua.
*   **Data Privacy:** Partitioning logic within the library's persistence layer is essential and must be rigorously tested. The consuming application defines the `PartitionContext`.
*   **Input Sanitization:** While the library might perform basic checks, the consuming application should ideally sanitize inputs before passing them to the library, especially `Content` fields if they might be rendered elsewhere.

**11. Future Considerations & Scalability**

*   **Database Scaling:** Scaling strategies (read replicas, partitioning) are implemented at the infrastructure level managed by the consuming application, but the library's queries should be designed to be compatible with such strategies where possible.
*   **Asynchronous Operations:** For long-running operations (complex scripts, bulk updates), the library could offer asynchronous method variants (e.g., returning channels or requiring callbacks).
*   **Alternative Backends:** The layered architecture allows, in theory, for swapping the Postgres persistence layer with other implementations if needed, by creating new implementations of the repository interfaces.
*   **Visualization:** Provide clear documentation and potentially helper functions or data export methods to enable consuming applications to build their own visualization tools by querying the library or its database directly (respecting partitioning).
