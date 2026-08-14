-- 0003_performance_indexes.down.sql — rollback performance indexes.
DROP INDEX idx_post_tags_tag_id;
DROP INDEX idx_blog_posts_status_published;
DROP INDEX idx_projects_status_sort;
DROP INDEX idx_experiences_sort;
DROP INDEX idx_project_links_project_id;
DROP INDEX idx_exp_highlights_exp_id;
