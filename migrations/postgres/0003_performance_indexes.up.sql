-- 0003_performance_indexes.up.sql — performance indexes for PostgreSQL.

-- Index for many-to-many tag lookups in blog_posts queries
CREATE INDEX IF NOT EXISTS idx_post_tags_tag_id ON post_tags (tag_id);

-- Composite index for published posts listing and ordering
CREATE INDEX IF NOT EXISTS idx_blog_posts_status_published ON blog_posts (status, published_at DESC);

-- Composite index for published projects ordering
CREATE INDEX IF NOT EXISTS idx_projects_status_sort ON projects (status, sort_order ASC, published_at DESC);

-- Index for experience sorting and display
CREATE INDEX IF NOT EXISTS idx_experiences_sort ON experiences (sort_order ASC, start_date DESC);

-- Foreign key lookup indexes
CREATE INDEX IF NOT EXISTS idx_project_links_project_id ON project_links (project_id);
CREATE INDEX IF NOT EXISTS idx_exp_highlights_exp_id ON experience_highlights (experience_id);
