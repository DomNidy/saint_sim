-- +goose Up
ALTER TABLE
    simulation
ADD
    COLUMN IF NOT EXISTS owner_id text REFERENCES "user" (id) ON DELETE
SET
    NULL;

-- +goose Down
ALTER TABLE
    simulation DROP COLUMN IF EXISTS owner_id;