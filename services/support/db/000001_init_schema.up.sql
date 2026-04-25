CREATE TYPE support_event_type AS ENUM ('ticket_created', 'message', 'status_changed', 'reassigned', 'rated');
CREATE TYPE support_author_role AS ENUM ('user', 'support', 'system');
CREATE TYPE support_agent_status AS ENUM ('online', 'offline');

CREATE TABLE "support_category" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	name TEXT NOT NULL UNIQUE
		CHECK (char_length(name) >= 1 AND char_length(name) <= 100),
	description TEXT
		CHECK (char_length(description) <= 500),
		
	-- К какой линии по умолчанию относится категория (1 или 2)
	default_line INT DEFAULT 1 NOT NULL
		CHECK (default_line >= 1 AND default_line <= 3),

	is_active BOOLEAN DEFAULT TRUE NOT NULL,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);


CREATE TABLE "support_template" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	
	name TEXT NOT NULL
		CHECK (char_length(name) >= 1 AND char_length(name) <= 100),
	content TEXT NOT NULL
		CHECK (char_length(content) >= 1),
		
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "support_agent_profile" (
	account_id BIGINT PRIMARY KEY,
	
	status support_agent_status NOT NULL DEFAULT 'offline',
	support_line INT NOT NULL DEFAULT 1
		CHECK (support_line >= 1 AND support_line <= 3),
		
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE "support_ticket" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	public_id UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
	
	client_account_id BIGINT,
	guest_id UUID,
	contact_email TEXT NOT NULL
		CHECK (contact_email = LOWER(contact_email)),
	
	category_id BIGINT NOT NULL,
	current_status TEXT NOT NULL DEFAULT 'open',
	
	support_line INT NOT NULL DEFAULT 1,
	assignee_id BIGINT,
	
	resolution_rating INT
		CHECK (resolution_rating >= 1 AND resolution_rating <= 5),
	
	client_meta JSONB,
	creator_role support_author_role NOT NULL,
	idempotency_key TEXT UNIQUE,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,

	CONSTRAINT fk_ticket_category
		FOREIGN KEY (category_id)
		REFERENCES "support_category"(id)
		ON DELETE RESTRICT,
		
	CONSTRAINT fk_ticket_assignee
		FOREIGN KEY (assignee_id)
		REFERENCES "support_agent_profile"(account_id)
		ON DELETE SET NULL
);

CREATE TABLE "support_event" (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	ticket_id BIGINT NOT NULL,
	
	author_account_id BIGINT,
	author_role support_author_role NOT NULL,
	
	event_type support_event_type NOT NULL,
	payload JSONB NOT NULL,
	
	idempotency_key TEXT UNIQUE,
	
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
	
	CONSTRAINT fk_event_ticket
		FOREIGN KEY (ticket_id)
		REFERENCES "support_ticket"(id)
		ON DELETE CASCADE
);

CREATE INDEX idx_support_event_ticket_id ON "support_event"(ticket_id);
CREATE INDEX idx_support_ticket_client_id ON "support_ticket"(client_account_id) WHERE client_account_id IS NOT NULL;
CREATE INDEX idx_support_ticket_guest_id ON "support_ticket"(guest_id) WHERE guest_id IS NOT NULL;
CREATE INDEX idx_support_agent_online ON "support_agent_profile"(status) WHERE status = 'online';

INSERT INTO "support_category" (name, description, default_line) VALUES 
('Баг/техническая ошибка', 'Ошибки на сайте или в приложении', 2),
('Вопрос по заказу', 'Где курьер, отмена заказа, изменения', 1),
('Жалоба на доставку/продукт/ресторан', 'Невкусно, холодное, недовес', 1),
('Предложение', 'Идеи по улучшению сервиса', 1)
ON CONFLICT (name) DO UPDATE SET default_line = EXCLUDED.default_line;