-- Seed payment cards for user_id=3 (first client).
-- external_id uses fake yookassa-style IDs.

INSERT INTO "payment_method" (user_id, external_id, first6, last4, expiry_month, expiry_year, card_type, issuer_name, is_default) VALUES
(3, 'pm-card-sber-001',   '427600', '4532', '09', '2028', 'Visa',       'Сбербанк',    TRUE),
(3, 'pm-card-tink-002',   '521324', '7891', '03', '2027', 'Mastercard', 'Тинькофф',    FALSE),
(3, 'pm-card-alfa-003',   '220070', '2245', '11', '2029', 'Mir',        'Альфа-Банк',  FALSE),
(3, 'pm-card-vtb-004',    '427629', '0019', '06', '2028', 'Visa',       'ВТБ',         FALSE),
(3, 'pm-card-yandex-005', '510621', '3377', '01', '2027', 'Mastercard', 'Яндекс Банк', FALSE),
(3, 'pm-card-ozon-006',   '220054', '8800', '12', '2029', 'Mir',        'Озон Банк',   FALSE)
ON CONFLICT DO NOTHING;
