-- migrate:up
CREATE TYPE race_registrations__status AS ENUM ('registered', 'submitted', 'approved');

ALTER TABLE
  race_registrations
ADD
  COLUMN status race_registrations__status NOT NULL DEFAULT 'registered';

ALTER TABLE
  race_registrations
ALTER COLUMN
  status DROP DEFAULT;

-- migrate:down
ALTER TABLE
  race_registrations DROP COLUMN status;

DROP TYPE race_registrations__status;