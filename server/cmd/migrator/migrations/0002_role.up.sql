-- Operator authorization (control-plane RBAC). A role grants access either to
-- ALL products (all_products) or to a specific set (role_product). Operators
-- are bound to roles via operator_role. The gateway checks these before
-- proxying to a product's admin API.
CREATE TABLE foundry.role (
    slug         TEXT PRIMARY KEY,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    name         TEXT NOT NULL,
    all_products BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE foundry.role_product (
    role_slug    TEXT NOT NULL REFERENCES foundry.role (slug) ON DELETE CASCADE,
    product_slug TEXT NOT NULL,
    PRIMARY KEY (role_slug, product_slug)
);

CREATE TABLE foundry.operator_role (
    operator_id UUID NOT NULL REFERENCES foundry.operator (id) ON DELETE CASCADE,
    role_slug   TEXT NOT NULL REFERENCES foundry.role (slug) ON DELETE CASCADE,
    PRIMARY KEY (operator_id, role_slug)
);
