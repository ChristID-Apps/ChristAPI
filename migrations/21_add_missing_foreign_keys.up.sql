DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'public.users'::regclass
          AND confrelid = 'public.contacts'::regclass
          AND conname IN ('users_contact_id_fkey', 'fk_christapi_users_contact_id')
    ) THEN
        ALTER TABLE public.users
            ADD CONSTRAINT fk_christapi_users_contact_id
            FOREIGN KEY (contact_id) REFERENCES public.contacts(id)
            ON DELETE SET NULL NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'public.users'::regclass
          AND confrelid = 'public.roles'::regclass
          AND conname IN ('users_role_id_fkey', 'fk_christapi_users_role_id')
    ) THEN
        ALTER TABLE public.users
            ADD CONSTRAINT fk_christapi_users_role_id
            FOREIGN KEY (role_id) REFERENCES public.roles(id)
            ON DELETE SET NULL NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'public.users'::regclass
          AND confrelid = 'public.sites'::regclass
          AND conname IN ('users_site_id_fkey', 'fk_christapi_users_site_id')
    ) THEN
        ALTER TABLE public.users
            ADD CONSTRAINT fk_christapi_users_site_id
            FOREIGN KEY (site_id) REFERENCES public.sites(id)
            ON DELETE SET NULL NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'public.contacts'::regclass
          AND confrelid = 'public.sites'::regclass
          AND conname IN ('contacts_site_id_fkey', 'fk_christapi_contacts_site_id')
    ) THEN
        ALTER TABLE public.contacts
            ADD CONSTRAINT fk_christapi_contacts_site_id
            FOREIGN KEY (site_id) REFERENCES public.sites(id)
            ON DELETE SET NULL NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'public.roles'::regclass
          AND confrelid = 'public.sites'::regclass
          AND conname IN ('roles_site_id_fkey', 'fk_christapi_roles_site_id')
    ) THEN
        ALTER TABLE public.roles
            ADD CONSTRAINT fk_christapi_roles_site_id
            FOREIGN KEY (site_id) REFERENCES public.sites(id)
            ON DELETE SET NULL NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'public.news'::regclass
          AND confrelid = 'public.users'::regclass
          AND conname = 'fk_christapi_news_author_id'
    ) THEN
        ALTER TABLE public.news
            ADD CONSTRAINT fk_christapi_news_author_id
            FOREIGN KEY (author_id) REFERENCES public.users(id)
            ON DELETE SET NULL NOT VALID;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'public.news'::regclass
          AND confrelid = 'public.sites'::regclass
          AND conname = 'fk_christapi_news_site_id'
    ) THEN
        ALTER TABLE public.news
            ADD CONSTRAINT fk_christapi_news_site_id
            FOREIGN KEY (site_id) REFERENCES public.sites(id)
            ON DELETE SET NULL NOT VALID;
    END IF;
END $$;
