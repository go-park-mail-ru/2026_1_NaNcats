-- Оптимизация листинга ресторанов (итерация 2).
-- Запросы GetRestaurantBrandsList / GetRestaurantBrandsByCategory используют:
--   ORDER BY promotion_tier DESC, id ASC LIMIT $1 OFFSET $2
-- что без индекса даёт Seq Scan + Top-N Sort всей таблицы на каждый запрос.
-- Составной btree-индекс в точном порядке сортировки позволяет планировщику
-- отдавать страницу обычным Index Scan без сортировки.

CREATE INDEX IF NOT EXISTS idx_restaurant_brand_tier_id
    ON "restaurant_brand" (promotion_tier DESC, id ASC);
