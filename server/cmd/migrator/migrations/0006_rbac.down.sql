-- Best-effort reversal (dev-only). Recreate the binary model structures and
-- re-derive all_apps from a granted '*.*' permission; per-app/role bindings are
-- not reconstructed.
ALTER TABLE foundry.role ADD COLUMN all_apps BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE foundry.role SET all_apps = true
    WHERE slug IN (SELECT role_slug FROM foundry.role_permission WHERE permission_id = '*.*');

CREATE TABLE foundry.role_app (
    role_slug TEXT NOT NULL REFERENCES foundry.role (slug) ON DELETE CASCADE,
    app_slug  TEXT NOT NULL,
    PRIMARY KEY (role_slug, app_slug)
);

CREATE TABLE foundry.user_role (
    user_id   UUID NOT NULL REFERENCES foundry.app_user (id) ON DELETE CASCADE,
    role_slug TEXT NOT NULL REFERENCES foundry.role (slug) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_slug)
);

CREATE TABLE foundry.service_account_role (
    service_account_id UUID NOT NULL REFERENCES foundry.service_account (id) ON DELETE CASCADE,
    role_slug          TEXT NOT NULL REFERENCES foundry.role (slug) ON DELETE CASCADE,
    PRIMARY KEY (service_account_id, role_slug)
);

INSERT INTO foundry.user_role (user_id, role_slug)
    SELECT subject_id, role_slug FROM foundry.subject_role WHERE subject_type = 'USER' ON CONFLICT DO NOTHING;

INSERT INTO foundry.service_account_role (service_account_id, role_slug)
    SELECT subject_id, role_slug FROM foundry.subject_role WHERE subject_type = 'SERVICE_ACCOUNT' ON CONFLICT DO NOTHING;

DROP TABLE foundry.subject_role;
DROP TABLE foundry.role_permission;
ALTER TABLE foundry.role DROP COLUMN kind;
DROP TABLE foundry.permission;
