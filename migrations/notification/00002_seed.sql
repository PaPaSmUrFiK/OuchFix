-- +goose Up
INSERT INTO notification.notification (id, user_id, channel, title, body, status, related_type, related_id, sent_at) VALUES
  (1, 5, 'in_app', 'Запись подтверждена', 'Вы записаны к кардиологу на 02.11.2026 14:00', 'sent', 'appointment', 2, '2026-10-28 12:00+03'),
  (2, 5, 'email',  'Оплата получена',     'Оплата 35.00 BYN за приём №1 прошла успешно',   'read', 'payment',     1, '2026-11-02 09:26+03'),
  (3, 5, 'push',   'Напоминание',         'Завтра в 14:00 приём у кардиолога',             'pending', 'appointment', 2, NULL),
  (4, 6, 'in_app', 'Оплата не прошла',    'Платёж за приём №3 отклонён',                   'failed', 'payment', 3, NULL),
  (5, 2, 'in_app', 'Новая запись',        'К вам записан пациент на 02.11.2026 09:00',     'sent', 'schedule', 1, '2026-10-28 10:00+03');

SELECT setval('notification.notification_id_seq', 5);

-- +goose Down
TRUNCATE notification.notification RESTART IDENTITY CASCADE;
