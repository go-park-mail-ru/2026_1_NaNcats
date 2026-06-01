--
-- PostgreSQL database dump
--

\restrict adXetY7FIvRFWTp3u95JGcHFe7iQgbgxE1F4iuKgnLqIjfLPTYoB9p9dlLdNUHz

-- Dumped from database version 18.4
-- Dumped by pg_dump version 18.4

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.category (
    id bigint NOT NULL,
    name text NOT NULL,
    emoji text DEFAULT ''::text NOT NULL,
    idempotency_key text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: category_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.category ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.category_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: dish; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.dish (
    id bigint NOT NULL,
    restaurant_brand_id bigint NOT NULL,
    name text NOT NULL,
    description text,
    image_url text,
    price bigint NOT NULL,
    idempotency_key text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    section text,
    CONSTRAINT dish_description_check CHECK ((char_length(description) <= 1000)),
    CONSTRAINT dish_image_url_check CHECK ((char_length(image_url) <= 2048)),
    CONSTRAINT dish_name_check CHECK ((char_length(name) <= 50)),
    CONSTRAINT dish_price_check CHECK ((price >= 1000000)),
    CONSTRAINT dish_section_check CHECK ((char_length(section) <= 60))
);


--
-- Name: dish_category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.dish_category (
    dish_id bigint NOT NULL,
    category_id bigint NOT NULL,
    idempotency_key text
);


--
-- Name: dish_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.dish ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.dish_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: restaurant_branch; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.restaurant_branch (
    id bigint NOT NULL,
    restaurant_brand_id bigint NOT NULL,
    location_id bigint NOT NULL,
    open_time time without time zone,
    close_time time without time zone,
    idempotency_key text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: restaurant_branch_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.restaurant_branch ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.restaurant_branch_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: restaurant_brand; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.restaurant_brand (
    id bigint NOT NULL,
    owner_profile_id bigint NOT NULL,
    name text NOT NULL,
    description text,
    promotion_tier integer DEFAULT 0 NOT NULL,
    logo_url text,
    idempotency_key text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    banner_url text,
    CONSTRAINT restaurant_brand_banner_url_check CHECK ((char_length(banner_url) <= 2048)),
    CONSTRAINT restaurant_brand_description_check CHECK ((char_length(description) <= 500)),
    CONSTRAINT restaurant_brand_logo_url_check CHECK ((char_length(logo_url) <= 2048)),
    CONSTRAINT restaurant_brand_name_check CHECK ((char_length(name) <= 60)),
    CONSTRAINT restaurant_brand_promotion_tier_check CHECK (((promotion_tier >= 0) AND (promotion_tier <= 3)))
);


--
-- Name: restaurant_brand_category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.restaurant_brand_category (
    restaurant_brand_id bigint NOT NULL,
    category_id bigint NOT NULL,
    idempotency_key text
);


--
-- Name: restaurant_brand_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.restaurant_brand ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.restaurant_brand_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: category category_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: category category_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_name_key UNIQUE (name);


--
-- Name: category category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.category
    ADD CONSTRAINT category_pkey PRIMARY KEY (id);


--
-- Name: dish_category dish_category_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dish_category
    ADD CONSTRAINT dish_category_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: dish_category dish_category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dish_category
    ADD CONSTRAINT dish_category_pkey PRIMARY KEY (dish_id, category_id);


--
-- Name: dish dish_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dish
    ADD CONSTRAINT dish_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: dish dish_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dish
    ADD CONSTRAINT dish_pkey PRIMARY KEY (id);


--
-- Name: restaurant_branch restaurant_branch_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restaurant_branch
    ADD CONSTRAINT restaurant_branch_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: restaurant_branch restaurant_branch_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restaurant_branch
    ADD CONSTRAINT restaurant_branch_pkey PRIMARY KEY (id);


--
-- Name: restaurant_brand_category restaurant_brand_category_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restaurant_brand_category
    ADD CONSTRAINT restaurant_brand_category_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: restaurant_brand_category restaurant_brand_category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restaurant_brand_category
    ADD CONSTRAINT restaurant_brand_category_pkey PRIMARY KEY (restaurant_brand_id, category_id);


--
-- Name: restaurant_brand restaurant_brand_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restaurant_brand
    ADD CONSTRAINT restaurant_brand_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: restaurant_brand restaurant_brand_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restaurant_brand
    ADD CONSTRAINT restaurant_brand_name_key UNIQUE (name);


--
-- Name: restaurant_brand restaurant_brand_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restaurant_brand
    ADD CONSTRAINT restaurant_brand_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: dish_category fk_dish_category_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dish_category
    ADD CONSTRAINT fk_dish_category_category FOREIGN KEY (category_id) REFERENCES public.category(id) ON DELETE RESTRICT;


--
-- Name: dish_category fk_dish_category_dish; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dish_category
    ADD CONSTRAINT fk_dish_category_dish FOREIGN KEY (dish_id) REFERENCES public.dish(id) ON DELETE RESTRICT;


--
-- Name: dish fk_dish_restaurant_brand; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dish
    ADD CONSTRAINT fk_dish_restaurant_brand FOREIGN KEY (restaurant_brand_id) REFERENCES public.restaurant_brand(id) ON DELETE CASCADE;


--
-- Name: restaurant_branch fk_restaurant_branch_restaurant_brand; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restaurant_branch
    ADD CONSTRAINT fk_restaurant_branch_restaurant_brand FOREIGN KEY (restaurant_brand_id) REFERENCES public.restaurant_brand(id) ON DELETE CASCADE;


--
-- Name: restaurant_brand_category fk_restaurant_brand_category_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restaurant_brand_category
    ADD CONSTRAINT fk_restaurant_brand_category_category FOREIGN KEY (category_id) REFERENCES public.category(id) ON DELETE CASCADE;


--
-- Name: restaurant_brand_category fk_restaurant_brand_category_restaurant_brand; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.restaurant_brand_category
    ADD CONSTRAINT fk_restaurant_brand_category_restaurant_brand FOREIGN KEY (restaurant_brand_id) REFERENCES public.restaurant_brand(id) ON DELETE RESTRICT;


--
-- PostgreSQL database dump complete
--

\unrestrict adXetY7FIvRFWTp3u95JGcHFe7iQgbgxE1F4iuKgnLqIjfLPTYoB9p9dlLdNUHz

