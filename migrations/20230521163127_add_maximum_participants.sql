-- migrate:up
ALTER TABLE
  races
ADD
  COLUMN maximum_participants INTEGER NOT NULL DEFAULT 0;

-- migrate:down
ALTER TABLE
  DROP COLUMN maximum_participants;