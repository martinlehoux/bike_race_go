-- migrate:up
ALTER TABLE
  race_registrations
ADD
  COLUMN medical_certificate_id UUID NULL;

ALTER TABLE
  race_registrations
ADD
  COLUMN is_medical_certificate_approved BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE
  race_registrations
ALTER COLUMN
  is_medical_certificate_approved DROP DEFAULT;

-- migrate:down
ALTER TABLE
  race_registrations DROP COLUMN medical_certificate_id;

ALTER TABLE
  race_registrations DROP COLUMN is_medical_certificate_approved;