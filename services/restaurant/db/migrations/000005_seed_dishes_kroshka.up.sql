-- Блюда бренда «Крошка Картошка».
-- Источник: market-delivery.yandex.ru (ресторан «Крошка Картошка», Москва).
-- Изображения залиты в S3 (nancats-bucket/foods/). Цена в микрорублях (1 ₽ = 1 000 000).
-- Комбо/сеты исключены. section — раздел меню для группировки на фронте.

WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Крошка Картошка' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Печёный картофель
	((SELECT id FROM rb), 'Крошка Картошка с сыром', 'Печёный картофель на сырной основе.', 265000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e009cf31-760c-5939-9f35-c08f097cdf03.webp', 'Печёный картофель'),
	((SELECT id FROM rb), 'Крошка Картошка с растительным маслом', 'Печёный картофель на растительном масле.', 265000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a2b8ed3e-b2cb-5bb7-9ac2-f112653a2b28.webp', 'Печёный картофель'),
	((SELECT id FROM rb), 'Крошка Картошка с укропом и растительным', 'Печёный картофель с укропом и маслом.', 265000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/50eac810-163f-5233-ac1c-d4adb3444659.webp', 'Печёный картофель'),
	((SELECT id FROM rb), 'Крошка Картошка со сливочным маслом', 'Печёный картофель на сливочном масле.', 265000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e26f27b6-fa83-5f09-99a1-f079e46905e2.webp', 'Печёный картофель'),
	-- Наполнители
	((SELECT id FROM rb), 'Крабовое мясо с майонезом', 'Крабовые палочки с майонезом.', 159000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bdb2e620-e242-5ee2-8987-780d2a8d59c7.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Сырный с ветчиной', 'Ветчина, сыр и маринованные огурчики.', 159000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fdb9a444-53f2-5b02-8d5a-cacbe5b8575e.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Брынзовый с укропом', 'Творожная брынза с укропом.', 159000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b7e5a956-0796-5c8e-9493-f81926d106ed.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Мясное ассорти', 'Сытное ассорти из мясных деликатесов.', 159000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6cf847a0-07b8-5320-9dab-9a66ecf0ba8d.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Закусочный с грибами', 'Маринованные опята и огурчики.', 159000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/180a076c-9614-508c-9bed-67c029cea58a.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Том Ям с курицей и креветкой', 'Курица и креветки в соусе том-ям.', 195000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4d2c6b74-c3e6-5612-87ec-87278e166bb3.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Цыпленок песто', 'Нежная курица в соусе песто.', 195000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c5973a8f-bea1-576e-a10e-1530626212e4.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Красная рыбка', 'Солёная красная рыба с маслом.', 245000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/37c7a621-4917-51b9-bf97-1d9b176c0c60.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Треска по-северному', 'Филе трески с яичным омлетом.', 159000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d52b1e67-d333-59c0-a404-5aa02dec0f27.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Цыпленок с жареными грибами', 'Курица с жареными грибами и луком.', 159000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/56494d29-f981-5bf6-86af-6c73c4d344bb.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Сосиски в горчичном соусе', 'Сосиски в пряном горчичном соусе.', 159000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/103a8cf7-8be8-5fef-b1e3-65c11059205b.webp', 'Наполнители'),
	((SELECT id FROM rb), 'Охотничьи Колбаски в бургерном соусе', 'Колбаски-гриль с беконом и сыром.', 195000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5168a800-b764-5ba6-b33a-292bf6d9967e.webp', 'Наполнители'),
	-- Гриль
	((SELECT id FROM rb), 'Гриль Чиз Корн Бекон', 'Картофель с кукурузой, беконом и сыром.', 615000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7ff255a0-cc58-5888-bfd8-061c8537239e.webp', 'Гриль'),
	((SELECT id FROM rb), 'Гриль Чиз Пепперони', 'Картофель с колбасками под сыром.', 659000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dc44bcf2-1ac7-5749-8cf5-e9c3866ac369.webp', 'Гриль'),
	((SELECT id FROM rb), 'Гриль Чиз Креветки Том Ям', 'Картофель с креветками в том-яме.', 625000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fc89e302-c85d-52df-a397-65e0c0a00529.webp', 'Гриль'),
	((SELECT id FROM rb), 'Гриль Чиз с фрикадельками и сырным соусом', 'Картофель с фрикадельками под сырным соусом.', 625000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4cca421b-be2f-5799-b5b9-05bc9f3c7c9a.webp', 'Гриль'),
	-- Мэш
	((SELECT id FROM rb), 'Мэш большой', 'Четыре шарика картофеля с наполнителями.', 519000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/150e2385-6d6a-58e4-96e2-ef99e7a9c57a.webp', 'Мэш'),
	((SELECT id FROM rb), 'Мэш стандарт', 'Два шарика картофеля с наполнителем.', 325000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/94f5e5c5-9a6e-5879-9a5f-15fa08034269.webp', 'Мэш'),
	-- Пататы
	((SELECT id FROM rb), 'Патата Бургер', 'Котлета, сыр и пюре в тортилье.', 399000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7b8ca673-47a1-5f4c-a4e4-cf8910353295.webp', 'Пататы'),
	((SELECT id FROM rb), 'Патата Люля Кебаб', 'Куриный люля и пюре в тортилье.', 399000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/833d6f88-a914-586a-903a-de5e1b061cf8.webp', 'Пататы'),
	((SELECT id FROM rb), 'Патата Сосиска Голландский сыр', 'Сосиска и пюре в тортилье.', 269000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3a2beebe-225f-5f76-be83-7a73b8599575.webp', 'Пататы'),
	-- Сэндвичи
	((SELECT id FROM rb), 'Сэндвич Пепперони с сыром', 'Пепперони с сыром и томатным соусом.', 269000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2f95979a-8ae8-5c05-b70c-09dd12cdc069.webp', 'Сэндвичи'),
	-- Супы
	((SELECT id FROM rb), 'Щи из квашеной капусты с мясом большая порция', 'Кислые щи с мясом, большая порция.', 499000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3162c23a-9b7d-505d-b62f-beab409c739e.webp', 'Супы'),
	((SELECT id FROM rb), 'Лапша куриная', 'Лёгкий суп с курицей и лапшой.', 399000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/222217f2-0386-59b8-b8ab-353a1636f506.webp', 'Супы'),
	((SELECT id FROM rb), 'Суп гороховый', 'Насыщенный гороховый суп.', 399000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/70c628a7-7aa1-5a59-9915-0d02c7a62772.webp', 'Супы'),
	((SELECT id FROM rb), 'Щи из квашеной капусты с мясом', 'Кислые щи из квашеной капусты с мясом.', 399000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/26656c7f-3162-5a3b-8640-e9474c6a83c8.webp', 'Супы'),
	((SELECT id FROM rb), 'Суп гороховый Большая порция', 'Гороховый суп, большая порция.', 499000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ae051bdf-57ca-52c8-9070-5b63e8c23552.webp', 'Супы'),
	-- Салаты
	((SELECT id FROM rb), 'Цезарь Боул', 'Курица, сухарики и соус цезарь.', 389000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/57004509-7f1d-5ffa-b6bc-5d2487341ef0.webp', 'Салаты'),
	((SELECT id FROM rb), 'Овощной Боул', 'Свежие овощи с брынзой и укропом.', 345000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0aca2c36-a515-5f46-b438-a6489a9199a6.webp', 'Салаты'),
	-- Хлеб
	((SELECT id FROM rb), 'Гренки  "8 злаков"', 'Злаковые гренки с чесноком и зеленью.', 105000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1512197e-dbc2-5c2f-9a2c-3af3dae0c654.webp', 'Хлеб'),
	((SELECT id FROM rb), 'Хлеб с чесночным маслом', 'Поджаренный хлеб с чесночным маслом.', 105000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a60b0f99-8e61-523b-ab08-3b795926b44d.webp', 'Хлеб'),
	-- Десерты
	((SELECT id FROM rb), 'Пирожное картошка "Халва"', 'Пирожное «картошка» с халвой.', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a9f09f05-fba3-5329-be08-c6b72f20a701.webp', 'Десерты'),
	((SELECT id FROM rb), 'Пирожное Картошка', 'Классическое бисквитное пирожное «картошка».', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/04b471ed-a539-5435-9ddf-7aa627867682.webp', 'Десерты'),
	((SELECT id FROM rb), 'Пирожное картошка "Орех в карамели"', 'Пирожное «картошка» с орехом в карамели.', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1142dcd2-c67c-57e6-898d-9d6d0daee2ca.webp', 'Десерты'),
	((SELECT id FROM rb), 'Чизкейк маракуйя-облепиха', 'Чизкейк с маракуйей и облепихой.', 325000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8984c59f-5a9d-534a-bd9b-e4c6ac63abd6.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Добрый Кола', 'Газированный напиток кола, 0,5 л.', 225000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/aef629ff-9ce2-5d29-b681-42984a5d8480.webp', 'Напитки'),
	((SELECT id FROM rb), 'Домашний Лимонад Розмарин Мята в бутылке 0,5л', 'Домашний лимонад с розмарином и мятой.', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1a479bd2-96f7-5472-b385-0f3bf073a957.webp', 'Напитки'),
	((SELECT id FROM rb), 'Морс ягодный 0,5л  (в бутылке)', 'Натуральный ягодный морс.', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/29e2b500-e5dc-596a-a9a1-0c3255082de4.webp', 'Напитки'),
	((SELECT id FROM rb), 'Домашний Лимонад Розмарин Мята в бутылке 1л', 'Домашний лимонад с розмарином и мятой.', 375000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1a479bd2-96f7-5472-b385-0f3bf073a957.webp', 'Напитки'),
	((SELECT id FROM rb), 'Морс ягодный 1л (в бутылке)', 'Натуральный ягодный морс.', 375000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/29e2b500-e5dc-596a-a9a1-0c3255082de4.webp', 'Напитки'),
	((SELECT id FROM rb), 'Добрый Кола без сахара', 'Кола без сахара, 0,5 л.', 225000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f34e37c6-f01a-50c1-b175-2243dc7214f8.webp', 'Напитки'),
	((SELECT id FROM rb), 'БонаАква негаз', 'Питьевая вода без газа.', 199000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/01a46888-9d62-59cb-a680-d49e85f02c71.webp', 'Напитки'),
	((SELECT id FROM rb), 'БонаАква газ', 'Питьевая вода с газом.', 199000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8715231e-c6ef-56a1-8de3-e79aad8493db.webp', 'Напитки'),
	((SELECT id FROM rb), 'Добрый апельсин', 'Апельсиновый напиток, 0,5 л.', 225000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/712b44a9-8e10-5323-840f-ec5dc0c8439d.webp', 'Напитки'),
	((SELECT id FROM rb), 'Добрый лимон-лайм', 'Напиток лимон-лайм, 0,5 л.', 225000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c3aa63e9-d364-55c8-b4d2-951ed29c8550.webp', 'Напитки'),
	((SELECT id FROM rb), 'Rich черный чай с лимоном 0,5', 'Холодный чёрный чай с лимоном.', 249000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d49b8dfc-5494-5ddd-80e4-bedddfa44bbe.webp', 'Напитки'),
	((SELECT id FROM rb), 'Rich зеленый чай с манго 0,5', 'Холодный зелёный чай с манго.', 249000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f95658b8-2294-5416-baf4-2c7b4a259aed.webp', 'Напитки'),
	((SELECT id FROM rb), 'Сок Добрый АПЕЛЬСИН МАНДАРИН 0,3', 'Сок апельсин-мандарин, 0,3 л.', 169000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e10a08d5-501f-536b-af19-ed1f67b30f54.webp', 'Напитки'),
	((SELECT id FROM rb), 'Сок Добрый ЯБЛОКО 0,3', 'Яблочный сок, 0,3 л.', 169000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/118da3b9-6e27-53c8-a0d0-bf45fde6a761.webp', 'Напитки'),
	((SELECT id FROM rb), 'Добрый PULPY апельсин', 'Апельсиновый напиток с мякотью.', 239000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9245f74f-9c62-5676-bd5b-bd26d33134ba.webp', 'Напитки'),
	((SELECT id FROM rb), 'Добрый PULPY ананас', 'Ананасовый напиток с мякотью.', 239000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3c6ac6e9-75c2-5365-8906-da4af822e2d7.webp', 'Напитки'),
	((SELECT id FROM rb), 'Квас Яхонт', 'Традиционный хлебный квас.', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c95d2e26-7ed5-5bf8-b382-88320af7744d.webp', 'Напитки'),
	((SELECT id FROM rb), 'Пиво Bud Б/А', 'Безалкогольное пиво Bud.', 249000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2adf74ea-c38b-567b-88b6-a3dc9793c80a.webp', 'Напитки'),
	-- Соусы и приправы
	((SELECT id FROM rb), 'Сметана', 'Порционная сметана.', 41000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c53724a6-f433-5b2d-8a82-709cf9b926f2.webp', 'Соусы и приправы'),
	((SELECT id FROM rb), 'Майонез', 'Порционный майонез.', 41000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/14a09063-8875-5968-9e73-2ee03f89b17a.webp', 'Соусы и приправы');
