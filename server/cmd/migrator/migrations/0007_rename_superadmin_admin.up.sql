-- Rename the built-in all-access role superadmin → admin (Grafana's term).
-- The FKs on role_permission/subject_role reference role(slug) ON DELETE
-- CASCADE with no ON UPDATE CASCADE, so we can't rename the PK in place:
-- insert the new role, repoint the children, then drop the old role.
INSERT INTO foundry.role (slug, name, kind)
SELECT 'admin', 'Admin', kind FROM foundry.role WHERE slug = 'superadmin'
ON CONFLICT (slug) DO NOTHING;

UPDATE foundry.role_permission SET role_slug = 'admin' WHERE role_slug = 'superadmin';
UPDATE foundry.subject_role SET role_slug = 'admin' WHERE role_slug = 'superadmin';

DELETE FROM foundry.role WHERE slug = 'superadmin';
