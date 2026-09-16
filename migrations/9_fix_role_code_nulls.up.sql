UPDATE public.roles
SET code = CASE lower(name)
    WHEN 'super admin' THEN CASE WHEN NOT EXISTS (SELECT 1 FROM public.roles existing WHERE existing.code = 'super_admin') THEN 'super_admin' ELSE 'role_' || id::text END
    WHEN 'admin' THEN CASE WHEN NOT EXISTS (SELECT 1 FROM public.roles existing WHERE existing.code = 'admin') THEN 'admin' ELSE 'role_' || id::text END
    WHEN 'church staff' THEN CASE WHEN NOT EXISTS (SELECT 1 FROM public.roles existing WHERE existing.code = 'church_staff') THEN 'church_staff' ELSE 'role_' || id::text END
    WHEN 'public' THEN CASE WHEN NOT EXISTS (SELECT 1 FROM public.roles existing WHERE existing.code = 'public') THEN 'public' ELSE 'role_' || id::text END
    ELSE 'role_' || id::text
END
WHERE code IS NULL OR code = '';

ALTER TABLE public.roles
    ALTER COLUMN code SET NOT NULL;