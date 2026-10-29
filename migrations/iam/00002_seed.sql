-- +goose Up
-- +goose StatementBegin
-- Тестовые данные ЛР3. Пароли не настоящие: вход появится в ЛР5.
INSERT INTO iam.role (id, name, description) VALUES
  (1, 'patient', 'Пациент'),
  (2, 'doctor',  'Врач'),
  (3, 'admin',   'Администратор');

INSERT INTO iam.permission (id, code, description) VALUES
  (1, 'appointment:create', 'Запись на приём'),
  (2, 'appointment:read',   'Просмотр приёмов'),
  (3, 'appointment:complete', 'Проведение приёма'),
  (4, 'schedule:manage',    'Ведение расписания'),
  (5, 'medication:manage',  'Учёт медикаментов'),
  (6, 'payment:create',     'Оплата');

INSERT INTO iam.role_permission (role_id, permission_id) VALUES
  (1, 1), (1, 2), (1, 6),
  (2, 2), (2, 3),
  (3, 1), (3, 2), (3, 4), (3, 5), (3, 6);

INSERT INTO iam.app_user (id, role_id, email, password_hash, full_name, phone) VALUES
  (1, 3, 'admin@ouchfix.by',    'not-a-real-hash', 'Администратор Центра', '+375291000001'),
  (2, 2, 'ivanov@ouchfix.by',   'not-a-real-hash', 'Иванов Пётр Сергеевич', '+375291000002'),
  (3, 2, 'petrova@ouchfix.by',  'not-a-real-hash', 'Петрова Анна Игоревна', '+375291000003'),
  (4, 2, 'sidorov@ouchfix.by',  'not-a-real-hash', 'Сидоров Олег Викторович', '+375291000004'),
  (5, 1, 'kozlov@mail.by',      'not-a-real-hash', 'Козлов Артём Андреевич', '+375291000005'),
  (6, 1, 'smirnova@mail.by',    'not-a-real-hash', 'Смирнова Мария Олеговна', '+375291000006'),
  (7, 1, 'novik@mail.by',       'not-a-real-hash', 'Новик Денис Павлович', NULL);

SELECT setval('iam.role_id_seq', 3);
SELECT setval('iam.permission_id_seq', 6);
SELECT setval('iam.app_user_id_seq', 7);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM iam.app_user WHERE id IN (1, 2, 3, 4, 5, 6, 7);
DELETE FROM iam.role_permission WHERE (role_id, permission_id) IN (
  (1, 1), (1, 2), (1, 6),
  (2, 2), (2, 3),
  (3, 1), (3, 2), (3, 4), (3, 5), (3, 6));
DELETE FROM iam.permission WHERE id IN (1, 2, 3, 4, 5, 6);
DELETE FROM iam.role WHERE id IN (1, 2, 3);
SELECT setval('iam.app_user_id_seq', 1, false);
SELECT setval('iam.permission_id_seq', 1, false);
SELECT setval('iam.role_id_seq', 1, false);
-- +goose StatementEnd
