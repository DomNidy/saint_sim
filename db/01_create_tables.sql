
-- Create sequence for simulation_data ID if it does not exist
CREATE SEQUENCE IF NOT EXISTS public.simulation_data_id_seq;

-- Create simulation_request table
CREATE TABLE IF NOT EXISTS public.simulation_request
(
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    received_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT simulation_request_pkey PRIMARY KEY (id)
);

-- Create simulation_data table
CREATE TABLE IF NOT EXISTS public.simulation_data
(
    id integer NOT NULL DEFAULT nextval('simulation_data_id_seq'::regclass),
    from_request UUID NOT NULL REFERENCES public.simulation_request(id) ON DELETE CASCADE,
    sim_result text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT simulation_data_pkey PRIMARY KEY (id)
);

-- Set the table owners to 'saint'
ALTER TABLE IF EXISTS public.simulation_data
    OWNER to saint;

ALTER TABLE IF EXISTS public.simulation_request
    OWNER to saint;



-- WITH rows AS (
-- 	insert into public.simulation_request default values
-- 	returning id as "rowid"
-- ) insert into public.simulation_data (from_request, sim_result) values ((select * from rows), 'asd');

-- select * from public.simulation_data 
-- join  public.simulation_request
-- on public.simulation_data.from_request = public.simulation_request.id;