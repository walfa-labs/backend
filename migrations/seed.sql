-- seed.sql — demo data for the portfolio backend (Oracle ATP schema).
-- Applied MANUALLY against the ATP database; it is NOT part of the golang-migrate
-- migration sequence. Assumes 0001/0002 are applied (the singleton profiles row
-- created by 0002_profile.up.sql is updated at the bottom).
--
-- Parent IDs are hardcoded UUID string literals: Oracle does not allow scalar
-- subqueries in VALUES, and fixed UUIDs make the seed deterministic. Child rows
-- (highlights, links, post_tags) reference those literals directly.
--
-- Notes:
-- - '' is NULL in Oracle: the empty demo_url values below insert as NULL (the
--   column is nullable, and the Go layer reads NULL as '').
-- - If you run this file in SQLcl/SQL*Plus, execute SET SQLBLANKLINES ON first so
--   the blank lines inside the markdown string literals don't terminate statements.

-- Experiences
INSERT INTO experiences (experience_id, experience_type, organization, role_title, location, start_date, end_date, "current", summary_markdown, sort_order) VALUES
  ('e1111111-1111-4111-8111-111111111111', 'work', 'Google', 'Senior Software Engineer', 'Singapore', DATE '2022-01-01', NULL, TRUE, 'Building distributed backend systems and leading backend architecture initiatives.', 0);
INSERT INTO experiences (experience_id, experience_type, organization, role_title, location, start_date, end_date, "current", summary_markdown, sort_order) VALUES
  ('e2222222-2222-4222-8222-222222222222', 'work', 'Stripe', 'Software Engineer II', 'Remote', DATE '2019-06-01', DATE '2021-12-31', FALSE, 'Worked on payment infrastructure processing billions of dollars in transactions.', 1);
INSERT INTO experiences (experience_id, experience_type, organization, role_title, location, start_date, end_date, "current", summary_markdown, sort_order) VALUES
  ('e3333333-3333-4333-8333-333333333333', 'work', 'Vercel', 'Software Engineer Intern', 'Remote', DATE '2018-05-01', DATE '2018-08-31', FALSE, 'Contributed to the Next.js framework and internal tooling.', 2);
INSERT INTO experiences (experience_id, experience_type, organization, role_title, location, start_date, end_date, "current", summary_markdown, sort_order) VALUES
  ('e4444444-4444-4444-8444-444444444444', 'education', 'National University of Singapore', 'B.Sc. Computer Science', 'Singapore', DATE '2015-08-01', DATE '2019-05-31', FALSE, 'First Class Honours. Focus on distributed systems and compilers.', 3);

-- Experience highlights
INSERT INTO experience_highlights (experience_highlight_id, experience_id, body_markdown, sort_order) VALUES
  ('d1111111-1111-4111-8111-111111111111', 'e1111111-1111-4111-8111-111111111111', 'Led migration of monolithic PHP backend to Go microservices, reducing p99 latency by 40%', 0);
INSERT INTO experience_highlights (experience_highlight_id, experience_id, body_markdown, sort_order) VALUES
  ('d2222222-2222-4222-8222-222222222222', 'e1111111-1111-4111-8111-111111111111', 'Designed and implemented a real-time analytics pipeline processing 10M+ events/day', 1);
INSERT INTO experience_highlights (experience_highlight_id, experience_id, body_markdown, sort_order) VALUES
  ('d3333333-3333-4333-8333-333333333333', 'e1111111-1111-4111-8111-111111111111', 'Mentored 3 junior engineers and drove adoption of TDD across the team', 2);
INSERT INTO experience_highlights (experience_highlight_id, experience_id, body_markdown, sort_order) VALUES
  ('d4444444-4444-4444-8444-444444444444', 'e2222222-2222-4222-8222-222222222222', 'Redesigned the dispute resolution API used by 50k+ businesses', 0);
INSERT INTO experience_highlights (experience_highlight_id, experience_id, body_markdown, sort_order) VALUES
  ('d5555555-5555-4555-8555-555555555555', 'e2222222-2222-4222-8222-222222222222', 'Improved payment retry logic, recovering $2M+ in failed transactions annually', 1);

-- Projects
INSERT INTO projects (project_id, slug, title, tagline, description_markdown, repo_url, demo_url, tech_stack, status, featured, sort_order, published_at) VALUES
  ('a1111111-1111-4111-8111-111111111111', 'realtime-chat-go', 'Realtime Chat Engine', 'A WebSocket-based chat platform built in Go with 10k concurrent connections.', '## Overview

A high-performance realtime chat engine built in Go, using WebSocket connections and Redis pub/sub for message distribution.

## Features

- WebSocket connections with heartbeat
- Redis pub/sub for horizontal scaling
- JWT authentication
- Message persistence in PostgreSQL
- Typing indicators and presence

## Performance

Benchmarked at **10,000 concurrent connections** on a single instance with <50ms message latency.', 'https://github.com/walfa/realtime-chat-go', 'https://chat-demo.walfa.dev', '["Go","WebSocket","Redis","PostgreSQL","Docker"]', 'published', TRUE, 0, CURRENT_TIMESTAMP);
INSERT INTO projects (project_id, slug, title, tagline, description_markdown, repo_url, demo_url, tech_stack, status, featured, sort_order, published_at) VALUES
  ('a2222222-2222-4222-8222-222222222222', 'nuxt-portfolio-cms', 'Portfolio CMS', 'A headless CMS for personal portfolios with SSR frontend and Go backend.', '## Overview

A full-stack portfolio CMS built with Nuxt 4 and Go/Fiber. Features hybrid SSR for SEO, a dashboard with WYSIWYG editor, and S3-compatible asset management.

## Highlights

- Hybrid SSR + SPA rendering
- JWT auth with refresh tokens
- Admin dashboard with CRUD for all entities
- Terminal-style homepage design', 'https://github.com/walfa/nuxt-portfolio-cms', '', '["Nuxt","Go","Fiber","PostgreSQL","TailwindCSS"]', 'published', TRUE, 1, CURRENT_TIMESTAMP);
INSERT INTO projects (project_id, slug, title, tagline, description_markdown, repo_url, demo_url, tech_stack, status, featured, sort_order, published_at) VALUES
  ('a3333333-3333-4333-8333-333333333333', 'url-shortener-rust', 'URL Shortener in Rust', 'A blazing-fast URL shortener written in Rust with SQLite storage.', '## Overview

A minimal URL shortener built in Rust, focusing on performance and simplicity. Uses SQLite for storage and Axum for the web server.

## Benchmarks

Handles **100k+ requests/sec** on a single core.', 'https://github.com/walfa/url-shortener-rust', 'https://s.walfa.dev', '["Rust","Axum","SQLite"]', 'published', TRUE, 2, CURRENT_TIMESTAMP);
INSERT INTO projects (project_id, slug, title, tagline, description_markdown, repo_url, demo_url, tech_stack, status, featured, sort_order, published_at) VALUES
  ('a4444444-4444-4444-8444-444444444444', 'cli-task-manager', 'CLI Task Manager', 'A terminal-based task manager written in Go with bubble tea TUI.', '## Overview

A keyboard-driven task manager for the terminal, built with Bubble Tea TUI framework.

## Features

- Kanban board view
- Vim-style keybindings
- Local SQLite storage
- Export to JSON/Markdown', 'https://github.com/walfa/cli-task-manager', '', '["Go","Bubble Tea","SQLite"]', 'published', FALSE, 3, CURRENT_TIMESTAMP);
INSERT INTO projects (project_id, slug, title, tagline, description_markdown, repo_url, demo_url, tech_stack, status, featured, sort_order, published_at) VALUES
  ('a5555555-5555-4555-8555-555555555555', 'graphql-mesh-gateway', 'GraphQL Mesh Gateway', 'A unified GraphQL gateway aggregating multiple REST APIs.', '## Overview

A GraphQL gateway that aggregates data from multiple REST APIs into a single unified GraphQL schema. Built with Go and gqlgen.

## Features

- Schema stitching
- Automatic REST to GraphQL mapping
- Query batching and caching
- Rate limiting per client', 'https://github.com/walfa/graphql-mesh-gateway', '', '["Go","GraphQL","gqlgen","Redis"]', 'published', FALSE, 4, CURRENT_TIMESTAMP);

-- Project links
INSERT INTO project_links (project_link_id, project_id, label, url, kind) VALUES
  ('f1111111-1111-4111-8111-111111111111', 'a1111111-1111-4111-8111-111111111111', 'GitHub', 'https://github.com/walfa/realtime-chat-go', 'repo');
INSERT INTO project_links (project_link_id, project_id, label, url, kind) VALUES
  ('f2222222-2222-4222-8222-222222222222', 'a1111111-1111-4111-8111-111111111111', 'Live Demo', 'https://chat-demo.walfa.dev', 'demo');
INSERT INTO project_links (project_link_id, project_id, label, url, kind) VALUES
  ('f3333333-3333-4333-8333-333333333333', 'a2222222-2222-4222-8222-222222222222', 'GitHub', 'https://github.com/walfa/nuxt-portfolio-cms', 'repo');
INSERT INTO project_links (project_link_id, project_id, label, url, kind) VALUES
  ('f4444444-4444-4444-8444-444444444444', 'a3333333-3333-4333-8333-333333333333', 'GitHub', 'https://github.com/walfa/url-shortener-rust', 'repo');
INSERT INTO project_links (project_link_id, project_id, label, url, kind) VALUES
  ('f5555555-5555-4555-8555-555555555555', 'a3333333-3333-4333-8333-333333333333', 'Live Demo', 'https://s.walfa.dev', 'demo');
INSERT INTO project_links (project_link_id, project_id, label, url, kind) VALUES
  ('f6666666-6666-4666-8666-666666666666', 'a4444444-4444-4444-8444-444444444444', 'GitHub', 'https://github.com/walfa/cli-task-manager', 'repo');

-- Blog posts
INSERT INTO blog_posts (blog_post_id, slug, title, excerpt, body_markdown, status, view_count, published_at) VALUES
  ('b1111111-1111-4111-8111-111111111111', 'building-realtime-chat-go-websockets', 'Building a Realtime Chat in Go with WebSockets', 'A deep dive into building a production-grade WebSocket chat server in Go, covering connection pooling, Redis pub/sub, and horizontal scaling.', '# Building a Realtime Chat in Go

When I set out to build a chat system that could handle **10,000 concurrent connections**, I knew Go was the right choice.

## The Architecture

The system uses a hub-and-spoke model:

- **Hub**: Manages all active connections and broadcasts messages
- **Client**: Wraps a single WebSocket connection
- **Redis pub/sub**: Enables horizontal scaling across instances

## Scaling with Redis

A single instance can handle ~10k connections, but what if you need more? Redis pub/sub lets multiple instances share messages.

## Results

- **10,000** concurrent connections on a single $5 VPS
- **<50ms** message delivery latency
- **Zero** downtime during deploys

## Conclusion

Go makes WebSocket servers straightforward to build. The goroutine-per-connection model is elegant, and with Redis pub/sub, horizontal scaling is trivial.', 'published', 1247, CURRENT_TIMESTAMP - INTERVAL '7' DAY);
INSERT INTO blog_posts (blog_post_id, slug, title, excerpt, body_markdown, status, view_count, published_at) VALUES
  ('b2222222-2222-4222-8222-222222222222', 'why-i-chose-go-for-backend', 'Why I Chose Go Over Node.js for Backend', 'After 5 years of Node.js, I switched to Go. Here is what I learned about performance, developer experience, and trade-offs.', '# Why I Chose Go Over Node.js

After five years of building backends in Node.js, I made the switch to Go. This was not a decision I took lightly.

## Performance

Go is faster. Not just in microbenchmarks — in real-world throughput:

| Metric | Node.js | Go |
|---|---|---|
| Requests/sec | 8,000 | 45,000 |
| Memory (1k reqs) | 120MB | 18MB |
| Cold start | 400ms | 15ms |

## Type Safety

TypeScript is great, but Go is *compiled*. There is no `any` escape hatch.

## Concurrency

Goroutines are game-changing. No callback hell, no async/await pyramid.

## What I Miss

- NPM ecosystem (Go modules are catching up)
- JSON handling (Go requires explicit struct tags)
- Quick prototyping (Node.js is still faster to start)

## Conclusion

Go wins for production backends. Node.js wins for rapid prototyping. I use both.', 'published', 3412, CURRENT_TIMESTAMP - INTERVAL '14' DAY);
INSERT INTO blog_posts (blog_post_id, slug, title, excerpt, body_markdown, status, view_count, published_at) VALUES
  ('b3333333-3333-4333-8333-333333333333', 'postgres-indexing-strategies', 'Postgres Indexing Strategies I Wish I Knew Earlier', 'A practical guide to choosing the right Postgres index types — B-tree, GIN, BRIN, and partial indexes — with real examples.', '# Postgres Indexing Strategies

PostgreSQL has several index types, and choosing the right one can mean the difference between a 2ms and 2000ms query.

## B-Tree (Default)

Good for equality and range queries. This is what you get with `CREATE INDEX`.

## GIN Indexes

Perfect for `JSONB` and full-text search.

## Partial Indexes

Only index the rows you actually query. This saves space and is faster:

```sql
CREATE INDEX idx_active_users ON users(email) WHERE active = true;
```

## BRIN Indexes

For large tables with naturally ordered data (like timestamps), BRIN is incredibly space-efficient.

## Key Takeaways

1. **Always check your query plan** with `EXPLAIN ANALYZE`
2. **Partial indexes** are underused — they save space and are faster
3. **GIN** for JSONB and arrays
4. **BRIN** for time-series data', 'published', 892, CURRENT_TIMESTAMP - INTERVAL '21' DAY);
INSERT INTO blog_posts (blog_post_id, slug, title, excerpt, body_markdown, status, view_count, published_at) VALUES
  ('b4444444-4444-4444-8444-444444444444', 'terminal-ui-design-principles', 'Terminal UI Design Principles', 'Designing beautiful terminal interfaces is an art. Here are the principles I follow when building TUI apps with Bubble Tea.', '# Terminal UI Design Principles

A good terminal UI follows the same principles as a good web UI: hierarchy, contrast, and whitespace.

## Use Color Sparingly

Color should guide attention, not decorate. Use one accent color for interactive elements.

## Whitespace Matters

Even in a terminal, padding and margins make a huge difference.

## Keyboard First

Every action should have a keyboard shortcut. The mouse is secondary.

## Conclusion

A well-designed TUI is faster than any web interface for the right tasks.', 'published', 543, CURRENT_TIMESTAMP - INTERVAL '30' DAY);
INSERT INTO blog_posts (blog_post_id, slug, title, excerpt, body_markdown, status, view_count, published_at) VALUES
  ('b5555555-5555-4555-8555-555555555555', 'docker-multi-stage-builds-go', 'Docker Multi-Stage Builds for Go', 'Shrink your Go Docker images from 800MB to 15MB with multi-stage builds and scratch images.', '# Docker Multi-Stage Builds for Go

A default Go Docker image is ~800MB. With multi-stage builds, you can get it down to **15MB**.

## The Solution

```dockerfile
FROM golang:1.21 AS builder
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o app .

FROM scratch
COPY --from=builder /app /app
CMD ["/app"]
```

Result: **15MB** image with zero attack surface.

## Tips

- Use `CGO_ENABLED=0` for static binaries
- `-ldflags="-s -w"` strips debug info
- `scratch` base has no shell, no libraries
- Use `gcr.io/distroless/static` if you need CA certificates

## Results

| Metric | Before | After |
|---|---|---|
| Image size | 800MB | 15MB |
| Pull time | 30s | 2s |
| CVEs | 47 | 0 |', 'published', 1876, CURRENT_TIMESTAMP - INTERVAL '45' DAY);

-- Tags
INSERT INTO tags (tag_id, name, slug) VALUES ('c1111111-1111-4111-8111-111111111111', 'Go', 'go');
INSERT INTO tags (tag_id, name, slug) VALUES ('c2222222-2222-4222-8222-222222222222', 'WebSockets', 'websockets');
INSERT INTO tags (tag_id, name, slug) VALUES ('c3333333-3333-4333-8333-333333333333', 'Redis', 'redis');
INSERT INTO tags (tag_id, name, slug) VALUES ('c4444444-4444-4444-8444-444444444444', 'PostgreSQL', 'postgresql');
INSERT INTO tags (tag_id, name, slug) VALUES ('c5555555-5555-4555-8555-555555555555', 'Rust', 'rust');
INSERT INTO tags (tag_id, name, slug) VALUES ('c6666666-6666-4666-8666-666666666666', 'Docker', 'docker');
INSERT INTO tags (tag_id, name, slug) VALUES ('c7777777-7777-4777-8777-777777777777', 'TypeScript', 'typescript');
INSERT INTO tags (tag_id, name, slug) VALUES ('c8888888-8888-4888-8888-888888888888', 'Nuxt', 'nuxt');
INSERT INTO tags (tag_id, name, slug) VALUES ('c9999999-9999-4999-8999-999999999999', 'Performance', 'performance');
INSERT INTO tags (tag_id, name, slug) VALUES ('caaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', 'DevOps', 'devops');

-- Post-tag associations
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b1111111-1111-4111-8111-111111111111', 'c1111111-1111-4111-8111-111111111111');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b1111111-1111-4111-8111-111111111111', 'c2222222-2222-4222-8222-222222222222');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b1111111-1111-4111-8111-111111111111', 'c3333333-3333-4333-8333-333333333333');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b1111111-1111-4111-8111-111111111111', 'c9999999-9999-4999-8999-999999999999');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b2222222-2222-4222-8222-222222222222', 'c1111111-1111-4111-8111-111111111111');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b2222222-2222-4222-8222-222222222222', 'c7777777-7777-4777-8777-777777777777');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b2222222-2222-4222-8222-222222222222', 'c9999999-9999-4999-8999-999999999999');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b3333333-3333-4333-8333-333333333333', 'c4444444-4444-4444-8444-444444444444');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b3333333-3333-4333-8333-333333333333', 'c9999999-9999-4999-8999-999999999999');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b4444444-4444-4444-8444-444444444444', 'c1111111-1111-4111-8111-111111111111');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b4444444-4444-4444-8444-444444444444', 'c9999999-9999-4999-8999-999999999999');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b5555555-5555-4555-8555-555555555555', 'c1111111-1111-4111-8111-111111111111');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b5555555-5555-4555-8555-555555555555', 'c6666666-6666-4666-8666-666666666666');
INSERT INTO post_tags (blog_post_id, tag_id) VALUES ('b5555555-5555-4555-8555-555555555555', 'caaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa');

-- Update profile (singleton row created by 0002_profile.up.sql)
UPDATE profiles SET
  name = 'Walfa',
  email = 'hello@walfa.dev',
  tagline = 'Software engineer building fast, reliable systems with Go and TypeScript.',
  bio_markdown = '## About Me

I am a software engineer with a passion for building fast, reliable backend systems. I work primarily with **Go** and **TypeScript**, focusing on distributed systems and developer tooling.

When I am not coding, you can find me contributing to open source, writing about software engineering, or exploring new coffee shops.',
  location = 'Singapore',
  github_url = 'https://github.com/walfa',
  linkedin_url = 'https://linkedin.com/in/walfa',
  twitter_url = 'https://twitter.com/walfa',
  updated_at = CURRENT_TIMESTAMP
WHERE profile_id = 1;

-- Admin user (admin / admin123)
INSERT INTO admin_users (admin_user_id, username, password_hash) VALUES
  ('00000000-0000-0000-0000-000000000001', 'admin', '$2a$10$Qwz/cxG0s36.a/3nwRf7kuvYI9/V.5ab0Y8C6LNE723211RNOsbdO');
