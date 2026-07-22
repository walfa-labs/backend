-- 0001_init.up.sql — initial schema for the portfolio backend.
-- All UUID PKs use gen_random_uuid() (built into Postgres 13+).

CREATE TYPE experience_type AS ENUM ('work', 'education');
CREATE TYPE content_status AS ENUM ('draft', 'published');

CREATE TABLE experience (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    experience_type experience_type NOT NULL,
    organization     TEXT,
    role_title       TEXT,
    location         TEXT,
    start_date       DATE         NOT NULL,
    end_date         DATE,
    current          BOOLEAN      NOT NULL DEFAULT FALSE,
    summary_markdown TEXT         NOT NULL DEFAULT '',
    sort_order       INTEGER      NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE experience_highlight (
    id             UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    experience_id  UUID    NOT NULL REFERENCES experience(id) ON DELETE CASCADE,
    body_markdown  TEXT,
    sort_order     INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE project (
    id                   UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                 TEXT            UNIQUE NOT NULL,
    title                TEXT            NOT NULL,
    tagline              TEXT,
    description_markdown TEXT            NOT NULL DEFAULT '',
    cover_image_url      TEXT,
    repo_url             TEXT,
    demo_url             TEXT,
    tech_stack           TEXT[]          NOT NULL DEFAULT '{}',
    status               content_status  NOT NULL DEFAULT 'draft',
    featured             BOOLEAN         NOT NULL DEFAULT FALSE,
    sort_order           INTEGER         NOT NULL DEFAULT 0,
    published_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ     NOT NULL DEFAULT now()
);

CREATE TABLE project_link (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID    NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    label       TEXT,
    url         TEXT,
    kind        TEXT    NOT NULL CHECK (kind IN ('repo', 'demo', 'docs', 'other'))
);

CREATE TABLE blog_post (
    id               UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    slug             TEXT            UNIQUE NOT NULL,
    title            TEXT            NOT NULL,
    excerpt          TEXT,
    body_markdown   TEXT            NOT NULL DEFAULT '',
    cover_image_url  TEXT,
    status           content_status  NOT NULL DEFAULT 'draft',
    view_count       INTEGER         NOT NULL DEFAULT 0,
    published_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ     NOT NULL DEFAULT now()
);

CREATE INDEX idx_blog_post_published_at
    ON blog_post (published_at DESC)
    WHERE status = 'published';

CREATE TABLE tag (
    id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name  TEXT UNIQUE NOT NULL,
    slug  TEXT UNIQUE NOT NULL
);

CREATE TABLE post_tag (
    post_id UUID NOT NULL REFERENCES blog_post(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tag(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

CREATE TABLE asset (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    key           TEXT        UNIQUE NOT NULL,
    url           TEXT,
    content_type  TEXT,
    size_bytes    BIGINT,
    uploaded_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE admin_user (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username      TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
