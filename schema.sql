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
-- Name: race_organizers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.race_organizers (
    race_id uuid NOT NULL,
    user_id uuid NOT NULL
);


--
-- Name: race_registered_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.race_registered_users (
    race_id uuid NOT NULL,
    user_id uuid NOT NULL,
    registered_at timestamp with time zone NOT NULL
);


--
-- Name: races; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.races (
    id uuid NOT NULL,
    name character varying(255) NOT NULL,
    start_at timestamp with time zone NOT NULL,
    is_open_for_registration boolean NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(128) NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    username character varying(255) NOT NULL,
    password_hash bytea NOT NULL
);


--
-- Name: race_organizers race_organizers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.race_organizers
    ADD CONSTRAINT race_organizers_pkey PRIMARY KEY (race_id, user_id);


--
-- Name: race_registered_users race_registered_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.race_registered_users
    ADD CONSTRAINT race_registered_users_pkey PRIMARY KEY (race_id, user_id);


--
-- Name: races races_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.races
    ADD CONSTRAINT races_name_key UNIQUE (name);


--
-- Name: races races_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.races
    ADD CONSTRAINT races_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: race_organizers race_organizers_race_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.race_organizers
    ADD CONSTRAINT race_organizers_race_id_fkey FOREIGN KEY (race_id) REFERENCES public.races(id);


--
-- Name: race_organizers race_organizers_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.race_organizers
    ADD CONSTRAINT race_organizers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: race_registered_users race_registered_users_race_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.race_registered_users
    ADD CONSTRAINT race_registered_users_race_id_fkey FOREIGN KEY (race_id) REFERENCES public.races(id);


--
-- Name: race_registered_users race_registered_users_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.race_registered_users
    ADD CONSTRAINT race_registered_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20230504203320'),
    ('20230505202311'),
    ('20230507193814'),
    ('20230521110535');
