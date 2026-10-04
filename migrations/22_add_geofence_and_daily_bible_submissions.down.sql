DROP TABLE IF EXISTS public.bible_reading_submissions;

ALTER TABLE public.attendance_records
    DROP CONSTRAINT IF EXISTS attendance_matched_location_type_valid,
    DROP CONSTRAINT IF EXISTS attendance_location_accuracy_valid,
    DROP CONSTRAINT IF EXISTS attendance_distance_valid,
    DROP COLUMN IF EXISTS matched_location_type,
    DROP COLUMN IF EXISTS matched_location_id,
    DROP COLUMN IF EXISTS matched_location_name,
    DROP COLUMN IF EXISTS distance_meters,
    DROP COLUMN IF EXISTS location_accuracy_m;

ALTER TABLE public.activity_occurrences
    DROP CONSTRAINT IF EXISTS activity_occurrences_coordinates_pair,
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS longitude;

ALTER TABLE public.sites
    DROP CONSTRAINT IF EXISTS sites_attendance_coordinates_pair,
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS longitude;
