-- migrate:up
CREATE TABLE race_registered_users (
  race_id UUID NOT NULL REFERENCES races(id),
  user_id UUID NOT NULL REFERENCES users(id),
  registered_at TIMESTAMP WITH TIME ZONE NOT NULL,
  PRIMARY KEY (race_id, user_id)
);

-- migrate:down
DROP TABLE race_registered_users;