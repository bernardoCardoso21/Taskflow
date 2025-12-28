# TaskFlow

TaskFlow is a **task-tracking backend API** built to demonstrate real-world backend engineering practices in Go.  
It focuses on clean architecture, secure data access, and scalable API design.

This project is intentionally backend-only and designed to be consumed by a web or mobile frontend.

---

## Why this project exists

I built TaskFlow to practice and showcase **production-grade backend development in Go**, beyond simple CRUD demos.

The goal was to design an API that:
- models a realistic business domain
- enforces proper authentication and ownership
- scales with growing data
- follows patterns commonly used in Go production systems

TaskFlow represents the kind of backend service many companies build internally or as the foundation of a SaaS product.

---

## Business problem it solves

TaskFlow provides a backend for **organizing work into projects and tasks**.

It allows users to:
- register and authenticate securely
- create projects
- manage tasks within projects
- track progress over time
- retrieve data efficiently and safely

This type of system is commonly used for:
- personal productivity tools
- internal team dashboards
- startup MVP backends
- project or work tracking services

---

## Core domain model

- **User**
    - Owns projects and tasks
- **Project**
    - Groups related work
    - Belongs to a single user
- **Task**
    - Represents a unit of work
    - Belongs to a project (and therefore a user)

Ownership is enforced at every layer so users can only access their own data.

---

## API features (current)

- JWT-based authentication
- User registration and login
- Project CRUD operations
- Ownership enforcement at the SQL level
- Cursor-based pagination
- Consistent JSON error responses
- Health check endpoint

---

## Technical decisions & architecture

### Language
- **Go**
- Chosen for simplicity, performance, and strong ecosystem for backend services

---

### Architecture
The project follows a **layered architecture**:

