-- Блюда остальных брендов (25 ресторанов).
-- Источник: market-delivery.yandex.ru (точное совпадение бренда). Меню до 40 позиций/ресторан.
-- Изображения залиты в S3 (nancats-bucket/foods/, webp). Цена в микрорублях (1 ₽ = 1 000 000).
-- Комбо/сеты и промо-разделы исключены. section — раздел меню для группировки на фронте.

-- Братья Караваевы (id 5)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Братья Караваевы' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Закуски
	((SELECT id FROM rb), 'Бутер сыр-ветчина-омлет', 'Омлет, сочная ветчина из индейки, сыр чеддер…', 400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6c0fa594-159e-5507-8700-11fbeba2e922.webp', 'Закуски'),
	((SELECT id FROM rb), 'Круассан с ветчиной и сыром Моцарелла', 'Воздушный круассан, с сочной ветчиной из мяса…', 320000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/898bd131-76af-54a6-bb20-d70d8fe03190.webp', 'Закуски'),
	((SELECT id FROM rb), 'Блин Рваная говядина', 'Блин из постного теста с начинкой из…', 165000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7c8448dc-da31-594d-acdd-87779d28136d.webp', 'Закуски'),
	((SELECT id FROM rb), 'Блин овощной', 'Нежные блинчики с овощами и грибами порадуют…', 130000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3b253f02-7288-5e3c-9070-a9d6fc335e2d.webp', 'Закуски'),
	((SELECT id FROM rb), 'Сэндвич с брискетом', 'Изысканный сэндвич с брискеттом: нежная говядина на…', 415000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8bc043a2-989d-54fb-bf32-cd3016da0eaf.webp', 'Закуски'),
	((SELECT id FROM rb), 'Блин шоколад-апельсин', 'Шоколадный блин с бархатистыми сливками, сочетанием тёмного…', 170000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/31af098a-2244-58f6-b30a-8762682fcca5.webp', 'Закуски'),
	-- Выпечка
	((SELECT id FROM rb), 'Ватрушка венгерская с творогом', 'Ватрушка из слоеного теста с нежной творожной начинкой', 165000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a013bf8a-9dbc-544f-b087-7b52b7023925.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Слойка шоколад-апельсин', 'Аппетитная выпечка, украшенная шоколадными молочными каплями, которые…', 245000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7974b6c1-92f8-540f-b6df-2d2ed4463f29.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Слойка с голубикой', 'Хрустящая слоёная корочка и нежная начинка из…', 365000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f7be299b-8c71-5a3f-bc90-804237d2ed3a.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Сладкое сердце', 'Традиционная российская выпечка.', 160000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/79b4cedc-7af9-5806-b968-149379387bb3.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Маритоццо', 'Пышная булочка с воздушным кремом на основе…', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7d75a1f4-effb-5e8d-8de3-3829cb50c978.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Круассан французский', 'Классический французский круассан - это отличное начало дня.', 160000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8d163b06-7f30-508c-8251-15caf18670b3.webp', 'Выпечка'),
	-- Супы
	((SELECT id FROM rb), 'Окрошка', 'Освежающая прохлада окрошки, приготовленной на нежном кефире…', 200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fd3207f0-0712-5da9-be49-dd492c84cd70.webp', 'Супы'),
	-- Домашняя еда
	((SELECT id FROM rb), 'Кордон блю', 'Сочное филе куриной грудки и бедра создают…', 260000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6db27325-99ff-58f5-9aa2-75d60350556a.webp', 'Домашняя еда'),
	((SELECT id FROM rb), 'Домашняя куриная котлета', 'Нежное мясо цыплёнка в сочетании с ароматным…', 125000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4578bd46-f6cf-5924-923a-acabda1498fc.webp', 'Домашняя еда'),
	((SELECT id FROM rb), 'Котлета мясная', 'Мясные котлеты, приготовленные из свинины и говядины…', 195000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0616fd6f-a0a0-50a4-9934-2c948f3eacf0.webp', 'Домашняя еда'),
	-- Гарниры
	((SELECT id FROM rb), 'Картофельный оладушек', 'Вкусное и сытное блюдо, приготовленное из свежего картофеля', 55000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/797dbdc7-4092-5ab1-8754-0e71a7de2ed6.webp', 'Гарниры'),
	-- Десерты
	((SELECT id FROM rb), 'Маковый сметанник', 'Воздушная текстура, пропитанная сливочной нежностью сметаны, гармонично…', 255000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/92ea952b-466b-59ad-b799-8f40b16ddf73.webp', 'Десерты'),
	((SELECT id FROM rb), 'Тарталетка карамельный пекан', 'Тарталетка с карамелизованным пеканом: миндальное финансье, кленовая…', 235000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/25f261da-c9c0-5c1d-8c4f-8a3f6cdb9c9d.webp', 'Десерты'),
	((SELECT id FROM rb), 'Кольцо Творожное', 'Кольцо из заварного теста с нежным творогом…', 210000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/84fa634f-603e-5da9-b797-f6cd429af6d1.webp', 'Десерты'),
	((SELECT id FROM rb), 'Торт Три шоколада', 'Симфония вкусов: нежный бисквит, слои из тёмного…', 315000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4cd905e1-0ddc-5508-87d7-4626dcdd749b.webp', 'Десерты'),
	((SELECT id FROM rb), 'Эклер пряный латте', 'Нежное тесто, заварной крем на топлёном молоке…', 195000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d3168dcc-c8cb-52e8-bf99-73215bf4ef23.webp', 'Десерты'),
	((SELECT id FROM rb), 'Эклер Шоколадный', 'Изысканный десерт, который станет настоящим украшением любого стола.', 195000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bdc05fe2-b746-5439-b1ae-41970d6ad3c8.webp', 'Десерты'),
	-- Торты
	((SELECT id FROM rb), 'Торт Вацлавский', 'Мягкий шоколадный бисквитс ванильным кремом, добавлением Компотной…', 980000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b4db3e6c-575e-55cb-9c5d-61c8a246243f.webp', 'Торты'),
	((SELECT id FROM rb), 'Торт Ягодный бархат', 'Изысканный бисквитный торт сочетающий в себе нежность…', 980000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c7cc0374-24db-5b22-ba82-f6145a2a62f8.webp', 'Торты'),
	-- Хлеб
	((SELECT id FROM rb), 'Хлеб гречишный', 'Гречишный хлеб, с лёгкой пряностью кориандра и…', 105000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3208a577-3672-5203-9356-9c294aba98b7.webp', 'Хлеб'),
	((SELECT id FROM rb), 'Хлеб Чесночный', 'Хлеб пропитанный маслом с чесноком', 115000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/72e17f21-26c6-5828-8dbc-63033fe8c526.webp', 'Хлеб'),
	((SELECT id FROM rb), 'Хлеб Бородинский', 'Хлеб из ржаной муки с кориандром и медом', 130000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2ad3242e-9310-5c93-abfc-364cc213b6d6.webp', 'Хлеб'),
	-- Печенье
	((SELECT id FROM rb), 'Слойка голландская', 'Золотистая корочка песочного теста на сметане идеально…', 320000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1387bbff-39cf-5836-97a1-768027b79371.webp', 'Печенье'),
	-- Холодные напитки
	((SELECT id FROM rb), 'Лимонад груша', 'Лимонад „Груша: основа „Груша-липа, свежая груша, мята, лёд.', 430000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/585c297f-90c0-588d-b97e-50af457c7e52.webp', 'Холодные напитки'),
	((SELECT id FROM rb), 'Лимонад клубника-базилик', 'Освежающий лимонад с нотами клубники и базилика.', 430000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/505d5c1d-7f34-5e48-93df-dedc61ff8a48.webp', 'Холодные напитки'),
	((SELECT id FROM rb), 'Морс облепиха-апельсин', 'Яркий и ароматный напиток, который сочетает в…', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/03d65a9b-6050-52b9-a2cf-6d3c179a4411.webp', 'Холодные напитки'),
	((SELECT id FROM rb), 'Морс черная смородина', 'Восхитительный натуральный морс из чёрной смородины с…', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/354c34bf-6da5-5aed-a66e-06a86c8ac8eb.webp', 'Холодные напитки'),
	((SELECT id FROM rb), 'Морс клюква-брусника', 'Восхитительный натуральный морс из клюквы и брусники…', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/02502a79-0f71-51f1-8fc8-45cae7ad20d7.webp', 'Холодные напитки'),
	-- Кофе в упаковке
	((SELECT id FROM rb), 'Кофе жареный в зернах Эфиопия/Кения', '100% Арабика класса Specialty coffee - это…', 3410000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/cd21c4d0-a774-5185-a4ba-612c74ed85bf.webp', 'Кофе в упаковке'),
	((SELECT id FROM rb), 'Кофе жареный в зернах Эфиопия', 'Specialty coffee натуральный, 100% арабика', 845000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a8f1191f-9328-5844-b953-6de6eb155ee8.webp', 'Кофе в упаковке'),
	((SELECT id FROM rb), 'Кофе жареный в зернах Бразилия', '100% Арабика класса Specialty coffee - это…', 2995000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a9a7031b-2e75-5039-84f8-d3675a0b767e.webp', 'Кофе в упаковке'),
	((SELECT id FROM rb), 'Кофе жареный в зернах Эфиопия/Бразилия', '100% Арабика класса Speciality', 3080000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2cabcf7c-03e5-5e55-90a1-a93a1b053f32.webp', 'Кофе в упаковке'),
	((SELECT id FROM rb), 'Кофе в дрип-пакетах Эфиопия', 'Кофе натуральный жареный, молотый, арабика высший сорт…', 465000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c27c2223-b971-5b04-b076-1e63bdea485e.webp', 'Кофе в упаковке'),
	((SELECT id FROM rb), 'Кофе в дрип-пакетах в ассортименте', 'Кофе натуральный жареный молотый, арабика высший сорт…', 465000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6c55d5c1-155c-565d-9677-ea3248f18f0a.webp', 'Кофе в упаковке');

-- Technikum (id 7)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Technikum' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Торты
	((SELECT id FROM rb), 'Блинный торт', 'Блины (яйца, молоко, мука, сахар, сливочное масло)…', 6500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/56b44eab-4469-56d8-9f4c-532373423347.webp', 'Торты'),
	((SELECT id FROM rb), 'Вишневый торт', 'Бисквитные коржи с морковью в вишневом соке…', 6500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d276ecdf-4a50-547d-8e2c-97162f725d5e.webp', 'Торты'),
	-- Десерты
	((SELECT id FROM rb), 'Чизкейк Сан-Себастьян', 'Творог, яйца, сливки, творожный сыр со сливочным…', 5900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/11876b31-508b-503d-93ee-5241fc29b35f.webp', 'Десерты'),
	((SELECT id FROM rb), 'Эклер с маком', 'Заварное тесто с заварным кремом из вареной…', 670000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4fe6dfed-9c2d-557d-9878-1820fcb11af8.webp', 'Десерты'),
	-- Завтрак
	((SELECT id FROM rb), 'Зож-завтрак', 'Полезный завтрак из белкового омлета (яичный белок)…', 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2ae4e095-8715-5159-8687-a8716f4e6bc3.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Завтрак на одного', 'Завтрак на одного из сэндвича с ветчиной…', 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/28796bf7-fa09-508e-8057-5075a8b8812d.webp', 'Завтрак'),
	-- Клубничное лето
	((SELECT id FROM rb), 'Страчателла, клубника и лавандовый мед', 'Нежный сыр страчателла, подушка из шпината, соус…', 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e1ddcfcb-0645-55ae-835a-6f1e35c8c155.webp', 'Клубничное лето'),
	((SELECT id FROM rb), 'Брускетта со страчателлой и клубникой', 'Домашний хлеб, сыр страчателла, клубника в собственном…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4747715d-05df-5f2c-b640-be00e94cc732.webp', 'Клубничное лето'),
	-- Холодные закуски
	((SELECT id FROM rb), 'Мясное плато', 'Мясное плато из пастрами из говядины, пастрами…', 1350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/48a4b43d-53c9-5007-ae36-4500f74c8439.webp', 'Холодные закуски'),
	((SELECT id FROM rb), 'Хумус с питой', 'Отварной нут, оливковое масло, чили, кунжутная паста…', 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/630f653d-693e-5387-9677-3be662d0e694.webp', 'Холодные закуски'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Креветки гриль по-тайски', 'Креветки в тайском соусе терияки с ароматным…', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ef22e041-8c75-5246-9b1e-8d9a494d2257.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Батат с креветками и соусом BBQ', 'Запеченный батат, креветки, соус чимичурри, соус терияки…', 720000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/09a0fa72-1a01-5ced-97f0-906985021068.webp', 'Горячие закуски'),
	-- Бутеры
	((SELECT id FROM rb), 'Брускетта с баклажанами и страчателлой', 'Домашний хлеб, нежный сыр страчателла, баклажаны, кедровый…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1dba808b-ac71-5972-9069-a00ec27155e7.webp', 'Бутеры'),
	((SELECT id FROM rb), 'Сморреброд с креветками и гуакамоле', 'Домашний хлеб, соус гуакамоле: авокадо, томаты, лук…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8df237b2-d01e-5ca2-bb83-37d12897a2a7.webp', 'Бутеры'),
	-- Салаты
	((SELECT id FROM rb), 'Креветки, авокадо и рукола', 'Обжаренные креветки, рукола, базилик, пармезан, авокадо, томаты…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8c8a8720-8a37-5d16-971c-5cd9d8e0d6d5.webp', 'Салаты'),
	((SELECT id FROM rb), 'Помидоры, авокадо и киноа', 'Помидоры, киноа, огурцы, авокадо, красный лук, соус…', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d4a20e78-4416-5273-93a0-119447df1a94.webp', 'Салаты'),
	-- Супы
	((SELECT id FROM rb), 'Борщ с говядиной', 'Борщ из говяжего бульона, свекла, морковь, сметана…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/62f78674-91c0-593e-abb8-e13dbcc7eb75.webp', 'Супы'),
	((SELECT id FROM rb), 'Крем-суп из спаржи', 'Сельдерей, репчатый лук, спаржа, кабачки, куриный бульон…', 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/889de5be-7a18-5e33-9bfc-73361379fdbc.webp', 'Супы'),
	-- Поке
	((SELECT id FROM rb), 'Поке с угрем', 'Традиционное гавайское блюдо из белого риса, угря…', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/681765e8-0236-5a47-b78a-156e1aff15c8.webp', 'Поке'),
	((SELECT id FROM rb), 'Поке с тунцом', 'Традиционное гавайское блюдо на подушке из белого…', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6028360a-74b1-50ca-b5c8-eac11a9434bb.webp', 'Поке'),
	-- Роллы
	((SELECT id FROM rb), 'Ролл Калифорния с крабом', 'Рис, нори, тобико, сливочный сыр, авокадо и…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d88ef3fc-6eec-58f1-ba6e-4b46074f30b1.webp', 'Роллы'),
	((SELECT id FROM rb), 'Ролл Филадельфия с лососем', 'Рис, нори, сливочный сыр, авокадо и лосось', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7e52e571-1cfa-565e-8195-d78cbf3e2ff0.webp', 'Роллы'),
	-- Пицца
	((SELECT id FROM rb), 'Пицца Техникум', 'Наша авторская пицца из ассорти вкусных кусочков…', 1290000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/891be315-8785-52e3-ae22-a631dc1520da.webp', 'Пицца'),
	((SELECT id FROM rb), 'Пицца Маргарита', 'Соус из спелых итальянских томатов с базиликом…', 840000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b1b419fc-1b67-5fe0-b018-0af750baa54d.webp', 'Пицца'),
	-- Паста
	((SELECT id FROM rb), 'Спагетти арабьята с креветками', 'Спагетти в сливочно-томатном соусе, креветки, зеленый лук…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0513264d-d02f-5156-9a39-fe7d46d10966.webp', 'Паста'),
	((SELECT id FROM rb), 'Лазанья с телятиной в соусе бешамель', 'Фарш из телятины, томаты конкассе, соус бешамель…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5d4d14e1-1d3f-5e52-baf0-a40232ff19f0.webp', 'Паста'),
	-- Junk Food
	((SELECT id FROM rb), 'Супербоул с цыпленком', 'Белый рис, шпинат, бамбук, чили, яйцо, нори…', 820000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b62abc7a-85e7-5a2b-b170-caae22aa3682.webp', 'Junk Food'),
	((SELECT id FROM rb), 'Шаверма с курицей в пите', 'Филе цыпленка, капуста, соленый огурец, помидор, маринованный…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2f91cf16-df5b-5a6d-b64e-5a325907cd1f.webp', 'Junk Food'),
	-- Рыба и морепродукты
	((SELECT id FROM rb), 'Лосось на гриле, авокадо и терияки', 'Стейк из лосося, авокадо, соус терияки, соль', 1250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6911a57d-bc69-5b72-acb3-bbb9364551da.webp', 'Рыба и морепродукты'),
	((SELECT id FROM rb), 'Гребешки, трюфель и белые грибы', 'Воздушное картофельное пюре (kартофель, молоко, соль) в…', 1250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f8694095-15e5-5abb-b8a5-5bb3bac3eaf6.webp', 'Рыба и морепродукты'),
	-- Мясо и птица
	((SELECT id FROM rb), 'Томленый рибай, белые грибы и устричный соус', 'Томленый рибай, картофельное пюре (Картофель, молоко, соль)…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/633c70c7-168d-5d0c-b3d6-1bf309dadb61.webp', 'Мясо и птица'),
	((SELECT id FROM rb), 'Котлеты из индейки с маринованными огурцами', 'Нежные котлетки из индейки (Филе индейки, хлеб…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e3a12261-6cbd-5297-b0f9-a9738c2628f3.webp', 'Мясо и птица'),
	-- Гарниры
	((SELECT id FROM rb), 'Белый рис', 'Отварной рис в заправке из мирина и мицукана', 100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fc9f96f1-eec7-5ada-b467-0e717e5ef938.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Картофель фри', 'Картофель фри с солью и бургерным соусом', 350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/44f8f335-67b8-5893-bee9-aae5d693b0cf.webp', 'Гарниры'),
	-- Дополнительно
	((SELECT id FROM rb), 'Хлеб домашний', 'Хлеб домашний', 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/55f9d475-4e80-55d0-9633-b13b045abcc1.webp', 'Дополнительно'),
	((SELECT id FROM rb), 'Угорь', 'Свежий угорь', 400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f72a331c-87ec-59e7-b52b-6d82f49da7bd.webp', 'Дополнительно'),
	-- Напитки
	((SELECT id FROM rb), 'Безалкогольное пиво Крушовице [ат]', NULL, 520000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a296260c-e3b7-5bb2-9612-5d2248149b83.webp', 'Напитки'),
	-- Утро в Техникуме
	((SELECT id FROM rb), 'Омлет с сыром и зеленью', 'Омлет из трех яиц с добавлением сыра…', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f43a2f0f-b749-5538-a425-e0ca438b3334.webp', 'Утро в Техникуме'),
	-- Выпечка
	((SELECT id FROM rb), 'Картофельный драник с пастрами из индейки', 'На золотистом картофельном дранике - нежная пастрами…', 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6e495dfa-7fcb-5cc1-ae8e-c336eb5cf727.webp', 'Выпечка'),
	-- Сэндвичи
	((SELECT id FROM rb), 'Авокадо-тост с рикоттой', 'Хлеб бездрожжевой, гуакамоле (Авокадо, томаты, лук, оливковое…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3465cce2-b923-5a07-a5f2-72d52e778272.webp', 'Сэндвичи');

-- Лепим и Варим (id 8)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Лепим и Варим' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Пельмени вареные
	((SELECT id FROM rb), 'Мамин сибиряк', 'Фарш: окорок свиной, оковалок говяжий, лук репчатый…', 515000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d6072cfe-9f1c-519a-a2a4-04bc11e64d26.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Итальянцы в россии', 'Фарш: филе бедра куриное, лук пассерованный, сыр…', 515000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6bdee0d3-7d56-578c-9cca-a2be9da25a29.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Большие мясные пельмени', 'Фарш: окорок свиной, грудинка говяжья, баранина мякоть…', 755000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d95697c1-17a3-5a01-9b90-f4478e2157af.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Классика жанра 15 шт. (300г)', 'Фарш: филе бедра куриное, говядина оковалок, лук…', 647000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1cffbc34-82f1-5703-9e51-885debafc457.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Бульба бегинс', 'Фарш: картофель отварной, лук фри, укроп свежий…', 347000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e8035dbe-8580-508b-84ce-70c76deca4f7.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Мятрица 10 шт. (180г) 10 шт', NULL, 395000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/36c0686a-d36b-5db2-b229-57c42652efca.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Трое в тесте 10 шт. (200 гр)', NULL, 389000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a71fc433-320f-577e-91bc-ce4bd25fff9e.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Трое в тесте 15 шт. (200 гр)', NULL, 549000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a71fc433-320f-577e-91bc-ce4bd25fff9e.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Поросячий восторг 10 шт. (200 гр)', NULL, 389000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7bbd3e5d-700a-523a-bc03-2bdc3785ee3a.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Поросячий восторг 15 шт. (300 гр)', NULL, 549000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7bbd3e5d-700a-523a-bc03-2bdc3785ee3a.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Щука-любовь 10 шт. (200г) 10 шт', 'Состав: филе судака, филе щуки, лук, масло…', 659000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/748c11c9-37cb-5b8f-ac08-1fbbb9831fa6.webp', 'Пельмени вареные'),
	((SELECT id FROM rb), 'Щука-любовь 15 шт. (300г) 15 шт', 'Состав: филе судака, филе щуки, лук, масло…', 919000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/748c11c9-37cb-5b8f-ac08-1fbbb9831fa6.webp', 'Пельмени вареные'),
	-- Супы
	((SELECT id FROM rb), 'Куриный бульон', 'Куриный бульон, укроп, соль, перец', 199000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/cf88fc7d-33d6-5a0d-8461-39430b33d9e2.webp', 'Супы'),
	-- Салаты
	((SELECT id FROM rb), 'Коул слоу', 'Капуста, морковь, соль, сахар, майонез, сок лимона', 407000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/06c7e9d7-b672-5a0b-a9d1-b9a50f4e8f40.webp', 'Салаты'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Пончики товрожные 5 шт', 'Небольшие шаровидные пончики из творожного теста, посыпанные…', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dce46345-f6db-57a7-a9e9-0c56d55400cc.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Картофель фри', NULL, 159000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a3f71f4a-7aca-566a-8d85-f130b640eb8f.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Пельмени фри три мяса 4 шт', 'Аррр!', 395000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/75df11e4-93f1-59e8-bd7d-38f979317bd1.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Пельмени фри три мяса 8 шт', 'Аррр!', 719000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/75ee32dd-e25a-5d59-abfa-23fb588e2b05.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Пельмени фри сыыыр 4 шт', 'Преступно-притягательное сыыырное удовольствие!', 395000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6202fff0-196c-53ce-bcee-74168357cc8e.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Пельмени фри сыыыр 8 шт', 'Преступно-притягательное сыыырное удовольствие!', 719000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/020aab95-95d4-5558-b665-a7fa5749e0d8.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Пельмени фри Сыр и мясо 4 шт', '4 шт', 395000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a0353e08-8cdc-507f-b9d9-211d1e8d12ef.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Пельмени фри Сыр и мясо 8 шт', '8 шт', 719000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fb997694-5918-583a-8243-47b9de32bd42.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Пончики с дыркой 2 шт', 'Классические пончики из пышного теста, посыпанные сахарной пудрой.', 167000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0830a7e4-6e81-585f-874a-68d64ddf5f45.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Пончики с дыркой 3 шт', 'Классические пончики из пышного теста, посыпанные сахарной пудрой.', 215000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c6c07287-ead2-5013-8ec7-4f6aeda3973f.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Пончики товрожные 7 шт', 'Небольшие шаровидные пончики из творожного теста, посыпанные…', 251000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/738b3465-4a7b-5293-b6f2-ba662b484da5.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Посикунчики классик 5 шт', 'Посикунчики - это традиционное пермское блюдо: что-то…', 287000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1acb1a9a-4a0e-557c-8afc-5dc370a46558.webp', 'Горячие закуски'),
	-- Напитки
	((SELECT id FROM rb), 'Морс из клюквы', 'Клюква, сахар, вода', 239000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6d039207-b01c-5bf9-bf5d-31b962acef75.webp', 'Напитки'),
	((SELECT id FROM rb), 'Компот из сухофруктов', 'Сухофрукты, сахар, вода', 239000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/02704ce3-e53d-522d-9ce1-5591527a26aa.webp', 'Напитки'),
	-- Соусы
	((SELECT id FROM rb), 'Сметана', 'Сметана 30%', 109000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/769e6b84-e048-59f9-854b-39d5e88d1136.webp', 'Соусы'),
	-- Пельмени шоковой заморозки
	((SELECT id FROM rb), 'Трое в тесте 480г я', NULL, 809000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8c769dfa-7b63-5699-99cd-e6e12681ebc8.webp', 'Пельмени шоковой заморозки'),
	((SELECT id FROM rb), 'Поросячий восторг 480г я', NULL, 809000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/52c16e40-5f1c-573a-91cc-c30eb9d2e7cd.webp', 'Пельмени шоковой заморозки'),
	((SELECT id FROM rb), 'Щука-любовь 30 шт. (480г)', 'Филе судака, филе щуки, лук, масло подсолнечное…', 1079000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3e5031e7-386b-55de-b5ea-00fe4482b6f1.webp', 'Пельмени шоковой заморозки'),
	((SELECT id FROM rb), 'Мамин сибиряк 30 шт.(480г)', 'Фарш: окорок свиной, оковалок говяжий, лук репчатый…', 1079000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0d2b0fd8-3181-5824-889b-0d4c7668f9d2.webp', 'Пельмени шоковой заморозки'),
	((SELECT id FROM rb), 'Итальянцы в России 30 шт.(480)', 'Фарш: филе бедра куриное, лук пассерованный, сыр…', 1079000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7e7fca92-fb2a-530d-9411-a222757f98fc.webp', 'Пельмени шоковой заморозки'),
	((SELECT id FROM rb), 'Синяя птица 30 шт.(480г)', 'Фарш: филе бедра индейки, мясо птицы, лук…', 1069000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/01b2f065-0c7c-509a-8752-9d96cf911935.webp', 'Пельмени шоковой заморозки'),
	((SELECT id FROM rb), 'Конечно, Уася! 30 шт.(480г)', 'Фарш: баранина мякоть, лук репчатый, жир бараний…', 1349000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/723d1ec9-3c36-5578-bf06-948bd0c4c3dc.webp', 'Пельмени шоковой заморозки'),
	((SELECT id FROM rb), 'Большие мясные пельмени 24 шт.(600г)', 'Фарш: окорок свиной, грудинка говяжья, баранина мякоть…', 1484000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f78c1e58-b8c6-57a4-94db-da72e0ca3ef2.webp', 'Пельмени шоковой заморозки'),
	((SELECT id FROM rb), 'Скажите Сыыыр! 30 шт.(480г)', 'Фарш: cыр Гауда, сыр Сулугуни, сыр Пармезан…', 1349000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/28fa2fcc-107f-54e4-973c-0bbdc170227d.webp', 'Пельмени шоковой заморозки'),
	((SELECT id FROM rb), 'Знаменитая креветка 30 шт.(480г)', 'Фарш: филе бедра куриное, креветки очищенные б/г…', 1349000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7ba50655-f264-50d9-a8b0-e88de152d29c.webp', 'Пельмени шоковой заморозки'),
	((SELECT id FROM rb), 'Дядя с Камчатки 30 шт. (480г)', 'Фарш: мясо краба в/м, сметана 20%, масло…', 3359000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/29324e5d-a0e7-5f09-96d4-db5589e26968.webp', 'Пельмени шоковой заморозки');

-- Китайские новости (id 9)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Китайские новости' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Мясо и птица
	((SELECT id FROM rb), 'Утка по-пекински половина', 'Хрустящая утиная кожа и нежное мясо, приготовленные…', 2900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ae0b95f6-3d75-5280-a03d-4fc53def2923.webp', 'Мясо и птица'),
	((SELECT id FROM rb), 'Курица по-сычуаньски', 'Курица, обжаренная с сушеным перцем чили, ароматным…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/874d3497-1df8-5671-8459-d9021bc51655.webp', 'Мясо и птица'),
	((SELECT id FROM rb), 'Говядина по-сычуаньски', 'Говядина, обжаренная с пастой бобовой, сушеным перцем…', 1290000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/eeb08766-570a-574e-9056-a3b1616b1115.webp', 'Мясо и птица'),
	((SELECT id FROM rb), 'Баранина по-пекински с паровыми булочками бао', 'Нежная баранина, приправленная солью и перцем, подается…', 1390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e813c1f5-da8a-5421-97b8-be8107f9f6d4.webp', 'Мясо и птица'),
	-- Лапша и рис
	((SELECT id FROM rb), 'Жареная широкая рисовая лапша Хэ-фан с говядиной', 'Широкая рисовая лапша, обжаренная на сильном огне…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5b8f6b13-6a90-56da-b2ba-df7ef539e963.webp', 'Лапша и рис'),
	((SELECT id FROM rb), 'Лапша по-гонконгски с уткой', 'Яичная лапша, обжаренная с луком и хрустящими…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/386d5cea-ef69-5205-92b9-b9a73b25001a.webp', 'Лапша и рис'),
	((SELECT id FROM rb), 'Жареный рис с креветками', 'Рассыпчатый рис, обжаренный на растительном масле до…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0d6e3059-83ab-5c59-8a44-16810da458dd.webp', 'Лапша и рис'),
	-- Холодные закуски и салаты
	((SELECT id FROM rb), 'Огурцы по-тайваньски', 'Хрустящие огурцы в пикантном маринаде из соевого…', 540000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/870d21d3-c326-5f42-bcdc-ae8e55e56994.webp', 'Холодные закуски и салаты'),
	((SELECT id FROM rb), 'Салат с креветками и ананасом', 'Сочный микс салатных листьев дополнен сладким болгарским…', 1190000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6441a4a8-c974-5966-9ead-0d6f05ca36cd.webp', 'Холодные закуски и салаты'),
	((SELECT id FROM rb), 'Китайский салат с баклажанами', 'Нежные баклажаны дополнены хрустящим огурцом и сочными…', 820000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/11677247-d750-5173-9d72-ea4b71aa0046.webp', 'Холодные закуски и салаты'),
	-- Пельмени и спринг-роллы
	((SELECT id FROM rb), 'Жареные вонтоны с крабом и сливочным сыром (6шт.)', 'Хрустящие вонтоны из тонкого пшеничного теста с…', 810000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4b845031-085e-58f5-b0ed-78584661a929.webp', 'Пельмени и спринг-роллы'),
	((SELECT id FROM rb), 'Дим-самы ассорти (8 шт.)', 'Паровые пельмени с морепродуктами (креветки) (2 шт)…', 1090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ef8f1231-dc6b-55b9-812c-d9415f5519c8.webp', 'Пельмени и спринг-роллы'),
	((SELECT id FROM rb), 'Вонтоны в пряном соусе с говядиной (5шт.)', 'Нежные вонтоны из пшеничного теста с начинкой…', 620000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/99d49d12-f178-5421-99a6-723cd1f0cc4e.webp', 'Пельмени и спринг-роллы'),
	-- Уникальные супы
	((SELECT id FROM rb), 'Том Ям', 'Ароматный бульон на основе томленых галангала, лемонграсса…', 980000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0527f308-2c25-5467-8964-c18008552f8a.webp', 'Уникальные супы'),
	((SELECT id FROM rb), 'Кисло-острый суп', 'Насыщенный бульон с яркой кислинкой рисового уксуса…', 540000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/004b1ca9-d4b0-5839-a6ef-9f75513e3135.webp', 'Уникальные супы'),
	((SELECT id FROM rb), 'Целебный суп-бульон с курицей и кореньями', 'Легкий, прозрачный бульон, томленый на курином бедре…', 560000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f2f41a10-8296-5f3e-a8e9-8eba8c3f88ce.webp', 'Уникальные супы'),
	-- Летнее меню
	((SELECT id FROM rb), 'Битая редиска', 'Битый редис пропитывается заправкой из уксуса мицукан…', 490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/16298c7a-af5a-5a3f-a8f1-3758e424ca07.webp', 'Летнее меню'),
	((SELECT id FROM rb), 'Салат Банг-Банг', 'Отварная куриная грудка, хрустящие овощи (морковь, огурец…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/79a09f0f-a786-5e8e-8395-764b265451ce.webp', 'Летнее меню'),
	((SELECT id FROM rb), 'Холодный азиатский суп с лапшой', 'Пшеничная лапша в освежающем бульоне на основе…', 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b579602d-7186-5533-90df-07c3687c31bd.webp', 'Летнее меню'),
	-- Рыба и морепродукты
	((SELECT id FROM rb), 'Ассорти из морепродуктов с овощами', 'Сочетание сочных креветок, нежного кальмара и сладковатых…', 1590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/228a8712-29e3-5b45-9421-85b3376d7042.webp', 'Рыба и морепродукты'),
	((SELECT id FROM rb), 'Креветки в кисло-сладком соусе', 'Креветки в густом кисло-сладком соусе на основе…', 1090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2a54626a-021f-53ca-8224-f0a9750ec05a.webp', 'Рыба и морепродукты'),
	((SELECT id FROM rb), 'Креветки васаби', 'Нежные креветки в хрустящей панировке из муки…', 1090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d6d99bfe-d2d5-50d2-bd6b-ef5b785909bf.webp', 'Рыба и морепродукты'),
	-- Роллы
	((SELECT id FROM rb), 'Угорь с авокадо', 'Водоросли, рис, угорь, авокадо, икра тобико, соус…', 1090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2b5f50f0-2d4d-5590-a7aa-88098c00b295.webp', 'Роллы'),
	((SELECT id FROM rb), 'Ассорти из гунканов (3 шт)', 'Острый лосось (1 шт), острый угорь (1…', 840000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/70aa1940-bd2b-5089-82c9-1b136303b6f5.webp', 'Роллы'),
	((SELECT id FROM rb), 'Дракон с угрем', 'Водоросли, рис, угорь копченый, краб, снежный краб…', 1090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0c3f6770-a89d-523a-af1f-98732b0f345a.webp', 'Роллы'),
	-- Овощи
	((SELECT id FROM rb), 'Тофу в стиле Три стакана', 'Тофу обжаренный до золотистой корочки, томлённый в…', 830000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/63f3d851-69a1-559f-8b0b-f724703f0b6e.webp', 'Овощи'),
	((SELECT id FROM rb), 'Баклажаны в чесночном соусе', 'Нежные баклажаны, томленые до мягкости в насыщенном…', 630000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/780ba626-9540-5ee1-a0cf-de201562d08c.webp', 'Овощи'),
	((SELECT id FROM rb), 'Овощи Восемь сокровищ', 'Богатое ассорти из тофу, древесных грибов, кабачков…', 670000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/40cd07f8-7fa8-57b5-a8da-97dc1ac838f0.webp', 'Овощи'),
	-- Десерты
	((SELECT id FROM rb), 'Жареное молоко', 'Нежные брусочки, обжаренные до золотистой корочки, скрывают…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e6e8ca75-409e-5d4c-a979-8212421aee0e.webp', 'Десерты'),
	((SELECT id FROM rb), 'Китайские пончики с арахисом', 'Нежные рисовые шарики с хрустящей кунжутной корочкой…', 570000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/eaf4265d-46bf-58a0-b82d-7a35609875f6.webp', 'Десерты'),
	((SELECT id FROM rb), 'Китайские пончики с красной фасолью', 'Нежные рисовые шарики с хрустящей кунжутной корочкой…', 570000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/05b3e979-8f47-5896-909f-90bc713f208a.webp', 'Десерты'),
	-- Детское меню
	((SELECT id FROM rb), 'Кукуруза (1 шт)', 'Кукуруза', 400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9bae6a15-e350-54e6-bb58-8f3db07d87d9.webp', 'Детское меню'),
	((SELECT id FROM rb), 'Наггетсы', 'Наггетсы куриные', 400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/beb50b8d-8b9c-56cd-847a-33f36072b12a.webp', 'Детское меню'),
	((SELECT id FROM rb), 'Кетчуп', 'Кетчуп', 80000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4ae82202-3a74-5a53-a48f-fcd0d2ccd647.webp', 'Детское меню'),
	-- Напитки
	((SELECT id FROM rb), 'Добрый кока-кола (250мл.)', '250 мл', 375000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c7c83010-4c77-5816-ae8b-2b8b99a52791.webp', 'Напитки'),
	((SELECT id FROM rb), 'Добрый кока-кола б-с', '250 мл', 375000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7e82c8fa-bd2e-53d1-bff1-93359bc70386.webp', 'Напитки'),
	((SELECT id FROM rb), 'Добрый Апельсин', NULL, 375000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f04cb552-31a0-5650-9545-c5efa44eb93c.webp', 'Напитки'),
	-- Дополнительно
	((SELECT id FROM rb), 'Печенье с предсказанием', 'Хрустящее печенье из пшеничной муки, сахара и…', 70000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c47e8303-f93d-5d3c-85a0-355e014c2af7.webp', 'Дополнительно'),
	((SELECT id FROM rb), 'Луковая лепешка', 'Тонкое тесто из пшеничной муки с добавлением…', 290000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f2590ebc-d07b-5533-9a89-1dafa553d2b1.webp', 'Дополнительно'),
	((SELECT id FROM rb), 'Китайский паровой хлеб Сан Дон', 'Мягкий паровой хлеб на основе пшеничной муки…', 200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3e98bfa6-8f59-5398-9bd5-25b9f457dd65.webp', 'Дополнительно');

-- Tutta La Vita (id 10)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Tutta La Vita' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Пицца на тонком тесте
	((SELECT id FROM rb), 'Пицца Баварская', NULL, 1080000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b6e1afae-ca6d-5a55-9e5e-b3f5a76b961e.webp', 'Пицца на тонком тесте'),
	((SELECT id FROM rb), 'Пицца грибная с трюфельным маслом', NULL, 1170000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e37fa7bb-bb80-574a-9e5a-ef301559f06d.webp', 'Пицца на тонком тесте'),
	((SELECT id FROM rb), 'Пицца Цезарь', NULL, 1050000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f467bcb4-8bf0-5ba2-85db-8bb88d00a061.webp', 'Пицца на тонком тесте'),
	((SELECT id FROM rb), 'Пицца Мясная', NULL, 1220000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3b09b220-7f8c-5b56-8a23-af06173a5bfa.webp', 'Пицца на тонком тесте'),
	-- Домашний хлеб бездрожжевой
	((SELECT id FROM rb), 'Пшеничный хлеб', NULL, 450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2007ffda-3861-5fb2-bcca-da8c3ae0d0ce.webp', 'Домашний хлеб бездрожжевой'),
	((SELECT id FROM rb), 'Бородинский хлеб', NULL, 470000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/32e9791d-3093-5783-b8de-2201e7c387f7.webp', 'Домашний хлеб бездрожжевой'),
	((SELECT id FROM rb), 'Хлеб Бородинский', NULL, 350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/52be8474-0db5-58f1-b63d-ab68c2933af6.webp', 'Домашний хлеб бездрожжевой'),
	-- Салаты
	((SELECT id FROM rb), 'Салат Греческий', NULL, 1000000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c380e9b2-53da-508a-baa5-0cea3f223de9.webp', 'Салаты'),
	((SELECT id FROM rb), 'Цезарь с креветками', NULL, 1080000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dc823690-b604-5ab9-b3c9-dff6e08ff406.webp', 'Салаты'),
	((SELECT id FROM rb), 'Капрезе', NULL, 1080000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9c565b1e-f177-5706-91ac-ecbf4a695222.webp', 'Салаты'),
	-- Закуски
	((SELECT id FROM rb), 'Пармеджано баклажано с соусом из печеного перца', NULL, 1270000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f5ad1775-2644-523e-9686-6072f9a76432.webp', 'Закуски'),
	((SELECT id FROM rb), 'Вителло тоннато по-пьемонтски', NULL, 1370000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1163e8f7-e805-54ea-8d24-0334f9d45e2e.webp', 'Закуски'),
	((SELECT id FROM rb), 'Сельдь с картофелем', NULL, 820000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a93390d4-c413-552d-8b3f-97dd34f04f85.webp', 'Закуски'),
	-- Супы
	((SELECT id FROM rb), 'Суп из щавеля', NULL, 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fca2d609-5764-53db-9385-aa0c797bce88.webp', 'Супы'),
	((SELECT id FROM rb), 'Похлебка Рыбацкая', NULL, 900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/70c2fa38-3b05-50f5-b8ae-8827ceae7ba2.webp', 'Супы'),
	((SELECT id FROM rb), 'Крем-суп из лосося', NULL, 980000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a9a3e92e-8969-5df6-aa77-38740429d96e.webp', 'Супы'),
	-- Паста
	((SELECT id FROM rb), 'Феттуччине с лососем и креветками', NULL, 1180000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/89ca4958-f033-51c1-96b0-6b9b59263b2a.webp', 'Паста'),
	((SELECT id FROM rb), 'Феттуччине с белыми грибами', NULL, 1120000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b690ad49-45e7-55b3-b871-afe7bee88f4b.webp', 'Паста'),
	((SELECT id FROM rb), 'Лазанья с уткой', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8063d7d1-d3f9-5770-b2c1-bbd51f4a4544.webp', 'Паста'),
	-- Бургеры
	((SELECT id FROM rb), 'Шаурма куриная с картофелем фри', NULL, 970000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b0e31b22-7e42-5059-b408-b1d3211b1239.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Биф бургер с беконом с картофелем фри', NULL, 1270000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/619e04cd-f2fc-537f-84d3-786cdd29916a.webp', 'Бургеры'),
	-- Шашлык
	((SELECT id FROM rb), 'Шашлык из курицы с картофелем фри', NULL, 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f32b2fad-343e-5134-8544-339d3df94ee9.webp', 'Шашлык'),
	((SELECT id FROM rb), 'Шашлык из свинины с картофелем фри', NULL, 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bfba1f8f-93c3-5414-83f5-11da891f4727.webp', 'Шашлык'),
	((SELECT id FROM rb), 'Шашлык из баранины с картофелем фри', NULL, 1650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bbf21928-9192-54a1-94bd-37d9bd1456a1.webp', 'Шашлык'),
	-- Рыба и морепродукты
	((SELECT id FROM rb), 'Крабовые кейки с черным рисом и устричным соусом', NULL, 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9f6937a8-2042-5dbd-b82d-46ae947e38a5.webp', 'Рыба и морепродукты'),
	((SELECT id FROM rb), 'Дорадо на открытом огне с овощами', NULL, 1600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2214f439-104a-5f61-b3d6-86f399074d3b.webp', 'Рыба и морепродукты'),
	((SELECT id FROM rb), 'Кальмар, фаршированный моцареллой, рагу из перца', NULL, 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c2095538-8cb3-577d-8b55-a5741f1a4a6b.webp', 'Рыба и морепродукты'),
	-- Стейки
	((SELECT id FROM rb), 'Стейк Ангус с печеным мини-картофелем', NULL, 2250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ce540cf2-26e8-5305-aa95-28e5a61ac9df.webp', 'Стейки'),
	-- Горячие блюда
	((SELECT id FROM rb), 'Бефстроганов из говядины с картофельным пюре', NULL, 1480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/075e6cac-c10a-5d19-aa23-6be4a633a934.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Утиная грудка с картофельным гратеном и грибным', NULL, 1450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f10c3571-6d72-5196-9185-9cdb11bb92a5.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Бифштекс из говядины с яйцом, картофелем пай и', NULL, 1400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/03d90df9-5045-5bcd-bf60-e599eb49face.webp', 'Горячие блюда'),
	-- Десерты
	((SELECT id FROM rb), 'Медовик с кремом из соленой карамели', NULL, 770000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/22bd2a37-a580-5a99-b08a-6e8b9d77dbc8.webp', 'Десерты'),
	((SELECT id FROM rb), 'Профитроли с манговым курдом', NULL, 720000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6e5fa351-2da1-5186-b05f-cef0ee8f596a.webp', 'Десерты'),
	((SELECT id FROM rb), 'Чизкейк с кремом из вареной сгущенки', NULL, 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/333eb58e-3903-5ac9-b28b-37a37f4375d3.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Морс Клюква', NULL, 400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f863586b-e8f0-52ea-bfcc-6bc487d7b47a.webp', 'Напитки'),
	((SELECT id FROM rb), 'Компот из сухофруктов', NULL, 370000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6b5369c7-e836-5e36-80b1-4a29ec7e6acf.webp', 'Напитки'),
	((SELECT id FROM rb), 'Coca-Cola', NULL, 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0b1f97af-64b1-5acc-ba26-7ab2fd7b4a65.webp', 'Напитки'),
	-- Завтрак
	((SELECT id FROM rb), 'Каша овсяная с клубничным соусом', NULL, 620000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bfd6b330-0592-51d7-896a-c63257d3235b.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Каша рисовая с манго', NULL, 620000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ee2f9a56-373a-54b9-9480-950b804e3f31.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Шакшука с сыром фета и гренками', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c085b133-7e38-57cb-8d7a-8f2e34c1d58b.webp', 'Завтрак');

-- KFC (id 11)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'KFC' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Баскеты
	((SELECT id FROM rb), 'Баскет L 24 Острых Крылышка', 'Баскет L с острыми Крылышками — это…', 1077000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ec99d4cf-6947-5d84-a581-9da64507a398.webp', 'Баскеты'),
	((SELECT id FROM rb), 'Баскет М 18 острых крылышек', 'Баскет М с острыми Крылышками в ведерке…', 898000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/be2bea88-ad8d-5ba6-9169-aee147aad77c.webp', 'Баскеты'),
	((SELECT id FROM rb), 'Баскет S 12 Острых Крылышек', 'В Баскете S для вас приготовлены острые…', 687000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dda6818b-5fa7-5462-8fb1-25e2891c5a7c.webp', 'Баскеты'),
	((SELECT id FROM rb), 'Шеф Баскет со стрипсами', 'Что входит в Шеф Баскет со стрипсами?', 402000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f2ffb084-3138-5f1f-acda-41aef9beea12.webp', 'Баскеты'),
	-- Сочная курица
	((SELECT id FROM rb), '3 Оригинальных Ножки', 'Оригинальные куриные Ножки с аппетитно хрустящей корочкой…', 309000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/06929021-7b61-565f-9bb3-9a1c178410eb.webp', 'Сочная курица'),
	((SELECT id FROM rb), '8 Острых Крылышек', 'Что мы о них знаем?', 487000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/49cb7c75-d284-5494-ab53-c366e770add1.webp', 'Сочная курица'),
	((SELECT id FROM rb), '2 Оригинальных Ножки', 'Оригинальные куриные Ножки с аппетитно хрустящей корочкой…', 220000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/725496c5-469d-533c-b6f8-ba4a634b58ec.webp', 'Сочная курица'),
	((SELECT id FROM rb), '3 Острых Крылышка', 'Что мы о них знаем?', 244000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a307e3cd-95b5-5b1a-ade6-09c009ebbcc1.webp', 'Сочная курица'),
	-- Бургеры
	((SELECT id FROM rb), 'Чизбургер де люкс', 'Воздушная глазированная булочка, сочная куриная котлета, свежие…', 152000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dd1c6a99-addf-557a-b6a2-9dbacd9b2761.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Шефбургер Джуниор оригинальный', 'Еще не пробовали ШефБургер Джуниор?', 161000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f51949f5-779a-587a-a4e0-9a19c8e55c64.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Шефбургер Джуниор острый', 'Два сочных стрипса в острой панировке, сочные…', 161000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ba683739-c7bb-5550-a4af-dfaca962f7b1.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Шефбургер Де Люкс оригинальный', 'Булочка с кунжутом, сочное куриное филе в…', 256000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d814426b-20a1-50d8-b368-8334894084e8.webp', 'Бургеры'),
	-- День Рождения Rostic’s
	((SELECT id FROM rb), 'Шефролл оригинальный', 'Шефролл оригинальный — аппетитная и сочная еда…', 199000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3b23d16e-e199-58ee-bce0-589142ead6c1.webp', 'День Рождения Rostic’s'),
	((SELECT id FROM rb), 'Коктейль молочный Шейк Лесная Клубника 0,25 л', 'Хотите освежиться и зарядиться хорошим настроением?', 123000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9f70aae7-8c0b-5c6e-b01b-0a8039617c91.webp', 'День Рождения Rostic’s'),
	((SELECT id FROM rb), 'Коктейль молочный Шейк Лесная Клубника 0,4 л', 'Хотите освежиться и зарядиться хорошим настроением?', 187000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/05533a87-eb80-543b-900b-344cbb3d79e7.webp', 'День Рождения Rostic’s'),
	((SELECT id FROM rb), 'Коктейль молочный Шейк Нежная Ваниль 0,25 л', 'Идеальные ванильные отношения могут быть только с…', 123000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/515f7aed-5ddf-57ec-81be-2b1909075a1a.webp', 'День Рождения Rostic’s'),
	-- Только в доставке
	((SELECT id FROM rb), 'Баскет М х 2', '*Продукция содержит или может содержать ракообразных, рыбу…', 1606000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1886ae2a-8494-5f0b-b5da-63e97f653b82.webp', 'Только в доставке'),
	-- Сэндвичи
	((SELECT id FROM rb), 'Шеф Сэндвич оригинальный', 'Воздушная булочка чиабатта с хрустящей корочкой, сытная…', 568000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/be3c7141-5681-5ae1-bb97-27239f2d4b28.webp', 'Сэндвичи'),
	((SELECT id FROM rb), 'Шеф Сэндвич острый', 'Воздушная булочка чиабатта с хрустящей корочкой, сытная…', 568000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dc154b1c-1eaf-5707-b012-b64b26c85848.webp', 'Сэндвичи'),
	((SELECT id FROM rb), 'Шеф Сэндвич Лайт', 'Воздушная булочка чиабатта с хрустящей корочкой, начинка…', 440000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0648e0e0-5737-5303-bc25-950b4e42e935.webp', 'Сэндвичи'),
	-- Роллы
	((SELECT id FROM rb), 'Шефролл Джуниор', 'Шефролл Джуниор — это сырное наслаждение с…', 168000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e5b1ca61-c3dd-5406-8f6f-610d64364984.webp', 'Роллы'),
	((SELECT id FROM rb), 'Шефролл Де Люкс оригинальный', 'Кусочки нежнейшего куриного филе в хрустящей оригинальной…', 300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8a8a990f-8f73-5572-92bb-33f6a05407f2.webp', 'Роллы'),
	((SELECT id FROM rb), 'Шефролл Де Люкс острый', 'Кусочки нежнейшего куриного филе в хрустящей острой…', 300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d57ae6aa-6609-5d2a-8e28-58a493447482.webp', 'Роллы'),
	((SELECT id FROM rb), 'Ростмастер оригинальный', 'Невероятно вкусный и сытный Ростмастер оригинальный ждет тебя.', 350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/05b60b3e-c951-5221-8361-2f0180642930.webp', 'Роллы'),
	-- Картофель и Снэки
	((SELECT id FROM rb), 'Малый Картофель фри', 'Закажите Картофель фри по вкусной цене!', 104000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0e707f86-cada-592d-a03c-05971dba1a89.webp', 'Картофель и Снэки'),
	((SELECT id FROM rb), 'Баскет Картофель Фри', 'Закажите Картофель фри по вкусной цене!', 258000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/38fb04fa-15b8-5bae-b1fc-503160948b99.webp', 'Картофель и Снэки'),
	((SELECT id FROM rb), 'Средний Картофель по-деревенски', 'Никто не готовит Картофель по-деревенски так, как…', 158000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4791e7fe-81ed-5007-81af-8ce59c8afbf3.webp', 'Картофель и Снэки'),
	((SELECT id FROM rb), 'Средний Картофель фри', 'Закажите Картофель фри по вкусной цене!', 122000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1e548f75-2b82-52cf-ab79-df5567d0fa53.webp', 'Картофель и Снэки'),
	-- Холодные Напитки и Милкшейки
	((SELECT id FROM rb), 'Чай Lipton Зеленый 0,4 л', '*Внешний вид и упаковка продукта могут отличаться…', 155000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bf1692d6-7fe3-54e4-8839-877033abf711.webp', 'Холодные Напитки и Милкшейки'),
	((SELECT id FROM rb), 'Чай Lipton Лимон 0,4 л', '*Внешний вид и упаковка продукта могут отличаться…', 155000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b7ca39c1-d797-5bc3-861e-0dfde5aac121.webp', 'Холодные Напитки и Милкшейки'),
	((SELECT id FROM rb), 'Аква Минерале Актив Цитрус 0,5 л', '*Внешний вид и упаковка продукта могут отличаться…', 160000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8dc1bfdc-ea5b-5aa4-9eab-1c11ffbb1e25.webp', 'Холодные Напитки и Милкшейки'),
	-- Кофе и чай
	((SELECT id FROM rb), 'Кофе Латте малый', 'Больше нежности, больше молока!', 118000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4bdc3385-68bd-5cbe-9c16-2115590930c9.webp', 'Кофе и чай'),
	((SELECT id FROM rb), 'Кофе Латте средний', 'Больше нежности, больше молока!', 158000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/22675d0c-a49d-537e-b9e7-78f744e292fc.webp', 'Кофе и чай'),
	((SELECT id FROM rb), 'Кофе Латте большой', 'Больше нежности, больше молока!', 197000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/70c5a248-7c85-582c-8f40-e803d2cdebe9.webp', 'Кофе и чай'),
	-- Десерты и Мороженое
	((SELECT id FROM rb), 'Донат Клубничный', 'Донат Клубничный – это не просто десерт.', 146000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/06965ac3-1728-5e41-a6c1-c04541780b8e.webp', 'Десерты и Мороженое'),
	((SELECT id FROM rb), 'Мини-тарт карамельно-арахисовый', 'Мини-тарт с карамельно-арахисовой начинкой: хрустящее тесто, сладкая…', 159000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5fce2afb-8171-5dd8-aa57-900a60270eee.webp', 'Десерты и Мороженое'),
	((SELECT id FROM rb), 'Пирожок манго-маракуйя крем-чиз', 'Пирожок с манго-маракуйей и крем-чизом - это…', 117000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6dfe9249-9289-5d33-a7a1-f1c8bd33d88c.webp', 'Десерты и Мороженое'),
	-- Соусы
	((SELECT id FROM rb), 'Кетчуп Томатный', '*Масса/объем, пищевая и энергетическая ценность указана на упаковке.', 62000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6195c942-4e5a-5b02-ac2b-bf1c608f621f.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Барбекю', '*Масса/объем, пищевая и энергетическая ценность указана на упаковке.', 62000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6503eae4-fff2-5c26-8b72-5bd400db2b5b.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Кисло-Сладкий Чили', '*Масса/объем, пищевая и энергетическая ценность указана на упаковке.', 62000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ed63c6e5-a76c-5f12-98f3-49c1a28daa21.webp', 'Соусы');

-- Аист (id 12)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Аист' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Роллы
	((SELECT id FROM rb), 'Ролл филадельфия с лососевой икрой', 'Сыр филадельфия, огурец.', 2750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c95e9677-bdda-5ede-b147-f99cd62534a8.webp', 'Роллы'),
	((SELECT id FROM rb), 'Ролл Калифорния с крабом', 'Рис, краб, авокадо, японский майонез, тобико, нори…', 1950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1fedbdfc-355e-5f45-952e-7c33607ef7b9.webp', 'Роллы'),
	-- Салаты
	((SELECT id FROM rb), 'Овощной салат, семена, подсолнечное масло', 'Огурцы, томаты, слайсы редиса, укроп, петрушка, семена…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/97340d41-a5ac-5daf-a8a8-575cc6b78274.webp', 'Салаты'),
	((SELECT id FROM rb), 'Авокадо, салат романо, семечки', 'Листья салата романо, авокадо, орехи, цукини, лимон…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d3de243c-b33c-5211-8c4b-74e601c0d19c.webp', 'Салаты'),
	-- Мясо и птица
	((SELECT id FROM rb), 'Куриное филе по-милански', 'Куриное филе в панировке, обжаренное на топленом масле.', 1350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b24f1994-a155-5274-9b54-52ebe24265a9.webp', 'Мясо и птица'),
	((SELECT id FROM rb), 'Бефстроганов с картофелем', 'Обжаренная говядина тушится в сметанно-огуречном соусе с…', 1650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e8a8b9e2-d929-55bb-b63b-0bcde269e3ac.webp', 'Мясо и птица'),
	-- Поке
	((SELECT id FROM rb), 'Поке Лосось', 'Лосось, рис, бобы, огурцы веча, красная капуста…', 1550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bf71b7c9-4607-52cc-b5e1-7d4d20c428e0.webp', 'Поке'),
	((SELECT id FROM rb), 'Поке Овощной', 'Рис, микс-салат, гамадари ореховый соус, салат чука…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8a86a4c8-8db0-5bbb-a2a2-57add667ceda.webp', 'Поке'),
	-- Супы
	((SELECT id FROM rb), 'Борщ', 'Говядина, свёкла, картофель, морковь, сметана, укроп, чеснок…', 1150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/36f70f81-5034-57b8-b2f8-c48b0ec3ae53.webp', 'Супы'),
	((SELECT id FROM rb), 'Куриный бульон с лапшой и яйцом', 'Куриный бульон, лапша домашняя, яйцо куриное, куриная…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d68d591b-3673-591f-95fe-053151066228.webp', 'Супы'),
	-- Пицца
	((SELECT id FROM rb), 'Маргарита', 'Томатный соус, моцарелла, орегано, базилик, оливковое масло…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f7ca1c15-5164-5fca-8aa1-c7d709dafad7.webp', 'Пицца'),
	((SELECT id FROM rb), 'Груша, горгонзола, орех пекан', 'Груша, горгонзола, моцарелла, орех пекан (30 см)', 1150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/227fdf14-f6df-5253-bc3f-2796c98cd7c5.webp', 'Пицца'),
	-- Японские закуски
	((SELECT id FROM rb), 'Татаки лосось, кунжут, юдзу', 'Лосось обожженный, соус юдзу, оливковое масло, соевый…', 1250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9874bcf7-6894-5556-80ea-93cff21956de.webp', 'Японские закуски'),
	((SELECT id FROM rb), 'Гребешок, соус понзу, трюфель', 'Нежный морской гребешок под соусом трюфельный понзу.', 1550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/570f1295-2717-502d-a2da-fa931cafd82e.webp', 'Японские закуски'),
	-- Суши Нигири
	((SELECT id FROM rb), 'Суши Лосось', 'Рис, лосось, васаби, лосось смазывается соевым соусом.', 300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e27973f7-39a1-5a06-ab00-5d54fa4f8e11.webp', 'Суши Нигири'),
	((SELECT id FROM rb), 'Суши Угорь', 'Рис, угорь, васаби, угорь смазывается соусом унаги, кунжут.', 250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6884ba02-bbff-57c6-9559-0fe095a3d812.webp', 'Суши Нигири'),
	-- Сашими
	((SELECT id FROM rb), 'Сашими Лосось', 'Свежий лосось (50 грамм), дайкон, лайм', 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f1b78707-8aec-57d4-9b91-60ff0cb6b55f.webp', 'Сашими'),
	((SELECT id FROM rb), 'Сашими Гребешок', 'Свежий гребешок (50 грамм), тобико, дайкон, лимон', 1600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1ce95507-8742-56fb-b808-3dee48d82904.webp', 'Сашими'),
	-- Гунканы
	((SELECT id FROM rb), 'Гункан Угорь', 'Рис, нори, угорь, спаси соус', 300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7595cc95-a0ca-5ad5-aa56-51f0309ef91e.webp', 'Гунканы'),
	((SELECT id FROM rb), 'Гункан Лосось', 'Рис, нори, лосось, спаси соус', 350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a7373217-7821-557a-ae0f-47d5575b1088.webp', 'Гунканы'),
	-- Закуски
	((SELECT id FROM rb), 'Половина бурраты, помидоры, базилик', 'Буррата собственного приготовления, томаты без кожи, оливковое…', 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b064d351-a1c2-5e6a-a0dd-1595f3875a82.webp', 'Закуски'),
	((SELECT id FROM rb), 'Сыр, мёд, конфитюр', 'Сыр азиаго, таледжио, сыр пекорино, романо, горгондзола…', 1750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/06fec1d9-aa51-5403-b08e-b5a1f0b82aa0.webp', 'Закуски'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Стейк из цветной капусты, соус зеленый перец', 'Жареная и запеченная цветная капуста с соусом…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ed74fa17-0a78-5b88-aa11-2216e009fab9.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Скрамбл, черный трюфель', NULL, 1450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/666c7360-87e5-539d-ba22-a09ea78cd31d.webp', 'Горячие закуски'),
	-- Ризотто
	((SELECT id FROM rb), 'Ризотто, белые грибы, травы', NULL, 2150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1222011b-4fbe-5a7c-b29e-3a56bb52e5cd.webp', 'Ризотто'),
	((SELECT id FROM rb), 'Ризотто, креветки, лимон', NULL, 1450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/07359741-ca65-5213-8e9d-5aed56f0c982.webp', 'Ризотто'),
	-- Домашняя паста
	((SELECT id FROM rb), 'Тальолини, сыр, чёрный перец', 'Домашняя паста готовится на основе куриного бульона…', 1250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f0e4d7a7-4bc0-5af4-86e0-98821f662952.webp', 'Домашняя паста'),
	((SELECT id FROM rb), 'Тальолини, помидор, базилик', 'Томатная база и овощной бульон, помидоры двух…', 1150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/659a8fcc-c685-5545-9cf2-7aaf5066eedb.webp', 'Домашняя паста'),
	-- Паста Граньяно
	((SELECT id FROM rb), 'Спагетти с морепродуктами и томатами даттерини', 'Спагетти, соус каччукко, томаты, базилик, петрушка, осьминог…', 3550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7653814e-4f28-5da8-b72d-b13525c47606.webp', 'Паста Граньяно'),
	((SELECT id FROM rb), 'Лингвине, вонголе, петрушка 1 эт', NULL, 1750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/827d8bc5-fe86-51f3-88dc-52e97a61521d.webp', 'Паста Граньяно'),
	-- Крудо
	((SELECT id FROM rb), 'Тартар лосось, лук сибулет', 'Лосось, лук сибулет.', 1950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/52d84c09-809f-52f3-aaf4-e98da88d44bc.webp', 'Крудо'),
	-- Рыба и морепродукты
	((SELECT id FROM rb), 'Вонголе, мартини драй, чеснок', 'Спагетти, соус каччукко, томаты, базилик, петрушка, осьминог…', 1550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1ee81154-033c-5165-bede-84b0994d86d8.webp', 'Рыба и морепродукты'),
	((SELECT id FROM rb), 'Угольная треска, картофель, помидоры, оливки', 'Спагетти, соус каччукко, томаты, базилик, петрушка, осьминог…', 2100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c2ec5f1d-1bee-56e9-a46e-e9d3d1dab6c2.webp', 'Рыба и морепродукты'),
	-- Овощи
	((SELECT id FROM rb), 'Молодой картофель с розмарином', 'Картофель, розмарин, оливковое масло', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e2b7f4fd-c619-5075-8af4-824c60c72d9d.webp', 'Овощи'),
	-- Выпечка
	((SELECT id FROM rb), 'Улитка', 'Свежевыпеченная улитка с изюмом и заварным кремом', 250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ca9ca5c0-b56f-5493-9996-ad140618117b.webp', 'Выпечка'),
	-- Десерты
	((SELECT id FROM rb), 'Картошка', 'Пышно взбитое сливочное масло со сгущенным молоком…', 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3b1a2612-7da7-5b1d-b68e-b34360192a97.webp', 'Десерты'),
	-- Конфеты
	((SELECT id FROM rb), 'Пралине-орех', 'Пралине трюфельное в хрустящей вафле с цельным фундуком', 250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3ad0ad72-601c-5fb7-8954-506e0a9e9fda.webp', 'Конфеты'),
	-- Свежевыжатый сок
	((SELECT id FROM rb), 'Апельсиновый фреш 200мл', 'Свежевыжатый апельсин', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a139c21a-acf3-5a5d-a675-02dabb9a1886.webp', 'Свежевыжатый сок'),
	-- Чай, кофе
	((SELECT id FROM rb), 'Манго-маракуйя', 'Пюре из манго, пюре из маракуйи, семечки…', 1050000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b79e3a8d-df7c-5b70-89fe-496114174b25.webp', 'Чай, кофе'),
	-- Соусы
	((SELECT id FROM rb), 'Соус Аджика', 'Томаты без кожицы, чеснок, кинза, перец чили…', 250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/77c155b7-84f0-5887-b875-962bd2b9c281.webp', 'Соусы');

-- Anna (id 13)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Anna' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Салаты
	((SELECT id FROM rb), 'Зеленый салат', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3e293a31-8c84-5996-909b-f5f22cf26e89.webp', 'Салаты'),
	((SELECT id FROM rb), 'Утиный салат с пряной грушей', NULL, 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/052c0b7c-e765-5854-837c-6f203cb7931c.webp', 'Салаты'),
	((SELECT id FROM rb), 'Нисуаз с артишоками и анчоусами', NULL, 1500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/53ac2910-ea61-5a69-82f8-616d9d8226fd.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с телятиной', NULL, 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/11bb6f2e-96c1-5829-9f20-7bc4f8bbf3c5.webp', 'Салаты'),
	((SELECT id FROM rb), 'Оливье с томленой телятиной', NULL, 900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1d816bda-147c-5f97-b8f0-9401c1af8380.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с крабом', 'С авокадо и грейпфрутом', 1700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3193b18a-c021-5562-8c67-131b48d62508.webp', 'Салаты'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Тигровые креветки', 'В соусе рокфор', 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8d14bbc4-b3fc-50f8-9e7e-34613e9132c0.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Фуа-гра с пряной грушей', NULL, 2350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0c3e66a1-d6da-5911-a2fb-eaf215999087.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Рататуй', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0c5ba868-0498-59c1-a388-b4e2000b6c66.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Гратен картофельный', 'С соусом из пармезана', 720000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/72f53243-6769-55e8-8a92-d6c0d7b2d5e4.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Картофель фри', 'С трюфельным маслом и пармезаном', 600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a9633d8f-7097-5075-a81e-0586faafb3e2.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Спаржа', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d5bea402-6ef7-53c9-abe8-188b40e5750f.webp', 'Горячие закуски'),
	-- Завтрак
	((SELECT id FROM rb), 'Сырники', 'Со сметаной и вареньем из клубники', 680000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4d995363-ef27-52e5-aa2f-751c64f7ac87.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Омлет', NULL, 390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b3332fe8-fb86-576e-be34-748863b4da97.webp', 'Завтрак'),
	-- Горячие рыбные блюда
	((SELECT id FROM rb), 'Черная треска', 'Со спаржей и соусом голландез', 1900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c98dd47a-1e2c-57d9-b4a0-70838057d5a7.webp', 'Горячие рыбные блюда'),
	((SELECT id FROM rb), 'Осьминог гриль с бок-чоем и фенхелем', 'Запеченный на углях, осьминог подается c пикантным…', 2300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/cbb2d9b8-2e8a-5beb-8ae1-7b1b3edd0cef.webp', 'Горячие рыбные блюда'),
	((SELECT id FROM rb), 'Лосось со шпинатом', 'Под икорным соусом', 1900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5e853abd-e9a6-5558-903d-73ad751b8521.webp', 'Горячие рыбные блюда'),
	((SELECT id FROM rb), 'Матлот из черной трески и гребешка', NULL, 1400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/782b7aea-27f2-585f-a62d-335e8d147170.webp', 'Горячие рыбные блюда'),
	-- Горячие мясные блюда
	((SELECT id FROM rb), 'Чизбургер гриль', NULL, 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c02dee41-af21-5b55-9485-04d514bdcb4f.webp', 'Горячие мясные блюда'),
	((SELECT id FROM rb), 'Котлета де-воляй', 'Котлета де-воляй — горячее мясное блюдо, в…', 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1782dde4-1124-51cd-882c-97c225ec5e10.webp', 'Горячие мясные блюда'),
	((SELECT id FROM rb), 'Бёф-бургиньон (на 2 персоны)', 'Бёф-бургиньон для двоих – классическое французское горячее…', 2300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6d7d29fa-0475-5b2d-899f-551224ddc3db.webp', 'Горячие мясные блюда'),
	((SELECT id FROM rb), 'Антрекот из говядины под соусом фюме', 'Сочное говяжье мясо, обжаренное до золотистой корочки…', 1700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c1a957cf-43c8-5be5-85da-02b29cc8e551.webp', 'Горячие мясные блюда'),
	-- Супы
	((SELECT id FROM rb), 'Буйабес', 'Густой, ароматный и насыщенный суп — классика…', 1500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c4d1cde4-11a5-5a5d-a16d-128b35d8331b.webp', 'Супы'),
	((SELECT id FROM rb), 'Луковый, с запеченным тостом под сыром грюйер', NULL, 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/742c87ab-fd99-5735-bf0e-05ec8361d804.webp', 'Супы'),
	((SELECT id FROM rb), 'Грибной крем-потаж с крутонами', NULL, 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/05aedadc-cb52-5e7b-b0e4-b1a3fd4aed64.webp', 'Супы'),
	((SELECT id FROM rb), 'Бульон куриный', NULL, 800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c2139f2f-bec2-500a-b1bd-e4c8e565e7b0.webp', 'Супы'),
	-- Холодные закуски
	((SELECT id FROM rb), 'Пате из фуа-гра, бриошь, желе из игристого', 'Изящная холодная закуска с нежным паштетом из…', 1750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a3ce0ea3-c9d4-5911-9608-bf13c848d820.webp', 'Холодные закуски'),
	((SELECT id FROM rb), 'Вителло тонато с руколой', 'Холодная закуска из телятины с лёгким и…', 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f52478d6-c957-5371-8463-94ff00d839ce.webp', 'Холодные закуски'),
	((SELECT id FROM rb), 'Риет из осьминога', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5cf776d3-0a5c-5fcf-a3b5-8d1bdff22f93.webp', 'Холодные закуски'),
	((SELECT id FROM rb), 'Сыры', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/03447555-f5c5-5390-a00b-69a76c2b99de.webp', 'Холодные закуски'),
	-- Паста
	((SELECT id FROM rb), 'Пенне с томленой говядиной', NULL, 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5e825186-b0dc-53b6-8e05-cd4d83c4b0a0.webp', 'Паста'),
	((SELECT id FROM rb), 'Орзо по-провански', NULL, 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6ac95f3d-8327-574a-b58c-b0ec240bec0b.webp', 'Паста'),
	-- Десерты
	((SELECT id FROM rb), 'Профитроли с горячим шоколадом', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3f92bed1-aea2-5c95-85aa-6fe9902bfeab.webp', 'Десерты'),
	((SELECT id FROM rb), 'Креп сюзетт', NULL, 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/be444170-6f1c-5d75-ab6b-cdc63664f18f.webp', 'Десерты'),
	((SELECT id FROM rb), 'Трюфельная конфета', NULL, 300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/01cd8c34-3b68-5141-9a74-f8b9d617c292.webp', 'Десерты'),
	-- Сэндвичи
	((SELECT id FROM rb), 'Крок Мадам', NULL, 960000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ccd32a23-2182-5e7d-ac80-39b8e71897e0.webp', 'Сэндвичи'),
	((SELECT id FROM rb), 'Крок Месьё', NULL, 930000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/35d2fdd3-0a7e-507d-b9a4-9364777f803f.webp', 'Сэндвичи'),
	((SELECT id FROM rb), 'Крок Пацан', NULL, 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e30da9b6-0c5c-52b6-818c-2ccf2451faab.webp', 'Сэндвичи'),
	((SELECT id FROM rb), 'Крок ЗОЖ', NULL, 820000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7fc1340b-87a4-5d92-90e0-f1227526aa59.webp', 'Сэндвичи'),
	((SELECT id FROM rb), 'Крокчик', NULL, 800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4e43d7d2-1596-5b98-b937-76e2fb75ec8b.webp', 'Сэндвичи');

-- Честная Рыба (id 15)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Честная Рыба' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Роллы
	((SELECT id FROM rb), 'Филадельфия', 'Охлаждённая сёмга, огурец, кремчиз, соус терияки, рис, нори', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3ba0a812-f2b8-5671-a827-e68a73811762.webp', 'Роллы'),
	((SELECT id FROM rb), 'Калифорния', 'Снежный краб сурими в соусе майо, авокадо…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bca01d8e-428f-5b92-a838-577af062aef1.webp', 'Роллы'),
	((SELECT id FROM rb), 'Опаленный крем-угорь', 'Внимание: это не горячие роллы.', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ffcb40c2-4e9d-52c9-b2f4-8204b8a5ff4f.webp', 'Роллы'),
	((SELECT id FROM rb), 'Много рыбы', 'Подкопченная форель, креветка попкорн, вьетнамский тунец, соус…', 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/56f60a18-6e42-53b0-926a-604adefbef88.webp', 'Роллы'),
	-- Устрицы, ежи, крабы, гребешки
	((SELECT id FROM rb), 'Морской еж', NULL, 470000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9f742d61-bbec-5e03-9731-dc53c42fde19.webp', 'Устрицы, ежи, крабы, гребешки'),
	((SELECT id FROM rb), 'Гребешки на огне 3 шт', 'Гребешок, мисо масло, микрозелень', 1080000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/224406dd-8371-5aa2-b496-d9140f0d3a66.webp', 'Устрицы, ежи, крабы, гребешки'),
	((SELECT id FROM rb), 'Гребешок сашими в ракушке', 'Гребешок, соус понзу, микрозелень', 290000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/07c8aa6f-3ee6-591a-913e-45b4a6f36e22.webp', 'Устрицы, ежи, крабы, гребешки'),
	((SELECT id FROM rb), '12 шт Пафосные устрицы', 'Что поймали: джоли, дибба бэй, касабланка и другие', 5780000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3ca9dcce-f977-53c5-be87-d218ee493f6f.webp', 'Устрицы, ежи, крабы, гребешки'),
	-- Суши
	((SELECT id FROM rb), 'Охлажденный лосось суши 3 шт', 'Охлажденный лосось, рис, соус терияки', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/183b0724-02df-586a-9351-8f415ef4c171.webp', 'Суши'),
	((SELECT id FROM rb), 'Хендроллы с гребешком', 'Гребешок, зелёный лук, трюфельный соус, кунжут, рис, нори', 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/644d70fe-dc9c-584b-98f6-18aff25b87ef.webp', 'Суши'),
	((SELECT id FROM rb), 'Хендроллы с лососем', 'Лосось, подкопченная форель, рис, нори, соус спайси…', 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/07ca546d-0131-545e-b688-3055ad6addae.webp', 'Суши'),
	((SELECT id FROM rb), 'Форель и чеддер суши 3 шт', 'Подкопченная форель, сыр чеддер, соус терияки, фурикакэ, рис', 490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/867a705e-5e28-51b6-8afe-bbb513ccfc82.webp', 'Суши'),
	-- Супы
	((SELECT id FROM rb), 'Суп Том-Ям', 'Внимание: возможно попадание осколков раковин и жемчужин.', 800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8b3be9fa-07a5-50b8-a7b6-99392be3f1b4.webp', 'Супы'),
	((SELECT id FROM rb), 'Суп Честный Рамен', 'Креветки ваннамей, мидии, грибы муэр маринованные, яйцо…', 800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/eb1cf625-cefe-560a-845b-d3ca83077283.webp', 'Супы'),
	((SELECT id FROM rb), 'Честная уха', 'Треска, картофель, морковь, чесночное масло, рыбный, бульон…', 480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/09970199-4945-58bf-b4f5-e29aa64bfb37.webp', 'Супы'),
	((SELECT id FROM rb), 'Том ям с лососем', 'Обжаренный лосось, грибы вешенки, томаты черри, бульон…', 800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bccc80fa-2a9a-5d36-947a-c1fc0660c24f.webp', 'Супы'),
	-- Закуски
	((SELECT id FROM rb), 'Бургер с лососем и трюфельным соусом', 'Картофельная бриошь, котлета из лосося и кеты…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3961dd8c-df91-5018-84cc-d990152023cc.webp', 'Закуски'),
	((SELECT id FROM rb), 'Сашими Гребешок', 'Гребешок, лист шисо, лимон', 660000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/816c742a-f8d6-52ef-b5e7-d84467b1455b.webp', 'Закуски'),
	((SELECT id FROM rb), 'Сашими Лосось', 'Охлаждённый лосось, лист шисо, лимон', 610000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4ed335ed-b750-5b38-a9ef-05a47f1acca0.webp', 'Закуски'),
	((SELECT id FROM rb), 'Сашими Угорь', 'Угорь жареный, лист шисо, лимон, соус унаги', 470000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7759f775-8fbe-5eaf-9582-1a74ff4c86b8.webp', 'Закуски'),
	-- Мидии
	((SELECT id FROM rb), 'Мидии сливочно-чесночные', 'Внимание: возможно попадание осколков раковин и жемчужин.', 1340000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3a2fe817-0cf4-580f-9e4e-d8ba3d4e0869.webp', 'Мидии'),
	((SELECT id FROM rb), 'Мидии блючиз', 'Внимание: возможно попадание осколков раковин и жемчужин.', 1340000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/99b11afd-49ba-5893-ba79-3c627899975d.webp', 'Мидии'),
	((SELECT id FROM rb), 'Мидии том ям', 'Внимание: возможно попадание осколков раковин и жемчужин.', 1340000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4881726f-433d-5632-9246-33b7b14ae295.webp', 'Мидии'),
	((SELECT id FROM rb), 'Хлебная корзина', 'Ломтики белого и черного багета с чесночным маслом', 240000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ba37e307-5f22-54f5-9472-5403810e3747.webp', 'Мидии'),
	-- Паста
	((SELECT id FROM rb), 'Ригатони с пармезаном', 'Пармезан, паста ригатони', 310000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a329b229-194e-50af-94d6-96c8e9059778.webp', 'Паста'),
	((SELECT id FROM rb), 'Паста с мидиями и креветками', 'Внимание: возможно попадание осколков раковин и жемчужин.', 920000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1f560bc3-271a-5533-b8e1-312cd12fe6e8.webp', 'Паста'),
	((SELECT id FROM rb), 'Паста с крабом и трюфельным соусом', 'Мясо камчатского краба, паста ригатони, трюфельный соус, микрозелень', 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/37475fee-5de7-519d-ba95-335d2797eca6.webp', 'Паста'),
	((SELECT id FROM rb), 'Паста с тунцом и сливками', 'Тунец, паста ригатони, бобы эдамаме, сливки, соус…', 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/906464a4-152a-5b02-a7ea-2650e62993b8.webp', 'Паста'),
	-- Основные
	((SELECT id FROM rb), 'Треска с пюре и вешенками', 'Треска, грибы вешенки, картофельное пюре, сливки, сливочный…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0692bb82-d8ed-542d-8374-2bf1fac7a6dd.webp', 'Основные'),
	((SELECT id FROM rb), 'Тунец с авокадо гриль', 'Тунец, авокадо, огурец, микс салата, соус терияки…', 1050000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c290ce04-2c63-5f09-9264-69732904c0a1.webp', 'Основные'),
	((SELECT id FROM rb), 'Лангустины с чесноком', 'Аргентинские креветки, чесночное масло, петрушка, лимон', 1030000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7c434df9-1e18-5194-9372-f8d988fb2a08.webp', 'Основные'),
	-- Десерты
	((SELECT id FROM rb), 'Эклер амаретто-миндаль', NULL, 480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bb65d70a-c07d-5fa2-af75-8f74e2ca2cd0.webp', 'Десерты'),
	-- Гарниры
	((SELECT id FROM rb), 'Трюфельное пюре', 'Картофельное пюре, трюфельный соус', 310000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f8140671-e24b-554d-a604-701165cc495d.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Картофельное пюре', 'Картофельное пюре, пармезан', 270000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e6e395ce-63b8-5c11-be1e-f9d2a9ee181b.webp', 'Гарниры'),
	-- Напитки
	((SELECT id FROM rb), 'Ягодный негрони', 'Арбуз, ягоды, игристое', 670000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/41c72224-53b2-5cb8-8789-174fe66488bb.webp', 'Напитки'),
	((SELECT id FROM rb), 'Честный Лимонад 0,4', 'Мята, лимон, содовая', 480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ba4ddff2-37cc-5fc6-b026-46feca77bb4f.webp', 'Напитки'),
	((SELECT id FROM rb), 'Апельсиновый фреш 0,25л', NULL, 510000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/203da5b6-0668-5d4a-954b-defe2069967c.webp', 'Напитки'),
	-- Завтраки до 15:00
	((SELECT id FROM rb), 'Дальневосточный омлет', 'Мясо камчатского краба, томат, яйцо куриное, микс…', 780000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6f9fe19b-d5fb-5d81-8a74-a4788fe4df07.webp', 'Завтраки до 15:00'),
	((SELECT id FROM rb), 'Драники с лососем', 'Опаленный лосось, драники, яйцо куриное, соус терияки…', 780000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8d8481ca-dfe5-53c4-8cbf-5be059c9b832.webp', 'Завтраки до 15:00'),
	((SELECT id FROM rb), 'Бейгл с лососем', 'Охлажденный лосось в медово-горчичном соусе, яйцо куриное…', 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6d73e125-20eb-5fa9-a2e1-e77cfed206bd.webp', 'Завтраки до 15:00');

-- DiDi (id 16)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'DiDi' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Выпечка
	((SELECT id FROM rb), 'Хачапури по-мегрельски', 'Закрытый пирог круглой формы из дрожжевого теста…', 910000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/cf8bd01d-169a-5176-b035-e91c84e7401d.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Хачапури по-аджарски', 'Открытый, как большое грузинское сердце, хачапури-лодочка с…', 860000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c17a4a24-c15b-5159-b686-bf5388ccfd7b.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Хачапури с грушей и горгонзолой', 'Хачапури с домашним сыром сулугуни.', 1050000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/53e5cd9f-dc3d-5132-bef2-f8c2a08db934.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Мчади 3 шт.', NULL, 350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1f199c7a-a121-5e71-b790-a70c69cd1705.webp', 'Выпечка'),
	-- Шашлык & Гриль
	((SELECT id FROM rb), 'Шашлык из свинины', 'Сочный шашлык из свиной шеи, замаринованной в…', 1190000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e8ace8dd-4711-5007-a65c-f1a9de62b709.webp', 'Шашлык & Гриль'),
	((SELECT id FROM rb), 'Шашлык из курицы', 'Шашлык из сочного куриного бедра, маринованный в…', 980000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/23c59156-553c-547f-9b82-edda07af7100.webp', 'Шашлык & Гриль'),
	((SELECT id FROM rb), 'Люля-кебаб из курицы', 'Сочный люля-кебаб из рубленого куриного бедра, зажаренный…', 980000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b6f809fa-d83f-5d09-bf74-e91f73119a77.webp', 'Шашлык & Гриль'),
	((SELECT id FROM rb), 'Люля-кебаб из телятины', NULL, 1390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/21b31582-5492-56b1-b539-ae2d2cb6fad9.webp', 'Шашлык & Гриль'),
	-- Салаты
	((SELECT id FROM rb), 'Салат по-грузински с орехами', 'Салат по-грузински из свежих овощей и ароматной зеленью.', 920000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5bf6a684-28d1-58e8-945f-5db648bf96c3.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с креветками и копченым сулугуни', NULL, 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2a845b3d-3203-5d11-a218-c04654eace87.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с ростбифом и томатами', 'Свежие листья салата с добавлением ростбифа и…', 930000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/407b855d-b56e-579a-bdc8-715d7fe2c1de.webp', 'Салаты'),
	-- Закуски
	((SELECT id FROM rb), 'Гебжалия', NULL, 720000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4f8cee93-bbd1-5358-aea0-7a3e876253b2.webp', 'Закуски'),
	((SELECT id FROM rb), 'Пхали из болгарского перца', NULL, 600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8344122d-5075-5000-80f8-af839976ab0b.webp', 'Закуски'),
	((SELECT id FROM rb), 'Ростбиф в медово-горчичном соусе', 'Закуска «Ростбиф в медово-горчичном соусе» — это…', 810000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2b0d836c-2775-5f22-86ad-1266c14bd13d.webp', 'Закуски'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Долма со свининой и говядиной', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/338add14-76c0-5eed-a174-e9f8be30d782.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Батат фри с сырным соусом', 'Батат панируется в кукурузном крахмале и обжаривается…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/df19ac82-b97a-59f1-be95-26e0a72b9977.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Креветки в кляре с соусом тобико', 'Горячая закуска с хрустящей корочкой и сочной начинкой.', 1250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/22e4c605-99de-57f5-b52b-9f5dd0e089ae.webp', 'Горячие закуски'),
	-- Супы
	((SELECT id FROM rb), 'Грузинский рыбный суп', NULL, 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0fb66cfd-ca28-565a-b62e-25b530117207.webp', 'Супы'),
	((SELECT id FROM rb), 'Крем-суп из тыквы', NULL, 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2dfc9400-6a4e-5091-8cf8-021c5d236d3b.webp', 'Супы'),
	((SELECT id FROM rb), 'Хашлама', NULL, 1350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9398836f-02c7-5d7e-bda5-0be598454228.webp', 'Супы'),
	-- Хинкали
	((SELECT id FROM rb), 'Хинкали свинина-говядина', 'Большие хинкали со свининой, говядиной, зеленью и…', 630000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/16e74bf4-fc0a-5c4e-8e0c-2acf01c49f94.webp', 'Хинкали'),
	((SELECT id FROM rb), 'Хинкали с бараниной', 'Большие хинкали с сочной рубленой бараниной, зеленью…', 630000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f02aa488-e01b-5aba-a169-e4cd6bb15a6b.webp', 'Хинкали'),
	((SELECT id FROM rb), 'Хинкали с телятиной', 'Большие хинкали с рубленой телятиной, зеленью и…', 630000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6604e79f-8a57-5bb8-b112-e79699d943a8.webp', 'Хинкали'),
	-- Горячие блюда
	((SELECT id FROM rb), 'Барабулька с соусом ткемали', NULL, 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/72007e4b-b79d-542a-a9f5-80061616c294.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Баклажаны с мясом по-рачински', 'Погрузитесь в мир ярких вкусов с нашим…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d5622e94-ba72-543f-8cbb-5c9cd42d5a93.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Бифштекс', NULL, 1350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8abdf2f4-8397-5b46-8df9-e7cb0133ad11.webp', 'Горячие блюда'),
	-- Гарниры
	((SELECT id FROM rb), 'Картофельное пюре', NULL, 450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fbd4a4aa-2b2a-5582-a5c8-72eef27994a4.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Гречка с грибами', NULL, 320000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/10479e42-7614-5dc9-b51a-d65665e32e8e.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Рис басмати', 'Отварной рис', 300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/efd284ee-ecf8-5968-9dad-6a55d5408876.webp', 'Гарниры'),
	-- Десерты
	((SELECT id FROM rb), 'Рулет Фисташковый', 'Фисташковый рулет с начинкой из свежей малины…', 810000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/48536864-6625-5436-bd94-c9cbd68f2d59.webp', 'Десерты'),
	((SELECT id FROM rb), 'Банановый торт', 'Банановый торт на основе песочного теста с…', 600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ebb7e8c8-c0c0-5009-880b-37dedc51c061.webp', 'Десерты'),
	((SELECT id FROM rb), 'Фисташковая пахлава', 'Национальный восточный десерт из слоеного теста с…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d13e3b29-75de-53c7-ad88-8844fedd09bc.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Домашний лимонад Тархун', NULL, 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7be41e71-abe9-57fd-bf0b-2cd64ddb7ec1.webp', 'Напитки'),
	((SELECT id FROM rb), 'Домашний лимонад Классический', NULL, 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/da7c44c7-e6a6-51f8-96dc-6cd787db551a.webp', 'Напитки'),
	((SELECT id FROM rb), 'Домашний лимонад Груша', NULL, 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2840f1ae-93ef-5dd4-8441-79f1a3146022.webp', 'Напитки'),
	-- Соусы
	((SELECT id FROM rb), 'Сметанно-чесночный соус', NULL, 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5dab2723-4ec8-5880-b790-fab343116b0c.webp', 'Соусы'),
	((SELECT id FROM rb), 'Ткемали', NULL, 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/37ed62db-260e-5934-a6ad-8dd71e36a336.webp', 'Соусы'),
	((SELECT id FROM rb), 'Наршараб', NULL, 250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/114958b9-727b-54ad-aa6e-c8c26efd88e7.webp', 'Соусы'),
	-- Завтрак
	((SELECT id FROM rb), 'Сырники с рикоттой', 'Воздушные сырники из нежнейшего творога.', 600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/04952ae1-c759-5d86-abb4-11bfc46b7455.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Оладьи с домашним вареньем и сметаной', 'Воздушные оладьи, подаются с вареньем из черешни…', 500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4c5ab04a-56d0-5221-ad6b-52080fe78100.webp', 'Завтрак');

-- Eshak (id 17)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Eshak' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Плов
	((SELECT id FROM rb), 'Плов Праздничный', 'Баранина, рис, морковь, лук, нут, изюм, куркума…', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6b609705-06ba-5482-99ed-c807c5862bd2.webp', 'Плов'),
	((SELECT id FROM rb), 'Ачичук к плову', 'Помидоры, базилик, перец чили, лук', 170000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/29a7abc4-e8be-5ccd-8d70-536df2d45dd4.webp', 'Плов'),
	((SELECT id FROM rb), 'Казы к плову', 'Казы, красный лук, укроп, петрушка', 305000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/70595029-7395-5cfa-b870-c0d1c4524401.webp', 'Плов'),
	-- Гриль
	((SELECT id FROM rb), 'Шашлык куриный', 'Куриное мясо, красный лук, армянский лаваш, петрушка…', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d4fee3a4-d697-5c0b-b4fd-305a39600e35.webp', 'Гриль'),
	((SELECT id FROM rb), 'Шашлык из телячьей вырезки', 'Телятина, армянский лаваш, лук, специи', 1450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/19d81d16-21f5-5084-bbe1-1e6e0f887544.webp', 'Гриль'),
	((SELECT id FROM rb), 'Шаурма классическая с курицей', 'Мясо куриное, армянский лаваш, паприка, огурцы, зеленый…', 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8fc6a70d-9f79-52ee-bba9-63366a4653f5.webp', 'Гриль'),
	-- Салаты
	((SELECT id FROM rb), 'Ачичук', 'Помидоры, базилик, перец чили, лук', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/444eabf0-c5ad-5f83-ab03-e3b7bd701d29.webp', 'Салаты'),
	((SELECT id FROM rb), 'Зеленый салат из овощей', 'Брокколи, авокадо, огурцы, бобы эдамаме, цукини, капуста…', 675000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6882b2e0-8fb1-5674-a652-0690845e264c.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с авокадо, шпинатом и креветками', 'Креветки, авокадо, соус чили свит, помидоры, мини-шпинат…', 980000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7135c84a-abe2-5765-9e17-7bf5b78f5c3f.webp', 'Салаты'),
	-- Тесто
	((SELECT id FROM rb), 'Хинкали телятина 3 шт', 'Телятина, лук, кинза, кориандр, мацони, специи', 520000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e30f5e7a-d762-5b4d-953d-c24509e34acb.webp', 'Тесто'),
	((SELECT id FROM rb), 'Пельмени жареные с телятиной', 'Говядина, лук, зира, кориандр, соус из сметаны…', 540000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9bc7baa4-b69c-52ac-a8d6-8f07dc204e99.webp', 'Тесто'),
	((SELECT id FROM rb), 'Тандырная лепешка', 'Мука, дрожжи, масло, сахар, кунжут', 180000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8b1f37bd-698b-5a7f-a9d5-e829b56f517a.webp', 'Тесто'),
	-- Супы
	((SELECT id FROM rb), 'Лагман', 'Баранина, лапша, пекинская капуста, болгарский перец, фасоль…', 710000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/adb22212-711e-5fc0-bd16-838001cf6232.webp', 'Супы'),
	((SELECT id FROM rb), 'Борщ', 'Говядина, картофель, болгарский перец, помидоры, морковь, свекла, сметана', 620000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7633ec38-673b-53b5-b23b-3b4b15aed017.webp', 'Супы'),
	((SELECT id FROM rb), 'Том Ям', 'Креветки, паста том ям, кокосовое молоко, томаты…', 840000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1852759c-b79c-5a3f-84a5-b238231b9aba.webp', 'Супы'),
	-- Летние супы
	((SELECT id FROM rb), 'Окрошка на тане с форелью', 'Тан, сметана, горчица дижонская, лосось юкола, картофель…', 1450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4599f6f8-1213-5fea-8e0f-e0ad0933b71a.webp', 'Летние супы'),
	((SELECT id FROM rb), 'Окрошка на тане', 'Квас, яйцо, сливочный хрен, горчица дижон, укроп…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5984a221-3f68-57fa-a5b5-7122c0b58ec4.webp', 'Летние супы'),
	((SELECT id FROM rb), 'Окрошка на квасе', 'Тан, сметана, горчица дижонская, язык говяжий, редис…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ab7805c9-8b56-58ea-8227-c335470c8a86.webp', 'Летние супы'),
	-- Закуски
	((SELECT id FROM rb), 'Ассорти из свежих овощей', 'Помидоры, огурцы, болгарский перец, редис, соус из…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/093b258e-41b0-5956-a696-02af6b375f53.webp', 'Закуски'),
	((SELECT id FROM rb), 'Баклажаны востока', 'Баклажаны, творожный сыр, чеснок, помидоры, соус песто…', 755000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3a046763-49ac-5b06-a02c-81e9f5ab4e6f.webp', 'Закуски'),
	-- Деликатесы из конины
	((SELECT id FROM rb), 'Казы', 'Казы, красный лук, укроп, петрушка', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/56077bc8-b7b5-51b1-b126-cd522a9ca0b9.webp', 'Деликатесы из конины'),
	((SELECT id FROM rb), 'Салат оливье с тушеной кониной', 'Конина, яйцо, картофель, морковь, огурцы, горошек, салат…', 760000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c597efe7-29b0-5b53-b037-bfb21dc101e6.webp', 'Деликатесы из конины'),
	-- Рыба и морепродукты
	((SELECT id FROM rb), 'Креветки васаби', 'Креветки, икра тобико, соус, васаби', 820000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/70bad9de-f184-57e5-abd1-3355e7871a56.webp', 'Рыба и морепродукты'),
	((SELECT id FROM rb), 'Креветки том ям', 'Креветки, соус том ям, соус соевый, икра тобико', 820000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4d16728b-e330-5fd1-9345-9ce95ed94ba2.webp', 'Рыба и морепродукты'),
	-- Мясные блюда
	((SELECT id FROM rb), 'Долма', 'Говядина, рис, виноградные листья, лук, помидоры, укроп…', 725000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a76c691d-05cb-5722-894f-6e15757d28e4.webp', 'Мясные блюда'),
	((SELECT id FROM rb), 'Котлеты по-домашнему с картофельным пюре', 'Куриный фарш, картофель, лук, сливки, соус с…', 780000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/63cfc1a9-68c3-5678-92f0-41a93146af4d.webp', 'Мясные блюда'),
	-- Гарниры
	((SELECT id FROM rb), 'Картофель печеный со сметаной', 'Картофель, соус из сметаны с чесноком, лук…', 390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e0c75e08-a507-52a6-8312-f5d284850dc9.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Картофель Фри', 'Картофель фри, соль, паприка', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e4921bae-c23b-5ff0-9787-132d7005bb44.webp', 'Гарниры'),
	-- Блюда к пиву
	((SELECT id FROM rb), 'Магаданские креветки', 'Креветки, соус, лимон', 675000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/094ce478-e8a0-5e43-b546-f85794ade4f9.webp', 'Блюда к пиву'),
	((SELECT id FROM rb), 'Куриные крылья с соусом маковый блю чиз', 'Куриные крылья в глазури из сладкого соуса…', 830000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e850a6fb-3da4-5540-81d7-3e6bc9cd195f.webp', 'Блюда к пиву'),
	-- Соусы
	((SELECT id FROM rb), 'Песто', 'Базилик, чеснок, соль, оливковое масло', 290000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/db556e9b-45e0-5291-8082-91f57ac5013f.webp', 'Соусы'),
	((SELECT id FROM rb), 'Аджика', 'Болгарский перец, яблоко, помидоры, лук, чеснок, соус…', 190000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/db556e9b-45e0-5291-8082-91f57ac5013f.webp', 'Соусы'),
	-- Детское меню
	((SELECT id FROM rb), 'Борщ со сметаной', 'Говяжий бульон, капуста, картофель, свекла, репчатый лук…', 290000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/04a91578-ca4f-5326-8dee-5cd50fdf350c.webp', 'Детское меню'),
	((SELECT id FROM rb), 'Бургер с куриной котлетой и фри', 'Булочка с кунжутом, куриная котлета, салат ромейн…', 520000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3ac300cf-93eb-5b76-8084-912e9a91378e.webp', 'Детское меню'),
	-- Десерты
	((SELECT id FROM rb), 'Восточный базар', 'Грецкий орех, миндаль, изюм, курага, финики, мята', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8797a042-5f6e-5e0a-ada1-96f8700bf97a.webp', 'Десерты'),
	((SELECT id FROM rb), 'Домашние орешки', 'Песочное тесто, сгущенное молоко, мята', 480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5588d25e-11d0-5b49-b805-dce1478e85d6.webp', 'Десерты'),
	-- Горячие напитки
	((SELECT id FROM rb), 'Облепиха цитрус', 'Яркий и солнечный напиток с сочной облепихой…', 620000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/79cce200-cfd3-5e2f-81bc-c1eeb69e6c49.webp', 'Горячие напитки'),
	((SELECT id FROM rb), 'Пряная малина', 'Насыщенный и согревающий чай на основе каркадэ…', 620000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b9eafffd-a2c7-5edd-a982-d18d6bf95ddb.webp', 'Горячие напитки'),
	-- Напитки
	((SELECT id FROM rb), 'Фреш Апельсиновый', 'Сок из свежевыжатых апельсинов', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/81ad2822-3a0f-5125-9d53-f55f006df62f.webp', 'Напитки'),
	((SELECT id FROM rb), 'Фреш Ананасовый', 'Сок из свежевыжатых ананасов', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/aac126c3-3d4d-5af1-93e4-c9d7dbe2b624.webp', 'Напитки');

-- Izumi (id 18)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Izumi' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Рис и Донбури
	((SELECT id FROM rb), 'Чикин кацу карэ райсу', 'Рис, куриная отбивная в хрустящей панировке, японский…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9f12202f-4b61-5277-b6f9-60c2d14cfb10.webp', 'Рис и Донбури'),
	((SELECT id FROM rb), 'Унагидон', 'Рис, японский угорь, сычуаньский перец, кунжут, зеленый…', 940000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/602df85d-5279-5755-9b02-7e7542c0becc.webp', 'Рис и Донбури'),
	((SELECT id FROM rb), 'Оякидон', NULL, 660000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8aba61fc-98a5-57e9-ab0b-bf99b815763e.webp', 'Рис и Донбури'),
	-- Закуски
	((SELECT id FROM rb), 'Такояки с осьминогом', 'Шарики из теста с начинкой из мини-осьминога…', 780000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/eae244ed-e824-55b3-a976-14e6cc8dc3e5.webp', 'Закуски'),
	((SELECT id FROM rb), 'Карааге курица', 'Жареное куриное бедро маринованное в саке, мирине…', 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fc6b40fc-64be-5c34-abd8-a419b389578e.webp', 'Закуски'),
	((SELECT id FROM rb), 'Гедза с тигровыми креветками', 'Тесто ручной работы, фарш из тигровых креветок…', 660000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f476cc15-d182-562c-9bb1-5247b458146f.webp', 'Закуски'),
	-- Тэйсёку
	((SELECT id FROM rb), 'Тейсеку-бокс Чикин кацу Карэ райсу', 'Куриное филе в паноровке, рис, мисо-суп, кимпира…', 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e0d4266a-c539-56c3-b470-7e100fcebdc1.webp', 'Тэйсёку'),
	((SELECT id FROM rb), 'Тейсеку-бокс Ханбага', 'Говяжья котлета хамбага с дайконом и соусом…', 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c637644c-9af4-55a4-a20a-b28b90cd0e18.webp', 'Тэйсёку'),
	((SELECT id FROM rb), 'Тейсеку-бокс Курица терияки', 'Курица терияки, рис, кимпира, маринванный дайкон, жареные…', 1270000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d6ebce16-d049-5ec7-bb09-cae81ca067f4.webp', 'Тэйсёку'),
	-- Рамен
	((SELECT id FROM rb), 'Камо рамен большой', 'Куриный бульон с соевым соусом, пшеничная лапша…', 920000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/06649bad-e171-5224-b201-c6035fcbbb12.webp', 'Рамен'),
	((SELECT id FROM rb), 'Хотатэ рамен', 'Сырный бульон, пшеничная лапша, морской гребешок, ошпаренный…', 1155000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ea20bec5-e46e-51f0-9af2-e276c4482bbb.webp', 'Рамен'),
	((SELECT id FROM rb), 'Биск рамен большой', 'Бульон на креветках и морском гребешке, пшеничная…', 1000000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e6e15907-f0d5-589d-a0dc-5330b05142ac.webp', 'Рамен'),
	-- Удон
	((SELECT id FROM rb), 'Эби темпура Удон', 'Пшеничная лапша, бульон на водорослях Комбу и…', 600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/25e40e48-7dff-5f56-aafa-3999271b6fec.webp', 'Удон'),
	((SELECT id FROM rb), 'Даси Удон', 'Пшеничная лапша, бульон на водорослях Комбу и…', 480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/01bd315a-a44a-5657-a965-6fae57d6b2c0.webp', 'Удон'),
	((SELECT id FROM rb), 'Какиаге Удон', 'Пшеничная лапша, бульонна водорослях Комбу и стружке…', 540000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5090187d-efd0-5dd9-b0cf-c7aaf56026ce.webp', 'Удон'),
	-- Соба
	((SELECT id FROM rb), 'Даси Соба', 'Гречневая лапша, бульон на водорослях Комбу и…', 480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c91fe233-c7fe-51c8-bfe6-6bf8100d0290.webp', 'Соба'),
	((SELECT id FROM rb), 'Какиаге Соба', 'Гречневая лапша, бульон на водорослях Комбу и…', 530000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dd67bf3a-32fe-508d-87ed-3f181c7a4df2.webp', 'Соба'),
	((SELECT id FROM rb), 'Эби темпура Соба', 'Гречневая лапша, бульон на водорослях Комбу и…', 600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/578583de-daef-5949-b669-02d58145e918.webp', 'Соба'),
	-- Салаты
	((SELECT id FROM rb), 'Тайский салат с креветками', 'Микс салатов, фунчоза, манго, креветки, свежий огурец…', 720000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/90b7469b-0563-536b-8d2d-be32aa629018.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с Вакамэ', 'Водоросли Вакамэ, маринованные древесные грибы, морковь, зеленый…', 510000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d0099711-37aa-5691-8ba4-7609ec3846a1.webp', 'Салаты'),
	((SELECT id FROM rb), 'Картофельный салат', 'Картофель, морковь, кукуруза, яйцо, свежие огурцы, майнез, петрушка', 360000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1e719f24-0b89-54c6-8a5e-a1ce0faa2da7.webp', 'Салаты'),
	-- Супы
	((SELECT id FROM rb), 'Мисо-суп с дайконом', 'Бульон на дайконе, дайкон, паста мисо, зеленый…', 370000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/983aee07-8dbd-5ced-b870-bf2ed25ffdf1.webp', 'Супы'),
	((SELECT id FROM rb), 'Отядзукэ', 'Рис, бульон Даси, икра, филе лосося, стружка…', 730000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/25a00364-9748-5d8c-ad29-50be95f24eb9.webp', 'Супы'),
	-- Готовим дома
	((SELECT id FROM rb), 'Якитори Нэгима замороженные', 'Полуфабрикат для приготовления в домашних условиях.', 370000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/30dd9e07-0fa7-5df7-be28-1a3e60e00e03.webp', 'Готовим дома'),
	((SELECT id FROM rb), 'Эдамаме замороженные', 'Полуфабрикат для приготовления в домашних условиях.', 280000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4f1b71d2-54b5-53cf-bbda-534048b23988.webp', 'Готовим дома'),
	((SELECT id FROM rb), 'Креветки Тартар', 'Креветки в кляре, соус тартар', 500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/68341cb0-fc68-59eb-b2c5-170e0b2ec128.webp', 'Готовим дома'),
	-- Сасими
	((SELECT id FROM rb), 'Лосось Сасими', 'Свежий лосось, соевый соус, имбирь, специальный соус', 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/36fe7fd8-0f86-5026-8cb6-5154a26c8159.webp', 'Сасими'),
	((SELECT id FROM rb), 'Угорь Сасими', 'Японский угорь, имбирь, васаби, специальный соус', 770000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6202ddd8-4d3f-55fc-871b-ada5cb8b2af2.webp', 'Сасими'),
	((SELECT id FROM rb), 'Тунец Сасими', 'Тунец, васаби, имбирь, сппециальный соус', 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2b67f02b-058e-51d5-9131-77124d2df0e5.webp', 'Сасими'),
	-- Роллы
	((SELECT id FROM rb), 'Ролл с лососем и чили', 'Рис, лосось, авокадо, острый соус Чили, соевый…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d21d8a59-8d74-5929-879d-55269400e948.webp', 'Роллы'),
	((SELECT id FROM rb), 'Ролл лосось и авокадо', 'Рис, лосось, авокадо, соевый соус, имбирь, васаби', 770000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1c08fcac-a9e6-5dbb-b6f7-0a640c868765.webp', 'Роллы'),
	((SELECT id FROM rb), 'Ролл креветка темпура', 'Рис, креветка темпура, соус майнез с тобико…', 770000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ad99298f-571e-529c-9f17-2431ec9100b9.webp', 'Роллы'),
	-- Моти
	((SELECT id FROM rb), 'Моти брусника', '1 шт', 460000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8548fc49-65a3-5c8a-8482-4c74351f85da.webp', 'Моти'),
	((SELECT id FROM rb), 'Моти шоколад с орео', '1 шт', 460000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/345bb57d-e00e-5c3e-aa59-4ecd0524e275.webp', 'Моти'),
	-- Десерты
	((SELECT id FROM rb), 'Данго Анко', 'Теплый шарики из молотого риса, сладкая паста…', 410000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5146bc66-0a80-5e53-b4ed-9bcab6d23d3a.webp', 'Десерты'),
	((SELECT id FROM rb), 'Данго Маття', 'Тёплые шарики из молотого риса с пастой…', 410000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/64cfb783-4c2c-5192-9870-825c0b70e496.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Сок Рич Томат', 'Состав: томатный сок, соль.', 300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f2537106-e57f-582c-b665-e0ea2424d139.webp', 'Напитки'),
	((SELECT id FROM rb), 'Сок Рич Яблоко', 'Состав: концентрированный яблочный сок, вода Бренд - Рич.', 300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0210d73e-744c-5e76-aaf1-71c093ba6f96.webp', 'Напитки'),
	-- Дополнительно
	((SELECT id FROM rb), 'Майонез', NULL, 130000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8ada9c73-b21b-5215-ac0d-ecf6bc2edbf3.webp', 'Дополнительно'),
	((SELECT id FROM rb), 'Кетчуп', NULL, 130000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8ada9c73-b21b-5215-ac0d-ecf6bc2edbf3.webp', 'Дополнительно');

-- Ketch Up (id 19)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Ketch Up' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Бургеры
	((SELECT id FROM rb), 'Бургер Оригинальный', NULL, 760000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3f2250b8-578c-597f-a7ae-49121520748d.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Бургер Бри', NULL, 935000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9a25da64-7f1b-5168-8c3c-f8c5ff24f1a2.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Бургер Фирменный Ketch up', NULL, 1125000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b2f8d5a6-2089-5f8c-8eb7-c84bcd18253b.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Мини-бургеры Кетч ап', NULL, 955000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/363be678-93e6-57ac-bc3f-b719b7b9c25d.webp', 'Бургеры'),
	-- Салаты
	((SELECT id FROM rb), 'Салат с хрустящим баклажаном', NULL, 740000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8870292f-6033-50c9-9863-877e975beba9.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат коул слоу', NULL, 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/92101248-eb1a-50ae-8dc7-5429dd8a3cdf.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат Цезарь с креветками и катаифи', NULL, 945000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/456b9f40-572a-5a4e-9196-527c3a34ef02.webp', 'Салаты'),
	((SELECT id FROM rb), 'Поке с опаленной форелью', NULL, 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/11fd1655-b679-513e-a7d9-4ac70479c3cd.webp', 'Салаты'),
	-- Горячие блюда
	((SELECT id FROM rb), 'Грудка цыплёнка на гриле', NULL, 775000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0df87181-d740-5b1a-a5fc-f7a9d3c158f0.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Колбаски гриль', NULL, 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/32369d77-8c66-5edf-b534-c4a842f8c3f0.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Бифштекс с соусом ромеско', 'Фарш из мраморной говядины, соль, перец, соус…', 1030000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1323ba5f-f231-5445-b5b5-0385c822af73.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Вок с цыпленком криспи', NULL, 900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c7ac241a-0a6c-52ce-9b0c-94b45222f7c4.webp', 'Горячие блюда'),
	-- Супы
	((SELECT id FROM rb), 'Суп сырный Чеддер', NULL, 720000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1f1ce8fa-6c75-59ba-97a0-5ce9bca3033f.webp', 'Супы'),
	((SELECT id FROM rb), 'Суп Том ям', NULL, 880000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/822f8ac7-a51f-5342-b404-cbdfbcc79599.webp', 'Супы'),
	((SELECT id FROM rb), 'Суп томатный с морепродуктами', NULL, 900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a5180528-bacd-59b6-9009-0818a5ac9a4f.webp', 'Супы'),
	((SELECT id FROM rb), 'Суп томатный со страчателлой', NULL, 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b4502231-09f0-59ec-904e-584ee9f81091.webp', 'Супы'),
	-- Ребра
	((SELECT id FROM rb), 'Свиные ребра с печеным картофелем и маринованной', NULL, 1180000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/67b346df-09ad-5eff-bc05-b455e05bf8e2.webp', 'Ребра'),
	((SELECT id FROM rb), 'Свиные ребра с печеным картофелем и маринованной', NULL, 1180000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2e01e66b-d83d-579d-8356-484c8ab88454.webp', 'Ребра'),
	((SELECT id FROM rb), 'Свиные ребра с печеным картофелем и маринованной', NULL, 1180000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0443e9a6-82e3-5bf3-84e9-5b01bc1e4ff4.webp', 'Ребра'),
	-- Закуски
	((SELECT id FROM rb), 'Сыр чеддер фри', NULL, 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7486186b-4eb8-5a39-a365-f79d1ac3e00e.webp', 'Закуски'),
	((SELECT id FROM rb), 'Брускетта с ростбифом', NULL, 625000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/22f2cacd-031f-54b3-88c4-9719de581c63.webp', 'Закуски'),
	((SELECT id FROM rb), 'Тартар из говядины с баклажановым муссом', NULL, 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/984e8530-11ab-5d96-a79c-2abe47cc1ffc.webp', 'Закуски'),
	((SELECT id FROM rb), 'Куриный паштет с вишней, орехом и хрустящей', NULL, 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/abc0e585-1124-52fb-8d3b-77eeb5eef1bc.webp', 'Закуски'),
	-- Гарниры
	((SELECT id FROM rb), 'Картофель Айдахо', NULL, 325000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c49568d4-1ff2-5ea7-8554-1bf4f06a264e.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Свежие овощи стикс', NULL, 375000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4e19f90f-a476-5725-a1db-61b6b4897517.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Кукуруза чипотле-пармезан', NULL, 475000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8bcf2b13-73aa-50c8-8693-7ddfcab76bd4.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Картофель фри', NULL, 325000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7a19a68a-5a8f-51b4-926d-60123d2d17cb.webp', 'Гарниры'),
	-- Десерты
	((SELECT id FROM rb), 'Мисо-брауни с маринованной вишней и кремом из', NULL, 635000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8e90c717-1440-50dc-9c27-ab9c6caa36ab.webp', 'Десерты'),
	((SELECT id FROM rb), 'Лаймовый тарт', NULL, 665000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ba2332ac-ce79-5699-a00e-4f1e494e489f.webp', 'Десерты'),
	((SELECT id FROM rb), 'Баноффи пай', NULL, 635000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/92c2b599-31ca-5780-9959-189e0f081738.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Домашний морс', 'Собственного приготовления', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8e3d4bef-91f9-51bd-902b-36646ccb7436.webp', 'Напитки'),
	((SELECT id FROM rb), 'Апельсиновый фреш', NULL, 535000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/58745c26-3bfb-5eff-a3dd-2fccb3429566.webp', 'Напитки'),
	((SELECT id FROM rb), 'Рич Тоник', NULL, 325000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/139fc7fe-b707-5036-abe3-f26039f0754c.webp', 'Напитки'),
	((SELECT id FROM rb), 'Добрый Апельсин', NULL, 325000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0b16b484-e96d-5cb8-8610-3465e5afd1d2.webp', 'Напитки'),
	-- Соусы
	((SELECT id FROM rb), 'Соус Кетчуп', NULL, 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/acba7008-33fc-53e4-8fce-b4c24b28d272.webp', 'Соусы'),
	((SELECT id FROM rb), 'Кетчуп Фирменный', 'Соус на основе томатов', 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0382614a-f0cb-5143-bc5c-f070ef160620.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Тартар', NULL, 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9cd49e01-d5f7-5b8a-a48b-be1cb1cb2eab.webp', 'Соусы'),
	-- Завтрак
	((SELECT id FROM rb), 'Большой английский завтрак', NULL, 945000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fd651d52-e5a7-569d-b1cb-6ddba4703aa0.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Большой норвежский завтрак', NULL, 1010000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4662b2d0-da63-5fcb-96c8-f42b1a6387d3.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Сырники с вареной сгущенкой и вишневым соусом', NULL, 560000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b309112e-fd96-5847-a9f1-fadc35b81cec.webp', 'Завтрак');

-- Moro (id 20)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Moro' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Салаты
	((SELECT id FROM rb), 'Греческий салат', 'Крупные дольки овощей с двумя видами оливок…', 1290000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2cbbd444-02c5-55aa-a334-e839710fb787.webp', 'Салаты'),
	((SELECT id FROM rb), 'Фатуш', 'Свежие овощи с зеленью и рассольным сыром…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/84800843-5439-5812-8af5-fa6e15bbff5d.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат из запеченной свеклы с фетой', 'Это гармоничное сочетание нежной запечённой свёклы, солоноватой…', 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a0bfb640-0791-5d6f-b78d-2e48b8457744.webp', 'Салаты'),
	-- Гриль
	((SELECT id FROM rb), 'Шашлык из куриного бедра с зеленью', 'Сочное куриное бедро, замаринованное и приготовленное на гриле.', 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6ad744fc-dfd4-5981-82b9-ca514376480d.webp', 'Гриль'),
	((SELECT id FROM rb), 'Овощи на гриле', 'Сезонные овощи готовятся на углях, сохраняя натуральную…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c3d61d69-2161-5f2d-bfe4-c9f63529b8ef.webp', 'Гриль'),
	((SELECT id FROM rb), 'Осьминог с печеными перцами', 'Нежное мясо осьминога со сладковатыми перцами и…', 2090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7153f76d-cf82-526b-b743-9c3166af29eb.webp', 'Гриль'),
	((SELECT id FROM rb), 'Сибас, кускус', 'Нежная рыба с хрустящей корочкой, приправленная оливковым…', 1590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0b042f1e-1b85-56d2-8567-4b81a15ed900.webp', 'Гриль'),
	-- Горячие блюда
	((SELECT id FROM rb), 'Гирос с цыпленком', 'Погрузитесь в атмосферу восточного базара с нашим…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/26f94406-b88f-510f-a811-d3b819e1b5db.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Долма с говядиной и мацони', 'Долма обладает насыщенным мясным ароматом и прекрасно…', 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7e010e0c-4b6d-56a9-b02c-4ff10e4fd2fd.webp', 'Горячие блюда'),
	-- Выпечка
	((SELECT id FROM rb), 'Пита', 'Традиционная лепешка с карманом для начинки пропитывается…', 290000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fc6f36a0-c92a-553b-806e-87a09ed0504d.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Пиде с ягненком, сулугуни и томатами', 'Ароматный фарш из баранины с мятой запекается…', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/49453795-8019-5b05-ba0e-868226993657.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Пиде со страчателлой и печеными овощами', 'Открытая лепешка с кремовой страчателлой внутри, печеные…', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d4512a84-adcf-56bc-9279-73e4526a7255.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Пиде с трюфельным кремом и пармезаном', 'Мини-пиде имеют хрустящую золотистую корочку.', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4334bc1b-a0d6-536c-9352-0151f0e9a533.webp', 'Выпечка'),
	-- Закуски
	((SELECT id FROM rb), 'Паштет из куриной печени, мусс Пряная вишня', 'Изысканное сочетание нежного паштета из куриной печени…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/87dcbcf0-dee8-5ad6-8c00-86c849814b9c.webp', 'Закуски'),
	((SELECT id FROM rb), 'Чипсы из цукини с дзадзики', 'Чипсы из тонко нарезанных цукини жарятся с…', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5c03d20b-0c92-55e4-aaa5-f0ef79cbd89b.webp', 'Закуски'),
	((SELECT id FROM rb), 'Маринованные греческие оливки', 'Три вида оливок разных размеров и оттенков.', 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e2d1821c-f017-5f50-b798-643d739b20e1.webp', 'Закуски'),
	((SELECT id FROM rb), 'Печеные перцы, крем фета', 'Мягкие полоски перца на белом сливочном креме.', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/85aa9af6-0daa-5e86-91c2-68c64c38aeb9.webp', 'Закуски'),
	-- Супы
	((SELECT id FROM rb), 'Турецкие пельмешки с куриным бульоном', 'Пельмешки с говядиной и куриным бедром, жареный…', 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/56e0b35e-d42f-59f8-b1c5-cb9fabc6e442.webp', 'Супы'),
	((SELECT id FROM rb), 'Цыпленок и орзо', 'Куриная грудка готовится в ароматном бульоне, паста…', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/93bed1bb-212f-572d-af00-9907a397044c.webp', 'Супы'),
	((SELECT id FROM rb), 'Чечевичный суп с белыми грибами и специями', 'Сытный суп с грибами и легкой цитрусовой…', 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e4e5bcda-2471-54d6-ae46-5c4b6d26925e.webp', 'Супы'),
	-- Паста
	((SELECT id FROM rb), 'Греческая лазанья, пармезан', 'Блюдо обладает насыщенным мясным вкусом с лёгкой…', 1090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b2c19b0f-2081-5ee3-97a7-fff1ccd19096.webp', 'Паста'),
	((SELECT id FROM rb), 'Спагетти, креветки, сливочный биск', 'Паста с соусом на основе креветочного биска…', 1390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/606cc0a2-d076-5ca1-860f-c18836c50157.webp', 'Паста'),
	((SELECT id FROM rb), 'Спагетти, форель, томат', 'Паста Спагетти с форелью и томатом…', 1450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f8fa16c5-0964-5434-a1bd-a9c7cbf60ed6.webp', 'Паста'),
	((SELECT id FROM rb), 'Паста казаречче, томленая говядина', 'Нежная паста казаречче в категории Паста гармонично…', 1290000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/92262469-c4a5-5452-8c92-1d80f5912ec0.webp', 'Паста'),
	-- Сувлаки
	((SELECT id FROM rb), 'Сувлаки из говядины', 'Обжаренные кусочки говядины с соусом и свежими овощами.', 1450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3dfb9e5f-c296-5fe3-85bf-ee8397dbb52e.webp', 'Сувлаки'),
	((SELECT id FROM rb), 'Сувлаки из баранины', 'Куски баранины на шпажках с овощным гарниром.', 1490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2d7e1170-804b-545a-bc1c-36245e7c69f1.webp', 'Сувлаки'),
	-- Десерты
	((SELECT id FROM rb), 'Пончики Ванильный крем-голубика', 'Воздушные дрожжевые пончики с ванильным кремом и…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e8aa380d-c8dc-5a08-91e2-3817904c4bd0.webp', 'Десерты'),
	((SELECT id FROM rb), 'Тирамису с дубайским шоколадом', 'Бисквитные савоярди пропитываются кофейным сиропом, фисташковое пралине…', 800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c941f3ba-eeaa-542b-ab7a-a20def3c4f2b.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Zuegg Сок персиковый', 'Сладкий персиковый сок', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/67941465-a650-5b53-9291-21824df10b7d.webp', 'Напитки'),
	((SELECT id FROM rb), 'Rich Сок яблочный', 'Сок из сладких сортов яблок', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4e2d463c-737a-5f9e-8030-5a58546f1742.webp', 'Напитки'),
	((SELECT id FROM rb), 'Rich Сок томатный', 'Густой насыщенный сок из спелых томатов', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/55320592-1bdf-5aac-999f-4a5103274672.webp', 'Напитки'),
	((SELECT id FROM rb), 'Rich Сок вишневый', 'Насыщенный ягодный сок с глубоким вкусом', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0d4c7032-cab4-5a52-b35f-a00987f280b4.webp', 'Напитки'),
	-- Дипы
	((SELECT id FROM rb), 'Хумус', 'Кремовая паста из нута с тахини, лимонный…', 500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8ac1db02-9b1f-593f-b3a3-eeae80f32917.webp', 'Дипы'),
	((SELECT id FROM rb), 'Дзадзики с орегано', 'Освежающий йогуртовый соус с тертыми огурцами, чеснок…', 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f414d39f-a64f-5d84-adfd-6291e2b96906.webp', 'Дипы'),
	((SELECT id FROM rb), 'Мухаммара с грецким орехом', 'Острый соус из печеного перца с грецкими…', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ad6f78bd-8fca-56ea-93e6-c8293dbe5217.webp', 'Дипы'),
	((SELECT id FROM rb), 'Бабагануш', 'Печеные баклажаны приобретают дымчатый привкус, тахини создает…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/13f8b542-134f-5bac-bf83-0ab9ffa65eb7.webp', 'Дипы'),
	-- Завтрак
	((SELECT id FROM rb), 'Овсяная каша', 'Густая каша на молоке с добавлением сливочного…', 350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2ead5ed0-d925-5954-91aa-6c673acbf4b2.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Омлет', 'Пышная яичная основа с равномерной структурой, сливки…', 250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/18eb1292-c66d-5864-aa94-5528fa36de2a.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Яйцо пашот', 'Нежный белок обволакивает жидкий желток, готовится без…', 250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b6900407-acd6-55d1-853a-0d3dfec4b74e.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Скрембл', 'Взбитые яйца со сливками томятся на медленном…', 250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/218f32d2-13bd-59a2-82a2-055fafe9099c.webp', 'Завтрак');

-- Pinsky Go (id 21)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Pinsky Go' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Салаты
	((SELECT id FROM rb), 'Греческий салат', 'Перец болгарский, лук красный, томаты, огурцы, оливки…', 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/cc9bda18-644f-5cc5-98c4-c221d3f86904.webp', 'Салаты'),
	((SELECT id FROM rb), 'Сельдь на тостах', 'Ломтики сельди укладываются на хрустящие бородинские тосты…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/37aa7765-4ee0-56ba-b822-4437bf32954a.webp', 'Салаты'),
	((SELECT id FROM rb), 'Винегрет с килькой', 'Вареная свекла, морковь, картофель смешивается с солеными…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/09f42db7-f044-574f-a0cf-5ce002eecc2c.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат Ava с креветками', NULL, 1190000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/93b00f52-ce06-508e-a46e-1687e42d16fd.webp', 'Салаты'),
	-- Горячее
	((SELECT id FROM rb), 'Котлеты из индейки с гречкой', 'Индейка, яйцо куриное, соль, масло сливочное, сливки…', 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0be151eb-f619-5b1e-b569-65b69464b4b4.webp', 'Горячее'),
	((SELECT id FROM rb), 'Жареный рис с говядиной', 'Рис, масло растительное, говядина, чеснок, лук зеленый…', 1190000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0daa0eb9-f39a-5b3c-abe5-311023744c46.webp', 'Горячее'),
	((SELECT id FROM rb), 'Жареный рис с креветками', 'Рис, масло растительное, чеснок, лук зеленый, соль…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e597af39-9c10-51ee-ad26-236253c4c16c.webp', 'Горячее'),
	((SELECT id FROM rb), 'Бургер с говяжьей котлетой и трюфельным соусом', 'Бриошь, говяжий фарш, соус барбекю, салат айсберг…', 1290000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f29998d1-a6e5-5bc5-86d7-7c3ef9d41c0b.webp', 'Горячее'),
	-- Япония
	((SELECT id FROM rb), 'Ролл с обожженным лососем', 'Лосось, зеленый лук, водросли Нори, рис, угорь…', 1090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3844c64f-ed6d-51b3-8535-05766c1669eb.webp', 'Япония'),
	((SELECT id FROM rb), 'Тартар ролл с гребешком', 'Нежный рубленый гребешок с пикантным соусом спайси…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e0d7839e-4d49-5c12-848e-4a2cfe3be0b4.webp', 'Япония'),
	((SELECT id FROM rb), 'Теплый ролл с угрем', 'Тёплый ролл с угрём, сливочным сыром и…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4ce3b37f-4bfd-5c81-a511-4bf0d26e439a.webp', 'Япония'),
	((SELECT id FROM rb), 'Хрустящий ролл с угрем', 'Угорь в хрустящей панировке панко с острым…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/52d2ba30-5e09-51c5-a6f3-b6d36e305269.webp', 'Япония'),
	-- Завтраки
	((SELECT id FROM rb), 'Куриная кесадилья', NULL, 640000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/388cb591-a324-5cbf-bf52-f24bac4b6b66.webp', 'Завтраки'),
	((SELECT id FROM rb), 'Тост Крок Мадам', NULL, 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a7d28785-13a3-5075-b8d5-705738e2e8a8.webp', 'Завтраки'),
	((SELECT id FROM rb), 'Завтрак Patriki', NULL, 940000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/94bd6b14-3eb7-5ae6-9115-11e06f95c65d.webp', 'Завтраки'),
	((SELECT id FROM rb), 'Блинчики с лососем и страчателлой', 'Изысканное сочетание: нежный блин с ломтиками слабосоленого…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c7321b02-28fa-5815-b9f3-34519206ada3.webp', 'Завтраки'),
	-- Детское меню
	((SELECT id FROM rb), 'Котлетки из индейки с пюре', 'Индейка, яйцо куриное, соль, масло сливочное, сливки…', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6ef1ea3a-bd40-5d8a-9a71-a295ff934248.webp', 'Детское меню'),
	((SELECT id FROM rb), 'Оливье с курочкой', 'Яйцо куриное, горошек зеленый, картофель, огурцы соленые…', 490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/66f85e49-3660-5fe3-884f-1264b79a9938.webp', 'Детское меню'),
	((SELECT id FROM rb), 'Супчик с фрикадельками', 'Морковь, лук репчатый, соль, сельдерей, лавровый лист…', 540000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e4fe2c8a-9c0a-5019-830f-6af5dde9b0b4.webp', 'Детское меню'),
	((SELECT id FROM rb), 'Мак энд Чиз', 'Мука пшеничная, сливки, сыр Моцарелла, сыр Пармезан…', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9a9aef9f-8aba-53dd-84b1-f0fb6be61c32.webp', 'Детское меню'),
	-- Супы
	((SELECT id FROM rb), 'Борщ с говядиной и сметаной', 'Свекла, морковь, лук, перец болгарский, капуста белокочанная…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4cc516f1-c79a-5af9-b2ee-ec0cc796b3bb.webp', 'Супы'),
	((SELECT id FROM rb), 'Уха', 'Ароматный суп с лососем, судаком и картофелем.', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/df976ba9-6283-59eb-aaac-bf98b501371d.webp', 'Супы'),
	((SELECT id FROM rb), 'Куриный бульон', 'Морковь, лук, курица, соль, лавровый лист, сельдерей…', 540000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5af35f8f-8e69-59e0-a0f9-109e7592101e.webp', 'Супы'),
	((SELECT id FROM rb), 'Солянка сборная мясная', 'Колбаса Чоризо, колбаса Мортаделла, Краковская колбаса, колбаса…', 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4db4cb4d-5e03-5541-90ea-70ebfd226fe3.webp', 'Супы'),
	-- Гарниры
	((SELECT id FROM rb), 'Авокадо на гриле', 'Авокадо, масло оливковое, лимон, соль', 740000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4e8887df-cdda-5e21-b900-6eb406a5d5a3.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Картофельное пюре', 'Картофель, молоко, масло сливочное, соль, сливки', 390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ce6f396f-97c0-5174-b4be-27416b42826d.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Салат Коул Слоу', 'Капуста, морковь, майонез, сметана, горчица, сахар, соль, уксус', 390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4775d675-891d-5b98-abda-d8141c28f3fd.webp', 'Гарниры'),
	-- Паста
	((SELECT id FROM rb), 'Диталини с креветками', 'Креветки, бульон куриный, томаты Пилатти, петрушка, соль…', 1090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0a3ce8bc-8ba3-5635-9276-4881d3cbf878.webp', 'Паста'),
	((SELECT id FROM rb), 'Карбонара', 'Паста, грудинка свиная, яйцо куриное, сыр Пармезан…', 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/555a6846-868e-515e-b6a0-c6289c77a90f.webp', 'Паста'),
	((SELECT id FROM rb), 'Орзо с говяжьими щёчками', 'Паста, бульон куриный, грибы белые, масло сливочное…', 1190000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/99756d06-c477-5ce9-a2b6-4a1b04719440.webp', 'Паста'),
	((SELECT id FROM rb), 'Спагетти Болоньезе', NULL, 740000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9c731456-ad24-5db8-9f2f-39ea1c35e453.webp', 'Паста'),
	-- Напитки
	((SELECT id FROM rb), 'Вода Аква Минерале газированная', 'Классическая газированная вода с мягким вкусом, идеально…', 225000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ee4727bd-5dc4-5fb7-bcff-14206b19b308.webp', 'Напитки'),
	((SELECT id FROM rb), 'Вода Аква Минерале негазированная', 'Чистая питьевая вода без газа - то…', 225000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/147c24a3-6718-50a5-925f-27692681731d.webp', 'Напитки'),
	((SELECT id FROM rb), 'Лимонад Bona с соком граната и экстрактом цветов', 'Насыщенный гранатовый вкус с изысканными цветочными нотками…', 375000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fb5e86f8-d03d-5115-bdea-4ca574090e2c.webp', 'Напитки'),
	((SELECT id FROM rb), 'Лимонад Bona с соком и мякотью апельсина', 'Яркий апельсиновый лимонад с натуральной мякотью и…', 375000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/76cf0692-4411-5f3d-9464-6b775c1b16d3.webp', 'Напитки'),
	-- Дополнительно
	((SELECT id FROM rb), 'Васаби', 'Васаби, вода', 45000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/87c7314e-3e5b-5074-b360-4889ca5eb1c8.webp', 'Дополнительно'),
	((SELECT id FROM rb), 'Имбирь маринованный', 'Имбирь маринованный', 45000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8f56d2d2-390b-5856-9c67-3a081b4365bf.webp', 'Дополнительно'),
	((SELECT id FROM rb), 'Соевый соус', 'Соевый соус, вода, водоросли, соус Мирин, сахар', 55000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5ac81576-d78f-51cc-b6bc-ee9ae647067a.webp', 'Дополнительно'),
	-- Десерты
	((SELECT id FROM rb), 'Медовик', 'Порция фирменного торта Медовика', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c7abe752-5076-5590-aaca-ae7c29c4a91c.webp', 'Десерты'),
	((SELECT id FROM rb), 'Наполеон', 'Порция фирменного торта Наполеон', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/98fc36ad-606b-5a7b-9be8-d682f72d7c7e.webp', 'Десерты');

-- Раменная Ru-Rik (id 22)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Раменная Ru-Rik' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Супы
	((SELECT id FROM rb), 'Рамен с говядиной', NULL, 960000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b8f403bc-c2f3-5af0-b1a6-1c27bd28556e.webp', 'Супы'),
	((SELECT id FROM rb), 'Рамен с рваной уткой', NULL, 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8d6ff9f1-a496-5eec-bb89-b1481808a48f.webp', 'Супы'),
	((SELECT id FROM rb), 'Мисо суп', NULL, 370000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/86a5dbfc-2730-56cb-8d70-5d00c0524c78.webp', 'Супы'),
	-- Вок
	((SELECT id FROM rb), 'Жареный рис с креветками', NULL, 610000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6c246807-6404-5eec-88dd-9af396bbbdb0.webp', 'Вок'),
	((SELECT id FROM rb), 'Удон с креветками', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/52df8e40-f1c0-5e56-b568-21fdc4dc4103.webp', 'Вок'),
	((SELECT id FROM rb), 'Жареный рис с курицей', NULL, 530000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/16380a07-6e79-55d7-8f24-2143b2e49cc1.webp', 'Вок'),
	-- Салаты
	((SELECT id FROM rb), 'Поке с креветками', NULL, 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/088046db-f4f1-5883-8c3a-42940372ab7c.webp', 'Салаты'),
	((SELECT id FROM rb), 'Хрустящие баклажаны с креветками', NULL, 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a633e650-394f-52ad-bbe3-c082a74ff7a8.webp', 'Салаты'),
	((SELECT id FROM rb), 'Поке с лососем', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b7be9a6c-d4e1-5312-a880-590bb7502c70.webp', 'Салаты'),
	-- Роллы
	((SELECT id FROM rb), 'Филадельфия', NULL, 960000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1d078fb3-249d-52e6-bfad-74dc49217551.webp', 'Роллы'),
	((SELECT id FROM rb), 'Канада', NULL, 960000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9bc96d07-d370-5351-a7e8-ff8764ac9c14.webp', 'Роллы'),
	((SELECT id FROM rb), 'Ролл с креветкой темпура', NULL, 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/92ab2351-1b6d-5478-a72f-262bfbd32c32.webp', 'Роллы'),
	-- Пекинская утка
	((SELECT id FROM rb), 'Утка по-пекински половина', 'Фирменная пекинская утка Гризби на две персоны', 2200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9cf14f0e-dd93-5137-996b-013441734aac.webp', 'Пекинская утка'),
	((SELECT id FROM rb), 'Утка по-пекински целая', 'Фирменная пекинская утка Гризби на четыре персоны', 3800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e76d7e47-337a-589a-a75f-d7b61aa4626a.webp', 'Пекинская утка'),
	((SELECT id FROM rb), 'Блинчики к утке по-пекински', 'Пшеничные китайские блинчики для утки по-пекински', 300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ab881b5c-db9c-5cd6-890d-84fd9e31f086.webp', 'Пекинская утка'),
	-- Димсамы и гедза
	((SELECT id FROM rb), 'Димсам с курицей', NULL, 490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b036c9f6-0231-5b0d-8ed6-c55e6814451f.webp', 'Димсамы и гедза'),
	((SELECT id FROM rb), 'Димсам с говядиной', NULL, 510000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/10b72a8b-8194-58a0-be4e-4e0970de3332.webp', 'Димсамы и гедза'),
	((SELECT id FROM rb), 'Димсам с креветкой', NULL, 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/21edd266-c515-5449-bad1-d31bf48a56dc.webp', 'Димсамы и гедза'),
	-- Суши
	((SELECT id FROM rb), 'Маки авокадо', NULL, 310000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/87b4fa49-b85c-5a1a-9260-7ffd10585f6d.webp', 'Суши'),
	((SELECT id FROM rb), 'Маки лосось', NULL, 470000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1857b63a-8047-5f3d-a324-cd4d5dadd3b2.webp', 'Суши'),
	-- Горячие роллы
	((SELECT id FROM rb), 'Горячий ролл краб', NULL, 870000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fb2609a4-7500-5345-b7fd-2028fec644fc.webp', 'Горячие роллы'),
	((SELECT id FROM rb), 'Горячий ролл лосось', NULL, 760000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9076cd81-3189-5c5a-8072-2d36525ece44.webp', 'Горячие роллы'),
	-- Горячие блюда
	((SELECT id FROM rb), 'Креветки в кисло-сладком соусе', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/21b0e455-fc5b-52b6-a8d5-00c019b49195.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Рисдог с лососем', NULL, 710000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9bc366b2-0523-51b0-9e30-5bc769fd69a4.webp', 'Горячие блюда'),
	-- Гарниры
	((SELECT id FROM rb), 'Картофель фри', NULL, 330000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fa09dc9b-2ea9-5ca8-8c38-27f8f0850959.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Японский рис', NULL, 250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/516ea91f-5e35-52cd-8c99-0ce9943bace7.webp', 'Гарниры'),
	-- Десерты
	((SELECT id FROM rb), 'Шоколадный брауни', NULL, 470000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/15b45c96-24f3-524e-a785-f0de2c7a2a0c.webp', 'Десерты'),
	((SELECT id FROM rb), 'Медовик', NULL, 450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/93cc7460-5629-5bed-9b50-b9d8eac870c7.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Апельсиновый фреш', NULL, 450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3239dfcc-61a6-5d2d-8fef-924f1c209065.webp', 'Напитки'),
	((SELECT id FROM rb), 'Фирменный лимонад Манго-маракуйя', NULL, 450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b87831d0-5012-553a-9815-69b381e4a33a.webp', 'Напитки'),
	-- Соусы
	((SELECT id FROM rb), 'Васаби', NULL, 60000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/de56631e-e8e3-50e8-8034-956c50be3e61.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соевый соус', NULL, 60000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/cf1bd14a-1581-52ff-b821-23a94fe04911.webp', 'Соусы'),
	-- Закуски
	((SELECT id FROM rb), 'Эдамаме с морской солью', NULL, 340000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2c8995b2-a539-5952-adbb-d73b96dd3725.webp', 'Закуски'),
	((SELECT id FROM rb), 'Капуста кимчи', NULL, 280000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/61a36d95-157f-5b7d-81b5-3615ab419732.webp', 'Закуски'),
	-- Дополнительно
	((SELECT id FROM rb), 'Помидоры черри', NULL, 100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/411c786e-31e3-5352-baaf-fffae6ad4510.webp', 'Дополнительно'),
	((SELECT id FROM rb), 'Яйцо маринованное', NULL, 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8a083664-6a90-5d1b-869f-595e885764ca.webp', 'Дополнительно'),
	-- Детское меню
	((SELECT id FROM rb), 'Детский рамен с курицей', NULL, 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8532149f-ee0a-57ff-a417-511eb3bf4b98.webp', 'Детское меню'),
	((SELECT id FROM rb), 'Детский Удон с курицей', NULL, 320000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b8f47ff9-174e-5b2f-99f8-9ba15e9402b0.webp', 'Детское меню'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Креветки васаби попкорн', NULL, 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/afa6b297-c706-5f9b-9591-41918089c62b.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Креветки темпура', NULL, 760000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/94ddb536-6ca6-5e83-8a6e-bfaa5b9bcae1.webp', 'Горячие закуски');

-- Империя Пиццы (id 23)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Империя Пиццы' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Пиццы 31 см
	((SELECT id FROM rb), 'Мега пепперони 31 см традиционное тесто', 'Двойная порция пепперони, Сыр моцарелла, Фирменный томатный соус', 809000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1859c382-772d-58b4-89c1-de42cb2b3206.webp', 'Пиццы 31 см'),
	((SELECT id FROM rb), 'Мега пепперони 31 см тонкое тесто', 'Двойная порция пепперони, Сыр моцарелла, Фирменный томатный соус', 809000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1859c382-772d-58b4-89c1-de42cb2b3206.webp', 'Пиццы 31 см'),
	((SELECT id FROM rb), 'Ветчина и грибы 31 см традиционное тесто', 'Ветчина, шампиньоны, томатный соус, сыр моцарелла', 759000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9dfc931e-85d4-5c28-a061-3b1d73e482c8.webp', 'Пиццы 31 см'),
	-- Закуски
	((SELECT id FROM rb), 'Наггетсы 9 штук', 'Куриные наггетсы в хрустящей панировке', 406000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ec0ed5e8-fc36-5b86-a7eb-8140f90fcd23.webp', 'Закуски'),
	((SELECT id FROM rb), 'Сырные палочки New', 'Сыр Моцарелла, Панировка', 413000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/93fa5908-c770-54a2-9081-f6c0d362a2e1.webp', 'Закуски'),
	((SELECT id FROM rb), 'Картофель фри', 'Картофель фри, Соль', 252000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6fbd8e51-a5e8-590c-a777-403be3861d17.webp', 'Закуски'),
	-- Роллы
	((SELECT id FROM rb), 'Калифорния', 'Сурими, Авокадо, Рис, Водоросли нори, Икра масаго, Майонез', 483000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9e265816-5759-57ab-a82c-c793e70c3c68.webp', 'Роллы'),
	((SELECT id FROM rb), 'Филадельфия', 'Лосось, Огурец, Сыр Филадельфия, Водоросли нори, 8…', 759000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/32b77196-d505-5065-b513-7bb43b93afe3.webp', 'Роллы'),
	((SELECT id FROM rb), 'Филадельфия 4 шт', 'Лосось, Огурец, Сыр Филадельфия, Водоросли нори', 538000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e44db28a-d49d-57dc-bf75-5cd7ec625b9e.webp', 'Роллы'),
	-- Пиццы 26 см
	((SELECT id FROM rb), 'Четыре сыра 26 см традиционное тесто', 'Сыр Дорблю, Сыр Моцарелла, Сыр Чеддер, Сыр…', 748000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2f7482f3-b976-52bd-b144-b840b6e73f47.webp', 'Пиццы 26 см'),
	((SELECT id FROM rb), 'Мега пепперони 26 см традиционное тесто', 'Двойная порция пепперони, Сыр моцарелла, Фирменный томатный соус', 611000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1859c382-772d-58b4-89c1-de42cb2b3206.webp', 'Пиццы 26 см'),
	((SELECT id FROM rb), 'Ветчина и грибы 26 см традиционное тесто', 'Ветчина, Шампиньоны, Сыр Моцарелла, Томатный соус', 472000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9dfc931e-85d4-5c28-a061-3b1d73e482c8.webp', 'Пиццы 26 см'),
	-- Суши
	((SELECT id FROM rb), 'Суши Сяке (2 шт.)', 'Лосось, Рис', 369000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/43205a48-a9c1-5b76-99b5-cdba6145e312.webp', 'Суши'),
	((SELECT id FROM rb), 'Суши Спайси Сяке (2 шт.)', 'Лосось, Рис, Водоросли нори, Соус спайси', 446000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e1be890a-4f3c-597a-804f-711082efe500.webp', 'Суши'),
	((SELECT id FROM rb), 'Суши Чука (2 шт.)', 'Чука, Рис, Водоросли нори', 160000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/54678e1c-c001-5f64-96db-b489b7071b4c.webp', 'Суши'),
	-- Горячее
	((SELECT id FROM rb), 'Хинкали с бараниной', 'Баранина, Бараний жир, Лук, Кориандр, Черный молотый…', 329000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/43795e88-74c3-5137-b66d-3be080535e55.webp', 'Горячее'),
	((SELECT id FROM rb), 'Хинкали со свининой и говядиной', 'Говядина, Свинина, говяжий жир, Лук репчатый, Кориандр…', 329000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b382d8bd-a293-5d1e-b8eb-6e07230cb417.webp', 'Горячее'),
	((SELECT id FROM rb), 'Мясные котлеты с картофелем по-деревенски', 'Говядина, Свинина, Лук репчатый, Хлеб, Яйцо куриное…', 461000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c869bf3d-e6c0-58e7-870f-b6ded2724e0c.webp', 'Горячее'),
	-- Воки
	((SELECT id FROM rb), 'Пшеничная лапша с курицей', 'Пшеничная лапша, Курица, Болгарский перец, Лук, Морковь…', 490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a0362afd-4060-5ea3-8516-189f652168d9.webp', 'Воки'),
	((SELECT id FROM rb), 'Пшеничная лапша с креветками под сливочным соусом', 'Пшеничная лапша, Креветки, Болгарский перец, Лук, Морковь…', 659000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/20f1babb-7ecc-5444-b1b0-d21dbaa59b2e.webp', 'Воки'),
	((SELECT id FROM rb), 'Пшеничная лапша с курицей под сливочным соусом', 'Пшеничная лапша, Курица, Болгарский перец, Лук, Морковь…', 523000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ea151fb6-5485-5a51-b92b-27eb31817a9b.webp', 'Воки'),
	-- Супы
	((SELECT id FROM rb), 'Том ям New', 'Тигровые креветки, Томаты черри, Шампиньоны, листья лайма…', 556000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/10f319a9-36c8-5366-b866-0afd5423162c.webp', 'Супы'),
	((SELECT id FROM rb), 'Окрошка', 'Огурцы, Ветчина, Картофель, Редис, Яйцо куриное, Зелень…', 395000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f3e7d09a-2a9e-560b-8db2-2bc9a717508a.webp', 'Супы'),
	((SELECT id FROM rb), 'Рамен с курицей New', 'Мисо бульон, Пшеничная лапша, Курица, Черри, нори', 468000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ca5947e3-acea-50bc-a8d1-ea65f3209055.webp', 'Супы'),
	-- Салаты
	((SELECT id FROM rb), 'Цезарь с курицей New', 'Курица, томаты черри, салат айсберг, пармезан, сухарики…', 527000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a04a0a95-50e7-5338-995d-d4561967bd38.webp', 'Салаты'),
	((SELECT id FROM rb), 'Цезарь с креветками New', 'Тигровые креветки, томаты черри, салат айсберг, пармезан…', 648000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/950401d9-d9ce-5ad6-930c-f0850f996dcc.webp', 'Салаты'),
	((SELECT id FROM rb), 'Поке с лососем', 'Рис, Соус поке фрай, Лосось, Свежий огурчик…', 699000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b71809e4-cbd7-5571-b2ef-0ab5ba4d4bde.webp', 'Салаты'),
	-- Детское
	((SELECT id FROM rb), 'Наггетсы 7 штук', 'Куриные наггетсы в хрустящей панировке', 329000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ff96baa3-5e96-512d-a0f2-ba045109e10a.webp', 'Детское'),
	((SELECT id FROM rb), 'Панакота с пюре манго', 'Манговое пюре, Сливки, Молоко, Ванильный сироп', 439000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c64eccd0-36f9-5d9b-acc7-21d1894c994a.webp', 'Детское'),
	((SELECT id FROM rb), 'Творожная запеканка', 'Творог, Крупа манная, Сахар, Сметана, Яйцо куриное…', 373000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/26c1cccf-14cd-576e-9024-9797760ef045.webp', 'Детское'),
	-- Десерты
	((SELECT id FROM rb), 'Наполеон', 'Слоеное бездрожжевое тесто, Крем из сливок и…', 424000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8d18e6b0-f1bf-5b2c-a865-4f52a6e92072.webp', 'Десерты'),
	((SELECT id FROM rb), 'Десерт Тирамису', 'Маскарпоне, Печенье Савоярди, Какао-порошок, Сахар, Яйца, Сливки…', 461000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/28640c3e-6539-5833-9cb6-26ce0c67188e.webp', 'Десерты'),
	((SELECT id FROM rb), 'Классический Чизкейк New York', 'Песочное тесто, Франжипан, Сыр сливочный, Сливки, Яйцо…', 373000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/95375248-9819-51bf-a34d-30027aa4d6ca.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Облепиховый морс 1 л', 'Облепиха, вода, сахар', 340000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7e10364f-8efa-54eb-aba4-d01437b8e7cd.webp', 'Напитки'),
	((SELECT id FROM rb), 'Ягодный морс 1 л', 'Клюква, чёрная смородина, вишня, сахар', 395000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9335092c-7a3d-5614-98ca-b09900af28e6.webp', 'Напитки'),
	((SELECT id FROM rb), 'Домашний лимонад 1 л', 'Лимон, лайм, сахар-песок, лимонная кислота, сахар тростниковый', 395000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/760d3b8d-aa81-5f43-8694-41b03c843923.webp', 'Напитки'),
	-- Соусы
	((SELECT id FROM rb), 'Соус сырный', 'Соус Heinz', 85000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8263dcd8-3383-53f5-8e44-f73965eb7e18.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус спайси', 'Паста кимчи, паста табаджан, чеснок, соус майонезный постный', 103000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0b244b8b-8be6-5b30-8bd9-586a52eea762.webp', 'Соусы'),
	-- Ланчи
	((SELECT id FROM rb), 'Техас ланч', 'Суп куриный, Рис с яйцом и беконом', 593000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b37bff3e-4f99-5583-a510-e7e906cbf491.webp', 'Ланчи'),
	((SELECT id FROM rb), 'Русский ланч', 'Суп куриный, картофель по-деревенски с котлетами', 622000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/69686343-7008-5dd4-8211-13a773ea77c3.webp', 'Ланчи');

-- El Chapo Burgers Tacos&Burritos (id 24)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'El Chapo Burgers Tacos&Burritos' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Тако и буррито
	((SELECT id FROM rb), 'Тако Говядина 3 лодочки', NULL, 1799000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/91bc8469-215c-5941-bbf5-bd99466863f4.webp', 'Тако и буррито'),
	((SELECT id FROM rb), 'Буррито с говядиной', NULL, 1480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/73856076-96d3-5730-8c19-29f454908130.webp', 'Тако и буррито'),
	((SELECT id FROM rb), 'Тако с курицей 3 лодочки', NULL, 1798000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/39a138e4-d54b-5eb3-ac9e-aa32ffbef839.webp', 'Тако и буррито'),
	((SELECT id FROM rb), 'Тако овощной 2 лодочки', NULL, 1397000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0068c581-3dc7-5482-95ec-5346b34618fb.webp', 'Тако и буррито'),
	((SELECT id FROM rb), 'Буррито с курицей', NULL, 1480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/87e72605-e6bf-5f53-bb45-d2bb02a96981.webp', 'Тако и буррито'),
	-- Бургеры
	((SELECT id FROM rb), 'Американский бургер', NULL, 1493000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ebf7b4ea-4804-547e-a381-6eb3237bd094.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Делюкс бургер', NULL, 1467000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b591318c-75e9-549c-b143-7c28c50c7195.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Разбитый бургер 3 котлеты', NULL, 1799000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f5dab96f-0ac7-5d12-81a1-a72ed30d8069.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Классический чизбургер', NULL, 1479000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/11b99229-e083-5512-af86-08b3c0aa937b.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Бургер с грибами', NULL, 1495000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b602a725-f116-5775-aea6-d1c3760fe086.webp', 'Бургеры'),
	-- Закуски
	((SELECT id FROM rb), 'Фри', NULL, 849000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d161d0ba-9e04-55d9-843d-8c8eea038ab2.webp', 'Закуски'),
	((SELECT id FROM rb), 'Мексиканские Начос', NULL, 799000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d50b4974-3fdd-5c2b-bc7a-6728ca313b68.webp', 'Закуски'),
	((SELECT id FROM rb), 'Картофельные дольки', NULL, 849000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3223e304-6017-5917-873f-393d8ca12f36.webp', 'Закуски'),
	-- Кесадилья
	((SELECT id FROM rb), 'Куриная кесадилья', NULL, 1490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ce7cc079-164f-5d68-9d3b-20473ea0794c.webp', 'Кесадилья'),
	((SELECT id FROM rb), 'Кесадилья с говядиной', NULL, 1493000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6389af96-db5b-5324-ac6b-ac3e81b5517c.webp', 'Кесадилья'),
	((SELECT id FROM rb), 'Сырная кесадилья Chease Сыр', NULL, 1498000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ebfc441d-00cb-557f-b0e3-cfac66c84d82.webp', 'Кесадилья'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Сырные Палки', NULL, 998000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ec5da0ae-02a1-5c5b-b517-74cb26e8a4f6.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Кольца из Кальмара', NULL, 997000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/26dcc09e-df20-58cf-8077-3e9df6dc8ad5.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Кольца из Лука в панировке', NULL, 899000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/98822ff0-b6f3-572d-a52a-1ad4c2881467.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Крылья Острые Хот', NULL, 999000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/42ec7551-25f8-5607-a5ad-2f2a1f65fb79.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Крылья с Копченым соусом барбекю', NULL, 999000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/17192535-0cd4-5368-a453-62765041f5b9.webp', 'Горячие закуски'),
	-- Супы
	((SELECT id FROM rb), 'Том Ям', NULL, 1399000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8d9ac399-8cf4-5b85-b5ee-eb372d8eae70.webp', 'Супы'),
	-- Вок
	((SELECT id FROM rb), 'Вок из премиальной говядины', NULL, 1395000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0b0f224d-1884-5da8-99c5-db5eac958efb.webp', 'Вок'),
	((SELECT id FROM rb), 'Вок с куриным филе', NULL, 1379000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2ccfd240-a173-53be-95c3-838abeebc8fa.webp', 'Вок'),
	((SELECT id FROM rb), 'Вок из Тигровых Креветок', NULL, 1390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/aab54e57-0cce-5902-82ae-69364de0f37a.webp', 'Вок'),
	((SELECT id FROM rb), 'Вок с мурманским лососем', NULL, 1399000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/baef04e3-6d46-559c-9139-569f7cec8e85.webp', 'Вок'),
	-- Напитки
	((SELECT id FROM rb), 'Добрый Кола', NULL, 369000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/52ade3ec-b6a3-5983-884f-e544ae834864.webp', 'Напитки'),
	((SELECT id FROM rb), 'Холодный Чай Пи Китайский', NULL, 769000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5a57447f-2c61-5bcf-9fa5-fc3ba4298985.webp', 'Напитки'),
	((SELECT id FROM rb), 'Вишневый морс', NULL, 349000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/aed50bbe-7d64-5b00-b22c-f305edc53666.webp', 'Напитки'),
	-- Холодные напитки
	((SELECT id FROM rb), 'Напиток из красных ягод', NULL, 299000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/53adccaf-4bf5-5a12-b0da-3405efc35f48.webp', 'Холодные напитки'),
	-- Соусы
	((SELECT id FROM rb), 'Сладкий Чили', NULL, 100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/539daa5b-8419-5e27-8457-44639f928eb7.webp', 'Соусы'),
	((SELECT id FROM rb), 'Блю чиз', NULL, 120000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/303cfa8c-032e-5772-8ad6-528acbf7ebac.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Сырный', NULL, 100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/69597cfd-3fa3-5ad0-9301-92f70b6701f9.webp', 'Соусы'),
	((SELECT id FROM rb), 'Барбекю соус', NULL, 100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1882bbdd-72ab-582a-a445-5da07629afdb.webp', 'Соусы'),
	-- Те самые пиццы
	((SELECT id FROM rb), '4 сыра', NULL, 1700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ef0a6fc7-2635-5862-9b8d-820961723009.webp', 'Те самые пиццы'),
	((SELECT id FROM rb), 'Вегетарианская', NULL, 1700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/71c6aa12-300b-57b4-aec3-5db70d5a9163.webp', 'Те самые пиццы'),
	((SELECT id FROM rb), 'Гавайская', NULL, 1700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0ae89674-b9e5-5d60-9dcd-b2a9900d8ae9.webp', 'Те самые пиццы'),
	((SELECT id FROM rb), 'Натощак', NULL, 1700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/cfbdb265-fe93-5b65-b459-17e144587da6.webp', 'Те самые пиццы'),
	-- Сэндвичи
	((SELECT id FROM rb), 'С ветчиной', NULL, 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6c8bb102-ed25-59c5-8f78-00c7b6240c29.webp', 'Сэндвичи'),
	((SELECT id FROM rb), 'С морепродуктами', NULL, 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/226a7747-5811-5305-a833-97d88a85f31d.webp', 'Сэндвичи');

-- Калифорния Дайнер (id 25)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Калифорния Дайнер' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Завтрак
	((SELECT id FROM rb), 'Стейк и яйца', NULL, 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/86349451-932c-5f0f-9bf0-42ef6543e11a.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Биф скиллет', NULL, 575000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/68ca7b1e-e192-51a4-9f89-e2f7aff11ac7.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Сэндвич с яйцом', NULL, 555000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a361d046-e4da-57b5-b4aa-c5170ef8566a.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Английский завтрак', NULL, 680000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/aa25bdb7-842b-5d70-8f5e-83090203e961.webp', 'Завтрак'),
	-- Салаты
	((SELECT id FROM rb), 'Боул', NULL, 730000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ef029838-3790-5bb5-a08b-fd2bc85da1aa.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с куриной печенью', NULL, 605000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/19fcca5b-3f94-5224-a6dc-2cf76e7a7c32.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат Цезарь с креветками', NULL, 785000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bf9f80a8-f78f-5848-b397-0b3a505dd179.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат Калифорния с говядиной', NULL, 720000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/41bc247c-bac1-52a6-9a68-4f1bc95c06ff.webp', 'Салаты'),
	-- Закуски
	((SELECT id FROM rb), 'Сырные шарики с халапенью', NULL, 480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/93863832-7227-5c40-add7-c51b68f05568.webp', 'Закуски'),
	((SELECT id FROM rb), 'Фиш энд чипс', NULL, 555000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1c5d0df3-2703-5ded-b770-c11bdcac7af6.webp', 'Закуски'),
	((SELECT id FROM rb), 'Жареные креветки', NULL, 865000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/11c464bb-e935-5714-8314-1a2564cac2eb.webp', 'Закуски'),
	((SELECT id FROM rb), 'Кесадилья', NULL, 520000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7f8c9a56-c484-541e-9b23-c22ec0c4c3e2.webp', 'Закуски'),
	-- Супы
	((SELECT id FROM rb), 'Фо-бо', NULL, 450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/58f14244-c416-5770-a42e-40bd0616a674.webp', 'Супы'),
	((SELECT id FROM rb), 'Классический борщ с курицей', NULL, 450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/861d92e7-61b8-52f9-9f73-d634a955e1a8.webp', 'Супы'),
	((SELECT id FROM rb), 'Куриный суп Гамбо', NULL, 450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/df95d707-7dfc-5d1f-bb1d-0d7e4ef36038.webp', 'Супы'),
	-- Паста
	((SELECT id FROM rb), 'Вегетарианская паста', NULL, 490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dae3f068-83a3-5b6e-8af2-7ac33fb5f0c0.webp', 'Паста'),
	((SELECT id FROM rb), 'Сливочная паста с креветками и песто', NULL, 920000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0369bd11-0bb5-51a9-b802-96992c565145.webp', 'Паста'),
	((SELECT id FROM rb), 'Паста по-кайенски с курицей', NULL, 575000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8411c6cc-1ffa-56c6-87f2-e1950e6c1a63.webp', 'Паста'),
	-- Бургеры
	((SELECT id FROM rb), 'Вегетарианский бургер Нью-Мехико', NULL, 580000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ec2828e2-1c68-5e06-be96-36d5f2f36e69.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Черный Адам', NULL, 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/117d2bba-c725-551b-b83a-657945783e5b.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Тако-бургер', NULL, 760000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/08a1ebc0-ef8c-5675-8512-bb1eb0e15d11.webp', 'Бургеры'),
	-- Сэндвичи
	((SELECT id FROM rb), 'Ориджинал пати мелтс', NULL, 685000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a90fe610-5e90-5625-9637-09248b1968e6.webp', 'Сэндвичи'),
	((SELECT id FROM rb), 'Сэндвич Рубен', NULL, 670000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5b0f9b44-8836-5dad-ba92-f5a6cabe7b90.webp', 'Сэндвичи'),
	((SELECT id FROM rb), 'Сэндвич Филадельфия-чиз с говядиной', NULL, 875000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d146e879-2a4f-5575-9358-df5b3f51d34f.webp', 'Сэндвичи'),
	-- Роллы
	((SELECT id FROM rb), 'Цезарь рэпс', NULL, 555000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/440eb74a-23ee-5016-9660-0ceca9e9081d.webp', 'Роллы'),
	-- Горячие блюда
	((SELECT id FROM rb), 'Куриный стейк на гриле', NULL, 555000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d778a7cc-3c2a-5234-9966-238a709831d4.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Стейк Медальоны с говяжьей вырезкой', NULL, 1330000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c282d07f-0487-5aaf-9378-7b80ed628bbe.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Говяжьи медальоны и хашбраун', NULL, 1315000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/26d1840a-aab3-56da-9fa2-834bacfda43d.webp', 'Горячие блюда'),
	-- Гарниры
	((SELECT id FROM rb), 'Батат фри', NULL, 475000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/04d8cbae-bfdd-52e2-8c0e-4f1b8d06b1ae.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Овощи гриль', NULL, 315000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8fe80949-60bf-557e-a6d2-b0f8271507bd.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Картофельное пюре', NULL, 235000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/84791921-c8ba-583f-9f65-c327227b8e4f.webp', 'Гарниры'),
	-- Десерты
	((SELECT id FROM rb), 'Торт Красный бархат', NULL, 415000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6472a6a3-2c31-5158-b4a1-86aa3e56293a.webp', 'Десерты'),
	((SELECT id FROM rb), 'Вишневый пирог', NULL, 445000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/520a39fe-d79c-5ae3-82d4-9f4c7c583024.webp', 'Десерты'),
	((SELECT id FROM rb), 'Ванильные панкейки', NULL, 480000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d3c3b875-ab79-5d5c-8e94-10b5edcebdee.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Rich Cola', NULL, 370000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/557011e4-3e3a-54bf-bbc2-7a864e060a6e.webp', 'Напитки'),
	((SELECT id FROM rb), 'Rich Cola Zero', NULL, 370000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/274dbae7-c843-56c5-8ffc-f77942d07dc7.webp', 'Напитки'),
	((SELECT id FROM rb), 'БонАква без газа', NULL, 270000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/12efde2d-ab45-55f7-8a0d-13a30e113ff3.webp', 'Напитки'),
	-- Соусы
	((SELECT id FROM rb), 'Соус Сырный', NULL, 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e9462643-92c1-5bc0-8b87-0a4b176a551f.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Блю чиз', NULL, 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/31c1c468-fb6f-5f47-84c9-5f0492966c35.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Тысяча островов', NULL, 110000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/055a8bed-0ed4-597c-8a2a-727465003226.webp', 'Соусы');

-- Машрумс (id 26)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Машрумс' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Обеденное меню
	((SELECT id FROM rb), 'Lunch №7', 'Салат овощной, цыпленок на гриле с брокколи, лимонад', 1650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f2072cd7-afbe-50df-92d9-0fc6c29e5ce7.webp', 'Обеденное меню'),
	((SELECT id FROM rb), 'Lunch №6', 'Паштет куриный с тостами, салат баклажанами и…', 2310000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fdbf3bbc-5006-5c62-9fe7-3c0d3fa6f909.webp', 'Обеденное меню'),
	((SELECT id FROM rb), 'Lunch №5', 'Пашет куриный с тостами, куриный суп с…', 1980000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d19e8aec-b395-5e78-9162-1e60720d1eba.webp', 'Обеденное меню'),
	-- Итальянские деликатесы
	((SELECT id FROM rb), 'Пекорино', NULL, 460000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1d2798c1-c1da-5dd6-8610-35cffb8a39ea.webp', 'Итальянские деликатесы'),
	((SELECT id FROM rb), 'Горгондзола', NULL, 460000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3d0a1f75-25da-5dc8-a3f1-7566862031a2.webp', 'Итальянские деликатесы'),
	((SELECT id FROM rb), 'Пармеджано реджано', NULL, 830000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/74522c9a-06b0-55f3-9822-b911ac3b00ba.webp', 'Итальянские деликатесы'),
	-- Брускетты
	((SELECT id FROM rb), 'Брускетта с мортаделлой и страчателлой', 'Хлеб пшеничный, сыр страчателла, колбаса мортаделла, соус…', 960000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9695acaa-0f42-5a73-838a-fc0ad3f39bda.webp', 'Брускетты'),
	((SELECT id FROM rb), 'Брускетта с креветками и вялеными томатами', 'Хлеб пшеничный, сыр страчателла, колбаса мортаделла, соус…', 910000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2082ab38-40de-5d0a-b8c2-cf01536b93ba.webp', 'Брускетты'),
	-- Холодные закуски
	((SELECT id FROM rb), 'Вителло тонатто, трюфельное масло и грибная пудра', 'Телятина, белые грибы соус (пармезан, анчоусы, каперсы…', 1270000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/63fa6773-b74f-5992-878a-21975ddf01ed.webp', 'Холодные закуски'),
	((SELECT id FROM rb), 'Запеченные перцы рамиро с соусом из тунца', 'Масло сливочное, яйца, сахар, соль, мука пшеничная…', 1090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4b084128-7ec8-5cde-9dd8-9ae5b7b26ae8.webp', 'Холодные закуски'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Баклажаны пармиджано', 'Баклажаны, томаты, лук, чеснок, орегано, тертый сыр…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/909fd000-7bcd-50f4-b63a-c01207d647b3.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Жульен из подберезовиков, боровиков и вешенок', 'Блюдо из жареных боровиков, подберезовиков, вешенки в…', 1770000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7187d47a-64cf-543c-a7ae-6415658508c7.webp', 'Горячие закуски'),
	-- Салаты
	((SELECT id FROM rb), 'Вяленая утка с грушей, горгондзола и малиновый', NULL, 1060000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e11c41f9-23ab-5edf-b058-eba314660c5e.webp', 'Салаты'),
	((SELECT id FROM rb), 'Хрустящие шиитаке, зеленый салат и кунжутный соус', 'Грибы шиитаке, шпинат, кресс-салат, кунжутные чипсы, заправка…', 1000000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8cdbbadc-7bb7-540b-8e60-0496f28e3537.webp', 'Салаты'),
	-- Супы
	((SELECT id FROM rb), 'Неопалетанский томатный суп с морепродуктами', NULL, 980000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/079a3c42-c1a1-5c8b-8274-eeff12085fc7.webp', 'Супы'),
	((SELECT id FROM rb), 'Крем-суп с белыми грибами и луком фри', 'Белые грибы, шампиньоны, чеснок, масло сливочное, бульон…', 1020000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5ab394e3-764b-5151-808c-ee7545626e88.webp', 'Супы'),
	-- Пицца
	((SELECT id FROM rb), 'Пицца Mushrooms', 'Тесто (мука пшеничная, дрожжи, соль, сахар, масло…', 1420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/52c1c52d-4b75-5fe8-919e-26e7230a6130.webp', 'Пицца'),
	((SELECT id FROM rb), 'Пицца Маргарита', 'Тесто (мука пшеничная, дрожжи, соль, сахар, масло…', 820000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4249d3a0-89c9-5d2b-b339-89cdb41e2a37.webp', 'Пицца'),
	-- Паста и Ризотто
	((SELECT id FROM rb), 'Лингвини алио олио', NULL, 870000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6b33b1b9-36c2-5c01-9d09-390dba4f6606.webp', 'Паста и Ризотто'),
	((SELECT id FROM rb), 'Казаречче с уткой и пармезаном', NULL, 1130000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c0ea5c54-0133-502c-a4ff-4a144af00e8f.webp', 'Паста и Ризотто'),
	-- Мясо и птица
	((SELECT id FROM rb), 'Шницель миланезе, листья романо и соус Цезарь', NULL, 1090000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8fb67011-7dd8-5c96-9f7f-e84e0d203d4e.webp', 'Мясо и птица'),
	((SELECT id FROM rb), 'Телячья печень по-венециански с картофельным пюре', NULL, 1010000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/484e4ede-525e-5e1b-9207-28f7579cf02e.webp', 'Мясо и птица'),
	-- Рыба и морепродукты
	((SELECT id FROM rb), 'Фритто мисто с креветками и кальмарами', NULL, 1140000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3895d6c5-4488-56c8-bd33-247db89b6aeb.webp', 'Рыба и морепродукты'),
	((SELECT id FROM rb), 'Краб-кейк, авокадо и яйцо пашот', 'Креветки, краб, филе лосося, филе трески, сливки…', 1530000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/18ecca0a-a93c-5ea5-8144-9c208126e8b8.webp', 'Рыба и морепродукты'),
	-- Овощи
	((SELECT id FROM rb), 'Батат фри, пармезан и трюфельный айоли', NULL, 830000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8659b558-4414-59e2-b771-28f761203d05.webp', 'Овощи'),
	((SELECT id FROM rb), 'Картофель фри с трюфельным маслом', 'Картофель фри, соль, масло трюфельное, сыр пармезан, кетчуп', 610000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1109edbf-0824-5c89-ae8c-ff9bc1c8f227.webp', 'Овощи'),
	-- Десерты
	((SELECT id FROM rb), 'Сметанник с клубникой', 'Шоколадный бисквит (масло сливочное, сахар, мука миндальная…', 870000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/240bc50d-698e-5efa-93a2-1b7f13bc1963.webp', 'Десерты'),
	((SELECT id FROM rb), 'Торта ди фармаджио', 'Печенье (пшеничная мука высшего сорта, сахар, яйцо…', 800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/59978169-0254-5a79-8d2c-3328133c7bc0.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Вода Nedra', 'Вода без газа в ассортименте', 780000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0624ca1e-14bd-5f08-b6da-372254eede5d.webp', 'Напитки'),
	((SELECT id FROM rb), 'Coca-cola', 'В ассортименте', 610000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/74d461cc-655d-5254-b5ec-cdc27b27ce51.webp', 'Напитки'),
	-- Утро в Mushrooms
	((SELECT id FROM rb), 'Атлантик с лососем', 'Бриошь, крем-сыр, яйцо пашот, лосось, шпинат, голландский…', 1160000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2e00cf62-fd5c-5db1-9c58-c5ac0843071b.webp', 'Утро в Mushrooms'),
	((SELECT id FROM rb), 'Трюфельный Флорентин', 'Бриошь, яйцо, масло растительное, горчица, масло сливочное…', 715000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d040afa2-041e-5bda-b7f6-0c95d568bb69.webp', 'Утро в Mushrooms'),
	-- Завтрак
	((SELECT id FROM rb), 'Омлет', 'Омлет из трех яиц, зелень, молоко', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c944f0f1-0d5a-57ab-9242-5e0906ad6809.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Яичница', 'Яичница из трех яиц', 460000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1f15b38f-7385-59a7-8d5c-ce3b055192a1.webp', 'Завтрак'),
	-- Здоровая еда
	((SELECT id FROM rb), 'Безглютеновая овсяная каша', 'Крупа геркулес без глютена, кокосовое молоко, соль, сахар', 640000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5f3ca006-a16f-5c38-8ec9-a8bc0b9a9766.webp', 'Здоровая еда'),
	((SELECT id FROM rb), 'Киноа боул с лососем', 'Киноа, лосось, редис, эдамамэ, яйцо пашот, гуакамоле…', 1380000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d5dc22d9-17aa-5aff-b522-df86d6bbd334.webp', 'Здоровая еда'),
	-- Выпечка
	((SELECT id FROM rb), 'Панкейки с клубникой и пеканом', 'Мука, яйцо, сахар, кефир, клубника, йогурт, сливки…', 610000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7206746e-ca3f-5288-a3b9-169c329b189c.webp', 'Выпечка'),
	((SELECT id FROM rb), 'Круассан с клубничным джемом', 'Круассан (мука, молоко, яйца, сахар, соль), сливочное…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c71da133-6664-557b-9a25-5d41306f1da1.webp', 'Выпечка'),
	-- Дополнительно к завтраку
	((SELECT id FROM rb), 'Шпинат', NULL, 330000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dcb412ae-f666-5681-adbf-e3c07a2687f9.webp', 'Дополнительно к завтраку'),
	((SELECT id FROM rb), 'Банан', NULL, 300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/71d018f7-fa9d-51f4-b319-e8a7fbceec97.webp', 'Дополнительно к завтраку');

-- Руки ВВерх! (id 27)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Руки ВВерх!' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Горячие блюда
	((SELECT id FROM rb), 'Сковорода от папы', 'Авторская сковородка от папы, Сергея Жукова!', 1150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3103ce5c-5348-515a-ac2e-6b55eb8b03f6.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Говяжий язык с картофелем (гриль)', NULL, 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3a332dc1-0a97-50b4-83be-d314f94f821d.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Бифштекс с яйцом пашот и бельгийским салатом', 'Нежный бифштекс из мраморной говядины, дополненный яйцом…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/43de2bc8-32a3-52f4-80bd-412884a40580.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Буженина с картофельным пюре и луком‑пореем', 'Нежная буженина из свиной шеи, дополненная бархатистым…', 910000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/48b9e060-b1a6-50f7-95c8-0cd5bf0e24c4.webp', 'Горячие блюда'),
	-- Супы
	((SELECT id FROM rb), 'Мамин борщ', 'Мамин борщ — это густой и ароматный…', 770000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7cb8d03b-1f1c-5e78-b6d9-068c99c5d51f.webp', 'Супы'),
	((SELECT id FROM rb), 'Том ям', 'Ароматный тайский суп, который покорит вас своим…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/82e7880e-4eb1-503a-94fb-f73d155020d7.webp', 'Супы'),
	-- Паста
	((SELECT id FROM rb), 'Паста Карбонара', 'Это блюдо из традиционной итальянской кухни, которое…', 770000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d49a76b2-5ce5-5d88-8c2f-082f72707edc.webp', 'Паста'),
	((SELECT id FROM rb), 'Пенне Четыре сыра', 'Нежный соус из четырёх видов сыра со…', 830000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e854a28c-9cb5-5e5a-aedb-1b6aaac7312e.webp', 'Паста'),
	((SELECT id FROM rb), 'Паста с креветками, пармезаном и соусом песто', NULL, 870000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/32a73ee4-80bd-56f2-813d-746ee6772a5e.webp', 'Паста'),
	((SELECT id FROM rb), 'Феттуччине с морепродуктами', 'Это блюдо из пасты с нежным сливочным…', 930000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/adc32900-fcaf-52ed-8a69-26aa9d7eba92.webp', 'Паста'),
	-- Бургеры
	((SELECT id FROM rb), 'Бургер с куриным филе', 'Сочная курица с насыщенным вкусом, густой кетчуп…', 830000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/69f1284a-db1d-51db-a674-b51677bd4ec7.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Кевин Спайси с картофелем фри', 'Сочная мраморная говядина, хрустящий бекон и сыр…', 1250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/df7f7672-6236-5a5d-93b5-4355bbdf1384.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Бургер с мраморной говядиной', 'Сочная говядина, обжаренная и положенная между свежими…', 930000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9a977015-73a6-51b2-b000-78457ea2deaf.webp', 'Бургеры'),
	-- Салаты
	((SELECT id FROM rb), 'Оливье по-советски', 'Хотите разнообразить меню?', 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9e1597ef-6496-5de0-9546-9db23044697a.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с ростбифом', 'Это блюдо подойдёт тем, кто хочет насладиться…', 890000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c675edf9-561b-53f8-8317-17c1a55b31d0.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат из морепродуктов', 'Нежный лосось, вкусные кальмары и креветки в…', 960000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d8469765-0860-5d43-ab80-440ec35e2875.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат из говяжьего языка с соусом айоли', 'Салат из говяжьего языка — это изысканное…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ba82a97a-1763-596a-b703-c0c8dd2078f8.webp', 'Салаты'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Чесночные гренки', 'Чесночные гренки – горячая закуска из поджаренного…', 410000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5099d07a-b119-5a17-b67c-091a46570673.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Куриные крылышки', 'Пикантные куриные крылышки с хрустящими овощами и…', 710000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f8c17d79-3e93-5272-8905-d9283db2adeb.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Кесадилья с курицей', NULL, 730000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6391e488-3a84-59e3-8247-25a0fb761fe1.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Индейка в темпуре с соусом спайси', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0f8f1240-ec0c-536d-86d7-53f53f222d5b.webp', 'Горячие закуски'),
	-- Закуски
	((SELECT id FROM rb), 'Брускетта со слабосоленым лососем', 'Брускетта со слабосолёной форелью – идеальный выбор…', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a807a67a-6ed7-5425-a46e-5d2d8079d13c.webp', 'Закуски'),
	((SELECT id FROM rb), 'Малосольная сельдь с отварным картофелем и', 'Это идеальное сочетание вкусов и текстур, которое…', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/787767fb-06ec-5f9c-b77b-1374dfc1b996.webp', 'Закуски'),
	((SELECT id FROM rb), 'Ассорти из закусок XXL', 'Фирменная закуска на компанию: рулетики из ветчины…', 1500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/affc6abd-f8e6-524b-b618-78afa0ae420f.webp', 'Закуски'),
	((SELECT id FROM rb), 'Сморреброды на бородинском хлебе', 'Нежный бородинский хлеб, сырная начинка, пикантные шпроты…', 540000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f2040e10-4a14-52b6-bcf6-de40ea37963e.webp', 'Закуски'),
	-- Гарниры
	((SELECT id FROM rb), 'Тосты из бородинского хлеба', 'Хрустящие тосты с насыщенным вкусом и ароматом.', 220000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e53dc1dc-701c-5b51-ae5e-528d63a28b59.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Картофель фри', 'Хрустящий картофель с кетчупом – беспроигрышный вариант…', 410000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/76e44587-a74e-575f-86ee-4cffe2834efc.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Овощи гриль', 'Универсальный гарнир к любому блюду.', 460000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/394c3f66-fd12-5407-bf34-9831637bedfe.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Картофель по‑деревенски', 'Картофельные дольки – идеальный гарнир к любому блюду.', 410000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/44881a97-651e-508e-ba14-909f1ac58ef6.webp', 'Гарниры'),
	-- Десерты
	((SELECT id FROM rb), 'Фруктовая тарелка', 'Этот десерт — настоящая феерия вкуса.', 1500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/165ded38-58f8-56ef-92c3-dfacc96893a2.webp', 'Десерты'),
	((SELECT id FROM rb), 'Классический Чизкейк', 'Этот десерт – настоящее наслаждение для гурманов.', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/66ddfde9-2115-5513-acff-59d4fd1055bf.webp', 'Десерты'),
	((SELECT id FROM rb), 'Медовик', 'Коржи, приготовленные с использованием мёда, в сочетании…', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d5aa8127-2a3c-51a9-b7a0-6820a0788388.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Сок J7 Яблоко-персик', 'Окунитесь в атмосферу летнего сада с напитком…', 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3c73b450-82a3-5cdd-aa84-ccd6653b9afe.webp', 'Напитки'),
	((SELECT id FROM rb), 'Сок J7 Яблоко', 'Освежающий напиток, воплощающий вкус сочных яблок…', 530000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dac366f8-98fe-57ce-a578-4452540a70f6.webp', 'Напитки'),
	((SELECT id FROM rb), 'Evervess Индиан Тоник', 'Лёгкий газированный напиток с характерной горчинкой хинина.', 280000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/32a15a61-9e31-5372-9662-82a24c1dca76.webp', 'Напитки'),
	((SELECT id FROM rb), 'Вода Aqua Minerale газированная', NULL, 200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b9b0ca5a-92bb-5130-8931-e03c4bae328b.webp', 'Напитки'),
	-- Соусы
	((SELECT id FROM rb), 'Цезарь', NULL, 79000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6d9d789e-b171-50b7-b418-7cfb5bf0bb63.webp', 'Соусы'),
	((SELECT id FROM rb), 'Терияки', NULL, 59000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bad415fe-857f-5441-a2c3-9ca942bfbaff.webp', 'Соусы'),
	((SELECT id FROM rb), 'Сладкий чили', NULL, 69000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a28df4c6-00a2-5c01-9707-cd9cd8b153c4.webp', 'Соусы'),
	((SELECT id FROM rb), 'Майонез', NULL, 59000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/07850780-5583-5671-96be-471b0b564223.webp', 'Соусы');

-- Такахули (id 28)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Такахули' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Тесто
	((SELECT id FROM rb), 'Хачапури Аджарский', 'Знаменитое грузинское хачапури из воздушного дрожжевого теста…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5c075215-d2ae-575d-b35f-393cb94f51e8.webp', 'Тесто'),
	((SELECT id FROM rb), 'Хачапури Мегрельский', 'Сдобная лепешка с двойной порцией сырной начинки…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e2b776e0-119c-59d3-8550-271de96b9455.webp', 'Тесто'),
	((SELECT id FROM rb), 'Ачма', 'Знаменитое грузинское сырное блюдо из слоёного теста…', 900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9a5f49e8-066b-5ccb-b8c1-8c23c4bfa2ca.webp', 'Тесто'),
	((SELECT id FROM rb), 'Хачапури с грибами и пудрой из трюфеля', 'Ароматное дрожжевое тесто с начинкой из шампиньонов…', 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ef310adf-490a-543f-b31b-6f0f7bea99fc.webp', 'Тесто'),
	((SELECT id FROM rb), 'Хинчапури с грибами', 'Пышная дрожжевая лепешка с тягучим сулугуни и моцареллой.', 1750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f8086182-62b6-5e66-9920-0bbfcc828712.webp', 'Тесто'),
	((SELECT id FROM rb), 'Хинчапури с сыром', 'Пышная дрожжевая лепешка с тягучим сулугуни и…', 1750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f8086182-62b6-5e66-9920-0bbfcc828712.webp', 'Тесто'),
	-- Салаты
	((SELECT id FROM rb), 'Салат грузинский с травами и овощами', 'Свежие помидоры и огурцы с грецкими орехами…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f915508d-095d-5883-970d-1daadb9c483f.webp', 'Салаты'),
	((SELECT id FROM rb), 'Рецепт от тети из Греции (Греческий салат)', 'Сочные помидоры и хрустящие огурцы, дополненные красным…', 1190000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/377db521-90ab-5924-adca-b8b5531e9fce.webp', 'Салаты'),
	((SELECT id FROM rb), 'Вяленая свекла, сырный мусс, карамельный арахис с', 'Нежная вяленая свекла с воздушным сырным муссом…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/30bb3af8-a464-5206-81ff-99ed69642917.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с томатами и джонджоли', 'Сочные помидоры с пикантными каперсами джонджоли и…', 970000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9975ab5c-3dc3-57ad-b930-239db1ef8958.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с цицмати и цыпленком', 'Нежное куриное филе с сочными помидорами, хрустящими…', 1050000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/835aba55-276e-5059-ae7c-e050c8fab045.webp', 'Салаты'),
	((SELECT id FROM rb), 'Цезарь с креветками', 'Креветки, салат романо, помидоры, соус цезарь, яйцо…', 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e863d50b-b342-5016-8692-c385edd0730d.webp', 'Салаты'),
	((SELECT id FROM rb), 'Цезарь с курицей', 'Куриная грудка филе, салат романо, помидоры, соус…', 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ed44e162-a384-5ff4-abdc-94ea327260e1.webp', 'Салаты'),
	-- Основное
	((SELECT id FROM rb), 'Каре, которое идет всем из ягненка', 'Нежное каре ягненка с ароматными травами подается…', 2800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/69a922c0-507b-52e5-94e2-ad3bd5c5a3c4.webp', 'Основное'),
	((SELECT id FROM rb), 'Окрыляющий шашлык из курицы', 'Сочное филе куриного бедра в ароматном маринаде…', 1050000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/23f5a0a7-ecdc-5d9e-9b4f-12fcb3474321.webp', 'Основное'),
	((SELECT id FROM rb), 'Долма с соусом тархун', 'Нежные виноградные листья, фаршированные ароматной смесью свинины…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4b5b2611-1dde-5f32-bf59-b813c3785e3c.webp', 'Основное'),
	((SELECT id FROM rb), 'Чахохбили', 'Нежное филе куриного бедра, тушённое с томатами…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/884e33db-7a04-5c35-a6cd-d10feaba70dc.webp', 'Основное'),
	((SELECT id FROM rb), 'Чашушули с томленой шеей бычка', 'Нежная шея бычка, томлённая с болгарским перцем…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e6291cac-656e-5e1a-a797-8a85ed34f2a0.webp', 'Основное'),
	((SELECT id FROM rb), 'Цыпленок по-чкмерски', 'Нежный цыпленок, запеченный в сливочно-чесночном соусе с…', 1550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/45c0131a-f254-5ea7-ab67-9043360c23a8.webp', 'Основное'),
	((SELECT id FROM rb), 'Томленая голяшка ягненка с булгуром и овощами', 'Нежная голяшка ягненка, томленная с овощами в…', 1900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3acc7c68-4673-55d1-bce9-8b22131cacdf.webp', 'Основное'),
	-- Супы
	((SELECT id FROM rb), 'Харчо с говядиной', 'Наваристый острый суп с говядиной, рисом, морковью…', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3335a07e-0741-572f-a311-ebc2f856a35e.webp', 'Супы'),
	((SELECT id FROM rb), 'Куриная чихиртма, кинза, тархун', 'Нежный куриный суп с ароматом эстрагона.', 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e40819d3-7330-5575-ae2a-514243ae72b7.webp', 'Супы'),
	((SELECT id FROM rb), 'Крем-суп из батата со страчателлой', 'Нежный крем-суп из сладкого батата с добавлением…', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a9051102-60c7-53ef-b388-12a7e39c1934.webp', 'Супы'),
	-- Закуски
	((SELECT id FROM rb), 'Сациви с курицей и пудрой из кинзы', 'Нежная куриная грудка в бархатном соусе из…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e7048895-12b3-54b4-878a-4075dc2e27ed.webp', 'Закуски'),
	((SELECT id FROM rb), 'Ассорти пхали', 'Классическая грузинская закуска - четыре нежных шарика…', 1350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2c5308a5-abc8-5367-90bd-1e2a3d927bde.webp', 'Закуски'),
	((SELECT id FROM rb), 'Хрустящие вешенки с зеленым ткемали', 'Хрустящие вешенки подаются с освежающим зеленым соусом ткемали', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/aa26d492-09fb-5547-bd3d-720b6e700554.webp', 'Закуски'),
	((SELECT id FROM rb), 'Печеный перец рамиро с соусом тоннато', 'Нежный печеный перец рамиро с изысканным соусом…', 1050000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/eedddb44-9e31-56bb-807d-8e0c0cafc468.webp', 'Закуски'),
	((SELECT id FROM rb), 'Хумус с фейхоа, фисташкой и эстрагоном', 'Нежный хумус из нута с экзотическими нотками…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0b72e5c1-538b-5f18-b28a-9ad8426d208a.webp', 'Закуски'),
	((SELECT id FROM rb), 'Домашние сыры с лавандовым медом', 'Ассорти из нежных домашних сыров сулугуни и моцарелла.', 1250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/db4db946-9306-5fa8-a13b-bce567f9d839.webp', 'Закуски'),
	((SELECT id FROM rb), 'Тартар из говядины с джонджоли', 'Нежная говяжья вырезка с пикантными каперсами джонджоли…', 1350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8abe3d0d-84bc-5d96-b89f-03607fb7e403.webp', 'Закуски'),
	-- Соусы
	((SELECT id FROM rb), 'Соус Мацони', 'Мацони (молоко 6%, закваска)', 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6b412e42-a755-53f0-9e92-56fd13bd834a.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Наршараб', 'Гранатовый сок, сахар, специи', 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f39b02e7-2187-551c-845e-f632448896c7.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Сацебели', 'Томаты пилати, вода питьевая, томатная паста, чеснок…', 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8c6ba877-7788-5078-8c5c-73847e7f00c5.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Чкмерули', 'Сливки 33%, лук репчатый, чеснок, тархун, масло…', 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/22a23b60-675e-5e71-a42b-9dd70688c82d.webp', 'Соусы'),
	((SELECT id FROM rb), 'Сметана', 'Сметана 20%', 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1cf17380-4dc8-51b9-9bad-0a92b65b382c.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Аджика', 'Перец чили дробленый, чеснок, специи шереули, соль, вода', 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d9555eec-8868-535a-b602-49e79ef672a8.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Ткемали зеленый', 'Соус ткемали зеленый, аджика острая, чеснок, шереули…', 150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1d09b546-57d8-51f3-911e-e78e867452da.webp', 'Соусы'),
	-- Десерты
	((SELECT id FROM rb), 'ПоШУшукаем', 'Молоко сгущеное, сыр креметте, сыр маскарпоне, яйцо…', 800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9e5a4203-283b-509b-beea-fd965e97a9f0.webp', 'Десерты'),
	((SELECT id FROM rb), 'Профитроли', 'Сливки 33%, молоко, яйцо куриное, масло сливочное…', 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a652f027-5f34-55f3-a1c5-a8ffd59eadbe.webp', 'Десерты'),
	((SELECT id FROM rb), 'Обожженый чизкейк с Изабеллой', 'Сыр креметте, сметана, яйцо куриное, сливки 33%…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/676b32e5-0e86-5901-b4f9-0784b99a7cd5.webp', 'Десерты');

-- Ванлав (id 29)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Ванлав' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Десерты в шоколаде
	((SELECT id FROM rb), 'Десерт Клубника в шоколаде', 'Нежное сочетание клубники и молочного бельгийского шоколада.', 948000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/acc59cc5-33f2-5dc6-8c0e-b40de119eca6.webp', 'Десерты в шоколаде'),
	((SELECT id FROM rb), 'Десерт Банан в шоколаде', 'Спелый банан, покрытый слоем молочного бельгийского шоколада Callebaut.', 384000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/175dcacf-773f-5a99-977d-82421bf72a78.webp', 'Десерты в шоколаде'),
	((SELECT id FROM rb), 'Десерт Малина в шоколаде', 'Спелая малина, покрытая белым и молочным бельгийским…', 828000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/df9958b3-e417-5cc5-9b51-8d6aae42dd1c.webp', 'Десерты в шоколаде'),
	-- Трайфлы
	((SELECT id FROM rb), 'Трайфл Сникерлав', 'Шоколадный бисквит с сырным кремом, соленой карамелью…', 588000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/88dec443-bb87-5499-b51b-b215805653c5.webp', 'Трайфлы'),
	((SELECT id FROM rb), 'Трайфл Дубай', 'Сочетание ярких вкусов и нежной текстуры из…', 828000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6fe71e32-7036-5f58-94e4-3af8b6a0e000.webp', 'Трайфлы'),
	((SELECT id FROM rb), 'Трайфл Наполеон', 'Нежное слоеное тесто со сливочным кремом, посыпанное…', 660000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/076b69e7-1bb4-514a-b006-16e58a032d11.webp', 'Трайфлы'),
	((SELECT id FROM rb), 'Трайфл Пинчер', 'Шоколадный бисквит и нежный сливочный крем со…', 660000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/346d65ea-6a37-5dcb-95c6-57cf34db685f.webp', 'Трайфлы'),
	((SELECT id FROM rb), 'Трайфл с вишней', 'Трайфл в основе которого шоколадный бисквит и…', 588000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0a982852-9587-5ef4-9c62-6054403b75cb.webp', 'Трайфлы'),
	-- Завтраки
	((SELECT id FROM rb), 'Сырники', 'Нежные сырники из творога с лёгкой ванильной…', 468000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/353b90db-cc55-5d60-b964-77855a3dda00.webp', 'Завтраки'),
	((SELECT id FROM rb), 'Каша Абрикос с урбечом', 'Нежная каша с тонким ароматом ванили и…', 348000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/19f97e2b-8bcf-5058-bf18-6f5cd70af858.webp', 'Завтраки'),
	((SELECT id FROM rb), 'Сэндвич Сандо', 'Сытный, сочный и многослойный взрыв вкуса с…', 588000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dd8fda8d-1d72-5f23-96fc-af2315e534cb.webp', 'Завтраки'),
	-- Десерты
	((SELECT id FROM rb), 'Чизкейк Сан-Себастьян порционный', 'Сливочный чизкейк с кремовой текстурой и карамельными нотками.', 708000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6a112048-43c0-5468-9f70-0974708ffc0c.webp', 'Десерты'),
	((SELECT id FROM rb), 'Десерт Сэндвич', 'Бисквитное печенье с нежным сырным кремом, шоколадным…', 588000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0d0feab9-05f9-5811-915f-c6fb5bffd615.webp', 'Десерты'),
	((SELECT id FROM rb), 'Десерт Тирамису', 'Слои воздушного крема на основе сыра маскарпоне…', 516000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e80cccb8-1fc5-5891-ba6b-6e5331e4b7b2.webp', 'Десерты'),
	((SELECT id FROM rb), 'Десерт Кейк-попс в кокосовой стружке', 'Шоколадный бисквит со сгущенным молоком в бельгийском…', 252000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b49e0628-bf74-540c-9721-c3605bc5da16.webp', 'Десерты'),
	((SELECT id FROM rb), 'Десерт Кейк-попс фисташковый', 'Нежный шоколадный бисквит со сгущённым молоком и…', 252000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/688acb6c-e3be-5998-93eb-b74dbf117298.webp', 'Десерты'),
	((SELECT id FROM rb), 'Десерт Кейк-попс в вафельной стружке', 'Шоколадный бисквит со сгущенным молоком в бельгийском…', 252000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/52738d61-dfb0-54b3-95b8-3144f3f16a75.webp', 'Десерты'),
	-- Торты
	((SELECT id FROM rb), 'Торт Шоколадный Ванлав', 'Шоколадный бисквит, нежный сырный крем.', 3348000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/74d9ae9c-4d4b-580c-9fbb-5c762d355e3e.webp', 'Торты'),
	((SELECT id FROM rb), 'Торт Бенто Love', 'Воздушный ванильный бисквит с сырным кремом, прослойкой…', 1668000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/5822f1f9-7d06-58c6-a0f4-5453c56434e3.webp', 'Торты'),
	((SELECT id FROM rb), 'Торт Бенто Весенний', 'Морковный бисквит с нежным сырным кремом.', 1668000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/730af3a1-e1d2-58c7-9eda-962e334ff075.webp', 'Торты'),
	((SELECT id FROM rb), 'Торт Бенто Конструктор', 'Нежный бисквит с малиновой начинкой, прослоенный воздушным…', 1668000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6aaf8c40-8bd7-576c-bd82-07c258edd642.webp', 'Торты'),
	-- Рулеты
	((SELECT id FROM rb), 'Рулет Меренга-малина', 'Этот рулет сочетает в себе легкую воздушную…', 2146000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f30508c2-215c-59c8-a9f0-0e6ffb166d65.webp', 'Рулеты'),
	-- Пп Десерты
	((SELECT id FROM rb), 'Пп Трайфл Клубника-банан', 'Бисквитная крошка без сахара, нежный сырный крем…', 708000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3bacf568-d5ca-5ea1-9cef-6b8eb8758b77.webp', 'Пп Десерты'),
	((SELECT id FROM rb), 'Пп Трайфл Вишня', 'Нежное сочетание ингредиентов без сахара.', 660000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a8bbe070-c797-5336-8d44-ee58177b7c22.webp', 'Пп Десерты'),
	-- Печенье и Макарон
	((SELECT id FROM rb), 'Печенье Брауни', 'Половинки хрустящего печенья брауни на основе темного…', 348000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e610c0ac-3f38-5eef-94a7-483a561f457c.webp', 'Печенье и Макарон'),
	((SELECT id FROM rb), 'Печенье Брауни с кремом', 'Мягкий, нежный брауни с насыщенным вкусом тёмного…', 348000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/57e5e450-5c38-5200-b271-c484e3a2f90c.webp', 'Печенье и Макарон'),
	-- Эклеры
	((SELECT id FROM rb), 'Эклер Шоколадный', 'Заварной эклер с шоколадным заварным кремом.', 348000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2dbbd91f-aabb-5d9f-9691-a9eeaad3d5ea.webp', 'Эклеры'),
	((SELECT id FROM rb), 'Эклер Кракелин ваниль', 'Золотистый эклер с хрустящей корочкой и сливочной начинкой.', 276000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/098619a5-d152-57d9-a393-9b030fbab377.webp', 'Эклеры'),
	-- Мерч Ванлав
	((SELECT id FROM rb), 'Термоккружка Ванлав & Termy', NULL, 3600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ac7102cd-60de-5671-8fb4-bd45ea4d671f.webp', 'Мерч Ванлав'),
	((SELECT id FROM rb), 'Пакет Ванлав Галактик', NULL, 228000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d6e8b316-bcef-593f-80ad-21ae852e13df.webp', 'Мерч Ванлав'),
	((SELECT id FROM rb), 'Шоппер Моей Ванлав', NULL, 2400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/959f7550-e817-5314-aed4-8d69442fa1a6.webp', 'Мерч Ванлав'),
	((SELECT id FROM rb), 'Шоппер Сердце', NULL, 2160000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c7746a07-82dc-5178-a561-786f22fe750e.webp', 'Мерч Ванлав'),
	-- Напитки
	((SELECT id FROM rb), 'Напиток Кинза Черная смородина', NULL, 252000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4c08226f-a207-5173-8bac-b591513054f4.webp', 'Напитки'),
	((SELECT id FROM rb), 'Напиток Кинза Лимон Без Сахара', NULL, 252000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/beb55f30-31a5-58e9-b425-87bac665b048.webp', 'Напитки'),
	((SELECT id FROM rb), 'Напиток Кинза Лимон', NULL, 252000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/90786be0-b10f-55c6-8e20-c135e425db76.webp', 'Напитки'),
	((SELECT id FROM rb), 'Напиток Кинза Кола Без Сахара', NULL, 252000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/59b57032-d915-5cff-a05d-24644c3dd139.webp', 'Напитки'),
	((SELECT id FROM rb), 'Напиток Кинза Гранат', NULL, 252000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/779d9863-3e72-5dd4-84b9-e92db5737b2b.webp', 'Напитки'),
	-- Колд Брю
	((SELECT id FROM rb), 'Колд Брю Гранат', 'Кофейный напиток с гранатом, в котором сочетаются…', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/58856641-495d-529b-8c63-2440b9a17244.webp', 'Колд Брю'),
	((SELECT id FROM rb), 'Колд Брю Малина', 'Прохладный колд брю с малиновым акцентом…', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e394e41d-ea32-548c-b423-956e236fa759.webp', 'Колд Брю'),
	((SELECT id FROM rb), 'Колд Брю Смородина', 'Нежное сочетание с кисло-сладкими нотами, где мягкая…', 420000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c6ef6b40-8b84-51f0-85da-aeefc088c6d4.webp', 'Колд Брю');

-- Varvarka III (id 30)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Varvarka III' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Супы
	((SELECT id FROM rb), 'Борщ из утки с вяленой грушей', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/55267994-e9ac-528b-837b-db9badb51df8.webp', 'Супы'),
	((SELECT id FROM rb), 'Крем-суп из белых грибов', 'Из серого хлеба мы нарезали и подсушили…', 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1abf439d-2720-539e-ba93-52832ca1b3d4.webp', 'Супы'),
	((SELECT id FROM rb), 'Том ям с крабом и грибами эноки', 'Готовим бульон из обжаренных панцирей креветок, лука…', 1990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a951995d-36a5-52c5-92ca-bb87f527aea7.webp', 'Супы'),
	-- Горячие блюда
	((SELECT id FROM rb), 'Котлеты из щуки', 'Мы смешали мякоть щуки, добавили немного яйца…', 1400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/66d036a2-26b9-5220-9a63-b1cc2dee4b1c.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Ростбиф из вырезки из ягненка', 'Обжариваем на сливочном масле мини-картофель с чесноком…', 2700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c0d3c52a-08d7-5db3-98ca-7537db73b23d.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Цыпленок жареный с соусом из кинзы и чеснока', NULL, 1590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3960fb26-e304-5aeb-9875-82d666c28c35.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Говяжьи щечки с птитимом', NULL, 1500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/948e1ed3-c507-5dfe-83ca-5b61f269ca18.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Филе дорадо со спаржей', NULL, 1900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/55f9c615-6ee0-5940-9d85-ee0b1b74f716.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Котлета по-киевски', 'Куриная котлета по-киевски с картофельным пюре с…', 1800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0410bdeb-03b3-530a-b49f-e0522e302d43.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Котлеты из краба с картофельным пюре со шпинатом', 'Котлеты из краба и креветок обжаривают на гриле.', 1900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a85cbad1-acf1-5cbb-842f-e07fb66e1ba3.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Pork стейк с печеным перцем и томатным соусом', 'Маринуем свиную шейку и обжариваем на гриле.', 1150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7d7eb9ed-737c-5db2-9876-1a25a176af4b.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Долма с гусем', 'Долма с сочным гусем в соусе из…', 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/db407ff0-6758-5be1-92ac-55be99701428.webp', 'Горячие блюда'),
	-- Вприкуску
	((SELECT id FROM rb), 'Паштет из печени цыпленка и утки', NULL, 1350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4623cf2d-be98-5464-bb33-8c7ff1e7a201.webp', 'Вприкуску'),
	((SELECT id FROM rb), 'Хумус и тёплая лепешка', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b238660b-220d-51b7-8ca7-5a4a98ef5bf4.webp', 'Вприкуску'),
	((SELECT id FROM rb), 'Бабагануш из баклажанов и теплая лепешка', NULL, 840000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/152ce207-6719-586b-883f-2e46bcd1fff8.webp', 'Вприкуску'),
	((SELECT id FROM rb), 'Мухаммара', NULL, 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bf29ba72-465d-50da-bbb0-1bd8e74cfcbc.webp', 'Вприкуску'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Голубцы из краба и креветок', NULL, 1950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1ce3c347-d2cd-5356-96d7-c57e7713b6ac.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Мини-чебуреки с рубленой бараниной', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/38fa9fc3-c338-5ea9-abb4-fed7a925be3f.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Блинчики с мясом по-кутаиски', NULL, 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1ba004e3-864b-5050-a653-7267dc29a204.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Хачапури по-мегрельски', NULL, 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/57ca7011-dcd3-52c4-8bde-fa090aed01bb.webp', 'Горячие закуски'),
	-- Салаты
	((SELECT id FROM rb), 'Салат с крабом', NULL, 1900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3f6ddcb3-81b1-5dbe-ae1c-6619f0f57019.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат из помидоров с халуми и крымским луком', NULL, 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b11c2077-fcf8-5e1e-802c-6879c1d2d24a.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с креветками, авокадо и томатами', 'Сочетание нежного авокадо, сочных томатов и креветок…', 1650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/82e744d8-e876-5d00-bec9-0058a10aa89b.webp', 'Салаты'),
	((SELECT id FROM rb), 'Овощной салат', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/46487b3e-ddb9-5388-b9f8-309168ca88cf.webp', 'Салаты'),
	((SELECT id FROM rb), 'Теплый салат с телятиной и кабачками', 'Горчичная заправка из зернистой горчицы, лука-шалота, прованских…', 1400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/24bbfc2e-cff5-51e8-b08c-b980d42a999b.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с магаданскими креветками и кальмаром', 'Отварные креветки соединяем с отварным яйцом и…', 1490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/41ab7875-d2ac-5b6b-9500-191d1a327c05.webp', 'Салаты'),
	((SELECT id FROM rb), 'Стейк-салат из вырезки из воронежского бычка', 'Перемешиваем помидоры, лук-шалот, зеленый лук, шпинат, кинзу…', 1940000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4659092f-c57a-5de1-918d-2f0ad158bd1b.webp', 'Салаты'),
	-- Лепка и ризотто
	((SELECT id FROM rb), 'Пельмени Домашние из фермерской говядины', 'В начинке мы смешали говядину, немного лука…', 900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8426a917-529e-5b15-8f4c-7787f8d38ccf.webp', 'Лепка и ризотто'),
	((SELECT id FROM rb), 'Черные пельмени с тремя видами рыб', NULL, 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fb70ee1a-72c8-5726-949e-4d60e660f171.webp', 'Лепка и ризотто'),
	((SELECT id FROM rb), 'Ризотто с копченым угрем и муссом из пармезана', NULL, 1400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/47f1e06c-94f3-5728-bba3-616c5f1a3b12.webp', 'Лепка и ризотто'),
	-- Гарниры
	((SELECT id FROM rb), 'Картофель фри', 'Жареный во фритюре картофель фри.', 350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9f7e6502-341b-5a46-97d0-4bca0afbe676.webp', 'Гарниры'),
	-- Десерты
	((SELECT id FROM rb), 'Сигара', NULL, 800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/83cee035-d29f-5acf-945c-414cf32fbdaa.webp', 'Десерты'),
	((SELECT id FROM rb), 'Сметанник с голубикой и лесным орехом', 'Воздушный сметанный бисквит, пропитанный ароматным кофе, чередуется…', 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6fe14330-8609-5350-8374-465a4ee9c28c.webp', 'Десерты');

-- Villa Pasta (id 31)
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Villa Pasta' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Пицца
	((SELECT id FROM rb), 'Пицца Пепперони', NULL, 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8aea7c2a-7220-5b3e-9156-909566e75908.webp', 'Пицца'),
	((SELECT id FROM rb), 'Пицца Маргарита', NULL, 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/57a698eb-d606-5f5a-8a89-f6fec6fa77e1.webp', 'Пицца'),
	((SELECT id FROM rb), 'Пицца Деревенская', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2a106b23-1bd8-5821-9e25-bdf6af05de99.webp', 'Пицца'),
	-- Горячие блюда из птицы
	((SELECT id FROM rb), 'Куриные котлетки', NULL, 900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/66a3263c-d526-52e6-9914-f07cc0873e83.webp', 'Горячие блюда из птицы'),
	-- Паста
	((SELECT id FROM rb), 'Спагетти Болоньезе', NULL, 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d53b0fd5-0a50-569d-aa3a-607581489079.webp', 'Паста'),
	((SELECT id FROM rb), 'Спагетти Карбонара', NULL, 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/90ea0532-7a23-5355-ba4d-032f82be72ef.webp', 'Паста'),
	((SELECT id FROM rb), 'Казаречче с томлёными щечками', NULL, 1250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3e92edc6-c26a-5ba5-9600-b365d1863780.webp', 'Паста'),
	-- Салаты
	((SELECT id FROM rb), 'Салат Греческий', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/96d68553-e71f-5576-be5e-588ec6782ef1.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат с тунцом, артишоками и мини-картофелем', NULL, 1400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e6479603-66ca-5a48-8dd9-1f76fa1fe414.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат Руккола с креветками', NULL, 1250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6f58e3b6-f1f6-59ef-9271-def02eb60aa2.webp', 'Салаты'),
	-- Рыбные горячие блюда
	((SELECT id FROM rb), 'Лосось на гриле с овощами и пюре из сельдерея', NULL, 2400000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/522f5469-cec9-52fb-b117-50c7b927917b.webp', 'Рыбные горячие блюда'),
	((SELECT id FROM rb), 'Камбала по-домашнему с овощами в соусе', NULL, 1350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c4426bd1-68c9-5f86-b3df-4d2e865f549b.webp', 'Рыбные горячие блюда'),
	((SELECT id FROM rb), 'Дорадо с молодым картофелем и шпинатом', NULL, 1950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6479bfd1-05a5-5386-9d0b-16a506cb3b02.webp', 'Рыбные горячие блюда'),
	-- Супы
	((SELECT id FROM rb), 'Куриный суп с домашней лапшой', NULL, 600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dbde1c83-8c6f-5519-bc47-6e85974a8cfa.webp', 'Супы'),
	((SELECT id FROM rb), 'Суп овощной Минестроне', NULL, 600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0d4bf230-e566-5590-be2a-b38caaaf3cb3.webp', 'Супы'),
	((SELECT id FROM rb), 'Крем-суп из тыквы со страчателлой', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/48757298-2eee-5a5b-9270-384504c238e6.webp', 'Супы'),
	-- Холодные закуски
	((SELECT id FROM rb), 'Севиче из сибаса', NULL, 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1b263882-91a2-5d59-8ad8-eceed27691dd.webp', 'Холодные закуски'),
	((SELECT id FROM rb), 'Прошутто с дыней', NULL, 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8c2f8f7b-bd56-5f92-8eb7-c1c5fc5ee39b.webp', 'Холодные закуски'),
	((SELECT id FROM rb), 'Ассорти рыбное', NULL, 1500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ed37727c-2d45-5885-ae64-65ce67f888b6.webp', 'Холодные закуски'),
	-- Ризотто
	((SELECT id FROM rb), 'Ризотто с морепродуктами', NULL, 1300000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/3a92a350-670b-5333-bf5a-0e5c64c53b0b.webp', 'Ризотто'),
	((SELECT id FROM rb), 'Ризотто с чернилами каракатиц и кальмарами', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e398c77e-4d96-5bfd-9b24-a414bdcf1b63.webp', 'Ризотто'),
	((SELECT id FROM rb), 'Ризотто с лесными грибами', NULL, 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/d737cb8f-9d84-56e8-bc1d-10b50714af1b.webp', 'Ризотто'),
	-- Мясные горячие блюда
	((SELECT id FROM rb), 'Стейк Стриплойн', NULL, 2900000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/46e93181-72a8-5911-bbfc-ab74475e3260.webp', 'Мясные горячие блюда'),
	((SELECT id FROM rb), 'Стейк Рибай', NULL, 3700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b37a7db8-54b1-5061-802d-db1c139565cd.webp', 'Мясные горячие блюда'),
	((SELECT id FROM rb), 'Бургер из мраморной говядины', NULL, 1200000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4a3daf78-ec38-5d88-ae9a-d68a1bc7fd36.webp', 'Мясные горячие блюда'),
	-- Горячие закуски
	((SELECT id FROM rb), 'Ассорти из морепродуктов', NULL, 3800000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ec7dfaae-fc62-54da-9454-b8e74a31f904.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Кальмары с чесночным маслом и гренками', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/edbde15e-0ba1-58d8-a189-f8a085a1b091.webp', 'Горячие закуски'),
	((SELECT id FROM rb), 'Хрустящий баклажан с муссом фета', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6a58b0a7-2752-5d8c-9e2b-d2a932fb3dd9.webp', 'Горячие закуски'),
	-- Гарниры
	((SELECT id FROM rb), 'Картофель фри', NULL, 500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/1e5fa343-7f93-52d8-aba5-51a6db1b3f07.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Картофель по-деревенски', NULL, 500000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2415ec0d-7b6f-569a-bed8-fb4401ddaff4.webp', 'Гарниры'),
	-- Десерты
	((SELECT id FROM rb), 'Десерт Анна Павлова', NULL, 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bfecaf45-74a4-5fca-89c9-c6da365ed0e4.webp', 'Десерты'),
	((SELECT id FROM rb), 'Панна котта с ягодным соусом', NULL, 600000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/713d1e16-0be0-51b5-a84c-9f32a945f5dc.webp', 'Десерты'),
	-- Напитки
	((SELECT id FROM rb), 'Coca-Cola', NULL, 450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/97775761-8fe2-5d66-8ca5-b1d3e7bc219a.webp', 'Напитки'),
	((SELECT id FROM rb), 'Coca-Cola Zero', NULL, 450000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/b7ee61b5-ecdc-5fa9-8738-162b83d4d396.webp', 'Напитки'),
	-- Соусы
	((SELECT id FROM rb), 'Майонез', NULL, 230000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/afb2c4b5-aa0e-57be-a956-211865108551.webp', 'Соусы'),
	((SELECT id FROM rb), 'Кетчуп', NULL, 230000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/afb2c4b5-aa0e-57be-a956-211865108551.webp', 'Соусы'),
	-- Летнее меню
	((SELECT id FROM rb), 'Салат из сладких томатов с базиликом', NULL, 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/f0554994-f2b0-5eee-b94c-fd2434d3660b.webp', 'Летнее меню'),
	((SELECT id FROM rb), 'Тартар из сладких креветок с цитрусовой заправкой', NULL, 1100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/24bfa816-ab5e-5d11-a0b2-92d58fe3a030.webp', 'Летнее меню'),
	-- Завтрак
	((SELECT id FROM rb), 'Омлет с крабом и авокадо', NULL, 1000000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/97b70d19-a7bc-5f23-bf18-ee26bc5e8bca.webp', 'Завтрак'),
	((SELECT id FROM rb), 'Омлет с мортаделлой и сыром страчателла', NULL, 700000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4bb61558-1977-5027-8493-54e71fc04ca4.webp', 'Завтрак');


-- Torro Grill (id 14) — заменил закрывшийся Arium Grill
WITH rb AS (SELECT id FROM "restaurant_brand" WHERE name = 'Torro Grill' LIMIT 1)
INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, section) VALUES
	-- Стейки
	((SELECT id FROM rb), 'Стейк Гурме Острый мачете', NULL, 2950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e6f069d1-355b-52b1-8098-3f0d6805f608.webp', 'Стейки'),
	((SELECT id FROM rb), 'Стейк Гурме Мачете', NULL, 2750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/770106c1-6e9d-5dc8-9c5d-be76d3d77152.webp', 'Стейки'),
	((SELECT id FROM rb), 'Стейк Классик Лонг Айленд', 'Стейк с жировыми вкраплениями из лопаточной части…', 2550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/debb24dd-da2b-5c55-951d-7b6429dee000.webp', 'Стейки'),
	((SELECT id FROM rb), 'Филе Миньон', NULL, 3990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c99c7d60-a13b-56f2-b69d-f085371c360d.webp', 'Стейки'),
	((SELECT id FROM rb), 'Стейк Классик рибай', 'Премиальный стейк из говядины.', 4990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/dab83899-7d65-59a3-8f97-dca79c0ac589.webp', 'Стейки'),
	-- Горячие блюда
	((SELECT id FROM rb), 'Свиные ребрышки барбекю', 'Запеченные в соусе BBQ свиные ребра.', 1390000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/717dc9fe-80e4-5e0b-9619-06f618583021.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Кальмары гриль', 'Кальмары, приготовленные на гриле с добавлением ароматного…', 790000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c91ac32e-9a18-5f46-9093-41f0f4af14c6.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Бифштекс из говядины с сыром', 'Бифштекс из 100%-й говядины с пикантными специями…', 1350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8c3b7c31-4dfb-545f-bb7a-b58d8a35c6c6.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Белый морской окунь (Баррамунди), жаренный на', 'Белый морской окунь Баррамунди, жаренный на гриле.', 1590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e7407811-02a6-5510-b689-0315d23441b4.webp', 'Горячие блюда'),
	((SELECT id FROM rb), 'Креветки, обжаренные на гриле', NULL, 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/76627f37-b5a6-55f2-af0a-f51fc6e821c7.webp', 'Горячие блюда'),
	-- Закуски
	((SELECT id FROM rb), 'Куриные крылья барбекю', NULL, 950000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0958e1d6-a419-5268-a738-e0e2817ec921.webp', 'Закуски'),
	((SELECT id FROM rb), 'Ростбиф Тоннато', 'Ростбиф представлен в виде тонко нарезанных ломтиков…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/77c13cbf-63a7-5876-a95c-daafbd7830c3.webp', 'Закуски'),
	((SELECT id FROM rb), 'Кукуруза на гриле', 'Сладкая кукуруза в специях (оливковое масло, соль…', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/a736c989-5da0-5773-851d-27d6d82ad338.webp', 'Закуски'),
	((SELECT id FROM rb), 'Брускетта с ростбифом', 'Слайсы ростбифа и листья руколы в фирменной…', 490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/caa07159-5492-59ab-9ae6-ee3c2f464842.webp', 'Закуски'),
	-- Салаты
	((SELECT id FROM rb), 'Стейк-салат из мраморной говядины', NULL, 1190000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/29ebb248-73ec-5a90-a542-b0d42bad9b7f.webp', 'Салаты'),
	((SELECT id FROM rb), 'Цезарь с курицей', NULL, 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/8a8ab97d-dc28-5bf8-a55b-50698b64db4d.webp', 'Салаты'),
	((SELECT id FROM rb), 'Цезарь с креветками', NULL, 750000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ebca479a-d205-57da-bdd2-45215345b72e.webp', 'Салаты'),
	((SELECT id FROM rb), 'Салат из свежих овощей', NULL, 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/60288adc-1ce1-5d37-b6bd-cebec6c45a6d.webp', 'Салаты'),
	-- Бургеры
	((SELECT id FROM rb), 'Бургер Дон Антонио', 'Котлета из 100% говядины, приготовленная на гриле…', 1150000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e969067f-9e59-5436-96d1-d20fae6c8e2e.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Чизбургер с релишем', 'Котлета из говядины, приготовленная на гриле, фирменная…', 850000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9f1f147b-f720-585c-b00d-175209df0162.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Фирменный бургер', 'Две котлеты из говядины, приготовленные на гриле…', 1490000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7b760cef-a75e-5d66-ba2a-9b48d3ae4b0c.webp', 'Бургеры'),
	((SELECT id FROM rb), 'Бургер с сыром раклет', 'Котлета из говядины, жаренная на гриле с…', 990000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/7c53b555-2f6a-50da-b099-fab5bd797c4e.webp', 'Бургеры'),
	-- Гарниры
	((SELECT id FROM rb), 'Овощи гриль', NULL, 690000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/289d16cd-f40c-5e26-bf5b-9c31809a55ed.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Коул слоу с кейлом', NULL, 350000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/ce0cb06a-d2d6-579c-b99f-f5ee8f600e2f.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Хрустящий батат фри', NULL, 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/9a719297-aff1-510c-a422-a09508fe0a32.webp', 'Гарниры'),
	((SELECT id FROM rb), 'Спаржа на гриле', NULL, 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/0485bce1-d96f-5b7b-ace1-ab3036047150.webp', 'Гарниры'),
	-- Десерты
	((SELECT id FROM rb), 'Эклеры', NULL, 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/4da590b2-cb27-5ece-9c3b-ea5f0f999479.webp', 'Десерты'),
	((SELECT id FROM rb), 'Малиновый суп с ягодами', NULL, 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/093ff7ce-0501-5f16-b89c-39d4b21d81dd.webp', 'Десерты'),
	((SELECT id FROM rb), 'Чизкейк', 'Классический баскский чизкейк с соусом из ягод малины.', 650000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fd5cef8a-60bd-5238-9d48-29023933256b.webp', 'Десерты'),
	((SELECT id FROM rb), 'Профитроли с карамелью', 'Профитроли (3 шт) с начинкой из крема…', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/600c8aff-5768-574b-8b79-edb0ea32b4ff.webp', 'Десерты'),
	-- Соусы
	((SELECT id FROM rb), 'Кетчуп', NULL, 100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/bbc32108-818f-5415-bfda-d0a923f3ecec.webp', 'Соусы'),
	((SELECT id FROM rb), 'Майонез', NULL, 100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/e04c4aec-371a-5435-9822-3e0cddfd1ba9.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус Барбекю', NULL, 100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c4899f75-e5e4-5ef6-98ae-8990e5825a2d.webp', 'Соусы'),
	((SELECT id FROM rb), 'Соус песто', NULL, 100000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/820f2c24-0598-5b1e-b616-2287b2e93c66.webp', 'Соусы'),
	-- Возьми Torro с собой!
	((SELECT id FROM rb), 'Чизбургер TG', 'Котлета из говядины замороженная, булочка бриошь, огуречный…', 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/fb2bb62d-9976-5293-b0fe-b08c36872197.webp', 'Возьми Torro с собой!'),
	((SELECT id FROM rb), 'Хрустящий багет замороженный', NULL, 210000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/c460bbda-e391-5bb1-8e75-b566a136f70a.webp', 'Возьми Torro с собой!'),
	((SELECT id FROM rb), 'Хрустящий багет запеченный', NULL, 250000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2d3a7147-6737-5760-b7cb-fc6f4d9c7bf4.webp', 'Возьми Torro с собой!'),
	((SELECT id FROM rb), 'Фантастический ростбиф Торро Гриль', 'Фирменный запеченный ростбиф Torro Grill', 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/6f9831a7-56fb-5b0d-b161-1c4d6f710de7.webp', 'Возьми Torro с собой!'),
	-- Супы
	((SELECT id FROM rb), 'Крем-суп капучино из лесных грибов', NULL, 590000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/2b34583e-b3d8-5eb4-87cd-689751daeebb.webp', 'Супы'),
	((SELECT id FROM rb), 'Суп с говяжьими щечками, зеленью и овощами', 'Фирменный густой суп с разварными говяжьими щёчками…', 550000000, 'https://nancats-bucket.storage.yandexcloud.net/foods/09ec0ad5-36e6-558c-a798-47c2a23dcaf0.webp', 'Супы');
