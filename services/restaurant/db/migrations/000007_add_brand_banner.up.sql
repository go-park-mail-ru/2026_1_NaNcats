-- Широкая обложка-баннер ресторана (отдельно от квадратного logo_url).
-- Источник обложек: market-delivery.yandex.ru (place.picture), залиты в S3 (banners/).
ALTER TABLE "restaurant_brand"
	ADD COLUMN banner_url TEXT
		CHECK (char_length(banner_url) <= 2048);

UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/68712586-a092-5ea0-823e-1bebc4ed9f59.webp' WHERE name = 'Anna';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/03f593a8-7f9c-5ae6-9e29-9d725d4ae539.webp' WHERE name = 'DiDi';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/9b7d0a44-0426-50cc-bb12-a3909065d6f6.webp' WHERE name = 'El Chapo Burgers Tacos&Burritos';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/f436016d-51b1-5f58-abac-97c3ece73e08.webp' WHERE name = 'Eshak';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/0755c782-0e12-5f59-9077-1692577723c5.webp' WHERE name = 'FoodBand';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/704ae8e8-608f-57fa-8f8a-a1e8e7257e66.webp' WHERE name = 'Izumi';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/636387f8-3713-538c-ad2d-e13a08c66994.webp' WHERE name = 'KFC';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/56ec57c7-49e5-5a3c-8b3f-858ca79bcf97.webp' WHERE name = 'Ketch Up';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/66923ed0-5afa-5548-9385-e7f8421c8d57.webp' WHERE name = 'Moro';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/70af8cae-4679-59e6-87a4-67742cbce51f.webp' WHERE name = 'Pinsky Go';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/6cfe7cc9-71d2-59d8-9434-547aa4a9fde1.webp' WHERE name = 'Subway';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/965fdd9e-3c9d-58c0-8ee3-e1b049abf4ac.webp' WHERE name = 'Technikum';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/e5f44a59-60f1-50fb-9b1e-a4b608b1bae1.webp' WHERE name = 'Torro Grill';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/a95664be-4efd-5069-810e-bce8dec2e60b.webp' WHERE name = 'Tutta La Vita';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/1678f627-6d3c-5ecf-956c-91260b89e4e9.webp' WHERE name = 'Varvarka III';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/b4904a1d-3ce7-5db4-b380-6c84496e930f.webp' WHERE name = 'Villa Pasta';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/f755cce4-71db-5878-99a1-0c4eca716a78.webp' WHERE name = 'Аист';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/41e9e89c-3592-558e-b708-8b4942e670f2.webp' WHERE name = 'Братья Караваевы';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/210d2aae-d563-56e7-a22e-3afa279685da.webp' WHERE name = 'Ванлав';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/92e59c95-f278-5da7-a832-70272bde0867.webp' WHERE name = 'Вкусно - и точка';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/9352ed0e-b783-53a5-94eb-f4794b6aedab.webp' WHERE name = 'Империя Пиццы';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/b9c02d82-bd00-5736-97f6-c273021ea996.webp' WHERE name = 'Калифорния Дайнер';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/dd1464ce-68f0-5d67-aec4-1de9dadab7be.webp' WHERE name = 'Китайские новости';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/266d009e-3be9-555e-9306-da1020a85e1b.webp' WHERE name = 'Крошка Картошка';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/612238e1-a9d7-5bfc-8d8e-3818df44d1ad.webp' WHERE name = 'Лепим и Варим';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/68481ba4-307d-58aa-97af-71212943e59e.webp' WHERE name = 'Машрумс';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/07166a90-4c19-51c1-a6d8-d835dd48865a.webp' WHERE name = 'Папа Джонс';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/62b9c97e-6ec7-5217-b48e-20faf6f862fe.webp' WHERE name = 'Раменная Ru-Rik';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/d3c97fd0-18d7-5368-9ef3-e0908adc204c.webp' WHERE name = 'Руки ВВерх!';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/5e436490-9c33-56cb-948d-8b2d6daf7c86.webp' WHERE name = 'Такахули';
UPDATE "restaurant_brand" SET banner_url = 'https://nancats-bucket.storage.yandexcloud.net/banners/96a5155d-f74f-5e07-bde9-11ac7aa74e7c.webp' WHERE name = 'Честная Рыба';
