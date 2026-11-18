CREATE TABLE IF NOT EXISTS books (
    id               BIGSERIAL PRIMARY KEY,
    title            VARCHAR(255) NOT NULL,
    author           VARCHAR(255) NOT NULL,
    category         VARCHAR(100) NOT NULL,
    isbn             VARCHAR(20)  NOT NULL UNIQUE,
    description      TEXT         NOT NULL DEFAULT '',
    cover_url        TEXT         NOT NULL DEFAULT '',
    published_year   INT          NOT NULL DEFAULT 0,
    total_copies     INT          NOT NULL CHECK (total_copies >= 1),
    available_copies INT          NOT NULL CHECK (available_copies >= 0),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CHECK (available_copies <= total_copies)
);

CREATE TABLE IF NOT EXISTS members (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(150) NOT NULL,
    email           VARCHAR(190) NOT NULL,
    phone           VARCHAR(25)  NOT NULL,
    address         TEXT         NOT NULL DEFAULT '',
    membership_type VARCHAR(20)  NOT NULL DEFAULT 'Standard',
    status          VARCHAR(20)  NOT NULL DEFAULT 'active',
    joined_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ  NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_members_email ON members (LOWER(email));

CREATE TABLE IF NOT EXISTS loans (
    id          BIGSERIAL PRIMARY KEY,
    book_id     BIGINT      NOT NULL REFERENCES books(id)   ON DELETE CASCADE,
    member_id   BIGINT      NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    issued_at   TIMESTAMPTZ NOT NULL,
    due_date    TIMESTAMPTZ NOT NULL,
    returned_at TIMESTAMPTZ,
    fine        NUMERIC(10,2) NOT NULL DEFAULT 0,
    renewals    INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS ix_loans_book   ON loans (book_id);
CREATE INDEX IF NOT EXISTS ix_loans_member ON loans (member_id);
CREATE INDEX IF NOT EXISTS ix_loans_open   ON loans (due_date) WHERE returned_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ux_loans_one_open_per_pair ON loans (book_id, member_id) WHERE returned_at IS NULL;
