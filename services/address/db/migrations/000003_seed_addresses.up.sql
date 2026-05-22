-- Seed client addresses for user id=3 (first client after owner+admin).
-- Adds extra locations and 6 client_address rows.

-- Extra Moscow locations
INSERT INTO "location" (address_text, coordinate) VALUES
('Москва, Пресненская наб., 12', ST_SetSRID(ST_MakePoint(37.5374, 55.7494), 4326)),
('Москва, Ленинский проспект, 38', ST_SetSRID(ST_MakePoint(37.5689, 55.7011), 4326)),
('Москва, Кутузовский пр-т, 24', ST_SetSRID(ST_MakePoint(37.5412, 55.7405), 4326)),
('Москва, ул. Бауманская, 11', ST_SetSRID(ST_MakePoint(37.6788, 55.7725), 4326)),
('Москва, Новый Арбат, 21', ST_SetSRID(ST_MakePoint(37.5867, 55.7531), 4326)),
('Москва, ул. Маросейка, 9', ST_SetSRID(ST_MakePoint(37.6367, 55.7566), 4326))
ON CONFLICT DO NOTHING;

-- Client addresses (labels: Дом, Работа, Офис, Спортзал, Дача, Друзья)
INSERT INTO "client_address" (location_id, client_account_id, label, apartment, entrance, floor_level, door_code) VALUES
(1, 3, 'Дом',      '42',  '2', '5',  '1234'),
(2, 3, 'Работа',   '301', '1', '3',  '4567'),
(4, 3, 'Офис',     '1508','А', '15', NULL),
(5, 3, 'Университет', NULL, NULL, NULL, NULL),
(6, 3, 'Спортзал', '5',   '3', '1',  '9090'),
(7, 3, 'Родители', '128', '4', '12', '5511'),
(8, 3, 'Друзья',   '77',  '1', '8',  NULL),
(9, 3, 'Бабушка',  '14',  '2', '3',  '3333')
ON CONFLICT DO NOTHING;
