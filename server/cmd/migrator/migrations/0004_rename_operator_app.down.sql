ALTER TABLE foundry.role_app RENAME COLUMN app_slug TO product_slug;
ALTER TABLE foundry.role_app RENAME TO role_product;

ALTER TABLE foundry.role RENAME COLUMN all_apps TO all_products;

ALTER TABLE foundry.app RENAME TO product;

ALTER TABLE foundry.session RENAME COLUMN user_id TO operator_id;

ALTER TABLE foundry.user_role RENAME COLUMN user_id TO operator_id;
ALTER TABLE foundry.user_role RENAME TO operator_role;

ALTER TABLE foundry.user_settings RENAME COLUMN user_id TO operator_id;
ALTER TABLE foundry.user_settings RENAME TO operator_settings;

ALTER TABLE foundry.user_credential RENAME COLUMN user_id TO operator_id;
ALTER TABLE foundry.user_credential RENAME TO operator_credential;

ALTER TABLE foundry.app_user RENAME TO operator;
