-- Откат сид-клиента. client_profile удаляется каскадом (ON DELETE CASCADE).
DELETE FROM "user" WHERE email = 'anna@foodcourt.fun';
