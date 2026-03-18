DROP TABLE IF EXISTS "cart_dish" CASCADE;
DROP TABLE IF EXISTS "cart" CASCADE;
DROP TABLE IF EXISTS "dish_category" CASCADE;
DROP TABLE IF EXISTS "restaurant_brand_category" CASCADE;
DROP TABLE IF EXISTS "promocode_category" CASCADE;
DROP TABLE IF EXISTS "promocode_restaurant_brand" CASCADE;
DROP TABLE IF EXISTS "order_dish" CASCADE;
DROP TABLE IF EXISTS "order_review" CASCADE;
DROP TABLE IF EXISTS "order" CASCADE;
DROP TABLE IF EXISTS "client_address" CASCADE;
DROP TABLE IF EXISTS "owner_profile" CASCADE;
DROP TABLE IF EXISTS "courier_profile" CASCADE;
DROP TABLE IF EXISTS "client_profile" CASCADE;
DROP TABLE IF EXISTS "dish" CASCADE;
DROP TABLE IF EXISTS "restaurant_branch" CASCADE;
DROP TABLE IF EXISTS "promocode" CASCADE;
DROP TABLE IF EXISTS "category" CASCADE;
DROP TABLE IF EXISTS "location" CASCADE;
DROP TABLE IF EXISTS "restaurant_brand" CASCADE;
DROP TABLE IF EXISTS "user" CASCADE;

CREATE TABLE "user" (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	phone TEXT UNIQUE
		CHECK (phone ~ '^(\+7|8)[\s\-]?\(?\d{3}\)?[\s\-]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}$'),
		
	name TEXT NOT NULL
		CHECK (char_length(name) >= 1 AND char_length(name) <= 39),
		
	email TEXT NOT NULL UNIQUE
		CHECK (email ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'),
		
	password_hash TEXT NOT NULL
,
		
	user_role TEXT NOT NULL
		CHECK (user_role IN ('client', 'courier', 'owner')),
	
	avatar_url TEXT
		CHECK (char_length(avatar_url) <= 2048),
		
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "restaurant_brand" (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	name TEXT NOT NULL
		CHECK (char_length(name) <= 30),
	description TEXT
		CHECK (char_length(description) <= 500),
		
	promotion_tier INTEGER
		CHECK (promotion_tier >= 0 AND promotion_tier <= 3),
		
	logo_url TEXT
		CHECK (char_length(logo_url) <= 2048),
	banner_url TEXT
		CHECK (char_length(banner_url) <= 2048),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "promocode" (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	code TEXT NOT NULL UNIQUE
		CHECK (char_length(code) >= 2 AND char_length(code) <= 50),
	
	discount_percent INTEGER
		CHECK (discount_percent > 0 AND discount_percent <= 100),
	discount_amount INTEGER
		CHECK (discount_amount > 0),
	
	is_global BOOLEAN DEFAULT FALSE NOT NULL,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
	
	CONSTRAINT check_discount_type
		CHECK (
			(discount_percent IS NOT NULL AND discount_amount IS NULL)
			OR
			(discount_percent IS NULL AND discount_amount IS NOT NULL)
		)
);

CREATE TABLE "category" (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	name TEXT NOT NULL UNIQUE,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "location" (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	address_text TEXT NOT NULL,
	
	latitude NUMERIC NOT NULL,
	longitude NUMERIC NOT NULL,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "client_profile" (
	account_id INTEGER PRIMARY KEY,
	
	bonus_balance INTEGER
		CHECK (bonus_balance >= 0),
	bonus_category_id INT,
	bonus_category_expires_at TIMESTAMP WITH TIME ZONE,
	bonus_expires_at TIMESTAMP WITH TIME ZONE,
	
	streak_count INT
		CHECK (streak_count >= 0),
	
	last_order_date TIMESTAMP WITH TIME ZONE,
	premium_expires_at TIMESTAMP WITH TIME ZONE,
	
	CONSTRAINT fk_client_profile_user
		FOREIGN KEY (account_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE --тут каскадное удаление, чтобы при удалении юзера удалялся и клиент
		
	CONSTRAINT fk_client_profile_category
		FOREIGN KEY (bonus_category_id)
		REFERENCES "category"(id)
		ON DELETE SET NULl
);

CREATE TABLE "courier_profile" (
	account_id INTEGER PRIMARY KEY,
	
	status TEXT NOT NULL
		CHECK (status IN ('offline', 'waiting', 'delivering')),
	
	CONSTRAINT fk_courier_profile_user
		FOREIGN KEY (account_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE --тут каскадное удаление, чтобы при удалении юзера удалялся и курьер
);

CREATE TABLE "owner_profile" (
	account_id INTEGER PRIMARY KEY,
	restaurant_brand_id INTEGER NOT NULL,
	
	CONSTRAINT fk_owner_profile_user
		FOREIGN KEY (account_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE,
	
	CONSTRAINT fk_owner_profile_restaurant_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "restaurant_branch" (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	restaurant_brand_id INTEGER NOT NULL,
	location_id INTEGER NOT NULL,
	
	open_time TIME,
	close_time TIME,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_restaurant_branch_restaurant_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE RESTRICT,
		
	CONSTRAINT fk_restaurant_branch_location
		FOREIGN KEY (location_id)
		REFERENCES "location"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "dish" (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	restaurant_brand_id INTEGER NOT NULL,
	
	name TEXT NOT NULL
		CHECK(char_length(name) <= 50),
	description TEXT
		CHECK(char_length(description) <= 1000),
	
	image_url TEXT 	
		CHECK (char_length(image_url) <= 2048),
	
	price INTEGER NOT NULL
		CHECK (price > 0),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_dish_restaurant_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "client_address" (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	location_id INTEGER NOT NULL,
	client_account_id INTEGER NOT NULL,
	
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
		
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_client_address_location
		FOREIGN KEY (location_id)
		REFERENCES "location"(id)
		ON DELETE RESTRICT,
	
	CONSTRAINT fk_client_address_client_profile
		FOREIGN KEY (client_account_id)
		REFERENCES "client_profile"(account_id)
		ON DELETE CASCADE
);

CREATE TABLE "order" (
	id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	client_account_id INTEGER NOT NULL,
	courier_account_id INTEGER,
	restaurant_branch_id INTEGER NOT NULL,
	client_address_id INTEGER NOT NULL,
	total_cost INTEGER
		CHECK (total_cost > 0)
	promocode_id INTEGER,
	
	status TEXT NOT NULL
		CHECK (status IN ('in_progress', 'waiting', 'delivering', 'finished')),
		
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_order_client_profile
		FOREIGN KEY (client_account_id)
		REFERENCES "client_profile"(account_id)
		ON DELETE RESTRICT,
	
	CONSTRAINT fk_order_courier_profile
		FOREIGN KEY (courier_account_id)
		REFERENCES "courier_profile"(account_id)
		ON DELETE SET NULL,
	
	CONSTRAINT fk_order_restaurant_branch
		FOREIGN KEY (restaurant_branch_id)
		REFERENCES "restaurant_branch"(id)
		ON DELETE RESTRICT,
	
	CONSTRAINT fk_order_client_address
		FOREIGN KEY (client_address_id)
		REFERENCES "client_address"(id)
		ON DELETE RESTRICT,
		
	CONSTRAINT fk_order_promocode
		FOREIGN KEY (promocode_id)
		REFERENCES "promocode"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "order_review" (
	order_id INTEGER PRIMARY KEY,
	restaurant_rating INTEGER
		CHECK (restaurant_rating >= 0 AND restaurant_rating <= 5), -- NULL означает, что отзыв не выставлен
	courier_rating INTEGER
		CHECK (courier_rating >= 0 AND courier_rating <= 5), -- NULL означает, что отзыв не выставлен
	
	client_comment TEXT
		CHECK (char_length(client_comment) <= 255),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_order_review_order
		FOREIGN KEY (order_id)
		REFERENCES "order"(id)
		ON DELETE CASCADE
);

CREATE TABLE "order_dish" (
	order_id INTEGER,
	dish_id INTEGER,
	PRIMARY KEY (order_id, dish_id),
	
	quantity INTEGER NOT NULL
		CHECK (quantity > 0),
	price INTEGER NOT NULL
		CHECK (price > 0),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_order_dish_order
		FOREIGN KEY (order_id)
		REFERENCES "order"(id)
		ON DELETE CASCADE,
		
	CONSTRAINT fk_order_dish_dish
		FOREIGN KEY (dish_id)
		REFERENCES "dish"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "promocode_restaurant_brand" (
	promocode_id INTEGER,
	restaurant_brand_id INTEGER,
	PRIMARY KEY (promocode_id, restaurant_brand_id),
	
	CONSTRAINT fk_promocode_restaurant_brand_promocode
		FOREIGN KEY (promocode_id)
		REFERENCES "promocode"(id)
		ON DELETE RESTRICT,
	
	CONSTRAINT fk_promocode_restaurant_brand_restaurand_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "promocode_category" (
	promocode_id INTEGER,
	category_id INTEGER,
	PRIMARY KEY (promocode_id, category_id),
	
	CONSTRAINT fk_promocode_category_promocode
		FOREIGN KEY (promocode_id)
		REFERENCES "promocode"(id)
		ON DELETE RESTRICT,
	
	CONSTRAINT fk_promocode_category_category
		FOREIGN KEY (category_id)
		REFERENCES "category"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "restaurant_brand_category" (
	restaurant_brand_id INTEGER,
	category_id INTEGER,
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
	dish_id INTEGER,
	category_id INTEGER,
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

CREATE TABLE "cart" (
	client_account_id INTEGER PRIMARY KEY,
	restaurant_brand_id INTEGER NOT NULL,
	
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_cart_client_profile
		FOREIGN KEY (client_account_id)
		REFERENCES "client_profile"(account_id)
		ON DELETE CASCADE,
	
	CONSTRAINT fk_cart_restaurant_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE CASCADE
);

CREATE TABLE "cart_dish" (
	cart_id INTEGER,
	dish_id INTEGER,
	PRIMARY KEY (cart_id, dish_id),
	
	quantity INTEGER NOT NULL
		CHECK (quantity > 0),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_cart_dish_cart
		FOREIGN KEY (cart_id)
		REFERENCES "cart"(client_account_id)
		ON DELETE CASCADE,
	
	CONSTRAINT fk_cart_dish_dish
		FOREIGN KEY (dish_id)
		REFERENCES "dish"(id)
		ON DELETE CASCADE
);

