ALTER TABLE public.point_streak_logs
    DROP CONSTRAINT IF EXISTS point_streak_logs_user_streak_date_activity_key;

ALTER TABLE public.point_streak_logs
    ADD CONSTRAINT point_streak_logs_user_id_streak_type_activity_date_key
    UNIQUE (user_id, streak_type, activity_date);
