CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE cart_mode AS ENUM('solo', 'shared');
CREATE TYPE cart_status AS ENUM('active', 'locked');

CREATE TABLE "cart" (
	cart_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id BIGINT NOT NULL, -- Создатель комнаты, имеет права на кик и смену оунеров
    restaurant_brand_id BIGINT NOT NULL,

	status cart_status NOT NULL DEFAULT 'active',
	mode cart_mode NOT NULL DEFAULT 'solo',
	
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "cart_dish" (
	cart_id UUID,
    dish_id BIGINT,
    PRIMARY KEY (cart_id, dish_id),
	
	owner_user_id BIGINT, 
    
    quantity INT NOT NULL CHECK (quantity > 0),
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_cart_dish_cart
		FOREIGN KEY (cart_id)
		REFERENCES "cart"(cart_id)
		ON DELETE CASCADE
);

CREATE TABLE "cart_member" (
    cart_id UUID REFERENCES "cart"(cart_id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (cart_id, user_id)
);

CREATE TABLE "cart_invite" (
    token TEXT PRIMARY KEY, -- хэш для url
    cart_id UUID NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

	CONSTRAINT fk_cart_invite_cart
		FOREIGN KEY (cart_id)
		REFERENCES "cart"(cart_id)
		ON DELETE CASCADE
);

CREATE TABLE "outbox_events" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);
