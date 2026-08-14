-- 0003_performance_indexes.down.sql — rollback performance indexes.
DROP INDEX IF EXISTS idx_post_tags_tag_id;
DROP INDEX IF EXISTS idx_blog_posts_status_published;
DROP INDEX IF EXISTS idx_projects_status_sort;
DROP INDEX IF EXISTS idx_experiences_sort;
DROP INDEX IF EXISTS idx_project_links_project_id;
DROP INDEX IF EXISTS idx_exp_highlights_exp_id;
