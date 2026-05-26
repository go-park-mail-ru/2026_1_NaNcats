USE analytics;

-- Таблица заказов
CREATE TABLE IF NOT EXISTS orders_report_log (
    event_time DateTime64(3, 'Europe/Moscow'),
    event_date Date DEFAULT toDate(event_time),

    order_public_id UUID,
    restaurant_id Int64,
    client_id Int64,

    total_cost_raw Int64,
    delivery_cost_raw Int64,
    service_fee_raw Int64,
    discount_raw Int64,
    -- Чистые деньги, которые достаются ресторану
    restaurant_revenue_raw Int64,

    status String,
    prev_status String,

    order_type String,
    members_count Int32,
    city String DEFAULT 'undefined',

    -- Будет 1 только при переходе в 'paid' чтобы не было дублей для денежных полей
    is_financial_impact Int8 DEFAULT 0 
) 
ENGINE = ReplacingMergeTree(event_time)
PARTITION BY toYYYYMM(event_date)
ORDER BY (restaurant_id, order_public_id, status);

-- Таблица блюд
CREATE TABLE IF NOT EXISTS order_items_report_log (
    event_time DateTime64(3, 'Europe/Moscow'),
    event_date Date DEFAULT toDate(event_time),

    order_public_id UUID,
    restaurant_id Int64,
    dish_id Int64,
    dish_name String,
    
    quantity Int32,
    price_raw Int64, 
    row_total_raw Int64,

    user_id Int64
) 
ENGINE = ReplacingMergeTree(event_time)
PARTITION BY toYYYYMM(event_date)
ORDER BY (restaurant_id, order_public_id, dish_id);