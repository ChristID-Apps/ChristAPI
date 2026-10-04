ALTER TABLE public.attendance_records
    DROP CONSTRAINT IF EXISTS attendance_matched_location_type_valid;

ALTER TABLE public.attendance_records
    ADD CONSTRAINT attendance_matched_location_type_valid
    CHECK (matched_location_type IS NULL OR matched_location_type IN ('site', 'activity', 'configured'));
