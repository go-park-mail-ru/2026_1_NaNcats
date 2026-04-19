CREATE TYPE order_status AS ENUM('in_progress', 'waiting', 'delivering', 'finished', 'canceled');

CREATE TABLE "order" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	public_id UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
	
	client_account_id BIGINT NOT NULL,
	courier_account_id BIGINT,
	restaurant_branch_id BIGINT NOT NULL,
	client_address_id TEXT NOT NULL,
	total_cost BIGINT
		CHECK (total_cost >= 1000000), -- 1 рубль
	promocode_id BIGINT,
	restaurant_name TEXT NOT NULL,

	payment_method_id TEXT,
	yookassa_payment_id TEXT,

	status order_status NOT NULL,

	idempotency_key TEXT UNIQUE,
		
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "order_review" (
	order_id BIGINT PRIMARY KEY,
	restaurant_rating INT NOT NULL
		CHECK (restaurant_rating >= 1 AND restaurant_rating <= 5),
	courier_rating INT
		CHECK (courier_rating >= 1 AND courier_rating <= 5),
	
	client_comment TEXT
		CHECK (char_length(client_comment) <= 255),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

	idempotency_key TEXT UNIQUE,
	
	CONSTRAINT fk_order_review_order
		FOREIGN KEY (order_id)
		REFERENCES "order"(id)
		ON DELETE CASCADE
);

CREATE TABLE "order_dish" (
	order_id BIGINT,
	dish_id BIGINT,
	PRIMARY KEY (order_id, dish_id),
	
	quantity INT NOT NULL
		CHECK (quantity > 0),
	price BIGINT NOT NULL
		CHECK (price >= 1000000),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

	idempotency_key TEXT UNIQUE,
	
	CONSTRAINT fk_order_dish_order
		FOREIGN KEY (order_id)
		REFERENCES "order"(id)
		ON DELETE CASCADE
);
