CREATE TABLE "restaurant_brand" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	owner_profile_id BIGINT NOT NULL,
	
	name TEXT UNIQUE NOT NULL
		CHECK (char_length(name) <= 60),
	description TEXT
		CHECK (char_length(description) <= 500),
		
	promotion_tier INT NOT NULL DEFAULT 0
		CHECK (promotion_tier >= 0 AND promotion_tier <= 3),
		
	logo_url TEXT
		CHECK (char_length(logo_url) <= 2048),

	idempotency_key TEXT UNIQUE,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "restaurant_branch" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	restaurant_brand_id BIGINT NOT NULL,
	location_id BIGINT NOT NULL,
	
	open_time TIME,
	close_time TIME,

	idempotency_key TEXT UNIQUE,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_restaurant_branch_restaurant_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "dish" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	restaurant_brand_id BIGINT NOT NULL,
	
	name TEXT NOT NULL
		CHECK (char_length(name) <= 50),
	description TEXT
		CHECK(char_length(description) <= 1000),
	
	image_url TEXT 	
		CHECK (char_length(image_url) <= 2048),
	
	price BIGINT NOT NULL
		CHECK (price >= 1000000), -- 1 рубль

	idempotency_key TEXT UNIQUE,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_dish_restaurant_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "category" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

	name TEXT NOT NULL UNIQUE,
	-- emoji-иконка для отображения в UI; пустая строка по умолчанию.
	emoji TEXT NOT NULL DEFAULT '',

	idempotency_key TEXT UNIQUE,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "restaurant_brand_category" (
	restaurant_brand_id BIGINT,
	category_id BIGINT,
	PRIMARY KEY (restaurant_brand_id, category_id),

	idempotency_key TEXT UNIQUE,
	
	CONSTRAINT fk_restaurant_brand_category_restaurant_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE RESTRICT,
	
	CONSTRAINT fk_restaurant_brand_category_category
		FOREIGN KEY (category_id)
		REFERENCES "category"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "dish_category" (
	dish_id BIGINT,
	category_id BIGINT,
	PRIMARY KEY (dish_id, category_id),

	idempotency_key TEXT UNIQUE,
	
	CONSTRAINT fk_dish_category_dish
		FOREIGN KEY (dish_id)
		REFERENCES "dish"(id)
		ON DELETE RESTRICT,
	
	CONSTRAINT fk_dish_category_category
		FOREIGN KEY (category_id)
		REFERENCES "category"(id)
		ON DELETE RESTRICT
);
