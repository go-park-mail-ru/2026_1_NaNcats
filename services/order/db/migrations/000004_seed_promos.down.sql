-- Откат сид-промокодов. promocode_restaurant_brand, promocode_category,
-- promocode_usage и user_promocode удаляются каскадом (ON DELETE CASCADE).
DELETE FROM "promocode" WHERE code IN (
    'WELCOME300', 'EPIC30', 'FREEBURGER', 'VKUSNO20', 'SUMMER15',
    'PIZZA500', 'FRIENDS', 'STUDENT', 'NEWUSER', 'WEEKEND25'
);
