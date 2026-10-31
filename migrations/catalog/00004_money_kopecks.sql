-- +goose Up
-- +goose StatementBegin
-- Цена услуги — целые копейки вместо NUMERIC(12,2).
ALTER TABLE catalog.medical_service ADD COLUMN price_kopecks BIGINT NOT NULL DEFAULT 0;
UPDATE catalog.medical_service SET price_kopecks = ROUND(price * 100)::BIGINT;
ALTER TABLE catalog.medical_service
  DROP COLUMN price,
  ADD CONSTRAINT medical_service_price_kopecks_check CHECK (price_kopecks >= 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE catalog.medical_service ADD COLUMN price NUMERIC(12,2) NOT NULL DEFAULT 0;
UPDATE catalog.medical_service SET price = price_kopecks / 100.0;
ALTER TABLE catalog.medical_service DROP COLUMN price_kopecks;
-- +goose StatementEnd
