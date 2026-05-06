-- 2. Локации
INSERT INTO "location" (address_text, coordinate) VALUES
('Москва, ул. Арбат, 1', ST_SetSRID(ST_MakePoint(37.599, 55.752), 4326)),
('Москва, Цветной бульвар, 15', ST_SetSRID(ST_MakePoint(37.622, 55.771), 4326)),
('Москва, ул. Тверская, 7', ST_SetSRID(ST_MakePoint(37.611, 55.758), 4326))
ON CONFLICT DO NOTHING;
