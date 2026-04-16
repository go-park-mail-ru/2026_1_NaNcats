CREATE TABLE "restaurant_brand" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	owner_profile_id INT NOT NULL,
	
	name TEXT UNIQUE NOT NULL
		CHECK (char_length(name) <= 60),
	description TEXT
		CHECK (char_length(description) <= 500),
		
	promotion_tier INT NOT NULL DEFAULT 0
		CHECK (promotion_tier >= 0 AND promotion_tier <= 3),
		
	logo_url TEXT
		CHECK (char_length(logo_url) <= 2048),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
);

CREATE TABLE "restaurant_branch" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	restaurant_brand_id INT NOT NULL,
	location_id INT NOT NULL,
	
	open_time TIME,
	close_time TIME,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_restaurant_branch_restaurant_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE RESTRICT,
);

CREATE TABLE "dish" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	restaurant_brand_id INT NOT NULL,
	
	name TEXT NOT NULL
		CHECK (char_length(name) <= 50),
	description TEXT
		CHECK(char_length(description) <= 1000),
	
	image_url TEXT 	
		CHECK (char_length(image_url) <= 2048),
	
	price BIGINT NOT NULL
		CHECK (price >= 1000000), -- 1 рубль
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_dish_restaurant_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "category" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	name TEXT NOT NULL UNIQUE,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "restaurant_brand_category" (
	restaurant_brand_id INT,
	category_id INT,
	PRIMARY KEY (restaurant_brand_id, category_id),
	
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
	dish_id INT,
	category_id INT,
	PRIMARY KEY (dish_id, category_id),
	
	CONSTRAINT fk_dish_category_dish
		FOREIGN KEY (dish_id)
		REFERENCES "dish"(id)
		ON DELETE RESTRICT,
	
	CONSTRAINT fk_dish_category_category
		FOREIGN KEY (category_id)
		REFERENCES "category"(id)
		ON DELETE RESTRICT
);
