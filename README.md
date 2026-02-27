# TaskFlow

TaskFlow is a **task-tracking backend API** built in Go to demonstrate **production-grade backend engineering practices**.

It is a backend-only service designed to be consumed by a web or mobile frontend and focuses on clean architecture, security, and scalability.

---

## Why this project exists

TaskFlow was built to go beyond simple CRUD demos and showcase how a real backend service is structured in Go.

The project demonstrates:
- layered architecture (HTTP → service → repository)
- explicit domain modeling
- ownership enforcement at the data layer
- secure authentication using JWT
- pagination strategies suitable for large datasets
- realistic testing strategies (unit + integration)

This mirrors the type of backend services commonly built inside companies or used as the foundation of SaaS products.

---

## Business problem it solves

TaskFlow provides a backend for **organizing work into projects and tasks**.

Users can:
- register and authenticate securely
- create and manage projects
- create, update, and complete tasks within projects
- list tasks efficiently with filtering and pagination
- access only their own data (strict ownership enforcement)

This type of system is commonly used for:
- personal productivity tools
- internal team dashboards
- startup MVP backends
- project and work tracking platforms

---

## System Architecture

```mermaid
graph TD
    Client["Client\n(Web / Mobile)"]
    SwaggerUI["Swagger UI\n:8081"]

    subgraph Docker["Docker Compose"]
        API["API Server\n:8080\n(Go + chi)"]

        subgraph Layers["Application Layers"]
            HTTP["HTTP Layer\nrouting · middleware · validation"]
            SVC["Service Layer\nbusiness logic · error mapping"]
            REPO["Repository Layer\nraw SQL · ownership checks"]
        end

        DB[("PostgreSQL 16\n:5432")]
        Migrate["Migration Runner\n(startup job)"]
    end

    Client -->|"REST / JSON"| API
    SwaggerUI -->|"REST / JSON"| API
    API --> HTTP --> SVC --> REPO --> DB
    Migrate -->|"runs migrations"| DB
```

---

## Domain Model

```mermaid
erDiagram
    USER {
        uuid id PK
        text email UK
        text password_hash
        timestamptz created_at
        timestamptz updated_at
    }

    PROJECT {
        uuid id PK
        uuid user_id FK
        text name
        timestamptz created_at
        timestamptz updated_at
    }

    TASK {
        uuid id PK
        uuid project_id FK
        text title
        boolean completed
        timestamptz created_at
        timestamptz updated_at
    }

    USER ||--o{ PROJECT : "owns"
    PROJECT ||--o{ TASK : "contains"
```

---

## Authentication Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant API as API Server
    participant SVC as Auth Service
    participant DB as PostgreSQL

    rect rgb(230, 240, 255)
        Note over C,DB: Registration
        C->>API: POST /v1/auth/register {email, password}
        API->>SVC: RegisterUser(email, password)
        SVC->>SVC: bcrypt hash password
        SVC->>DB: INSERT INTO users
        DB-->>SVC: user id
        SVC-->>API: UserID
        API-->>C: 201 { data: { id } }
    end

    rect rgb(230, 255, 230)
        Note over C,DB: Login
        C->>API: POST /v1/auth/login {email, password}
        API->>SVC: Login(email, password)
        SVC->>DB: SELECT user by email
        DB-->>SVC: user row
        SVC->>SVC: bcrypt compare password
        SVC->>SVC: sign JWT (24h, HS256)
        SVC-->>API: accessToken
        API-->>C: 200 { data: { accessToken } }
    end

    rect rgb(255, 245, 220)
        Note over C,DB: Protected Request
        C->>API: GET /v1/projects\nAuthorization: Bearer token
        API->>API: JWT middleware validates token
        API->>SVC: ListProjects(userID, cursor)
        SVC->>DB: SELECT WHERE user_id = $1
        DB-->>SVC: projects[]
        SVC-->>API: Page[Project]
        API-->>C: 200 { data: [...], meta: { nextCursor } }
    end
```

---

## Request Lifecycle

```mermaid
flowchart LR
    Req["Incoming\nRequest"]
    RID["Request ID\nMiddleware"]
    CORS["CORS\nMiddleware"]
    Auth["JWT Auth\nMiddleware"]
    Handler["HTTP\nHandler"]
    Service["Service\nLayer"]
    Repo["Repository\nLayer"]
    PG[("PostgreSQL")]
    Resp["JSON\nResponse"]

    Req --> RID --> CORS --> Auth
    Auth -->|"public route"| Handler
    Auth -->|"protected route\nvalidates Bearer token"| Handler
    Handler -->|"validates input"| Service
    Service -->|"business logic"| Repo
    Repo -->|"SQL + ownership check"| PG
    PG --> Repo --> Service --> Handler --> Resp
```

---

## Cursor-Based Pagination

```mermaid
sequenceDiagram
    participant C as Client
    participant API as API
    participant DB as PostgreSQL

    C->>API: GET /v1/tasks?limit=20
    API->>DB: SELECT ... ORDER BY created_at DESC, id DESC LIMIT 21
    Note over DB: fetches limit+1 to detect next page
    DB-->>API: 21 rows
    API-->>C: { data: [20 tasks], meta: { nextCursor: {createdAt, id} } }

    C->>API: GET /v1/tasks?limit=20&cursorCreatedAt=...&cursorId=...
    API->>DB: SELECT ... WHERE (created_at, id) < ($cursor) ORDER BY ... LIMIT 21
    DB-->>API: rows
    API-->>C: { data: [...], meta: { nextCursor: null } }
    Note over C: null nextCursor means last page
```

---

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/healthz` | - | Health check |
| `POST` | `/v1/auth/register` | - | Register user |
| `POST` | `/v1/auth/login` | - | Login, returns JWT |
| `GET` | `/v1/auth/me` | JWT | Get current user |
| `POST` | `/v1/projects` | JWT | Create project |
| `GET` | `/v1/projects` | JWT | List projects (paginated) |
| `GET` | `/v1/projects/{id}` | JWT | Get project |
| `PATCH` | `/v1/projects/{id}` | JWT | Update project name |
| `DELETE` | `/v1/projects/{id}` | JWT | Delete project |
| `POST` | `/v1/projects/{id}/tasks` | JWT | Create task |
| `GET` | `/v1/tasks` | JWT | List tasks (filtered, paginated) |
| `GET` | `/v1/tasks/{id}` | JWT | Get task |
| `PATCH` | `/v1/tasks/{id}` | JWT | Update task |
| `DELETE` | `/v1/tasks/{id}` | JWT | Delete task |

---

## Testing strategy

```mermaid
graph TD
    subgraph Unit["Unit Tests"]
        AuthSvc["auth_service_test.go\npassword hashing · JWT signing"]
        ProjSvc["project_service_test.go\nbusiness logic · error mapping"]
        TaskSvc["task_service_test.go\nbusiness logic · validation"]
    end

    subgraph Integration["Integration Tests (real PostgreSQL)"]
        ProjRepo["project_repo_integration_test.go\nownership · CRUD · pagination"]
        TaskRepo["task_repo_integration_test.go\nownership · filtering · cursor pagination"]
    end

    AuthSvc & ProjSvc & TaskSvc -.->|"mock repositories"| SvcLayer["Service Layer"]
    ProjRepo & TaskRepo -->|"real DB via docker"| PG[("PostgreSQL")]
```

---

## OpenAPI / Swagger

The API is fully documented using a handwritten OpenAPI specification.

Swagger UI is available via Docker:
```
http://localhost:8081/taskflow/swaggerui
```

The UI allows exploring endpoints and making authenticated requests using JWT tokens.

---

## Running the project

### Start all services

```bash
docker compose up -d
```

This starts:

- PostgreSQL
- database migrations
- API server (`localhost:8080`)
- Swagger UI (`localhost:8081`)

### Rebuild API after code changes

```bash
docker compose up -d --build api
```

### Stop services

```bash
docker compose down
```

### Stop services and wipe database

```bash
docker compose down -v
```

---

## Configuration

Environment variables are managed via `.env` and `.env.test` files.

| Variable | Description | Example |
|----------|-------------|---------|
| `HTTP_ADDR` | Address the API listens on | `:8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable` |
| `JWT_SECRET` | Secret for signing JWT tokens | `dev-secret-change-me` |

- `.env` — local development
- `.env.test` — integration tests

These files are intentionally excluded from version control. Copy `.env.example` to get started.

---

## Tech stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25 |
| HTTP Router | chi v5 |
| Database | PostgreSQL 16 |
| DB Driver | pgx/v5 |
| Auth | JWT (golang-jwt/v5) + bcrypt |
| Container | Docker & Docker Compose |
| API Docs | OpenAPI 3.0 + Swagger UI |

---

## Future Implementations

**Observability**
- Structured request logging (request id, latency)

**Security & robustness**
- Refresh tokens
- Password reset flow
- More Unit and Integration Tests

**Developer experience**
- GitHub Actions CI (tests + lint)
- Versioned API docs (/v1, /v2)

**Architecture evolution**
- Background job processing
- Event-driven task updates
