ALTER TABLE public.news DROP CONSTRAINT IF EXISTS fk_christapi_news_site_id;
ALTER TABLE public.news DROP CONSTRAINT IF EXISTS fk_christapi_news_author_id;
ALTER TABLE public.roles DROP CONSTRAINT IF EXISTS fk_christapi_roles_site_id;
ALTER TABLE public.contacts DROP CONSTRAINT IF EXISTS fk_christapi_contacts_site_id;
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS fk_christapi_users_site_id;
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS fk_christapi_users_role_id;
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS fk_christapi_users_contact_id;
