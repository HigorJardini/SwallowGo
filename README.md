🐦 SwallowGo – Collaborative Travel Planning API

SwallowGo is a backend REST API built with Go 1.21+, designed to support collaborative travel planning. Whether you're organizing a short trip with friends or a longer adventure, SwallowGo helps you plan and manage everything — from inviting participants to tracking activities and important trip links.

"Plan together. Travel better."

---------------------
✨ Features
---------------------

- Trip creation and management
- Invite and manage friends per trip
- Add and organize trip activities
- Save useful links (tickets, documents, etc.)
- Input validation and error handling
- JWT-ready architecture (if added)

---------------------
🛠️ Tech Stack
---------------------

Language:           Go 1.21+
HTTP Router:        github.com/go-chi/chi
Validation:         github.com/go-playground/validator/v10
Logging:            go.uber.org/zap
UUIDs:              github.com/google/uuid
PostgreSQL Driver:  github.com/jackc/pgx/v5
OpenAPI Support:    github.com/discord-gophers/goapi-gen
MIME Detection:     github.com/gabriel-vasile/mimetype
Email (optional):   github.com/wneessen/go-mail

---------------------
🚀 Getting Started
---------------------

Requirements:
- Go 1.21 or newer (recommended: 1.22+)
- PostgreSQL database
- Docker (optional)

Installation:
> git clone https://github.com/your-user/swallowgo.git
> cd swallowgo
> go mod tidy
> go run main.go

Note: You may need to adjust main.go depending on your structure.

---------------------
🧩 Project Structure (suggested)
---------------------

swallowgo/
├── cmd/               # App entrypoint
├── api/               # Route definitions & HTTP handlers
├── internal/          # Business logic
├── models/            # DTOs & domain models
├── repository/        # DB access (pgx)
├── config/            # App configs (env, flags)
├── docs/              # OpenAPI specs, if using
└── main.go            # Bootstrap

---------------------
📦 go.mod Highlights
---------------------

You’re using a solid, production-ready stack:
- chi + chi/v5 for routing
- goapi-gen and kin-openapi for API definition
- zap for structured logging
- pgx/v5 for PostgreSQL
- validator for field-level validation
- go-mail for potential email integrations
- gutils for personal utilities

---------------------
🧠 Roadmap
---------------------

[x] Trip creation
[x] Friend invitations
[x] Activities and link tracking
[ ] Role-based access control
[ ] Notifications
[ ] OAuth support
[ ] Admin panel

---------------------
🧪 Testing
---------------------

Run tests with:
> go test ./...

---------------------
📄 License
---------------------

Licensed under the MIT License

---------------------
🌍 About the Name
---------------------

The name "SwallowGo" is inspired by migratory birds (swallows)
that travel together — just like a group of friends planning
their next adventure. And of course, it’s written in Go.

