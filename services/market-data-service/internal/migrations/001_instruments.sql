CREATE TABLE IF NOT EXISTS instruments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    asset_class TEXT NOT NULL CHECK (asset_class IN ('stock', 'crypto', 'forex', 'other')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO instruments (symbol, name, asset_class)
VALUES
    ('AAPL', 'Apple Inc.', 'stock'),
    ('BTC-USD', 'Bitcoin', 'crypto'),
    ('EURUSD', 'Euro / US Dollar', 'forex')
ON CONFLICT (symbol) DO NOTHING;
