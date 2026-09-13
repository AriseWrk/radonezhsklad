ALTER TABLE users
  ADD COLUMN IF NOT EXISTS last_name   VARCHAR(100) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS first_name  VARCHAR(100) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS middle_name VARCHAR(100) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS phone       VARCHAR(50),
  ADD COLUMN IF NOT EXISTS login       VARCHAR(100),
  ADD COLUMN IF NOT EXISTS description TEXT;

UPDATE users SET
  last_name  = split_part(full_name, ' ', 1),
  first_name = COALESCE(NULLIF(split_part(full_name, ' ', 2), ''), ''),
  middle_name = COALESCE(NULLIF(split_part(full_name, ' ', 3), ''), '')
WHERE last_name = '';

UPDATE users SET login = split_part(email, '@', 1) WHERE login IS NULL OR login = '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_login ON users(login) WHERE login IS NOT NULL;