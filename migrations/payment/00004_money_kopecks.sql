-- +goose Up
-- +goose StatementBegin
-- Сумма платежа — целые копейки вместо NUMERIC(12,2). payment_amount_check
-- удаляется вместе со старой колонкой.
ALTER TABLE payment.payment ADD COLUMN amount_kopecks BIGINT;
UPDATE payment.payment SET amount_kopecks = ROUND(amount * 100)::BIGINT;
ALTER TABLE payment.payment
  ALTER COLUMN amount_kopecks SET NOT NULL,
  DROP COLUMN amount,
  ADD CONSTRAINT payment_amount_kopecks_check CHECK (amount_kopecks > 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payment.payment ADD COLUMN amount NUMERIC(12,2);
UPDATE payment.payment SET amount = amount_kopecks / 100.0;
ALTER TABLE payment.payment
  ALTER COLUMN amount SET NOT NULL,
  DROP COLUMN amount_kopecks,
  ADD CONSTRAINT payment_amount_check CHECK (amount > 0);
-- +goose StatementEnd
