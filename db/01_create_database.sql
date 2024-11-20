--
-- PostgreSQL database dump
--

-- Dumped from database version 16.4 (Debian 16.4-1.pgdg120+1)
-- Dumped by pg_dump version 16.4 (Debian 16.4-1.pgdg120+1)

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

ALTER TABLE IF EXISTS ONLY public.simulation_data DROP CONSTRAINT IF EXISTS simulation_data_request_id_fkey;
DROP TRIGGER IF EXISTS update_timestamp_trigger ON public.api_keys;
DROP TRIGGER IF EXISTS new_simulation_data ON public.simulation_data;
DROP TRIGGER IF EXISTS new_sim_result ON public.simulation_data;
ALTER TABLE IF EXISTS ONLY public.simulation_request DROP CONSTRAINT IF EXISTS simulation_request_pkey;
ALTER TABLE IF EXISTS ONLY public.simulation_data DROP CONSTRAINT IF EXISTS simulation_data_pkey;
ALTER TABLE IF EXISTS ONLY public.api_keys DROP CONSTRAINT IF EXISTS api_keys_pkey;
ALTER TABLE IF EXISTS public.api_keys ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.simulaton_data_id_seq;
DROP TABLE IF EXISTS public.simulation_request;
DROP TABLE IF EXISTS public.simulation_data;
DROP SEQUENCE IF EXISTS public.simulation_data_id_seq;
DROP SEQUENCE IF EXISTS public.api_keys_id_seq;
DROP TABLE IF EXISTS public.api_keys;
DROP FUNCTION IF EXISTS public.update_timestamp();
DROP FUNCTION IF EXISTS public.notify_simulation_data();
DROP FUNCTION IF EXISTS public.notify_sim_result();
--
-- Name: notify_sim_result(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.notify_sim_result() RETURNS trigger
    LANGUAGE plpgsql
    AS $$







BEGIN







	PERFORM pg_notify('new_sim_result', NEW.id::text);







	RETURN NEW;







END;







$$;


--
-- Name: notify_simulation_data(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.notify_simulation_data() RETURNS trigger
    LANGUAGE plpgsql
    AS $$







BEGIN







	PERFORM pg_notify('new_simulation_data', NEW.request_id::text);







	RETURN NEW;







END;







$$;


--
-- Name: update_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: api_keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.api_keys (
    id integer NOT NULL,
    api_key character varying(255) NOT NULL,
    service_name character varying(100) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: api_keys_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.api_keys_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: api_keys_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.api_keys_id_seq OWNED BY public.api_keys.id;


--
-- Name: simulation_data_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.simulation_data_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: simulation_data; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.simulation_data (
    id integer DEFAULT nextval('public.simulation_data_id_seq'::regclass) NOT NULL,
    request_id uuid NOT NULL,
    sim_result text NOT NULL
);


--
-- Name: simulation_request; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.simulation_request (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    received_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    options jsonb NOT NULL
);


--
-- Name: simulaton_data_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.simulaton_data_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: api_keys id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keys ALTER COLUMN id SET DEFAULT nextval('public.api_keys_id_seq'::regclass);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- Name: simulation_data simulation_data_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.simulation_data
    ADD CONSTRAINT simulation_data_pkey PRIMARY KEY (id);


--
-- Name: simulation_request simulation_request_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.simulation_request
    ADD CONSTRAINT simulation_request_pkey PRIMARY KEY (id);


--
-- Name: simulation_data new_sim_result; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER new_sim_result AFTER INSERT ON public.simulation_data FOR EACH ROW EXECUTE FUNCTION public.notify_sim_result();


--
-- Name: simulation_data new_simulation_data; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER new_simulation_data AFTER INSERT ON public.simulation_data FOR EACH ROW EXECUTE FUNCTION public.notify_simulation_data();


--
-- Name: api_keys update_timestamp_trigger; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_timestamp_trigger BEFORE UPDATE ON public.api_keys FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- Name: simulation_data simulation_data_request_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.simulation_data
    ADD CONSTRAINT simulation_data_request_id_fkey FOREIGN KEY (request_id) REFERENCES public.simulation_request(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

