ALTER TABLE public.users
ADD COLUMN IF NOT EXISTS points_balance BIGINT DEFAULT 0 NOT NULL;

CREATE TABLE IF NOT EXISTS public.user_points_ledger (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    change_amount BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    reason TEXT NOT NULL,
    reference_id TEXT,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_points_ledger_user_id ON public.user_points_ledger (user_id);
CREATE INDEX IF NOT EXISTS idx_user_points_ledger_created_at ON public.user_points_ledger (created_at DESC);
