-- 0001_init.down.sql — drop all tables created by 0001_init.up.sql (PostgreSQL).
-- Drop in reverse dependency order (FK children before parents).

DROP TABLE IF EXISTS fact_post_views;
DROP TABLE IF EXISTS dim_posts;
DROP TABLE IF EXISTS admin_users;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS post_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS blog_posts;
DROP TABLE IF EXISTS project_links;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS experience_highlights;
DROP TABLE IF EXISTS experiences;
