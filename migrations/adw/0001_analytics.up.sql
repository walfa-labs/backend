-- 0001_analytics.up.sql — analytics warehouse schema (star-schema style), ADW sequence.
-- Runs against the Autonomous Data Warehouse (ADW) database with a SEPARATE
-- golang-migrate invocation (-path migrations/adw -database <ADW url>) — NOT against
-- the ATP (OLTP) database, whose migrations live in migrations/atp/.

CREATE TABLE dim_posts (
    post_id VARCHAR2(36 CHAR) PRIMARY KEY,
    slug    VARCHAR2(500)  NOT NULL,
    title   VARCHAR2(1000) NOT NULL
);

CREATE TABLE fact_post_views (
    view_id   VARCHAR2(36 CHAR) PRIMARY KEY,
    post_id   VARCHAR2(36 CHAR) NOT NULL REFERENCES dim_posts(post_id),
    viewed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_fact_post_views_post_id ON fact_post_views (post_id);
CREATE INDEX idx_fact_post_views_viewed_at ON fact_post_views (viewed_at);
