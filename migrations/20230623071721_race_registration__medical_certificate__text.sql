-- migrate:up
ALTER TABLE
  race_registrations
ADD
  COLUMN medical_certificate TEXT;

ALTER TABLE
  race_registrations DROP COLUMN medical_certificate_id;

-- migrate:down
ALTER TABLE
  race_registrations DROP COLUMN medical_certificate;

ALTER TABLE
  race_registrations
ADD
  COLUMN medical_certificate_id UUID NULL;