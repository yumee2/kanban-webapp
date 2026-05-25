# Kanban Server

Backend API for a kanban-style task manager with authentication, board/list/card management, tagging, and Moodle course import.

This project is built as a portfolio-ready Go backend that demonstrates:

- layered application design
- REST API design with protected routes
- JWT-based authentication with refresh-token rotation
- PostgreSQL persistence with GORM
- Swagger/OpenAPI documentation
- third-party integration with Moodle

## What This Server Does

The server powers a kanban application where each user can:

- register and log in
- create personal boards
- work with default workflow lists created automatically for each board
- create, update, move, favorite, archive, and delete cards
- assign due dates and priority levels to cards
- create tags per board and attach/detach them from cards
- connect a Moodle account and import course activities into a board

When a board is created, the backend automatically creates three default lists so the client can start working immediately.

## Core Features

### 1. Authentication and Session Management

- `POST /api/auth/register` creates a new user
- `POST /api/auth/login` authenticates an existing user
- `POST /api/auth/refresh` rotates the refresh token and returns a new access token
- `POST /api/auth/logout` invalidates the current refresh token

How it works:

- access tokens are sent back to the client and used in the `Authorization: Bearer <token>` header
- refresh tokens are stored as HTTP-only cookies
- refresh tokens are hashed before being stored in PostgreSQL
- passwords are hashed with `bcrypt`

This gives the project a more realistic session model than storing everything client-side.

### 2. Board, List, Card, and Tag Management

The API supports a full kanban workflow:

- boards are owned by a user
- lists belong to a board
- cards belong to a list
- tags belong to a board and can be linked to cards through a many-to-many relation

Supported card behavior includes:

- title and description
- due dates
- priority: `low`, `medium`, `high`
- drag-and-drop friendly position updates
- cross-list movement
- favorites
- archive state

### 3. Moodle Integration

Users can connect a Moodle account and reuse that connection later.

The Moodle integration supports:

- saving a Moodle connection per user
- fetching the user’s enrolled courses
- importing a Moodle course as a kanban board

Current import behavior:

- the server creates a board from the selected Moodle course
- imported activities are added to the first default list
- supported Moodle modules are currently `assign` and `quiz`
- due dates are extracted from Moodle metadata when available
- Moodle access tokens are encrypted before being stored

This part of the project shows external API integration, credential handling, data transformation, and domain mapping.

## Architecture

The code is organized using a layered structure:

```text
cmd/app
  main.go                 # application bootstrap

internal/
  adapters/
    http_server/          # Gin controllers, route registration, middleware
    postgres/             # GORM repositories
    moodle/               # Moodle HTTP client
  config/                 # config loading
  domain/
    entity/               # domain models and domain errors
    services/             # business logic
```

### Request Flow

Typical request flow looks like this:

1. Gin route receives the request.
2. Middleware validates the JWT for protected endpoints.
3. Controller parses input and translates HTTP concerns into service calls.
4. Domain service applies business rules.
5. Repository persists or fetches data from PostgreSQL.
6. Controller returns a JSON response.

That separation keeps HTTP, business logic, and data access reasonably isolated.

## Technologies

- Go 1.24
- Gin
- GORM
- PostgreSQL
- JWT
- Swagger via `swaggo`
- Moodle REST integration

## Data Model Overview

Main entities:

- `User`
- `Board`
- `List`
- `Card`
- `Tag`
- `RefreshToken`
- `MoodleConnection`

Relationship summary:

- one user owns many boards
- one board has many lists
- one list has many cards
- one board has many tags
- cards and tags are connected through a join table
- one user can have one stored Moodle connection

## API Overview

### Auth

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`

### Boards

- `POST /api/boards`
- `GET /api/boards`
- `GET /api/boards/:id`
- `PATCH /api/boards/:id`
- `DELETE /api/boards/:id`

### Lists

- `POST /api/lists`
- `GET /api/lists/:id`
- `PUT /api/lists/:id`
- `DELETE /api/lists/:id`
- `PATCH /api/lists/:id/position`
- `GET /api/boards/:id/lists`

### Cards

- `POST /api/cards`
- `GET /api/cards/:id`
- `PATCH /api/cards/:id`
- `PATCH /api/cards/:id/position`
- `DELETE /api/cards/:id`
- `GET /api/lists/:id/cards`

### Tags

- `POST /api/tags`
- `GET /api/tags/:id`
- `PATCH /api/tags/:id`
- `DELETE /api/tags/:id`
- `GET /api/boards/:id/tags`
- `GET /api/cards/:id/tags`
- `POST /api/cards/:id/tags/:tag_id`
- `DELETE /api/cards/:id/tags/:tag_id`

### Moodle

- `POST /api/integrations/moodle/connect`
- `GET /api/integrations/moodle/connection`
- `GET /api/integrations/moodle/courses`
- `POST /api/integrations/moodle/import-board`

## Swagger Documentation

Swagger UI is exposed by the server at:

`/swagger/index.html`

After starting the server locally, the default URL is:

`http://localhost:8003/swagger/index.html`

## Configuration

The server reads configuration from:

- `config/config.yaml`
- `.env`

### Example `config/config.yaml`

```yaml
http_server:
  address: ":8003"
  timeout: 4s
  idle_timeout: 60s
postgres_storage:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your_password"
  dbname: "kanban"
```

### Example `.env`

```env
JWT_SECRET=change_me
MOODLE_TOKEN_ENCRYPTION_KEY=change_me
CLIENT_ORIGIN=http://localhost:5173
COOKIE_DOMAIN=localhost
COOKIE_SECURE=false
```

## Running Locally

### Prerequisites

- Go 1.24+
- PostgreSQL

### 1. Create the database

Create a PostgreSQL database named `kanban`, or update `config/config.yaml` to match your local setup.

### 2. Configure environment

Create:

- `server/.env`
- `server/config/config.yaml`

Use the example values above.

### 3. Start the API

From the `server` directory:

```bash
go run ./cmd/app
```

The API will start on:

`http://localhost:8003`

## Persistence Notes

At startup, the application runs `AutoMigrate` for the main entities. That makes local onboarding easier for a demo or portfolio project, because the schema is created automatically when the server boots.

## Security Notes

This project includes several production-style security ideas:

- password hashing with `bcrypt`
- refresh-token hashing before persistence
- JWT validation middleware for protected routes
- encrypted Moodle tokens before storage
- HTTP-only refresh cookies

For a real production deployment, the next steps would be:

- stronger environment and secret management
- stricter CORS policy management by environment
- HTTPS-only secure cookies in deployed environments
- structured observability and request tracing
- automated tests for critical flows

## Why This Project Works Well in a Portfolio

This backend is useful as a portfolio project because it shows more than CRUD:

- authentication with session renewal
- relational modeling
- layered backend design
- third-party integration
- API documentation
- practical business rules around cards, tags, ordering, and ownership

It demonstrates that the application is built around user workflows, not only database tables.

## Possible Next Improvements

If I continue developing this project, the next improvements I would prioritize are:

- Docker setup for API + PostgreSQL
- database migrations instead of relying only on `AutoMigrate`
- unit and integration tests
- role-based access control or board collaboration
- background jobs for external sync/import
- better observability and error reporting

## Entry Point

Start here if you want to review the backend quickly:

- `cmd/app/main.go`
- `internal/domain/services`
- `internal/adapters/http_server`
- `internal/adapters/postgres`

Those folders tell the story of how the system is wired from HTTP request to persistence.
