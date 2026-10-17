-- ============================================================
-- RadonezhSklad / auth / 0004_roles_v2
-- Роли: admin / warehouse (Склад-Бух) / pts (ПТС)
-- manager -> warehouse, user -> pts
-- ============================================================

UPDATE users SET role = 'warehouse' WHERE role = 'manager';
UPDATE users SET role = 'pts'       WHERE role = 'user';

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'warehouse', 'pts'));

-- Проверим, что не осталось "левых" ролей
UPDATE users SET role = 'pts'
WHERE role NOT IN ('admin', 'warehouse', 'pts');
