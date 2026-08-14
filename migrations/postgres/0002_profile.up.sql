-- 0002_profile.up.sql — singleton profiles table for the portfolio owner (PostgreSQL).
-- profile_id = 1 is the only valid row (enforced by CHECK constraint).
-- tagline and bio_markdown default to '' (PostgreSQL stores empty string, unlike Oracle).

CREATE TABLE profiles (
    profile_id   INTEGER      PRIMARY KEY CHECK (profile_id = 1) DEFAULT 1,
    name         TEXT,
    email        TEXT,
    tagline      TEXT         NOT NULL DEFAULT '',
    bio_markdown TEXT         NOT NULL DEFAULT '',
    location     TEXT,
    avatar_url   TEXT,
    github_url   TEXT,
    linkedin_url TEXT,
    twitter_url  TEXT,
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Seed the singleton row; demo content is applied by migrations/seed.sql.
INSERT INTO profiles (profile_id) VALUES (1);
