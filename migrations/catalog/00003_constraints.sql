-- +goose Up
-- +goose StatementBegin
ALTER TABLE catalog.medication
  ADD CONSTRAINT medication_stock_qty_check CHECK (stock_qty >= 0),
  ADD CONSTRAINT medication_name_uniq UNIQUE (name);

ALTER TABLE catalog.medication_dispense
  ADD CONSTRAINT medication_dispense_quantity_check CHECK (quantity > 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE catalog.medication_dispense
  DROP CONSTRAINT IF EXISTS medication_dispense_quantity_check;
ALTER TABLE catalog.medication
  DROP CONSTRAINT IF EXISTS medication_name_uniq,
  DROP CONSTRAINT IF EXISTS medication_stock_qty_check;
-- +goose StatementEnd
