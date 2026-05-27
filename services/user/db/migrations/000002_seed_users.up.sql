INSERT INTO "user" (name, email, password_hash, user_role)
VALUES ('Главный Ресторатор', 'owner@foodcourt.fun', '$argon2id$v=19$m=65536,t=3,p=2$VGKX6kuxIaCOIIoboAH4Jg$yUUplj1rz+/73fmnJ9/2nvCUTUKA8QTl7nwPDA/V6Ms', 'owner')
ON CONFLICT (email) DO NOTHING;

INSERT INTO "user" (name, email, password_hash, user_role)
VALUES ('Главный Админ', 'admin@foodcourt.fun', '$argon2id$v=19$m=65536,t=3,p=2$n/Lh7ygwqxPndVkvS2KvjQ$AL+BBFABO0C6ESRoYVJMYH5TdMPCKdWfeABDNXnxsWo', 'admin')
ON CONFLICT (email) DO NOTHING;

-- Привязываем профиль владельца
INSERT INTO "owner_profile" (account_id)
SELECT id FROM "user" WHERE email = 'owner@foodcourt.fun'
ON CONFLICT DO NOTHING;
