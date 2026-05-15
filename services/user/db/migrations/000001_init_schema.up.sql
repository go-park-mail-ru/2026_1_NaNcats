CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE courier_status AS ENUM('offline', 'waiting', 'delivering');

CREATE TABLE "user" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	public_id UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
	
	name TEXT NOT NULL
		CHECK (char_length(name) >= 1 AND char_length(name) <= 39),
	
	email TEXT NOT NULL UNIQUE
		CHECK (email = LOWER(email)),
	
	password_hash TEXT NOT NULL,
	
	user_role TEXT NOT NULL
		CHECK (user_role IN ('client', 'courier', 'owner', 'admin', 'support')),
	
	avatar_url TEXT
		CHECK (char_length(avatar_url) <= 2048),
	
	idempotency_key TEXT UNIQUE,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "owner_profile" (
	account_id BIGINT PRIMARY KEY,

	idempotency_key TEXT,
	
	CONSTRAINT fk_owner_profile_user
		FOREIGN KEY (account_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE
);

CREATE TABLE "client_profile" (
	account_id BIGINT PRIMARY KEY,
	
	bonus_balance BIGINT DEFAULT 0
		CHECK (bonus_balance >= 0),
	bonus_category_id BIGINT,
	bonus_category_expires_at TIMESTAMP WITH TIME ZONE,
	bonus_expires_at TIMESTAMP WITH TIME ZONE,
	
	streak_count INT DEFAULT 0
		CHECK (streak_count >= 0),
	
	last_order_date TIMESTAMP WITH TIME ZONE,
	premium_expires_at TIMESTAMP WITH TIME ZONE,

	idempotency_key TEXT UNIQUE,
	
	CONSTRAINT fk_client_profile_user
		FOREIGN KEY (account_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE
);

CREATE TABLE "courier_profile" (
	account_id BIGINT PRIMARY KEY,
	
	status courier_status NOT NULL,

	idempotency_key TEXT UNIQUE,
	
	CONSTRAINT fk_courier_profile_user
		FOREIGN KEY (account_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE
);
