-- +goose Up
-- +goose StatementBegin
CREATE TABLE notification.notification (
  id              BIGSERIAL PRIMARY KEY,
  user_id         BIGINT       NOT NULL,       -- IAM
  channel         VARCHAR(32)  NOT NULL DEFAULT 'in_app',  -- in_app | email | push
  title           VARCHAR(255) NOT NULL,
  body            TEXT         NOT NULL,
  status          VARCHAR(32)  NOT NULL DEFAULT 'pending', -- pending | sent | read | failed
  related_type    VARCHAR(64),
  related_id      BIGINT,
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  sent_at         TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notification.notification;
-- +goose StatementEnd
