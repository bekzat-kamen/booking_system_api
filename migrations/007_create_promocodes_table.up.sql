CREATE TABLE IF NOT EXISTS promocodes (
                                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    discount_type VARCHAR(20) NOT NULL DEFAULT 'percent' CHECK (discount_type IN ('percent', 'fixed')),
    discount_value DECIMAL(10, 2) NOT NULL DEFAULT 0,
    min_amount DECIMAL(10, 2) DEFAULT 0,
    max_uses INTEGER DEFAULT 0,
    used_count INTEGER DEFAULT 0,
    valid_from TIMESTAMPTZ DEFAULT NOW(),
    valid_until TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
    );

CREATE INDEX idx_promocodes_code ON promocodes(code);
CREATE INDEX idx_promocodes_is_active ON promocodes(is_active);
CREATE INDEX idx_promocodes_valid_until ON promocodes(valid_until);
CREATE INDEX idx_promocodes_created_by ON promocodes(created_by);

COMMENT ON TABLE promocodes is 'Промокоды для скидок на бронирования';