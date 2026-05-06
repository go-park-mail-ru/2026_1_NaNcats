INSERT INTO "user" (name, email, password_hash, user_role)
VALUES ('Главный Ресторатор', 'owner@foodcourt.fun', '$argon2id$v=19$m=65536,t=3,p=2$AHkE4ETIk466SP0Zn75HMQ$1sddMdeT8DiQglJz6iZgVzDMU/zdlsi8GJMOojMokbY', 'owner')
ON CONFLICT (email) DO NOTHING;

INSERT INTO "user" (name, email, password_hash, user_role)
VALUES ('Главный Админ', 'admin@foodcourt.fun', '$argon2id$v=19$m=65536,t=3,p=2$GfIMbyzKZm487Wdm9geXGg$Ubfq8BXJz5+B6rNOo1B/19rAwHUuGFdb+8c8knj83Qs', 'admin')
ON CONFLICT (email) DO NOTHING;

-- Привязываем профиль владельца
INSERT INTO "owner_profile" (account_id)
SELECT id FROM "user" WHERE email = 'owner@foodcourt.fun'
ON CONFLICT DO NOTHING;
