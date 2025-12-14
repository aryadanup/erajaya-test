
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price BIGINT NOT NULL,
    description TEXT,
    quantity INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) NULL,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    deleted_by VARCHAR(255) NULL
);


CREATE INDEX IF NOT EXISTS idx_products_search_gin 
ON products 
USING GIN (name gin_trgm_ops, description gin_trgm_ops)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_created_at_sort 
ON products (created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_price_sort 
ON products (price)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_name_sort 
ON products (name)
WHERE deleted_at IS NULL;