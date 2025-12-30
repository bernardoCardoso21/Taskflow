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

## Core domain model

- **User**
    - owns projects
- **Project**
    - belongs to a user
    - groups related tasks
- **Task**
    - belongs to a project
    - tracks completion state
    - supports efficient listing and filtering

---

## API features

- JWT-based authentication
- Ownership enforced at the SQL level
- Cursor-based pagination (stable and scalable)
- Filtering support on task lists
- Consistent JSON error responses
- OpenAPI (Swagger) documentation with live UI

---

## Architecture overview

- **HTTP layer**: routing, middleware, request/response handling
- **Service layer**: business logic, validation, error mapping
- **Repository layer**: raw SQL with explicit ownership checks
- **Database**: PostgreSQL with migrations

---

## Testing strategy

- **Unit tests**
  - Service-layer logic
  - Validation and error mapping
- **Integration tests**
  - Repository behavior against real Postgres
  - Ownership enforcement
  - Pagination and filtering correctness

Integration tests run against a real database using Docker and isolated environment variables.

---

## OpenAPI / Swagger

The API is fully documented using a handwritten OpenAPI specification.

Swagger UI was made available via Docker:
http://localhost:8081/taskflow/swaggerui 
(project doesn't have a public server as of now)


The UI allows exploring endpoints and making authenticated requests using JWT tokens.

---

## Running the project

### Start all services
docker compose up -d

This starts:

- PostgreSQL
- database migrations
- API server
- Swagger UI

### Rebuild API after code changes

docker compose up -d --build api

###  Stop services
docker compose down

### Stop services and wipe database
docker compose down -v

---

## Configuration

Environment variables are managed via .env and .env.test files.

- .env — local development

- .env.test — integration tests

These files are intentionally excluded from version control.

---

## Tech stack

- Go
- PostgreSQL
- chi router
- JWT authentication
- Docker & Docker Compose
- Swagger / OpenAPI

---

## Future Implementations

Observability:

- Structured request logging (request id, latency)

Security & robustness

- Refresh tokens
- Password reset flow
- More Unit and Integration Tests

Developer experience

- GitHub Actions CI (tests + lint)
- Versioned API docs (/v1, /v2)

Architecture evolution

- Background job processing
- Event-driven task updates
