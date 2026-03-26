CREATE TABLE IF NOT EXISTS payment_transactions (
                                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    transaction_id VARCHAR(255) UNIQUE, -- Внешний ID от платёжной системы
    amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'success', 'failed', 'refunded')),
    payment_method VARCHAR(50), -- 'card', 'sbp', 'cash'
    provider VARCHAR(50), -- 'yookassa', 'cloudpayments', 'mock'
    provider_response JSONB, -- Ответ от платёжной системы
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
    );

CREATE INDEX idx_payment_transactions_booking_id ON payment_transactions(booking_id);
CREATE INDEX idx_payment_transactions_status ON payment_transactions(status);
CREATE INDEX idx_payment_transactions_transaction_id ON payment_transactions(transaction_id);

COMMENT ON TABLE payment_transactions IS 'Платёжные транзакции';