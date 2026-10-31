-- +goose Up
-- +goose StatementBegin
ALTER TABLE payment.payment
  ADD CONSTRAINT payment_method_check CHECK (method IN ('cash', 'card', 'online')),
  ADD CONSTRAINT payment_status_check CHECK (status IN ('pending', 'paid', 'failed')),
  ADD CONSTRAINT payment_amount_check CHECK (amount > 0);

-- Приём оплачивается один раз.
CREATE UNIQUE INDEX payment_paid_appointment_uniq ON payment.payment (appointment_id)
  WHERE status = 'paid';
CREATE INDEX payment_paid_at_idx ON payment.payment (paid_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS payment.payment_paid_at_idx;
DROP INDEX IF EXISTS payment.payment_paid_appointment_uniq;
ALTER TABLE payment.payment
  DROP CONSTRAINT IF EXISTS payment_amount_check,
  DROP CONSTRAINT IF EXISTS payment_status_check,
  DROP CONSTRAINT IF EXISTS payment_method_check;
-- +goose StatementEnd
