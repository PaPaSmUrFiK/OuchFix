-- +goose Up
-- +goose StatementBegin
CREATE TABLE catalog.specialty (
  id            BIGSERIAL PRIMARY KEY,
  name          VARCHAR(128) NOT NULL UNIQUE
);

CREATE TABLE catalog.patient (
  id            BIGSERIAL PRIMARY KEY,
  user_id       BIGINT       NOT NULL UNIQUE,  -- id из IAM (без FK между сервисами)
  full_name     VARCHAR(255) NOT NULL,
  birth_date    DATE,
  gender        VARCHAR(16),
  phone         VARCHAR(32),
  address       VARCHAR(512),
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE catalog.doctor (
  id            BIGSERIAL PRIMARY KEY,
  user_id       BIGINT       NOT NULL UNIQUE,  -- id из IAM
  specialty_id  BIGINT       NOT NULL REFERENCES catalog.specialty(id),
  full_name     VARCHAR(255) NOT NULL,
  office        VARCHAR(64),
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE catalog.medical_service (
  id            BIGSERIAL PRIMARY KEY,
  name          VARCHAR(255) NOT NULL,
  price         NUMERIC(12,2) NOT NULL DEFAULT 0,
  duration_min  INT
);

CREATE TABLE catalog.medication (
  id            BIGSERIAL PRIMARY KEY,
  name          VARCHAR(255) NOT NULL,
  form          VARCHAR(128),
  unit          VARCHAR(32)  NOT NULL,
  stock_qty     NUMERIC(12,2) NOT NULL DEFAULT 0
);

CREATE TABLE catalog.medication_dispense (
  id              BIGSERIAL PRIMARY KEY,
  medication_id   BIGINT        NOT NULL REFERENCES catalog.medication(id),
  patient_id      BIGINT        NOT NULL REFERENCES catalog.patient(id),
  appointment_id  BIGINT,                       -- id приёма из Appointment
  quantity        NUMERIC(12,2) NOT NULL,
  dispensed_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catalog.medication_dispense;
DROP TABLE IF EXISTS catalog.medication;
DROP TABLE IF EXISTS catalog.medical_service;
DROP TABLE IF EXISTS catalog.doctor;
DROP TABLE IF EXISTS catalog.patient;
DROP TABLE IF EXISTS catalog.specialty;
-- +goose StatementEnd
