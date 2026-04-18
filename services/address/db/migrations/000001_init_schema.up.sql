CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE "location" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	address_text TEXT NOT NULL,
	
	coordinate GEOGRAPHY(Point, 4326) NOT NULL,

	idempotency_key TEXT,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "client_address" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	public_id UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
	
	location_id BIGINT NOT NULL,
	client_account_id BIGINT NOT NULL,
	
	apartment TEXT
		CHECK (char_length(apartment) <= 30),
	entrance TEXT
		CHECK (char_length(entrance) <= 30),
	floor_level TEXT
		CHECK (char_length(floor_level) <= 30),
	door_code TEXT
		CHECK (char_length(door_code) <= 30),
	courier_comment TEXT
		CHECK (char_length(courier_comment) <= 255),
	label TEXT
		CHECK (char_length(label) <= 60),

	is_active BOOLEAN DEFAULT true NOT NULL,

	idempotency_key TEXT,
		
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_client_address_location
		FOREIGN KEY (location_id)
		REFERENCES "location"(id)
		ON DELETE RESTRICT,
);
