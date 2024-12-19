CREATE TABLE sessions (
  id SERIAL PRIMARY KEY,
  user_id INT UNIQUE REFERENCES users (id) ON DELETE CASCADE,
  token_hash TEXT UNIQUE NOT NULL
);

-- OR:
CREATE TABLE sessions (
  id SERIAL PRIMARY KEY,
  user_id INT UNIQUE,
  token_hash TEXT UNIQUE NOT NULL
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Update existing table:
ALTER TABLE sessions
  ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

-- Creating indexes:
-- (In Postgres, have to create table first, then index. UNIQUE and PRIMARY KEY indexes created automatically by Postgres.)
CREATE INDEX sessions_token_hash_idx ON sessions (token_hash);

-- Index for multiple columns:
CREATE INDEX dogs_breed_age_idx ON dogs (breed, age);
