-- +goose Up
-- +goose StatementBegin
ALTER TABLE appointment.schedule_slot
  ADD CONSTRAINT schedule_slot_status_check CHECK (status IN ('free', 'busy', 'cancelled')),
  ADD CONSTRAINT schedule_slot_period_check CHECK (ends_at > starts_at);

ALTER TABLE appointment.appointment
  ADD CONSTRAINT appointment_status_check
    CHECK (status IN ('scheduled', 'queue', 'completed', 'cancelled', 'no_show')),
  ADD CONSTRAINT appointment_queue_type_check CHECK (queue_type IN ('scheduled', 'walk_in'));

CREATE INDEX schedule_slot_doctor_starts_idx ON appointment.schedule_slot (doctor_id, starts_at);
CREATE INDEX appointment_doctor_scheduled_idx ON appointment.appointment (doctor_id, scheduled_at);

-- Один слот — не более одного действующего приёма.
CREATE UNIQUE INDEX appointment_active_slot_uniq ON appointment.appointment (slot_id)
  WHERE slot_id IS NOT NULL AND status IN ('scheduled', 'completed');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS appointment.appointment_active_slot_uniq;
DROP INDEX IF EXISTS appointment.appointment_doctor_scheduled_idx;
DROP INDEX IF EXISTS appointment.schedule_slot_doctor_starts_idx;
ALTER TABLE appointment.appointment
  DROP CONSTRAINT IF EXISTS appointment_queue_type_check,
  DROP CONSTRAINT IF EXISTS appointment_status_check;
ALTER TABLE appointment.schedule_slot
  DROP CONSTRAINT IF EXISTS schedule_slot_period_check,
  DROP CONSTRAINT IF EXISTS schedule_slot_status_check;
-- +goose StatementEnd
