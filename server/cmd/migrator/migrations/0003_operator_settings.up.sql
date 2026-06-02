-- Per-operator UI preferences. Kept minimal (theme only) for now; the row is
-- created lazily on first write and defaults apply until then.
CREATE TABLE foundry.operator_settings (
    operator_id UUID PRIMARY KEY REFERENCES foundry.operator (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    theme       TEXT NOT NULL DEFAULT 'system'
);
