-- 0001_init.up.sql — initial schema for the portfolio backend.
-- All UUID PKs use gen_random_uuid() (built into Postgres 13+).
-- Naming: plural snake_case tables, {table}_id primary keys, idx_{table}_{column} indexes.

CREATE TYPE experience_type AS ENUM ('work', 'education');
CREATE TYPE content_status AS ENUM ('draft', 'published');

CREATE TABLE experiences (
    experience_id    UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    experience_type  experience_type NOT NULL,
    organization     TEXT,
    role_title       TEXT,
    location         TEXT,
    start_date       DATE            NOT NULL,
    end_date         DATE,
    current          BOOLEAN         NOT NULL DEFAULT FALSE,
    summary_markdown TEXT            NOT NULL DEFAULT '',
    sort_order       INTEGER         NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ     NOT NULL DEFAULT now()
);

CREATE TABLE experience_highlights (
    experience_highlight_id UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    experience_id           UUID    NOT NULL REFERENCES experiences(experience_id) ON DELETE CASCADE,
    body_markdown           TEXT,
    sort_order              INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE projects (
    project_id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                 TEXT           UNIQUE NOT NULL,
    title                TEXT           NOT NULL,
    tagline              TEXT,
    description_markdown TEXT           NOT NULL DEFAULT '',
    cover_image_url      TEXT,
    repo_url             TEXT,
    demo_url             TEXT,
    tech_stack           TEXT[]         NOT NULL DEFAULT '{}',
    status               content_status NOT NULL DEFAULT 'draft',
    featured             BOOLEAN        NOT NULL DEFAULT FALSE,
    sort_order           INTEGER        NOT NULL DEFAULT 0,
    published_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE TABLE project_links (
    project_link_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(project_id) ON DELETE CASCADE,
    label           TEXT,
    url             TEXT,
    kind            TEXT NOT NULL CHECK (kind IN ('repo', 'demo', 'docs', 'other'))
);

CREATE TABLE blog_posts (
    blog_post_id    UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            TEXT           UNIQUE NOT NULL,
    title           TEXT           NOT NULL,
    excerpt         TEXT,
    body_markdown   TEXT           NOT NULL DEFAULT '',
    cover_image_url TEXT,
    status          content_status NOT NULL DEFAULT 'draft',
    view_count      INTEGER        NOT NULL DEFAULT 0,
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT now()
);

CREATE INDEX idx_blog_posts_published_at
    ON blog_posts (published_at DESC)
    WHERE status = 'published';

CREATE TABLE tags (
    tag_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name   TEXT UNIQUE NOT NULL,
    slug   TEXT UNIQUE NOT NULL
);

CREATE TABLE post_tags (
    blog_post_id UUID NOT NULL REFERENCES blog_posts(blog_post_id) ON DELETE CASCADE,
    tag_id       UUID NOT NULL REFERENCES tags(tag_id) ON DELETE CASCADE,
    PRIMARY KEY (blog_post_id, tag_id)
);

CREATE TABLE assets (
    asset_id     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    key          TEXT        UNIQUE NOT NULL,
    url          TEXT,
    content_type TEXT,
    size_bytes   BIGINT,
    uploaded_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE admin_users (
    admin_user_id UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username      TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
