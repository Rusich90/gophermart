CREATE TABLE IF NOT EXISTS withdrawals (
    order_num VARCHAR(256) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sum NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для ускорения выборок
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);