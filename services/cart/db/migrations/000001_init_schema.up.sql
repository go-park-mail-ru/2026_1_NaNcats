CREATE TYPE cart_status AS ENUM('active', 'locked');

CREATE TABLE "cart" (
	client_account_id BIGINT PRIMARY KEY,
	restaurant_brand_id BIGINT NOT NULL,

	status cart_status NOT NULL DEFAULT 'active',
	
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "cart_dish" (
	cart_id BIGINT,
	dish_id BIGINT,
	PRIMARY KEY (cart_id, dish_id),
	
	quantity INT NOT NULL
		CHECK (quantity > 0),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_cart_dish_cart
		FOREIGN KEY (cart_id)
		REFERENCES "cart"(client_account_id)
		ON DELETE CASCADE
);
