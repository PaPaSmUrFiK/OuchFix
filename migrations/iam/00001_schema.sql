-- +goose Up
CREATE TABLE iam.role (
  id            BIGSERIAL PRIMARY KEY,
  name          VARCHAR(64)  NOT NULL UNIQUE,
  description   VARCHAR(255)
);

CREATE TABLE iam.permission (
  id            BIGSERIAL PRIMARY KEY,
  code          VARCHAR(128) NOT NULL UNIQUE,
  description   VARCHAR(255)
);

CREATE TABLE iam.role_permission (
  role_id       BIGINT NOT NULL REFERENCES iam.role(id),
  permission_id BIGINT NOT NULL REFERENCES iam.permission(id),
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE iam.app_user (
  id            BIGSERIAL PRIMARY KEY,
  role_id       BIGINT       NOT NULL REFERENCES iam.role(id),
  email         VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  full_name     VARCHAR(255) NOT NULL,
  phone         VARCHAR(32),
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS iam.app_user;
DROP TABLE IF EXISTS iam.role_permission;
DROP TABLE IF EXISTS iam.permission;
DROP TABLE IF EXISTS iam.role;
