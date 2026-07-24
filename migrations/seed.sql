-- Experiences
INSERT INTO experience (experience_type, organization, role_title, location, start_date, end_date, current, summary_markdown, sort_order) VALUES
  ('work', 'Stripe', 'Software Engineer II', 'Remote', '2019-06-01', '2021-12-31', false, 'Worked on payment infrastructure processing billions of dollars in transactions.', 1),
  ('work', 'Vercel', 'Software Engineer Intern', 'Remote', '2018-05-01', '2018-08-31', false, 'Contributed to the Next.js framework and internal tooling.', 2),
  ('education', 'National University of Singapore', 'B.Sc. Computer Science', 'Singapore', '2015-08-01', '2019-05-31', false, 'First Class Honours. Focus on distributed systems and compilers.', 0);

INSERT INTO experience_highlight (experience_id, body_markdown, sort_order)
SELECT id, body, ord FROM (VALUES
  ((SELECT id FROM experience WHERE organization='Google' AND role_title='Senior Software Engineer'), 'Led migration of monolithic PHP backend to Go microservices, reducing p99 latency by 40%', 0),
  ((SELECT id FROM experience WHERE organization='Google' AND role_title='Senior Software Engineer'), 'Designed and implemented a real-time analytics pipeline processing 10M+ events/day', 1),
  ((SELECT id FROM experience WHERE organization='Google' AND role_title='Senior Software Engineer'), 'Mentored 3 junior engineers and drove adoption of TDD across the team', 2),
  ((SELECT id FROM experience WHERE organization='Stripe' AND role_title='Software Engineer II'), 'Redesigned the dispute resolution API used by 50k+ businesses', 0),
  ((SELECT id FROM experience WHERE organization='Stripe' AND role_title='Software Engineer II'), 'Improved payment retry logic, recovering $2M+ in failed transactions annually', 1)
) AS t(id, body, ord);

-- Projects
INSERT INTO project (slug, title, tagline, description_markdown, repo_url, demo_url, tech_stack, status, featured, sort_order, published_at) VALUES
  ('realtime-chat-go', 'Realtime Chat Engine', 'A WebSocket-based chat platform built in Go with 10k concurrent connections.', '## Overview

A high-performance realtime chat engine built in Go, using WebSocket connections and Redis pub/sub for message distribution.

## Features

- WebSocket connections with heartbeat
- Redis pub/sub for horizontal scaling
- JWT authentication
- Message persistence in PostgreSQL
- Typing indicators and presence

## Performance

Benchmarked at **10,000 concurrent connections** on a single instance with <50ms message latency.', 'https://github.com/walfa/realtime-chat-go', 'https://chat-demo.walfa.dev', ARRAY['Go', 'WebSocket', 'Redis', 'PostgreSQL', 'Docker'], 'published', true, 0, now()),
  ('nuxt-portfolio-cms', 'Portfolio CMS', 'A headless CMS for personal portfolios with SSR frontend and Go backend.', '## Overview

A full-stack portfolio CMS built with Nuxt 4 and Go/Fiber. Features hybrid SSR for SEO, a dashboard with WYSIWYG editor, and S3-compatible asset management.

## Highlights

- Hybrid SSR + SPA rendering
- JWT auth with refresh tokens
- Admin dashboard with CRUD for all entities
- Terminal-style homepage design', 'https://github.com/walfa/nuxt-portfolio-cms', '', ARRAY['Nuxt', 'Go', 'Fiber', 'PostgreSQL', 'TailwindCSS'], 'published', true, 1, now()),
  ('url-shortener-rust', 'URL Shortener in Rust', 'A blazing-fast URL shortener written in Rust with SQLite storage.', '## Overview

A minimal URL shortener built in Rust, focusing on performance and simplicity. Uses SQLite for storage and Axum for the web server.

## Benchmarks

Handles **100k+ requests/sec** on a single core.', 'https://github.com/walfa/url-shortener-rust', 'https://s.walfa.dev', ARRAY['Rust', 'Axum', 'SQLite'], 'published', true, 2, now()),
  ('cli-task-manager', 'CLI Task Manager', 'A terminal-based task manager written in Go with bubble tea TUI.', '## Overview

A keyboard-driven task manager for the terminal, built with Bubble Tea TUI framework.

## Features

- Kanban board view
- Vim-style keybindings
- Local SQLite storage
- Export to JSON/Markdown', 'https://github.com/walfa/cli-task-manager', '', ARRAY['Go', 'Bubble Tea', 'SQLite'], 'published', false, 3, now()),
  ('graphql-mesh-gateway', 'GraphQL Mesh Gateway', 'A unified GraphQL gateway aggregating multiple REST APIs.', '## Overview

A GraphQL gateway that aggregates data from multiple REST APIs into a single unified GraphQL schema. Built with Go and gqlgen.

## Features

- Schema stitching
- Automatic REST to GraphQL mapping
- Query batching and caching
- Rate limiting per client', 'https://github.com/walfa/graphql-mesh-gateway', '', ARRAY['Go', 'GraphQL', 'gqlgen', 'Redis'], 'published', false, 4, now());

-- Project links
INSERT INTO project_link (project_id, label, url, kind) VALUES
  ((SELECT id FROM project WHERE slug='realtime-chat-go'), 'GitHub', 'https://github.com/walfa/realtime-chat-go', 'repo'),
  ((SELECT id FROM project WHERE slug='realtime-chat-go'), 'Live Demo', 'https://chat-demo.walfa.dev', 'demo'),
  ((SELECT id FROM project WHERE slug='nuxt-portfolio-cms'), 'GitHub', 'https://github.com/walfa/nuxt-portfolio-cms', 'repo'),
  ((SELECT id FROM project WHERE slug='url-shortener-rust'), 'GitHub', 'https://github.com/walfa/url-shortener-rust', 'repo'),
  ((SELECT id FROM project WHERE slug='url-shortener-rust'), 'Live Demo', 'https://s.walfa.dev', 'demo'),
  ((SELECT id FROM project WHERE slug='cli-task-manager'), 'GitHub', 'https://github.com/walfa/cli-task-manager', 'repo');

-- Blog posts
INSERT INTO blog_post (slug, title, excerpt, body_markdown, status, view_count, published_at) VALUES
  ('building-realtime-chat-go-websockets', 'Building a Realtime Chat in Go with WebSockets', 'A deep dive into building a production-grade WebSocket chat server in Go, covering connection pooling, Redis pub/sub, and horizontal scaling.', '# Building a Realtime Chat in Go

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

Go makes WebSocket servers straightforward to build. The goroutine-per-connection model is elegant, and with Redis pub/sub, horizontal scaling is trivial.', 'published', 1247, now() - interval '7 days'),
  ('why-i-chose-go-for-backend', 'Why I Chose Go Over Node.js for Backend', 'After 5 years of Node.js, I switched to Go. Here is what I learned about performance, developer experience, and trade-offs.', '# Why I Chose Go Over Node.js

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

Go wins for production backends. Node.js wins for rapid prototyping. I use both.', 'published', 3412, now() - interval '14 days'),
  ('postgres-indexing-strategies', 'Postgres Indexing Strategies I Wish I Knew Earlier', 'A practical guide to choosing the right Postgres index types — B-tree, GIN, BRIN, and partial indexes — with real examples.', '# Postgres Indexing Strategies

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
4. **BRIN** for time-series data', 'published', 892, now() - interval '21 days'),
  ('terminal-ui-design-principles', 'Terminal UI Design Principles', 'Designing beautiful terminal interfaces is an art. Here are the principles I follow when building TUI apps with Bubble Tea.', '# Terminal UI Design Principles

A good terminal UI follows the same principles as a good web UI: hierarchy, contrast, and whitespace.

## Use Color Sparingly

Color should guide attention, not decorate. Use one accent color for interactive elements.

## Whitespace Matters

Even in a terminal, padding and margins make a huge difference.

## Keyboard First

Every action should have a keyboard shortcut. The mouse is secondary.

## Conclusion

A well-designed TUI is faster than any web interface for the right tasks.', 'published', 543, now() - interval '30 days'),
  ('docker-multi-stage-builds-go', 'Docker Multi-Stage Builds for Go', 'Shrink your Go Docker images from 800MB to 15MB with multi-stage builds and scratch images.', '# Docker Multi-Stage Builds for Go

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
| CVEs | 47 | 0 |', 'published', 1876, now() - interval '45 days');

-- Tags
INSERT INTO tag (name, slug) VALUES
  ('Go', 'go'),
  ('WebSockets', 'websockets'),
  ('Redis', 'redis'),
  ('PostgreSQL', 'postgresql'),
  ('Rust', 'rust'),
  ('Docker', 'docker'),
  ('TypeScript', 'typescript'),
  ('Nuxt', 'nuxt'),
  ('Performance', 'performance'),
  ('DevOps', 'devops');

-- Post-tag associations
INSERT INTO post_tag (post_id, tag_id) SELECT p.id, t.id FROM blog_post p, tag t WHERE p.slug='building-realtime-chat-go-websockets' AND t.slug IN ('go', 'websockets', 'redis', 'performance');
INSERT INTO post_tag (post_id, tag_id) SELECT p.id, t.id FROM blog_post p, tag t WHERE p.slug='why-i-chose-go-for-backend' AND t.slug IN ('go', 'typescript', 'performance');
INSERT INTO post_tag (post_id, tag_id) SELECT p.id, t.id FROM blog_post p, tag t WHERE p.slug='postgres-indexing-strategies' AND t.slug IN ('postgresql', 'performance');
INSERT INTO post_tag (post_id, tag_id) SELECT p.id, t.id FROM blog_post p, tag t WHERE p.slug='terminal-ui-design-principles' AND t.slug IN ('go', 'performance');
INSERT INTO post_tag (post_id, tag_id) SELECT p.id, t.id FROM blog_post p, tag t WHERE p.slug='docker-multi-stage-builds-go' AND t.slug IN ('go', 'docker', 'devops');

-- Update profile
UPDATE profile SET
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
  updated_at = now()
WHERE id = 1;
