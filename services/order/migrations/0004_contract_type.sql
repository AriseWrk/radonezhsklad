ALTER TABLE contracts
  ADD COLUMN IF NOT EXISTS contract_type VARCHAR(100) NOT NULL DEFAULT 'Договор купли-продажи';