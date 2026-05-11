DROP INDEX IF EXISTS idx_user_points_ledger_created_at;
DROP INDEX IF EXISTS idx_user_points_ledger_user_id;
DROP TABLE IF EXISTS public.user_points_ledger;

ALTER TABLE public.users
DROP COLUMN IF EXISTS points_balance;
