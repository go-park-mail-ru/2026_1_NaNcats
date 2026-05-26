-- Откат сид-карт первого клиента (user_id = 3).
DELETE FROM "payment_method" WHERE user_id = 3;
