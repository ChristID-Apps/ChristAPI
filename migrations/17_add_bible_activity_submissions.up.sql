CREATE TABLE IF NOT EXISTS public.activity_bible_configs (
    activity_id BIGINT PRIMARY KEY REFERENCES public.activities(id) ON DELETE CASCADE,
    version_code VARCHAR(20) NOT NULL DEFAULT 'tb',
    book_code VARCHAR(50) NOT NULL,
    start_chapter INTEGER NOT NULL CHECK (start_chapter > 0),
    start_verse INTEGER NOT NULL CHECK (start_verse > 0),
    end_chapter INTEGER NOT NULL CHECK (end_chapter > 0),
    end_verse INTEGER NOT NULL CHECK (end_verse > 0),
    requires_reflection BOOLEAN NOT NULL DEFAULT false,
    reflection_prompt TEXT,
    reflection_min_length INTEGER NOT NULL DEFAULT 0 CHECK (reflection_min_length >= 0)
);

CREATE TABLE IF NOT EXISTS public.activity_submissions (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    activity_id BIGINT NOT NULL REFERENCES public.activities(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    answer_text TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'submitted' CHECK (status IN ('submitted', 'approved', 'rejected')),
    reviewer_note TEXT,
    reviewed_by BIGINT REFERENCES public.users(id),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reviewed_at TIMESTAMPTZ,
    UNIQUE (activity_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_activity_submissions_status ON public.activity_submissions(status);
CREATE INDEX IF NOT EXISTS idx_activity_submissions_user_id ON public.activity_submissions(user_id);