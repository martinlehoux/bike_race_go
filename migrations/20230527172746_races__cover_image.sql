-- migrate:up
ALTER TABLE
  races
ADD
  COLUMN cover_image_id UUID;

-- migrate:down
ALTER TABLE
  races DROP COLUMN cover_image_id;