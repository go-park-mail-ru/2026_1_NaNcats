Relation user:
{id} -> phone, name, email, password_hash, bonus_balance, bonus_category_id, bonus_category_expires_at, bonus_expires_at, streak_count, last_order_date, premium_expires_at, created_at, updated_at

Обоснование:
Отношение в 1НФ: все атрибуты атомарны
Отношение во 2НФ: первичный ключ `{id}` состоит из одного атрибута, поэтому частичных зависимостей неключевых атрибутов от PK быть не может
Отношение в 3НФ и НФБК: в таблице есть 3 потенциальных ключа (id, phone, email). Все остальные атрибуты напрямую зависят от конкретного пользователя и не зависят друг от друга. Транзитивных зависимостей нет


Relation restaurant:
{id} -> name, address, promotion_tier, created_at, updated_at

Обоснование:
Отношение в 1НФ: все атрибуты атомарны
Отношение во 2НФ: первичный ключ `{id}` состоит из одного атрибута, поэтому частичных зависимостей неключевых атрибутов от PK быть не может
Отношение в 3НФ и НФБК: в таблице есть 1 потенциальный ключ (id). Все остальные атрибуты напрямую зависят от конкретного предприятия и не зависят друг от друга. Транзитивных зависимостей нет


Relation order:
{id} -> user_id, restaurant_id, courier_id, promocode_id, delivery_address, status, created_at, updated_at

Обоснование:
Отношение в 1НФ: все атрибуты атомарны
Отношение во 2НФ: первичный ключ `{id}` состоит из одного атрибута, поэтому частичных зависимостей неключевых атрибутов от PK быть не может
Отношение в 3НФ и НФБК: в таблице есть 1 потенциальный ключ (id). Все остальные атрибуты напрямую зависят от конкретного заказа и не зависят друг от друга. Транзитивных зависимостей нет


Relation order_dish:
{order_id, dish_id} -> quantity, price, created_at

Обоснование:
Отношение в 1НФ: все атрибуты атомарны
Отношение во 2НФ: первичный ключ `{order_id}, {dish_id}` состоит из двух атрибутов, но неключевые атрибуты зависят одновременно от этих двух ключевых атрибутов (они не могут зависеть только от одного, в данном случае)
Отношение находится в  3НФ и НФБК, так как все функциональные зависимости сводятся к зависимости от ({order_id}) и ({dish_id}) одновременно. Также нет транзитивных зависимостей (ни один неключевой атрибут на зависит от другого неключевого атрибута)


updated_at тут не нужен, так как заказ создается 1 раз


Relation order_review:
{order_id} -> restaurant_rating, courier_rating, client_comment, created_at

Обоснование:
Отношение в 1НФ: все атрибуты атомарны
Отношение во 2НФ: первичный ключ `{order_id}` состоит из одного атрибута, поэтому частичных зависимостей неключевых атрибутов от PK быть не может
Отношение в 3НФ и НФБК: в таблице есть 1 потенциальный ключ (order_id). Все остальные атрибуты напрямую зависят от конкретного отзыва и не зависят друг от друга. Транзитивных зависимостей нет

updated_at тут не нужен, так как отзыв на заказ оставляется 1 раз


Relation dish:
{id} -> restaurant_id, category_id, name, description, price, created_at, updated_at

Обоснование:
Отношение в 1НФ: все атрибуты атомарны
Отношение во 2НФ: первичный ключ `{id}` состоит из одного атрибута, поэтому частичных зависимостей неключевых атрибутов от PK быть не может
Отношение в 3НФ и НФБК: в таблице есть 1 потенциальный ключ (id). Все остальные атрибуты напрямую зависят от конкретного блюда и не зависят друг от друга. Транзитивных зависимостей нет

price не нарушает 3НФ, поскольку в dish - это текущая цена блюда, а в order_dish - историческая стоимость блюда (в бывшем заказе)


Relation category:
{id} -> name, created_at, updated_at

Обоснование:
Отношение в 1НФ: все атрибуты атомарны
Отношение во 2НФ: первичный ключ `{id}` состоит из одного атрибута, поэтому частичных зависимостей неключевых атрибутов от PK быть не может
Отношение в 3НФ и НФБК: в таблице есть 2 потенциальных ключа (id, name). Все остальные атрибуты напрямую зависят от конкретной категории и не зависят друг от друга. Транзитивных зависимостей нет


Relation courier:
{id} -> name, phone, status, created_at, updated_at

Обоснование:
Отношение в 1НФ: все атрибуты атомарны
Отношение во 2НФ: первичный ключ `{id}` состоит из одного атрибута, поэтому частичных зависимостей неключевых атрибутов от PK быть не может
Отношение в 3НФ и НФБК: в таблице есть 2 потенциальных ключа (id, phone). Все остальные атрибуты напрямую зависят от конкретного курьера и не зависят друг от друга. Транзитивных зависимостей нет


Relation promocode:
{id} -> code, discount_percent, discount_amount, created_at, expires_at

Обоснование:
Отношение в 1НФ: все атрибуты атомарны
Отношение во 2НФ: первичный ключ `{id}` состоит из одного атрибута, поэтому частичных зависимостей неключевых атрибутов от PK быть не может
Отношение в 3НФ и НФБК: в таблице есть 2 потенциальных ключа (id, code). Все остальные атрибуты напрямую зависят от конкретного промокода и не зависят друг от друга. Транзитивных зависимостей нет


```mermaid
erDiagram
    %% Описание сущностей
    user {
        int id PK
        text phone UK "NOT NULL, UNIQUE"
        text name "NOT NULL"
        text email UK "NOT NULL, UNIQUE"
        text password_hash "NOT NULL"
        enum role "NOT NULL"
        datetime created_at "DEFAULT NOW(), NOT NULL"
        datetime updated_at "DEFAULT NOW(), NOT NULL"
    }

    client_profile {
        int account_id PK, FK
        int bonus_balance
        int bonus_category_id FK
        datetime bonus_category_expires_at
        datetime bonus_expires_at
        int streak_count
        datetime last_order_date
        datetime premium_expires_at
    }
    
    courier_profile {
        int account_id PK, FK
        text status "NOT NULL"
    }

    owner_profile {
        int account_id PK, FK
        int restaurant_brand_id FK
    }

    restaurant_brand {
        int id PK
        text name "NOT NULL"
        text description
        int promotion_tier
        datetime created_at "DEFAULT NOW()"
        datetime updated_at "DEFAULT NOW()"
    }

    restaurant_branch {
        int id PK
        int brand_id FK
        int location_id FK
        time open_time
        time close_time
        datetime created_at "DEFAULT NOW()"
        datetime updated_at "DEFAULT NOW()"
    }

    order {
        int id PK
        int client_account_id FK "NOT NULL"
        int courier_account_id FK "SET NULL"
        int restaurant_branch_id FK "NOT NULL"
        int client_address_id FK "NOT NULL"
        int promocode_id FK
        text status "NOT NULL"
        datetime created_at "DEFAULT NOW()"
        datetime updated_at "DEFAULT NOW()"
    }

    order_dish {
        int order_id PK, FK
        int dish_id PK, FK
        int quantity "NOT NULL"
        int price "NOT NULL"
        datetime created_at "DEFAULT NOW()"
    }

    order_review {
        int order_id PK, FK
        int restaurant_rating "NOT NULL"
        int courier_rating "NOT NULL"
        text client_comment
        datetime created_at "DEFAULT NOW()"
    }

    dish {
        int id PK
        int restaurant_brand_id FK "NOT NULL"
        text name "NOT NULL"
        text description
        int price "NOT NULL"
        datetime created_at "DEFAULT NOW()"
        datetime updated_at "DEFAULT NOW()"
    }


    promocode {
        int id PK
        text code UK "NOT NULL, UNIQUE"
        int discount_percent
        int discount_amount
        boolean is_global
        datetime created_at "DEFAULT NOW()"
        datetime expires_at "NOT NULL"
    }

    promocode_restaurant_brand {
        int promocode_id PK, FK
        int restaurant_brand_id PK, FK
    }

    promocode_category {
        int promocode_id PK, FK
        int category_id PK, FK
    }


    category {
        int id PK
        text name UK "UNIQUE"
        datetime created_at "DEFAULT NOW()"
        datetime updated_at "DEFAULT NOW()"
    }

    restaurant_brand_category {
        int restaurant_brand_id PK, FK
        int category_id PK, FK
    }

    dish_category {
        int dish_id PK, FK
        int category_id PK, FK
    }


    location {
        int id PK
        text address_text
        %% latitude - широта; longitude - долгота
        numeric latitude
        numeric longitude
        datetime created_at "DEFAULT NOW()"
        datetime updated_at "DEFAULT NOW()"
    }

    client_address {
        int id PK
        int client_account_id FK
        int location_id FK
        text apartment
        text entrance
        text floor
        text door_code
        text courier_comment
        text label "House, work, etc"
        datetime created_at "DEFAULT NOW()"
        datetime updated_at "DEFAULT NOW()"
    }


    %% Фейковая табличка редиса
    Redis_Session_Store {
        text info "Токены, сессии, корзина"
    }


    %% Описание связей
    user ||--|| client_profile: ""
    user ||--|| courier_profile: ""
    user ||--|| owner_profile: ""
    user ||--|{ Redis_Session_Store: ""

    courier_profile ||--o{ order: ""

    restaurant_brand ||--|{ restaurant_branch: ""
    restaurant_brand ||--|{ dish: ""
    restaurant_brand ||--o{ restaurant_brand_category: ""
    restaurant_brand ||--o{ promocode_restaurant_brand: ""
    restaurant_brand ||--|| owner_profile: ""

    restaurant_branch ||--o{ order: ""

    category ||--o{ client_profile: ""
    category ||--o{ dish_category: ""
    category ||--o{ promocode_category: ""
    category ||--o{ restaurant_brand_category: ""

    promocode ||--o{ order: ""
    promocode ||--o{ promocode_category: ""
    promocode ||--o{ promocode_restaurant_brand: ""

    order ||--|{ order_dish: ""
    order ||--o| order_review: ""

    dish ||--o{ order_dish: ""
    dish ||--|{ dish_category: ""

    location ||--o{ client_address: ""
    location ||--o{ restaurant_branch: ""

    client_profile ||--o{ client_address: ""
    client_profile ||--o{ order: ""

    client_address ||--o{ order: ""
```