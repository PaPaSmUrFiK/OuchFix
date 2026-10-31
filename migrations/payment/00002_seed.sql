-- +goose Up
-- +goose StatementBegin
INSERT INTO payment.payment (id, appointment_id, patient_id, service_id, amount, method, status, paid_at) VALUES
  (2, 2, 1, 2, 55.00, 'online', 'pending', NULL),
  (3, 3, 2, 4, 50.00, 'cash',   'failed',  NULL);

SELECT setval('payment.payment_id_seq', 3);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM payment.payment WHERE id IN (2, 3);
SELECT setval('payment.payment_id_seq', 1, false);
-- +goose StatementEnd
