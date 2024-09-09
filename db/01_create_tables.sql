CREATE SEQUENCE IF NOT EXISTS public.simulaton_data_id_seq;

CREATE TABLE IF NOT EXISTS public.simulation_data
(
    id integer NOT NULL DEFAULT nextval('simulaton_data_id_seq'::regclass),
    sim_result text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT simulaton_data_pkey PRIMARY KEY (id)
)


TABLESPACE pg_default;


ALTER TABLE IF EXISTS public.simulation_data
    OWNER to saint;


