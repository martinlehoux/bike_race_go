-- migrate:up
ALTER TABLE
  "users"
ADD
  COLUMN "language" VARCHAR(10) NOT NULL DEFAULT 'en';

ALTER TABLE
  "users"
ALTER COLUMN
  "language" DROP DEFAULT;

-- migrate:down
ALTER TABLE
  "users" DROP COLUMN "language";