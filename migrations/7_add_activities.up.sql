CREATE TABLE IF NOT EXISTS public.activity_categories (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS public.activities (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category_id BIGINT NOT NULL REFERENCES public.activity_categories(id),
    activity_type VARCHAR(20) NOT NULL CHECK (activity_type IN ('one_time', 'recurring')),
    site_id BIGINT REFERENCES public.sites(id),
    created_by BIGINT REFERENCES public.users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'cancelled', 'completed')),
    max_participants INTEGER,
    requires_registration BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS public.activity_schedules (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT NOT NULL UNIQUE REFERENCES public.activities(id) ON DELETE CASCADE,
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly')),
    interval_value INTEGER NOT NULL DEFAULT 1 CHECK (interval_value > 0),
    days_of_week JSONB NOT NULL DEFAULT '[]'::jsonb,
    day_of_month INTEGER CHECK (day_of_month BETWEEN 1 AND 31),
    start_date DATE NOT NULL,
    end_date DATE,
    start_time TIME,
    end_time TIME,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS public.activity_occurrences (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT NOT NULL UNIQUE REFERENCES public.activities(id) ON DELETE CASCADE,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ,
    location TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'open', 'completed', 'cancelled')),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

INSERT INTO public.activity_categories (code, name, description) VALUES
    ('OUTING', 'Outing', 'Kegiatan outing atau rekreasi'),
    ('IBADAH', 'Ibadah', 'Kegiatan ibadah'),
    ('GABUNGAN', 'Gabungan', 'Kegiatan gabungan'),
    ('NATAL', 'Natal', 'Perayaan Natal'),
    ('KOMSEL', 'Komsel', 'Komunitas sel'),
    ('BACA_ALKITAB', 'Baca Alkitab', 'Kegiatan baca Alkitab'),
    ('LAINNYA', 'Lainnya', 'Kategori kegiatan lainnya')
ON CONFLICT (code) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_activities_category_id ON public.activities(category_id);
CREATE INDEX IF NOT EXISTS idx_activities_site_id ON public.activities(site_id);
CREATE INDEX IF NOT EXISTS idx_activities_status ON public.activities(status);
CREATE INDEX IF NOT EXISTS idx_activity_occurrences_starts_at ON public.activity_occurrences(starts_at);
