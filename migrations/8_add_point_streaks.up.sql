ALTER TABLE public.activities
    ADD COLUMN IF NOT EXISTS streak_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS streak_type VARCHAR(50),
    ADD COLUMN IF NOT EXISTS streak_points BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS public.point_streaks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    streak_type VARCHAR(50) NOT NULL,
    current_streak INTEGER NOT NULL DEFAULT 0,
    longest_streak INTEGER NOT NULL DEFAULT 0,
    last_activity_date DATE,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, streak_type)
);

CREATE TABLE IF NOT EXISTS public.point_streak_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    streak_type VARCHAR(50) NOT NULL,
    activity_id BIGINT NOT NULL REFERENCES public.activities(id) ON DELETE CASCADE,
    activity_date DATE NOT NULL,
    points_earned BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, streak_type, activity_date)
);

CREATE INDEX IF NOT EXISTS idx_point_streaks_user_id ON public.point_streaks(user_id);
CREATE INDEX IF NOT EXISTS idx_point_streak_logs_user_id ON public.point_streak_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_point_streak_logs_activity_id ON public.point_streak_logs(activity_id);
