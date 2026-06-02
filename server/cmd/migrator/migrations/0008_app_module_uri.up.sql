-- Add the browser-reachable Module-Federation remote URL for each app's
-- console plugin (the Grafana-equivalent of a plugin's module.js). Nullable:
-- apps without a console remote (e.g. catalog, adoptions) leave it NULL.
ALTER TABLE foundry.app ADD COLUMN module_uri TEXT;
