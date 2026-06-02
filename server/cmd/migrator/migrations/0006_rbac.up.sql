-- Action+scope RBAC. Permissions use a "<resourceType>.<verb>" grammar and are
-- granted to roles; subjects (users + service accounts) bind to roles through a
-- single generalized subject_role table. Replaces the binary all_apps/role_app
-- model.
CREATE TABLE foundry.permission (
    id            TEXT PRIMARY KEY,
    resource_type TEXT NOT NULL,
    verb          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT ''
);

ALTER TABLE foundry.role ADD COLUMN kind TEXT NOT NULL DEFAULT 'CUSTOM';

CREATE TABLE foundry.role_permission (
    role_slug     TEXT NOT NULL REFERENCES foundry.role (slug) ON DELETE CASCADE,
    permission_id TEXT NOT NULL,
    PRIMARY KEY (role_slug, permission_id)
);

CREATE TABLE foundry.subject_role (
    subject_type TEXT NOT NULL,
    subject_id   UUID NOT NULL,
    role_slug    TEXT NOT NULL REFERENCES foundry.role (slug) ON DELETE CASCADE,
    PRIMARY KEY (subject_type, subject_id, role_slug)
);

-- Data migration: derive permissions + bindings from the old model.
INSERT INTO foundry.role_permission (role_slug, permission_id)
    SELECT slug, '*.*' FROM foundry.role WHERE all_apps = true ON CONFLICT DO NOTHING;

UPDATE foundry.role SET kind = 'SYSTEM' WHERE slug = 'superadmin';

INSERT INTO foundry.role_permission (role_slug, permission_id)
    SELECT role_slug, 'app:' || app_slug || '.read' FROM foundry.role_app ON CONFLICT DO NOTHING;

INSERT INTO foundry.subject_role (subject_type, subject_id, role_slug)
    SELECT 'USER', user_id, role_slug FROM foundry.user_role ON CONFLICT DO NOTHING;

INSERT INTO foundry.subject_role (subject_type, subject_id, role_slug)
    SELECT 'SERVICE_ACCOUNT', service_account_id, role_slug FROM foundry.service_account_role ON CONFLICT DO NOTHING;

DROP TABLE foundry.user_role;
DROP TABLE foundry.service_account_role;
ALTER TABLE foundry.role DROP COLUMN all_apps;
DROP TABLE foundry.role_app;
