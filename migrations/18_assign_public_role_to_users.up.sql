UPDATE public.users
SET role_id = public_role.id,
    updated_at = NOW()
FROM public.roles AS public_role
WHERE public_role.code = 'public'
  AND public.users.role_id IS NULL;