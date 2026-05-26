INSERT INTO "user" (name, email, password_hash, user_role) VALUES
('Анна Петрова', 'anna@foodcourt.fun',
 '$argon2id$v=19$m=65536,t=3,p=2$AHkE4ETIk466SP0Zn75HMQ$1sddMdeT8DiQglJz6iZgVzDMU/zdlsi8GJMOojMokbY',
 'client')
ON CONFLICT (email) DO NOTHING;

INSERT INTO "client_profile" (account_id, bonus_balance, streak_count, last_order_date,
                              bonus_expires_at, premium_expires_at)
SELECT id, 67000000, 4,
       NOW() - INTERVAL '1 day',
       '2026-04-01'::TIMESTAMPTZ,
       '2026-12-31'::TIMESTAMPTZ
FROM "user" WHERE email = 'anna@foodcourt.fun'
ON CONFLICT (account_id) DO UPDATE SET
    bonus_balance = EXCLUDED.bonus_balance,
    streak_count  = EXCLUDED.streak_count,
    last_order_date = EXCLUDED.last_order_date;
