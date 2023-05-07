-- migrate:up
ALTER TABLE
  races
ADD
  COLUMN start_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT '0001-01-01 00:00:00+00';

ALTER TABLE
  races
ALTER COLUMN
  start_at DROP DEFAULT;

ALTER TABLE
  races
ADD
  COLUMN is_open_for_registration BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE
  races
ALTER COLUMN
  is_open_for_registration DROP DEFAULT;

-- migrate:down
ALTER TABLE
  races DROP COLUMN start_at;

ALTER TABLE
  races DROP COLUMN is_open_for_registration;