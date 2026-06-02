-- Rename the domain vocabulary operator -> user and product -> app. Pure
-- rename, no data change. The user table is named app_user to avoid quoting
-- the reserved word "user" everywhere.
ALTER TABLE foundry.operator RENAME TO app_user;

ALTER TABLE foundry.operator_credential RENAME TO user_credential;
ALTER TABLE foundry.user_credential RENAME COLUMN operator_id TO user_id;

ALTER TABLE foundry.operator_settings RENAME TO user_settings;
ALTER TABLE foundry.user_settings RENAME COLUMN operator_id TO user_id;

ALTER TABLE foundry.operator_role RENAME TO user_role;
ALTER TABLE foundry.user_role RENAME COLUMN operator_id TO user_id;

ALTER TABLE foundry.session RENAME COLUMN operator_id TO user_id;

ALTER TABLE foundry.product RENAME TO app;

ALTER TABLE foundry.role RENAME COLUMN all_products TO all_apps;

ALTER TABLE foundry.role_product RENAME TO role_app;
ALTER TABLE foundry.role_app RENAME COLUMN product_slug TO app_slug;
