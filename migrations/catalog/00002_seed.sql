-- +goose Up
-- +goose StatementBegin
INSERT INTO catalog.specialty (id, name) VALUES
  (1, 'Терапевт'),
  (2, 'Кардиолог'),
  (3, 'Невролог');

INSERT INTO catalog.patient (id, user_id, full_name, birth_date, gender, phone, address) VALUES
  (1, 5, 'Козлов Артём Андреевич',  '1998-04-12', 'male',   '+375291000005', 'Минск, ул. Платонова, 39'),
  (2, 6, 'Смирнова Мария Олеговна', '1985-11-03', 'female', '+375291000006', 'Минск, пр. Независимости, 4'),
  (3, 7, 'Новик Денис Павлович',    '2001-07-21', 'male',   NULL,            NULL);

INSERT INTO catalog.doctor (id, user_id, specialty_id, full_name, office) VALUES
  (1, 2, 1, 'Иванов Пётр Сергеевич',   '101'),
  (2, 3, 2, 'Петрова Анна Игоревна',   '205'),
  (3, 4, 3, 'Сидоров Олег Викторович', NULL);

INSERT INTO catalog.medical_service (id, name, price, duration_min) VALUES
  (1, 'Первичная консультация терапевта', 35.00, 20),
  (2, 'Консультация кардиолога',          55.00, 30),
  (3, 'ЭКГ',                              25.50, 15),
  (4, 'Консультация невролога',           50.00, NULL);

INSERT INTO catalog.medication (id, name, form, unit, stock_qty) VALUES
  (1, 'Парацетамол 500 мг', 'таблетки', 'шт', 200),
  (2, 'Бисопролол 5 мг',    'таблетки', 'шт', 90),
  (3, 'Натрия хлорид 0,9%', 'ампулы',   'мл', 1500.5);

INSERT INTO catalog.medication_dispense (id, medication_id, patient_id, appointment_id, quantity) VALUES
  (1, 1, 1, 1, 10);

SELECT setval('catalog.specialty_id_seq', 3);
SELECT setval('catalog.patient_id_seq', 3);
SELECT setval('catalog.doctor_id_seq', 3);
SELECT setval('catalog.medical_service_id_seq', 4);
SELECT setval('catalog.medication_id_seq', 3);
SELECT setval('catalog.medication_dispense_id_seq', 1);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM catalog.medication_dispense WHERE id IN (1);
DELETE FROM catalog.medication WHERE id IN (1, 2, 3);
DELETE FROM catalog.medical_service WHERE id IN (1, 2, 3, 4);
DELETE FROM catalog.doctor WHERE id IN (1, 2, 3);
DELETE FROM catalog.patient WHERE id IN (1, 2, 3);
DELETE FROM catalog.specialty WHERE id IN (1, 2, 3);
SELECT setval('catalog.medication_dispense_id_seq', 1, false);
SELECT setval('catalog.medication_id_seq', 1, false);
SELECT setval('catalog.medical_service_id_seq', 1, false);
SELECT setval('catalog.doctor_id_seq', 1, false);
SELECT setval('catalog.patient_id_seq', 1, false);
SELECT setval('catalog.specialty_id_seq', 1, false);
-- +goose StatementEnd
