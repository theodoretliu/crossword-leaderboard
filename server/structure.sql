--
-- PostgreSQL database dump
--

-- Dumped from database version 15.2 (Debian 15.2-1.pgdg110+1)
-- Dumped by pg_dump version 15.4 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
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
-- Name: cookies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cookies (
    id bigint NOT NULL,
    value text NOT NULL
);


--
-- Name: cookies_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cookies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cookies_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cookies_id_seq OWNED BY public.cookies.id;


--
-- Name: examples; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.examples (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text,
    something integer NOT NULL
);


--
-- Name: examples_something_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.examples_something_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: examples_something_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.examples_something_seq OWNED BY public.examples.something;


--
-- Name: times; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.times (
    id bigint NOT NULL,
    time_in_seconds integer NOT NULL,
    date date,
    user_id bigint NOT NULL
);


--
-- Name: times_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.times_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: times_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.times_id_seq OWNED BY public.times.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    nyt_user_id bigint,
    name text
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: cookies id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookies ALTER COLUMN id SET DEFAULT nextval('public.cookies_id_seq'::regclass);


--
-- Name: examples something; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.examples ALTER COLUMN something SET DEFAULT nextval('public.examples_something_seq'::regclass);


--
-- Name: times id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.times ALTER COLUMN id SET DEFAULT nextval('public.times_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: cookies cookies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cookies
    ADD CONSTRAINT cookies_pkey PRIMARY KEY (id);


--
-- Name: examples examples_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.examples
    ADD CONSTRAINT examples_pkey PRIMARY KEY (id);


--
-- Name: times times_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.times
    ADD CONSTRAINT times_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: times_date_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX times_date_user_id_idx ON public.times USING btree (date, user_id);


--
-- Name: users_nyt_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX users_nyt_user_id_idx ON public.users USING btree (nyt_user_id);


--
-- PostgreSQL database dump complete
--

