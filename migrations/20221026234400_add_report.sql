-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS reports (
    report_id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    service_id bigint NOT NULL,
    amount bigint NOT NULL CHECK (amount > 0),
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reports;
-- +goose StatementEnd
