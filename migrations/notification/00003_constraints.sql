-- +goose Up
-- +goose StatementBegin
ALTER TABLE notification.notification
  ADD CONSTRAINT notification_status_check CHECK (status IN ('pending', 'sent', 'read', 'failed')),
  ADD CONSTRAINT notification_channel_check CHECK (channel IN ('in_app', 'email', 'push'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE notification.notification
  DROP CONSTRAINT IF EXISTS notification_channel_check,
  DROP CONSTRAINT IF EXISTS notification_status_check;
-- +goose StatementEnd
