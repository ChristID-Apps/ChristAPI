ALTER TABLE public.sites
    ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION;

ALTER TABLE public.activity_occurrences
    ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION;

ALTER TABLE public.attendance_records
    ADD COLUMN IF NOT EXISTS matched_location_type VARCHAR(20),
    ADD COLUMN IF NOT EXISTS matched_location_id BIGINT,
    ADD COLUMN IF NOT EXISTS matched_location_name TEXT,
    ADD COLUMN IF NOT EXISTS distance_meters DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS location_accuracy_m DOUBLE PRECISION;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'sites_attendance_coordinates_pair') THEN
        ALTER TABLE public.sites ADD CONSTRAINT sites_attendance_coordinates_pair
            CHECK ((latitude IS NULL AND longitude IS NULL) OR
                   (latitude IS NOT NULL AND longitude IS NOT NULL AND
                    latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'activity_occurrences_coordinates_pair') THEN
        ALTER TABLE public.activity_occurrences ADD CONSTRAINT activity_occurrences_coordinates_pair
            CHECK ((latitude IS NULL AND longitude IS NULL) OR
                   (latitude IS NOT NULL AND longitude IS NOT NULL AND
                    latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'attendance_matched_location_type_valid') THEN
        ALTER TABLE public.attendance_records ADD CONSTRAINT attendance_matched_location_type_valid
            CHECK (matched_location_type IS NULL OR matched_location_type IN ('site', 'activity'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'attendance_location_accuracy_valid') THEN
        ALTER TABLE public.attendance_records ADD CONSTRAINT attendance_location_accuracy_valid
            CHECK (location_accuracy_m IS NULL OR (location_accuracy_m >= 0 AND location_accuracy_m <= 100));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'attendance_distance_valid') THEN
        ALTER TABLE public.attendance_records ADD CONSTRAINT attendance_distance_valid
            CHECK (distance_meters IS NULL OR (distance_meters >= 0 AND distance_meters <= 500));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS public.bible_reading_submissions (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    reading_date DATE NOT NULL,
    payload JSONB NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    status VARCHAR(20) NOT NULL DEFAULT 'submitted'
        CHECK (status IN ('submitted', 'approved', 'rejected')),
    reviewer_note TEXT,
    reviewed_by BIGINT REFERENCES public.users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    points_awarded BIGINT NOT NULL DEFAULT 0 CHECK (points_awarded >= 0),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT bible_reading_approved_points_positive
        CHECK (status <> 'approved' OR points_awarded > 0)
);

CREATE INDEX IF NOT EXISTS idx_bible_reading_submissions_user_date
    ON public.bible_reading_submissions(user_id, reading_date DESC);
CREATE INDEX IF NOT EXISTS idx_bible_reading_submissions_review_queue
    ON public.bible_reading_submissions(status, submitted_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uq_bible_reading_one_active_per_user_day
    ON public.bible_reading_submissions(user_id, reading_date)
    WHERE status IN ('submitted', 'approved');
