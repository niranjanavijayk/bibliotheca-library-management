.PHONY: db api test web
db:   ; docker compose up -d db
api:  ; cd backend && DATABASE_URL="postgres://bib:bib@localhost:5432/bibliotheca?sslmode=disable" go run ./cmd/server
test: ; cd backend && go test -race -cover ./...
web:  ; cd frontend && npm start
