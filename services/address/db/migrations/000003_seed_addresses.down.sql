-- Откат сид-адресов первого клиента (client_account_id = 3).
DELETE FROM "client_address" WHERE client_account_id = 3;

-- Удаляем добавленные локации. FK location -> client_address стоит на
-- ON DELETE RESTRICT, поэтому только после удаления адресов выше.
DELETE FROM "location" WHERE address_text IN (
    'Москва, Пресненская наб., 12',
    'Москва, Ленинский проспект, 38',
    'Москва, Кутузовский пр-т, 24',
    'Москва, ул. Бауманская, 11',
    'Москва, Новый Арбат, 21',
    'Москва, ул. Маросейка, 9'
);
