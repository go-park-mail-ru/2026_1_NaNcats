CREATE ROLE app_read_only;
CREATE ROLE app_read_write;

GRANT SELECT ON ALL TABLES IN SCHEMA public TO app_read_only;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_read_write;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_read_write;

CREATE USER user_service_user WITH PASSWORD 'secure_password_1';
CREATE USER order_service_user WITH PASSWORD 'secure_password_2';
CREATE USER cart_service_user WITH PASSWORD 'secure_password_3';

GRANT app_read_write TO user_service_user;
GRANT app_read_write TO order_service_user;
GRANT app_read_write TO cart_service_user;
