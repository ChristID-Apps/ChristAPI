UPDATE public.users
SET role_id = NULL,
    updated_at = NOW()
WHERE role_id = (SELECT id FROM public.roles WHERE code = 'public' LIMIT 1);
