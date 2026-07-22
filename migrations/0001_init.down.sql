-- 0001_init.down.sql — reverse of 0001_init.up.sql.

DROP TABLE IF EXISTS admin_user;
DROP TABLE IF EXISTS asset;
DROP TABLE IF EXISTS post_tag;
DROP TABLE IF EXISTS tag;
DROP TABLE IF EXISTS blog_post;
DROP TABLE IF EXISTS project_link;
DROP TABLE IF EXISTS project;
DROP TABLE IF EXISTS experience_highlight;
DROP TABLE IF EXISTS experience;

DROP TYPE IF EXISTS content_status;
DROP TYPE IF EXISTS experience_type;
