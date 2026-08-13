-- 0002_profile.up.sql — singleton profiles table for the portfolio owner (Oracle ATP schema).
-- Oracle has no empty string ('' is stored as NULL), so the Postgres DEFAULT '' NOT NULL
-- columns are modeled as: CLOBs DEFAULT EMPTY_CLOB( NOT NULL) (empty LOB locator is a
-- valid, non-NULL default), short/URL text nullable. This keeps the singleton INSERT
-- below working; the Go layer should read NULL as '' (COALESCE) and always write values.

CREATE TABLE profiles (
    profile_id   NUMBER(3) DEFAULT 1 PRIMARY KEY CHECK (profile_id = 1),
    name         VARCHAR2(500),
    email        VARCHAR2(500),
    tagline      CLOB      DEFAULT EMPTY_CLOB() NOT NULL,
    bio_markdown CLOB      DEFAULT EMPTY_CLOB() NOT NULL,
    location     VARCHAR2(500),
    avatar_url   VARCHAR2(4000),
    github_url   VARCHAR2(4000),
    linkedin_url VARCHAR2(4000),
    twitter_url  VARCHAR2(4000),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Seed the singleton row; demo content is applied by migrations/seed.sql.
INSERT INTO profiles (profile_id) VALUES (1);
