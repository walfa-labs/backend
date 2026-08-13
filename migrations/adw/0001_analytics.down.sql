-- 0001_analytics.down.sql — reverse of 0001_analytics.up.sql.
-- Runs against the ADW database (separate golang-migrate invocation/URL), not ATP.
-- Fact table is dropped before the dimension table it references.

DROP TABLE fact_post_views;
DROP TABLE dim_posts;
