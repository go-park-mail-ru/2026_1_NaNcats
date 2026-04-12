CREATE TYPE order_status AS ENUM('in_progress', 'waiting', 'delivering', 'finished', 'canceled');
CREATE TYPE courier_status AS ENUM('offline', 'waiting', 'delivering');

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE "user" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		
	name TEXT NOT NULL
		CHECK (char_length(name) >= 1 AND char_length(name) <= 39),
		
	email TEXT NOT NULL UNIQUE
		CHECK (email = LOWER(email)),
		
	password_hash TEXT NOT NULL,
		
	user_role TEXT NOT NULL
		CHECK (user_role IN ('client', 'courier', 'owner')),
	
	avatar_url TEXT
		CHECK (char_length(avatar_url) <= 2048),
		
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "owner_profile" (
	account_id INT PRIMARY KEY,
	
	CONSTRAINT fk_owner_profile_user
		FOREIGN KEY (account_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE
);

CREATE TABLE "restaurant_brand" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	owner_profile_id INT NOT NULL,
	
	name TEXT UNIQUE NOT NULL
		CHECK (char_length(name) <= 30),
	description TEXT
		CHECK (char_length(description) <= 500),
		
	promotion_tier INT NOT NULL DEFAULT 0
		CHECK (promotion_tier >= 0 AND promotion_tier <= 3),
		
	logo_url TEXT
		CHECK (char_length(logo_url) <= 2048),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

	CONSTRAINT fk_restaurant_brand_owner_profile
		FOREIGN KEY(owner_profile_id)
		REFERENCES "owner_profile"(account_id)
		ON DELETE RESTRICT
);

CREATE TABLE "category" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	name TEXT NOT NULL UNIQUE,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "location" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	address_text TEXT NOT NULL,
	
	coordinate GEOGRAPHY(Point, 4326) NOT NULL,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "client_profile" (
	account_id INT PRIMARY KEY,
	
	bonus_balance BIGINT DEFAULT 0
		CHECK (bonus_balance >= 0),
	bonus_category_id INT,
	bonus_category_expires_at TIMESTAMP WITH TIME ZONE,
	bonus_expires_at TIMESTAMP WITH TIME ZONE,
	
	streak_count INT DEFAULT 0
		CHECK (streak_count >= 0),
	
	last_order_date TIMESTAMP WITH TIME ZONE,
	premium_expires_at TIMESTAMP WITH TIME ZONE,
	
	CONSTRAINT fk_client_profile_user
		FOREIGN KEY (account_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE, --тут каскадное удаление, чтобы при удалении юзера удалялся и клиент
		
	CONSTRAINT fk_client_profile_category
		FOREIGN KEY (bonus_category_id)
		REFERENCES "category"(id)
		ON DELETE SET NULL
);

CREATE TABLE "courier_profile" (
	account_id INT PRIMARY KEY,
	
	status courier_status NOT NULL,
	
	CONSTRAINT fk_courier_profile_user
		FOREIGN KEY (account_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE --тут каскадное удаление, чтобы при удалении юзера удалялся и курьер
);

CREATE TABLE "promocode" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id INT, -- NULL, если промокод не для конкретного юзера
	
	code TEXT NOT NULL UNIQUE
		CHECK (char_length(code) >= 2 AND char_length(code) <= 50),
	
	discount_percent INT
		CHECK (discount_percent > 0 AND discount_percent <= 100),
	discount_amount BIGINT
		CHECK (discount_amount >= 1000000), -- 1 рубль

	max_uses INT -- NULL будет означать безлимит
		CHECK (max_uses > 0),
	
	min_order_amount BIGINT
		CHECK (min_order_amount >= 1000000), -- 1 рубль
	
	
	is_global BOOL DEFAULT FALSE NOT NULL,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
	
	CONSTRAINT check_discount_type
		CHECK (
			(discount_percent IS NOT NULL AND discount_amount IS NULL)
			OR
			(discount_percent IS NULL AND discount_amount IS NOT NULL)
		),
	
	CONSTRAINT fk_promocode_client_profile
		FOREIGN KEY (user_id)
		REFERENCES "client_profile"(account_id)
		ON DELETE CASCADE
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
		
	CONSTRAINT fk_restaurant_branch_location
		FOREIGN KEY (location_id)
		REFERENCES "location"(id)
		ON DELETE RESTRICT
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

CREATE TABLE "client_address" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	public_id UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
	
	location_id INT NOT NULL,
	client_account_id INT NOT NULL,
	
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
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	public_id UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
	
	client_account_id INT NOT NULL,
	courier_account_id INT,
	restaurant_branch_id INT NOT NULL,
	client_address_id INT NOT NULL,
	total_cost BIGINT
		CHECK (total_cost >= 1000000), -- 1 рубль
	promocode_id INT,

	payment_method_id TEXT,
	yookassa_payment_id TEXT,

	status order_status NOT NULL,
		
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
	order_id INT PRIMARY KEY,
	restaurant_rating INT NOT NULL
		CHECK (restaurant_rating >= 1 AND restaurant_rating <= 5),
	courier_rating INT
		CHECK (courier_rating >= 1 AND courier_rating <= 5),
	
	client_comment TEXT
		CHECK (char_length(client_comment) <= 255),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_order_review_order
		FOREIGN KEY (order_id)
		REFERENCES "order"(id)
		ON DELETE CASCADE
);

CREATE TABLE "order_dish" (
	order_id INT,
	dish_id INT,
	PRIMARY KEY (order_id, dish_id),
	
	quantity INT NOT NULL
		CHECK (quantity > 0),
	price BIGINT NOT NULL
		CHECK (price >= 1000000),
	
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
	promocode_id INT,
	restaurant_brand_id INT,
	PRIMARY KEY (promocode_id, restaurant_brand_id),
	
	CONSTRAINT fk_promocode_restaurant_brand_promocode
		FOREIGN KEY (promocode_id)
		REFERENCES "promocode"(id)
		ON DELETE RESTRICT,
	
	CONSTRAINT fk_promocode_restaurant_brand_restaurant_brand
		FOREIGN KEY (restaurant_brand_id)
		REFERENCES "restaurant_brand"(id)
		ON DELETE RESTRICT
);

CREATE TABLE "promocode_category" (
	promocode_id INT,
	category_id INT,
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

CREATE TABLE "promocode_usage" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	promocode_id INT NOT NULL,
	order_id INT,
	
	client_account_id INT,
	UNIQUE (promocode_id, client_account_id),
	
	used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_promocode_usage_promocode
		FOREIGN KEY (promocode_id)
		REFERENCES "promocode"(id)
		ON DELETE CASCADE,
		
	CONSTRAINT fk_promocode_usage_order
		FOREIGN KEY (order_id)
		REFERENCES  "order"(id)
		ON DELETE SET NULL,
		
	CONSTRAINT fk_promocode_usage_client_profile
		FOREIGN KEY (client_account_id)
		REFERENCES "client_profile"(account_id)
		ON DELETE SET NULL
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

CREATE TABLE "cart" (
	client_account_id INT PRIMARY KEY,
	restaurant_brand_id INT NOT NULL,
	
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
	
	CONSTRAINT fk_cart_dish_dish
		FOREIGN KEY (dish_id)
		REFERENCES "dish"(id)
		ON DELETE CASCADE
);

CREATE TABLE "wordle_word" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	word TEXT NOT NULL UNIQUE
		CHECK (char_length(word) = 5 AND word = LOWER(word)),
		
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "wordle_daily" (
	word_of_day DATE PRIMARY KEY,
	word_id	INT NOT NULL,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_wordle_daily_wordle_word
		FOREIGN KEY (word_id)
		REFERENCES "wordle_word"(id)
		ON DELETE RESTRICT
);

-- игровая сессия юзера за какой-то день
CREATE TABLE "wordle_game" (
	user_id INT NOT NULL,
	game_date DATE NOT NULL,
	PRIMARY KEY(user_id, game_date),
	
	solved BOOL NOT NULL DEFAULT FALSE,
	attempt INT NOT NULL DEFAULT 0
		CHECK (attempt >= 0 AND attempt <= 6),
		
	finished_at TIMESTAMP WITH TIME ZONE, -- NULL пока игра не завершена
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_wordle_game_user
		FOREIGN KEY (user_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE,
	
	CONSTRAINT fk_wordle_game_wordle_daily
		FOREIGN KEY (game_date)
		REFERENCES "wordle_daily"(word_of_day)
		ON DELETE RESTRICT
);

-- история угадываний
CREATE TABLE "wordle_guess" (
	user_id INT NOT NULL,
	guess_date DATE NOT NULL,
	attempt_num INT NOT NULL
		CHECK (attempt_num >= 1 AND attempt_num <= 6),
	PRIMARY key(user_id, guess_date, attempt_num),
		
	word TEXT NOT NULL
		CHECK (char_length(word) = 5 AND word = LOWER(word)),
		
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_wordle_guess_wordle_game
		FOREIGN KEY (user_id, guess_date)
		REFERENCES "wordle_game"(user_id, game_date)
		ON DELETE CASCADE 
);

CREATE TABLE "wordle_streak" (
	user_id INT PRIMARY KEY,
	
	current_streak INT NOT NULL DEFAULT 0,
	last_played DATE, -- NULL если ещё не играл вообще
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_wordle_streak_client_profile
		FOREIGN KEY (user_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE
);

-- Это таблица для будущей игры, которую будем реализовывать
-- на странице ожидания заказа. Я её добавил, чтобы потом не
-- геморно было это делать
CREATE TABLE "game_session" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id INT NOT NULL,
	order_id INT,
	
	game_type TEXT NOT NULL, -- 'fruit_ninja', 'blockblast' и т.п., потом придумаем это
	score INT NOT NULL 
		CHECK (score >= 0),
	
	played_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL, -- то же, что и created_at
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_game_session_user
		FOREIGN KEY (user_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE,
	
	CONSTRAINT fk_game_session_order
		FOREIGN KEY (order_id)
		REFERENCES "order"(id)
		ON DELETE SET NULL
);

CREATE TABLE "achievement" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	code TEXT NOT NULL UNIQUE,
	
	name TEXT NOT NULL
		CHECK (char_length(name) <= 50),
	description TEXT
		CHECK (char_length(description) <= 100),
		
	icon_url TEXT
		CHECK (char_length(icon_url) <= 2048),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "user_achievement" (
	achievement_id INT,
	user_id INT,
	PRIMARY KEY(achievement_id, user_id),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_user_achievement_achievement
		FOREIGN KEY (achievement_id)
		REFERENCES "achievement"(id)
		ON DELETE CASCADE,
	
	CONSTRAINT fk_user_achievement_user
		FOREIGN KEY (user_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE
);

CREATE TABLE "payment_method" (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id INT NOT NULL,

	external_id TEXT NOT NULL UNIQUE,

	last4 TEXT NOT NULL
		CHECK (char_length(last4) = 4),
	card_type TEXT NOT NULL,
	issuer_name TEXT,

	is_default BOOLEAN DEFAULT FALSE NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

	CONSTRAINT fk_saved_card_user
		FOREIGN KEY (user_id)
		REFERENCES "user"(id)
		ON DELETE CASCADE
);
