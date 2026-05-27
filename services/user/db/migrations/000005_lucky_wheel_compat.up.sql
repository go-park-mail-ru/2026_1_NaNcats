-- Добивка схемы user-сервиса под Lucky Wheel: поля client_profile и ачивки.
-- Раньше эти изменения были засунуты прямо в 000001/000004; на проде, где
-- старые миграции уже накатились, новых столбцов и строк не появилось.
-- Этот файл навёрстывает разницу идемпотентно (IF NOT EXISTS / ON CONFLICT).

ALTER TABLE "client_profile"
    ADD COLUMN IF NOT EXISTS streak_freeze_active BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE "client_profile"
    ADD COLUMN IF NOT EXISTS last_wheel_spin_at TIMESTAMP WITH TIME ZONE;

INSERT INTO "achievement" (code, title, description, icon, sort_order) VALUES
    ('first_spin', 'Испытатель удачи', 'Запустите Колесо Пиццули в первый раз', '🌀', 4),
    ('lucky_wheel_winner', 'Любимчик Пиццули', 'Выиграйте эксклюзивную награду в Колесе Пиццули', '🎡', 5),
    ('streak_six', 'Постоянство', 'Поддерживайте серию заказов 6 недель подряд', '🔥', 6)
ON CONFLICT (code) DO NOTHING;
