-- migrate:up
CREATE TABLE races (
  id UUID PRIMARY KEY,
  name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE race_organizers (
  race_id UUID NOT NULL REFERENCES races(id),
  user_id UUID NOT NULL REFERENCES users(id),
  PRIMARY KEY (race_id, user_id)
);

-- migrate:down
DROP TABLE races;

DROP TABLE race_organizers;