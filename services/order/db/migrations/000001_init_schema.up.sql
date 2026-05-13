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
	PRIMARY KEY (order_id, dish_id),

	dish_name TEXT NOT NULL,
	
	quantity INT NOT NULL CHECK (quantity > 0),
	price BIGINT NOT NULL CHECK (price >= 1000000),
	
	owner_user_id BIGINT, 
	
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
