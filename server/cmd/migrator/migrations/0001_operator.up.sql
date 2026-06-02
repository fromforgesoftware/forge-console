-- Operators are the only accounts in Foundry: the people who administer the
-- console and its products. They are never application end-users (those live in
-- Aegis realms). Identity is global — no realm — and access is by role.
CREATE TABLE foundry.operator (
    id           UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    email        TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'ENABLED'
);

-- Local password credential (argon2id), one per operator. External-OIDC-only
-- operators simply have no row here.
CREATE TABLE foundry.operator_credential (
    operator_id UUID PRIMARY KEY REFERENCES foundry.operator (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    hash        TEXT NOT NULL
);

-- Server-side browser sessions (cookie carries the id) so logout is real.
CREATE TABLE foundry.session (
    id          UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    operator_id UUID NOT NULL REFERENCES foundry.operator (id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_session_operator ON foundry.session (operator_id);

-- The set of managed products (Aegis/Hallmark/Herald), with the in-cluster
-- admin API base the gateway proxies to. Seeded from config at boot.
CREATE TABLE foundry.product (
    slug           TEXT PRIMARY KEY,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    name           TEXT NOT NULL,
    kind           TEXT NOT NULL DEFAULT '',
    admin_base_url TEXT NOT NULL DEFAULT '',
    enabled        BOOLEAN NOT NULL DEFAULT TRUE
);
