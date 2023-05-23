-- migrate:up
ALTER TABLE
  race_registered_users RENAME TO race_registrations;

-- migrate:down
ALTER TABLE
  race_registrations RENAME TO race_registered_users;