CREATE TABLE IF NOT EXISTS public.bible_versions (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS public.bible_books (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    testament VARCHAR(2) NOT NULL CHECK (testament IN ('PL', 'PB')),
    book_order INTEGER NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS public.bible_verses (
    id BIGSERIAL PRIMARY KEY,
    version_id BIGINT NOT NULL REFERENCES public.bible_versions(id) ON DELETE CASCADE,
    book_id BIGINT NOT NULL REFERENCES public.bible_books(id) ON DELETE CASCADE,
    chapter INTEGER NOT NULL CHECK (chapter > 0),
    verse INTEGER NOT NULL CHECK (verse > 0),
    text TEXT NOT NULL,
    UNIQUE (version_id, book_id, chapter, verse)
);

CREATE INDEX IF NOT EXISTS idx_bible_verses_chapter
    ON public.bible_verses (version_id, book_id, chapter, verse);

CREATE INDEX IF NOT EXISTS idx_bible_verses_text_search
    ON public.bible_verses USING GIN (to_tsvector('simple', text));