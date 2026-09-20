CREATE TABLE IF NOT EXISTS public.rewards (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    points_required BIGINT NOT NULL CHECK (points_required > 0),
    stock INTEGER NOT NULL CHECK (stock >= 0),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_by BIGINT REFERENCES public.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.reward_redemptions (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    reward_id BIGINT NOT NULL REFERENCES public.rewards(id),
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    points_required BIGINT NOT NULL CHECK (points_required > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'completed')),
    admin_note TEXT,
    user_code VARCHAR(64) UNIQUE,
    admin_code VARCHAR(64) UNIQUE,
    approved_at TIMESTAMPTZ,
    rejected_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rewards_status ON public.rewards(status);
CREATE INDEX IF NOT EXISTS idx_reward_redemptions_user_id ON public.reward_redemptions(user_id);
CREATE INDEX IF NOT EXISTS idx_reward_redemptions_status ON public.reward_redemptions(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_one_pending_reward_per_user
    ON public.reward_redemptions(user_id, reward_id) WHERE status = 'pending';
