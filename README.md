# 📚 Bibliotheca – Library Management System

> **Your shelf, your story.** A full-stack library management system with an antique, wood-and-leather look.
> Manage books, members and loans; track due dates and overdue books; see everything live on a dashboard.

**Stack:** Angular 19 · TypeScript · SCSS · Go 1.22 · REST · PostgreSQL

---

## ✨ Features

### Required functionality
| Area | What you can do |
|---|---|
| **Dashboard** | Total books, total members, currently issued books, overdue books – updated live after every change. Also: shelf availability meter, books-by-category chart, recent activity, overdue list. |
| **Book management** | Add, edit, delete, view, **search** (title/author/ISBN), **filter** (category, availability), **sort** (title, author, category, year, availability, recently added), pagination. Stores title, author, category, ISBN, availability, copies, year, cover and description. |
| **Member management** | Register, edit, delete, view, search, filter by status, sort. Name, email, phone, address, membership type, status, joined/expiry dates. |
| **Issue & return** | Issue only *available* books, record issue and due dates, process returns (the book becomes available again), identify **overdue** loans. |
| **Details & history** | Book detail page and member detail page, each with full borrowing history. |
| **Persistence** | All data stored in a relational database (PostgreSQL) and survives refresh/restart. |
| **Validation & feedback** | Required-field validation, server-side validation mapped onto form fields, toast notifications for success/failure, confirmation pop-ups for destructive actions. |

### Additional features (beyond the brief)
- 🏛️ Cinematic landing page with an antique bookshelf, shown before the dashboard
- 📖 Visual **bookshelf view** with cover images (Open Library covers by ISBN, generated fallback cover when none exists) + table view toggle
- ⏳ **Loan renewals** (up to 2×, 7 days each, not for overdue loans)
- 💰 **Late fees** (₹2/day, configurable) calculated automatically
- 🚦 Borrowing rules: max 5 active loans per member; suspended/expired members cannot borrow
- 🔄 Live search with debounce and request cancellation
- 🌱 Demo data seeded on first run (22 books, 6 members, sample loans incl. overdue)
- 📱 Fully responsive layout, accessible labels and keyboard-friendly dialogs

## 🧱 Architecture

```
Angular app  ──HTTP/JSON──►  Go REST API  ──SQL──►  PostgreSQL
(components, services)       handler → service → repository
```

```
bibliotheca/
├── backend/                     Go module "bibliotheca"
│   ├── cmd/server/main.go       entry point, wiring, graceful shutdown
│   ├── internal/
│   │   ├── models/              entities, DTOs, filters (structs + methods)
│   │   ├── repository/          interfaces + PostgreSQL implementation (+ in-memory store used by unit tests)
│   │   ├── service/             business logic (validation, rules, concurrency)
│   │   ├── handler/             REST handlers + tests
│   │   ├── middleware/          logging, panic recovery, CORS
│   │   ├── apperror/            typed errors → HTTP responses
│   │   └── seed/                demo data
│   ├── migrations/              SQL schema (embedded, auto-applied)
│   └── Dockerfile
├── frontend/                    Angular 19 (standalone components)
│   └── src/
│       ├── styles/              _variables, _mixins, _base, _components (SCSS, BEM)
│       └── app/
│           ├── core/            models, injectable services, interceptor, pipe, utils
│           ├── shared/          reusable components (modal, cover, badge, toast, pagination…)
│           ├── layout/          shell with sidebar
│           └── features/        landing, dashboard, books, members, loans
├── docs/                        SDLC document, screenshots, submission email
└── docker-compose.yml           PostgreSQL (+ API)
```

### Concepts used
- **Go:** modules & packages, structs, methods, pointers, interfaces (repository/service layers), goroutines + `sync.WaitGroup` (parallel dashboard queries, background overdue watcher), `sync.Mutex` (serialised lending), `context`, graceful shutdown, table-driven & concurrency tests, `-race`.
- **Angular:** standalone components, `inject()`-based injectable services (`providedIn: 'root'`), signals, `@if` / `@for` / `@switch` control flow, reactive forms, HTTP interceptor, lazy-loaded routes, custom pipe, `input()` / `output()`, reusable components.
- **SCSS:** variables, maps, `@each`, mixins (`respond`, `card`, `button-variant`, `badge-colors`, `wood-plank`…), partials with `@use`, **BEM** naming with modifiers (`.btn--primary`, `.badge--overdue`, `.sidebar__link--active`).

## 🛠️ Prerequisites
- **Go 1.22+** – https://go.dev/dl
- **Node.js 20+** and npm
- **PostgreSQL 14+** (local install) *or* Docker Desktop

## 🚀 Setup & run

> Data is always stored in **PostgreSQL**. The backend will not start without it.

### 1. Database
**Option A – Docker (one command)**
```bash
docker compose up -d db      # user=bib password=bib db=bibliotheca on :5432
```

**Option B – local PostgreSQL** (open *SQL Shell (psql)* as the `postgres` user)
```sql
CREATE USER bib WITH PASSWORD 'bib' SUPERUSER;
CREATE DATABASE bibliotheca OWNER bib;
```
Both options match the backend's default connection string, so no environment variable is needed.

### 2. Backend (http://localhost:8080)
```bash
cd backend
go mod download
go run ./cmd/server
```
The log should show `using PostgreSQL` and `Bibliotheca API listening on :8080`.
Tables are created automatically and demo data is seeded on first start only (when the catalogue is empty).

To use a different database, set `DATABASE_URL`:
```powershell
$env:DATABASE_URL="postgres://USER:PASSWORD@localhost:5432/bibliotheca?sslmode=disable"
```

| Variable | Default | Purpose |
|---|---|---|
| `DATABASE_URL` | `postgres://bib:bib@localhost:5432/bibliotheca?sslmode=disable` | PostgreSQL connection string (matches docker-compose) |
| `PORT` | `8080` | API port |
| `CORS_ORIGIN` | `http://localhost:4200` | Allowed frontend origin |
| `SEED` | `true` | Seed demo data when the catalogue is empty |

### 3. Frontend (http://localhost:4200)
```bash
cd frontend
npm install
npm start
```
Open **http://localhost:4200** – the landing page – then press **Enter the Library**.

### 4. Tests
```bash
cd backend
go test -race -cover ./...
```

## 🔌 REST API
Base URL `http://localhost:8080/api` · responses: `{ "data": …, "meta": … }` or `{ "error": { "code", "message", "fields" } }`

| Method | Endpoint | Description |
|---|---|---|
| GET | `/dashboard` | Totals, issued, overdue, categories, recent loans |
| GET | `/books?q=&category=&available=&sort=&order=&page=&pageSize=` | Search / filter / sort / paginate |
| POST | `/books` | Create book |
| GET / PUT / DELETE | `/books/{id}` | View / update / delete |
| GET | `/books/{id}/history` | Borrowing history of a book |
| GET | `/categories` | Categories with counts |
| GET | `/members?q=&status=&sort=&order=&page=&pageSize=` | List members |
| POST | `/members` | Register member |
| GET / PUT / DELETE | `/members/{id}` | View / update / delete |
| GET | `/members/{id}/history` | Borrowing history of a member |
| GET | `/loans?status=all\|active\|overdue\|returned&q=` | List loans |
| POST | `/loans` | Issue a book `{bookId, memberId, days}` |
| POST | `/loans/{id}/return` | Process a return |
| POST | `/loans/{id}/renew` | Renew a loan |

**Status codes:** `201` created · `204` deleted · `400` bad request · `404` not found · `409` conflict (e.g. book unavailable, already returned) · `422` validation / rule violation · `500` internal.

## 🧠 Business rules
- ISBN and member email are unique (duplicates rejected); ISBN normalised (`978-0-…` → digits).
- A book can only be issued while `availableCopies > 0`; returning increments it again. Lending is atomic (mutex + `SELECT … FOR UPDATE`).
- A member cannot borrow the same book twice, may hold at most 5 loans, and must be active with a valid membership.
- Books with issued copies and members with active loans cannot be deleted.
- Overdue = not returned and past the due date; fine = ₹2 × days late.

## 🖼️ Screenshots
| | |
|---|---|
| ![Landing](docs/screenshots/01-landing.png) | ![Dashboard](docs/screenshots/02-dashboard.png) |
| ![Bookshelf](docs/screenshots/03-books-shelf.png) | ![Books table](docs/screenshots/04-books-table.png) |
| ![Add book](docs/screenshots/05-add-book.png) | ![Book detail](docs/screenshots/06-book-detail.png) |
| ![Members](docs/screenshots/07-members.png) | ![Member detail](docs/screenshots/08-member-detail.png) |
| ![Loans](docs/screenshots/09-loans.png) | ![Issue book](docs/screenshots/10-issue-book.png) |

## 📄 Documentation
- [SDLC document](docs/SDLC.md) – requirements, design, implementation, testing, deployment, maintenance
- [Submission email](docs/SUBMISSION_EMAIL.md) · [Submission guide](docs/SUBMISSION_GUIDE.md) · [Presentation](docs/Bibliotheca_Presentation.pptx) · [Talking notes](docs/PRESENTATION_NOTES.md)

## 🧯 Troubleshooting
| Problem | Fix |
|---|---|
| `cannot connect to PostgreSQL` / `connection refused` | Start PostgreSQL (`docker compose up -d db`, or start the `postgresql` Windows service). |
| `password authentication failed` | Create the `bib` user (Option B above) or fix the password in `DATABASE_URL`; special characters must be URL-encoded. |
| `ng is not recognized` | Run `npm install` inside `frontend` first. |
| Frontend shows "Cannot reach the server" | Start the backend and confirm `http://localhost:8080/api/health` returns OK. |
| CORS error | Set `CORS_ORIGIN` to the exact frontend URL. |
| Book covers not loading | Covers come from openlibrary.org; a generated cover is used when offline. |

## 👤 Author
**[Your Name]** · Beinex Internship · [GitHub](https://github.com/[username])
