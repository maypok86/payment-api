-- +goose Up
-- +goose StatementBegin
CREATE TYPE transaction_type AS ENUM ('enrollment', 'transfer', 'reservation', 'write-off');

CREATE TABLE IF NOT EXISTS transactions (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    type transaction_type NOT NULL,
    sender_id bigint NOT NULL REFERENCES accounts(id),
    receiver_id bigint NOT NULL REFERENCES accounts(id),
    amount bigint NOT NULL,
    description text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
-- +goose StatementEnd
