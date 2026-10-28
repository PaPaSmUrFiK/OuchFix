CREATE TABLE payment.payment (
  id              BIGSERIAL PRIMARY KEY,
  appointment_id  BIGINT         NOT NULL,     -- Appointment
  patient_id      BIGINT         NOT NULL,     -- Catalog
  service_id      BIGINT,                       -- Catalog
  amount          NUMERIC(12,2)  NOT NULL,
  method          VARCHAR(32)    NOT NULL,      -- cash | card | online
  status          VARCHAR(32)    NOT NULL,      -- pending | paid | failed
  created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
  paid_at         TIMESTAMPTZ
);
