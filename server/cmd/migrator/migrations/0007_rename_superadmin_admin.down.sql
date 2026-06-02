INSERT INTO foundry.role (slug, name, kind)
SELECT 'superadmin', 'Superadmin', kind FROM foundry.role WHERE slug = 'admin'
ON CONFLICT (slug) DO NOTHING;

UPDATE foundry.role_permission SET role_slug = 'superadmin' WHERE role_slug = 'admin';
UPDATE foundry.subject_role SET role_slug = 'superadmin' WHERE role_slug = 'admin';

DELETE FROM foundry.role WHERE slug = 'admin';
