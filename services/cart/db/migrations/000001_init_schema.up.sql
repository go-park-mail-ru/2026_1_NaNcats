CREATE TABLE "cart" (
	client_account_id INT PRIMARY KEY,
	restaurant_brand_id INT NOT NULL,
	
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
);

CREATE TABLE "cart_dish" (
	cart_id INT,
	dish_id INT,
	PRIMARY KEY (cart_id, dish_id),
	
	quantity INT NOT NULL
		CHECK (quantity > 0),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_cart_dish_cart
		FOREIGN KEY (cart_id)
		REFERENCES "cart"(client_account_id)
		ON DELETE CASCADE,
);
