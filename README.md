Week 1

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

Observability: request logging, metrics endpoint, pprof

Caching

Rate limiting middleware

Polish README + diagrams
