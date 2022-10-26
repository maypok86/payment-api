-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    order_id bigint PRIMARY KEY,
    account_id bigint NOT NULL REFERENCES accounts(account_id),
    service_id bigint NOT NULL,
    amount bigint NOT NULL,
    is_paid boolean NOT NULL DEFAULT false,
    is_cancelled boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER set_timestamp
AFTER UPDATE ON orders
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
