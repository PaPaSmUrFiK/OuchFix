-- +goose Up
-- +goose StatementBegin
-- Слоты на 2026-11-02 и 2026-11-03 (время по Минску, UTC+3).
INSERT INTO appointment.schedule_slot (id, doctor_id, starts_at, ends_at, status) VALUES
  (1, 1, '2026-11-02 09:00+03', '2026-11-02 09:20+03', 'busy'),
  (2, 1, '2026-11-02 09:20+03', '2026-11-02 09:40+03', 'free'),
  (3, 1, '2026-11-02 09:40+03', '2026-11-02 10:00+03', 'free'),
  (4, 1, '2026-11-02 10:00+03', '2026-11-02 10:20+03', 'cancelled'),
  (5, 2, '2026-11-02 14:00+03', '2026-11-02 14:30+03', 'busy'),
  (6, 2, '2026-11-02 14:30+03', '2026-11-02 15:00+03', 'free'),
  (7, 2, '2026-11-03 14:00+03', '2026-11-03 14:30+03', 'free'),
  (8, 3, '2026-11-03 11:00+03', '2026-11-03 11:30+03', 'busy');

INSERT INTO appointment.appointment
  (id, patient_id, doctor_id, slot_id, service_id, status, queue_type, scheduled_at, conclusion) VALUES
  (1, 1, 1, 1,    1, 'completed', 'scheduled', '2026-11-02 09:00+03', 'ОРВИ, лёгкое течение'),
  (2, 1, 2, 5,    2, 'scheduled', 'scheduled', '2026-11-02 14:00+03', NULL),
  (3, 2, 3, 8,    4, 'scheduled', 'scheduled', '2026-11-03 11:00+03', NULL),
  (4, 3, 1, NULL, 1, 'queue',     'walk_in',   NULL,                  NULL),
  (5, 2, 2, NULL, 3, 'cancelled', 'walk_in',   NULL,                  NULL);

INSERT INTO appointment.diagnosis (id, appointment_id, code, description) VALUES
  (1, 1, 'J06.9', 'Острая инфекция верхних дыхательных путей неуточнённая'),
  (2, 1, NULL,    'Субфебрильная температура');

INSERT INTO appointment.prescription (id, appointment_id, text, medication_id) VALUES
  (1, 1, 'Парацетамол 500 мг при температуре выше 38, не чаще 4 раз в сутки', 1),
  (2, 1, 'Обильное тёплое питьё, постельный режим 3 дня', NULL);

SELECT setval('appointment.schedule_slot_id_seq', 8);
SELECT setval('appointment.appointment_id_seq', 5);
SELECT setval('appointment.diagnosis_id_seq', 2);
SELECT setval('appointment.prescription_id_seq', 2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM appointment.prescription WHERE id IN (1, 2);
DELETE FROM appointment.diagnosis WHERE id IN (1, 2);
DELETE FROM appointment.appointment WHERE id IN (1, 2, 3, 4, 5);
DELETE FROM appointment.schedule_slot WHERE id IN (1, 2, 3, 4, 5, 6, 7, 8);
SELECT setval('appointment.prescription_id_seq', 1, false);
SELECT setval('appointment.diagnosis_id_seq', 1, false);
SELECT setval('appointment.appointment_id_seq', 1, false);
SELECT setval('appointment.schedule_slot_id_seq', 1, false);
-- +goose StatementEnd
