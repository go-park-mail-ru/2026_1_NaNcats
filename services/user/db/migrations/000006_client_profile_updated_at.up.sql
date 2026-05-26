-- Lucky Wheel-репозиторий пишет updated_at в client_profile, но колонки в схеме
-- никогда не было. Добавляем идемпотентно, дефолт NOW() заполнит существующие строки.

ALTER TABLE "client_profile"
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
