DO $$ BEGIN
    CREATE TYPE support_event_type AS ENUM ('TICKET_CREATED', 'MESSAGE', 'STATUS_CHANGED', 'REASSIGNED', 'RATED');
    CREATE TYPE support_author_role AS ENUM ('USER', 'SUPPORT', 'SYSTEM');
    CREATE TYPE support_agent_status AS ENUM ('ONLINE', 'OFFLINE');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS support_categories (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_templates (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_agent_profiles (
    account_id BIGINT PRIMARY KEY,
    status support_agent_status NOT NULL DEFAULT 'OFFLINE',
    support_line INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_tickets (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    
    client_id BIGINT NULL,
    guest_id UUID NULL,
    contact_email TEXT NOT NULL,
    
    category_id INT NOT NULL REFERENCES support_categories(id),
    current_status TEXT NOT NULL DEFAULT 'OPEN',
    
    support_line INT NOT NULL DEFAULT 1,
    assignee_id BIGINT NULL REFERENCES support_agent_profiles(account_id) ON DELETE SET NULL,
    resolution_rating INT NULL CHECK (resolution_rating >= 1 AND resolution_rating <= 5),
    
    client_meta JSONB NULL,
    creator_role support_author_role NOT NULL, 
    idempotency_key TEXT UNIQUE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    
    author_id BIGINT NULL,
    author_role support_author_role NOT NULL,
    
    event_type support_event_type NOT NULL,
    payload JSONB NOT NULL,
    
    idempotency_key TEXT UNIQUE,        
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_support_events_ticket_id ON support_events(ticket_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_client_id ON support_tickets(client_id) WHERE client_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_support_tickets_guest_id ON support_tickets(guest_id) WHERE guest_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_support_tickets_status ON support_tickets(current_status);
CREATE INDEX IF NOT EXISTS idx_support_tickets_assignee ON support_tickets(assignee_id) WHERE assignee_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_support_agents_status ON support_agent_profiles(status) WHERE status = 'ONLINE';

INSERT INTO support_categories (name, description) VALUES 
('Баг / Техническая ошибка', 'Ошибки на сайте или в приложении'),
('Вопрос по заказу', 'Где курьер, отмена заказа, изменения'),
('Жалоба на продукт/ресторан', 'Невкусно, холодное, недовес'),
('Предложение', 'Идеи по улучшению сервиса')
ON CONFLICT (name) DO NOTHING;