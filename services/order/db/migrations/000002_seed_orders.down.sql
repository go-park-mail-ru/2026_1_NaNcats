-- Откат сид-заказов: удаляем заказы первого клиента (admin_account_id = 3).
-- order_dish и order_review удаляются каскадом (ON DELETE CASCADE).
DELETE FROM "order" WHERE admin_account_id = 3;
