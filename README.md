1) Project spec: TaskFlow REST API (high-signal Go repo)
   Core features (MVP)

Auth: JWT (access token) + password hashing (bcrypt)

Users can manage their own:

Projects

Tasks

CRUD endpoints

Pagination + filtering on tasks

Proper validation + error model

Postgres with migrations

Docker compose for local dev

Tests: unit + integration (DB)

Non-functional “signals” (this is what gets interviews)

Clean architecture-ish separation: handler → service → repo

Context timeouts, request IDs

Structured logging (zap/slog)

Graceful shutdown

OpenAPI/Swagger (later)

CI: go test, golangci-lint

2) API design (endpoints)
   Auth

POST /v1/auth/register

POST /v1/auth/login

GET /v1/auth/me (requires JWT)

Projects

POST /v1/projects

GET /v1/projects (paginated)

GET /v1/projects/{id}

PATCH /v1/projects/{id}

DELETE /v1/projects/{id}

Tasks

POST /v1/projects/{projectId}/tasks

GET /v1/tasks
Query params:

status=todo|doing|done

projectId=...

q=search

limit, cursor (cursor pagination is a good signal)

GET /v1/tasks/{id}

PATCH /v1/tasks/{id}

DELETE /v1/tasks/{id}

Response conventions

Success: { "data": ... , "meta": ... }

Error:

{
"error": {
"code": "VALIDATION_ERROR",
"message": "invalid request",
"details": [{ "field": "email", "message": "must be valid" }]
}
}