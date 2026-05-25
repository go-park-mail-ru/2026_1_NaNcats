CREATE ROLE app_read_write;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_read_write;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_read_write;

CREATE USER user_svc_user WITH PASSWORD 'user_pass_123';
CREATE USER restaurant_svc_user WITH PASSWORD 'rest_pass_123';
CREATE USER cart_svc_user WITH PASSWORD 'cart_pass_123';
CREATE USER address_svc_user WITH PASSWORD 'addr_pass_123';
CREATE USER payment_svc_user WITH PASSWORD 'pay_pass_123';
CREATE USER order_svc_user WITH PASSWORD 'order_pass_123';
CREATE USER support_svc_user WITH PASSWORD 'supp_pass_123';

GRANT app_read_write TO user_svc_user;
GRANT app_read_write TO restaurant_svc_user;
GRANT app_read_write TO cart_svc_user;
GRANT app_read_write TO address_svc_user;
GRANT app_read_write TO payment_svc_user;
GRANT app_read_write TO order_svc_user;
GRANT app_read_write TO support_svc_user;
