-- Создание баз для RadonezhSklad.
-- Идемпотентно: повторный запуск не падает.

SELECT 'CREATE DATABASE radonezh_product'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'radonezh_product')\gexec

SELECT 'CREATE DATABASE radonezh_warehouse'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'radonezh_warehouse')\gexec

SELECT 'CREATE DATABASE radonezh_order'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'radonezh_order')\gexec

SELECT 'CREATE DATABASE radonezh_audit'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'radonezh_audit')\gexec
