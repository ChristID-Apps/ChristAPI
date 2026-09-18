DROP INDEX IF EXISTS idx_point_streak_logs_activity_id;
DROP INDEX IF EXISTS idx_point_streak_logs_user_id;
DROP INDEX IF EXISTS idx_point_streaks_user_id;
DROP TABLE IF EXISTS public.point_streak_logs;
DROP TABLE IF EXISTS public.point_streaks;
ALTER TABLE public.activities
    DROP COLUMN IF EXISTS streak_points,
    DROP COLUMN IF EXISTS streak_type,
    DROP COLUMN IF EXISTS streak_enabled;
