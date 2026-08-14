-- 0001_init.up.sql — initial schema for the portfolio backend (PostgreSQL).
-- UUID PKs use the native UUID type. Booleans use native BOOLEAN. Text uses
-- TEXT (unlimited) instead of VARCHAR2/CLOB. tech_stack uses JSONB for indexed
-- JSON storage. Enums are modeled as VARCHAR2-equivalent TEXT + CHECK constraints.
-- "current" is a reserved word in PostgreSQL and must stay double-quoted everywhere.
-- Naming: plural snake_case tables, {table}_id primary keys, idx_{table}_{column} indexes.

CREATE TABLE experiences (
    experience_id    UUID         PRIMARY KEY,
    experience_type  TEXT         NOT NULL CHECK (experience_type IN ('work', 'education')),
    organization     TEXT,
    role_title       TEXT,
    location         TEXT,
    start_date       DATE         NOT NULL,
    end_date         DATE,
    "current"        BOOLEAN      NOT NULL DEFAULT FALSE,
    summary_markdown TEXT         NOT NULL DEFAULT '',
    sort_order       INTEGER      NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE experience_highlights (
    experience_highlight_id UUID  PRIMARY KEY,
    experience_id           UUID  NOT NULL REFERENCES experiences(experience_id) ON DELETE CASCADE,
    body_markdown           TEXT  NOT NULL DEFAULT '',
    sort_order              INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE projects (
    project_id           UUID         PRIMARY KEY,
    slug                 TEXT         NOT NULL UNIQUE,
    title                TEXT         NOT NULL,
    tagline              TEXT,
    description_markdown TEXT         NOT NULL DEFAULT '',
    cover_image_url      TEXT,
    repo_url             TEXT,
    demo_url             TEXT,
    tech_stack           JSONB        NOT NULL DEFAULT '[]',
    status               TEXT         NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    featured             BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_order           INTEGER      NOT NULL DEFAULT 0,
    published_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE project_links (
    project_link_id UUID  PRIMARY KEY,
    project_id      UUID  NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    label           TEXT,
    url             TEXT,
    kind            TEXT  NOT NULL CHECK (kind IN ('repo', 'demo', 'docs', 'other'))
);

CREATE TABLE blog_posts (
    blog_post_id    UUID         PRIMARY KEY,
    slug            TEXT         NOT NULL UNIQUE,
    title           TEXT         NOT NULL,
    excerpt         TEXT,
    body_markdown   TEXT         NOT NULL DEFAULT '',
    cover_image_url TEXT,
    status          TEXT         NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    view_count      INTEGER      NOT NULL DEFAULT 0,
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Partial index on published_at for efficient public listing queries.
CREATE INDEX idx_blog_posts_published_at ON blog_posts (published_at DESC) WHERE status = 'published';

CREATE TABLE tags (
    tag_id UUID  PRIMARY KEY,
    name   TEXT  NOT NULL UNIQUE,
    slug   TEXT  NOT NULL UNIQUE
);

CREATE TABLE post_tags (
    blog_post_id UUID  NOT NULL REFERENCES blog_posts(blog_post_id) ON DELETE CASCADE,
    tag_id       UUID  NOT NULL REFERENCES tags(tag_id) ON DELETE CASCADE,
    PRIMARY KEY (blog_post_id, tag_id)
);

CREATE TABLE assets (
    asset_id     UUID         PRIMARY KEY,
    key          TEXT         NOT NULL UNIQUE,
    url          TEXT,
    content_type TEXT,
    size_bytes   BIGINT,
    uploaded_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE admin_users (
    admin_user_id UUID         PRIMARY KEY,
    username      TEXT         NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Analytics star schema (co-located with OLTP in PostgreSQL mode).
CREATE TABLE dim_posts (
    post_id UUID  PRIMARY KEY,
    slug    TEXT  NOT NULL,
    title   TEXT  NOT NULL
);

CREATE TABLE fact_post_views (
    view_id   UUID         PRIMARY KEY,
    post_id   UUID         NOT NULL REFERENCES dim_posts(post_id),
    viewed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fact_post_views_post_id   ON fact_post_views (post_id);
CREATE INDEX idx_fact_post_views_viewed_at ON fact_post_views (viewed_at);
