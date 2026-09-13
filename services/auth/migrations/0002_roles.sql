-- Обновляем существующего admin@radonezh.local до роли admin
UPDATE users SET role = 'admin' WHERE email = 'admin@radonezh.local' AND role <> 'admin';

-- Проверочный constraint на допустимые роли
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'manager', 'warehouse', 'user'));