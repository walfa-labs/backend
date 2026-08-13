-- 0001_init.down.sql — reverse of 0001_init.up.sql (Oracle ATP schema).
-- Children are dropped before parents; Oracle has no enum TYPE objects to drop.

DROP TABLE admin_users;
DROP TABLE assets;
DROP TABLE post_tags;
DROP TABLE tags;
DROP TABLE blog_posts;
DROP TABLE project_links;
DROP TABLE projects;
DROP TABLE experience_highlights;
DROP TABLE experiences;
