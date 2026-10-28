-- Выполняется один раз под администратором ouchfix при первом запуске контейнера
-- (или вручную: psql -U ouchfix -d ouchfix -f docker/postgres/init.sql).
-- У каждого сервиса свой пользователь, он владеет только своей схемой; таблицы и
-- сид создают миграции сервиса при старте.

REVOKE ALL ON SCHEMA public FROM PUBLIC;

CREATE ROLE ouchfix_iam          LOGIN PASSWORD 'ouchfix_iam';
CREATE ROLE ouchfix_catalog      LOGIN PASSWORD 'ouchfix_catalog';
CREATE ROLE ouchfix_appointment  LOGIN PASSWORD 'ouchfix_appointment';
CREATE ROLE ouchfix_payment      LOGIN PASSWORD 'ouchfix_payment';
CREATE ROLE ouchfix_notification LOGIN PASSWORD 'ouchfix_notification';

CREATE SCHEMA iam          AUTHORIZATION ouchfix_iam;
CREATE SCHEMA catalog      AUTHORIZATION ouchfix_catalog;
CREATE SCHEMA appointment  AUTHORIZATION ouchfix_appointment;
CREATE SCHEMA payment      AUTHORIZATION ouchfix_payment;
CREATE SCHEMA notification AUTHORIZATION ouchfix_notification;

ALTER ROLE ouchfix_iam          SET search_path = iam;
ALTER ROLE ouchfix_catalog      SET search_path = catalog;
ALTER ROLE ouchfix_appointment  SET search_path = appointment;
ALTER ROLE ouchfix_payment      SET search_path = payment;
ALTER ROLE ouchfix_notification SET search_path = notification;
