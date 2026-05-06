CREATE TABLE "payment_method" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id BIGINT NOT NULL,

	external_id TEXT NOT NULL UNIQUE,

	first6 TEXT NOT NULL 
        CHECK (char_length(first6) = 6),
    last4 TEXT NOT NULL
        CHECK (char_length(last4) = 4),
    expiry_month TEXT NOT NULL
        CHECK (char_length(expiry_month) = 2),
    expiry_year TEXT NOT NULL
        CHECK (char_length(expiry_year) = 4),
	
	card_type TEXT NOT NULL,
	issuer_name TEXT,

	is_default BOOLEAN DEFAULT FALSE NOT NULL,

	idempotency_key TEXT UNIQUE,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

	CONSTRAINT payment_method_user_id_external_id_key UNIQUE (user_id, external_id)
);
