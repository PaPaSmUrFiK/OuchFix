-- +goose Up
INSERT INTO payment.payment (id, appointment_id, patient_id, service_id, amount, method, status, paid_at) VALUES
  (1, 1, 1, 1, 35.00, 'card',   'paid',    '2026-11-02 09:25+03'),
  (2, 2, 1, 2, 55.00, 'online', 'pending', NULL),
  (3, 3, 2, 4, 50.00, 'cash',   'failed',  NULL);

SELECT setval('payment.payment_id_seq', 3);

-- +goose Down
TRUNCATE payment.payment RESTART IDENTITY CASCADE;
