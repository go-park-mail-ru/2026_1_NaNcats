CREATE TYPE order_status AS ENUM('created', 'cart_locked', 'payment_ready', 'paid', 'in_progress', 'waiting', 'delivering', 'finished', 'cancelled', 'failed');
--	 created       : создана запись в БД, запуск Саги
--	 cart_locked   : заблокировано изменение корзины
--	 payment_ready : создана ссылка на оплату, ждем оплаты
--	 paid          : юзер оплатил
--	 in_progress   : ресторан начал готовку
--	 waiting       : заказ готов, ждем пока курьер его заберет
--	 delivering    : курьер доставляет
--	 finished      : курьер доставил заказ, сфоткал его/передал пользователю
--	 cancelled     : заказ отменен (пользователь может отменить вплоть до статуса 'in_progress', дальше только через тех поддержку)
-- 	 failed          : по какой-то причине заказ не создался
CREATE TYPE split_status AS ENUM('pending', 'paid', 'failed', 'cancelled');

CREATE TABLE "order" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	public_id UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
	
	admin_account_id BIGINT NOT NULL, 
	
	courier_account_id BIGINT,
	restaurant_branch_id BIGINT NOT NULL,
	restaurant_brand_id BIGINT NOT NULL,
	client_address_id TEXT NOT NULL,
	
	total_cost BIGINT
		CHECK (total_cost >= 1000000), -- 1 рубль
	promocode_id BIGINT,
	discount_amount BIGINT DEFAULT 0 NOT NULL,
	promocode_code TEXT,

	restaurant_name TEXT NOT NULL,

	status order_status NOT NULL,
		
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
	
	CONSTRAINT fk_order_review_order
		FOREIGN KEY (order_id)
		REFERENCES "order"(id)
		ON DELETE CASCADE
);

CREATE TABLE "order_dish" (
	order_id BIGINT,
	dish_id BIGINT,
	-- owner_user_id входит в первичный ключ: одно блюдо у разных участников
	-- совместного заказа хранится отдельными строками. 0 = позиция ничья.
	owner_user_id BIGINT NOT NULL DEFAULT 0,
	PRIMARY KEY (order_id, dish_id, owner_user_id),

	dish_name TEXT NOT NULL,

	quantity INT NOT NULL CHECK (quantity > 0),
	price BIGINT NOT NULL CHECK (price >= 1000000),

	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_order_dish_order
		FOREIGN KEY (order_id)
		REFERENCES "order"(id)
		ON DELETE CASCADE
);

CREATE TABLE "order_split" (
	id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
	order_id BIGINT NOT NULL,
	user_id BIGINT NOT NULL, -- Плательщик
	
	base_amount BIGINT DEFAULT 0 NOT NULL,
	discount_amount BIGINT DEFAULT 0 NOT NULL,
	amount BIGINT NOT NULL CHECK (amount > 0),
	status split_status DEFAULT 'pending' NOT NULL,
	
	payment_method_id TEXT,
	yookassa_payment_id TEXT UNIQUE,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

	CONSTRAINT fk_order_split_order
		FOREIGN KEY (order_id)
		REFERENCES "order"(id)
		ON DELETE CASCADE
);

CREATE TABLE "idempotency_records" (
    user_id BIGINT NOT NULL,
    idempotency_key TEXT NOT NULL,
    grpc_method TEXT NOT NULL,
    response_payload JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    PRIMARY KEY (user_id, idempotency_key)
);

CREATE TABLE "promocode" (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code TEXT NOT NULL UNIQUE
        CHECK (char_length(code) >= 2 AND char_length(code) <= 50),
    title TEXT NOT NULL DEFAULT '',

    discount_percent INT
        CHECK (discount_percent > 0 AND discount_percent <= 100),
    discount_amount BIGINT
        CHECK (discount_amount > 0), 

    max_uses INT,
    current_uses INT DEFAULT 0 NOT NULL,
    min_order_amount BIGINT
        CHECK (min_order_amount > 0), 
    
    user_id BIGINT,
    restaurant_brand_id BIGINT,
    
    is_global BOOL DEFAULT FALSE NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    CONSTRAINT check_discount_type
        CHECK (
            (discount_percent IS NOT NULL AND discount_amount IS NULL)
            OR
            (discount_percent IS NULL AND discount_amount IS NOT NULL)
        )
);

CREATE TABLE "promocode_usage" (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    promocode_id BIGINT NOT NULL,
    -- order_id может быть NULL: использование промокода фиксируется и тогда,
    -- когда заказ ещё не сопоставлен, а при удалении заказа просто обнуляется.
    order_id BIGINT,
    user_id BIGINT NOT NULL,

    UNIQUE (promocode_id, user_id),

    used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

    CONSTRAINT fk_promocode_usage_promocode
        FOREIGN KEY (promocode_id)
        REFERENCES "promocode"(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_promocode_usage_order
        FOREIGN KEY (order_id)
        REFERENCES "order"(id)
        ON DELETE SET NULL
);

-- Промокоды, "сохранённые" пользователем в профиль.
CREATE TABLE "user_promocode" (
    user_id BIGINT NOT NULL,
    promocode_id BIGINT NOT NULL,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (user_id, promocode_id),

    CONSTRAINT fk_user_promocode_promocode
        FOREIGN KEY (promocode_id)
        REFERENCES "promocode"(id)
        ON DELETE CASCADE
);
