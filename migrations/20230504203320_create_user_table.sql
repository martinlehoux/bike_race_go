-- migrate:up
CREATE TABLE users (
  id UUID PRIMARY KEY,
  username VARCHAR(255) NOT NULL UNIQUE,
  password_hash BYTEA NOT NULL
);

-- migrate:down
DROP TABLE users;