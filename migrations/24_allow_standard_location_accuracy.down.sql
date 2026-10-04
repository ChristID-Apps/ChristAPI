ALTER TABLE public.attendance_records
    DROP CONSTRAINT IF EXISTS attendance_location_accuracy_valid;

ALTER TABLE public.attendance_records
    ADD CONSTRAINT attendance_location_accuracy_valid
    CHECK (location_accuracy_m IS NULL OR (location_accuracy_m >= 0 AND location_accuracy_m <= 100)) NOT VALID;
