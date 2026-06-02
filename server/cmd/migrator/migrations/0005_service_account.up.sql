-- Service accounts are machine identities that authenticate via
-- client-credentials (client_id + secret) and receive a bearer JWT to call
-- Foundry's gateway/admin APIs. The secret is stored only as an argon2 hash.
CREATE TABLE foundry.service_account (
    id           UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    name         TEXT NOT NULL,
    client_id    TEXT NOT NULL UNIQUE,
    secret_hash  TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'ENABLED',
    last_used_at TIMESTAMPTZ
);

-- Service accounts hold roles just like users do (a parallel binding table to
-- foundry.user_role). Phase C will generalize bindings across subject kinds.
CREATE TABLE foundry.service_account_role (
    service_account_id UUID NOT NULL REFERENCES foundry.service_account (id) ON DELETE CASCADE,
    role_slug          TEXT NOT NULL REFERENCES foundry.role (slug) ON DELETE CASCADE,
    PRIMARY KEY (service_account_id, role_slug)
);
