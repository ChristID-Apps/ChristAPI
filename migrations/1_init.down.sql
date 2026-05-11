-- Initial schema migration (down)
-- Drops objects created by 1_init.up.sql. Use with caution (destructive).

DROP TABLE IF EXISTS public.news;
DROP TABLE IF EXISTS public.roles;
DROP TABLE IF EXISTS public.sites;
DROP TABLE IF EXISTS public.contacts;
DROP TABLE IF EXISTS public.users;

-- Note: leaving extensions as-is; if you want to remove pgcrypto uncomment below
-- DROP EXTENSION IF EXISTS pgcrypto;
