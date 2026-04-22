INSERT INTO "user" (name, email, password_hash, user_role)
VALUES ('Главный Ресторатор', 'owner@foodcourt.fun', '$2a$10$12345678901234567890123456789012345678901234567890123', 'owner')
ON CONFLICT (email) DO NOTHING;

-- Привязываем профиль владельца
INSERT INTO "owner_profile" (account_id)
SELECT id FROM "user" WHERE email = 'owner@foodcourt.fun'
ON CONFLICT DO NOTHING;
