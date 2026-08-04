-- 0001_init.down.sql — reverse of 0001_init.up.sql.

DROP TABLE IF EXISTS admin_users;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS post_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS blog_posts;
DROP TABLE IF EXISTS project_links;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS experience_highlights;
DROP TABLE IF EXISTS experiences;

DROP TYPE IF EXISTS content_status;
DROP TYPE IF EXISTS experience_type;
