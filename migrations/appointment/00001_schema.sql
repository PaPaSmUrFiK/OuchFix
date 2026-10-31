-- +goose Up
-- +goose StatementBegin
CREATE TABLE appointment.schedule_slot (
  id            BIGSERIAL PRIMARY KEY,
  doctor_id     BIGINT       NOT NULL,         -- id врача из Catalog
  starts_at     TIMESTAMPTZ  NOT NULL,
  ends_at       TIMESTAMPTZ  NOT NULL,
  status        VARCHAR(32)  NOT NULL DEFAULT 'free'  -- free | busy | cancelled
);

CREATE TABLE appointment.appointment (
  id              BIGSERIAL PRIMARY KEY,
  patient_id      BIGINT       NOT NULL,       -- Catalog
  doctor_id       BIGINT       NOT NULL,       -- Catalog
  slot_id         BIGINT       REFERENCES appointment.schedule_slot(id),
  service_id      BIGINT,                       -- Catalog.medical_service
  status          VARCHAR(32)  NOT NULL,        -- scheduled | queue | completed | cancelled | no_show
  queue_type      VARCHAR(32)  NOT NULL,        -- scheduled | walk_in
  scheduled_at    TIMESTAMPTZ,
  conclusion      TEXT,
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE appointment.diagnosis (
  id              BIGSERIAL PRIMARY KEY,
  appointment_id  BIGINT       NOT NULL REFERENCES appointment.appointment(id),
  code            VARCHAR(64),
  description     TEXT         NOT NULL,
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE appointment.prescription (
  id              BIGSERIAL PRIMARY KEY,
  appointment_id  BIGINT       NOT NULL REFERENCES appointment.appointment(id),
  text            TEXT         NOT NULL,
  medication_id   BIGINT,                       -- Catalog.medication (ссылка-id)
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS appointment.prescription;
DROP TABLE IF EXISTS appointment.diagnosis;
DROP TABLE IF EXISTS appointment.appointment;
DROP TABLE IF EXISTS appointment.schedule_slot;
-- +goose StatementEnd
