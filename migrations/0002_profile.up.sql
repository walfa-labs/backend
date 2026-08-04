-- 0002_profile.up.sql — singleton profiles table for the portfolio owner.

CREATE TABLE profiles (
    profile_id   INTEGER     PRIMARY KEY DEFAULT 1 CHECK (profile_id = 1),
    name         TEXT        NOT NULL DEFAULT '',
    email        TEXT        NOT NULL DEFAULT '',
    tagline      TEXT        NOT NULL DEFAULT '',
    bio_markdown TEXT        NOT NULL DEFAULT '',
    location     TEXT        NOT NULL DEFAULT '',
    avatar_url   TEXT        NOT NULL DEFAULT '',
    github_url   TEXT        NOT NULL DEFAULT '',
    linkedin_url TEXT        NOT NULL DEFAULT '',
    twitter_url  TEXT        NOT NULL DEFAULT '',
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO profiles (profile_id) VALUES (DEFAULT);
