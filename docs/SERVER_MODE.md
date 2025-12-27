# MrRSS Server Mode Documentation

## Overview

MrRSS Server Mode provides a headless, API-only version of MrRSS that can run on servers without GUI dependencies. This mode is perfect for:

- **Server deployments**: Run RSS aggregation on dedicated servers
- **API integration**: Integrate RSS functionality into other applications
- **Headless environments**: Deploy in Docker containers or cloud environments
- **Background processing**: Automated RSS monitoring and processing

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/WCY-dt/MrRSS.git
cd MrRSS

# Build and run with Docker Compose
docker-compose up -d

# Access the API
curl http://localhost:1234/api/version
```

### Manual Build and Run

```bash
# Install Go 1.24+
# Install Node.js 18+

# Clone and build
git clone https://github.com/WCY-dt/MrRSS.git
cd MrRSS

# Install frontend dependencies
cd frontend && npm install && npm run build && cd ..

# Build server version
go build -tags server -o mrrss-server .

# Run server (default: 0.0.0.0:1234)
./mrrss-server

# Or specify host and port
./mrrss-server -host 127.0.0.1 -port 8080
```

## Configuration

### Command Line Flags

| Flag | Default | Description |
| ---- | ------- | ----------- |
| `-host` | `0.0.0.0` | Server bind address |
| `-port` | `1234` | Server port |
| `-server` | `false` | Force server mode (auto-detected with `-tags server`) |

### Environment Variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `MRRSS_DEBUG` | `false` | Enable debug logging |

### Data Directory

The server stores all data in the `./data` directory:

- `feeds.db` - SQLite database
- `cache/` - Media and content cache
- `scripts/` - Custom feed scripts

## API Reference

### Base URL

```url
http://localhost:1234/api/
```

### Authentication

Currently, the API does not require authentication. Consider adding authentication middleware for production deployments.

### Response Format

All API responses are in JSON format. Successful responses return HTTP 200, errors return appropriate HTTP status codes.

---

## Feeds API

### GET /api/feeds

Get all feeds.

**Response:**

```json
[
  {
    "id": 1,
    "title": "Example Feed",
    "url": "https://example.com/feed.xml",
    "link": "https://example.com",
    "description": "Feed description",
    "category": "Technology",
    "image_url": "https://example.com/image.jpg",
    "last_updated": "2024-01-01T12:00:00Z",
    "last_error": null,
    "discovery_completed": true
  }
]
```

### POST /api/feeds/add

Add a new feed.

**Request Body:**

```json
{
  "url": "https://example.com/feed.xml",
  "category": "Technology"
}
```

**Response:**

```json
{
  "id": 1,
  "message": "Feed added successfully"
}
```

### POST /api/feeds/delete

Delete a feed.

**Request Body:**

```json
{
  "id": 1
}
```

### POST /api/feeds/update

Update feed information.

**Request Body:**

```json
{
  "id": 1,
  "title": "New Title",
  "category": "New Category"
}
```

### POST /api/feeds/refresh

Refresh a specific feed.

**Request Body:**

```json
{
  "id": 1
}
```

### POST /api/feeds/reorder

Reorder feeds.

**Request Body:**

```json
{
  "feed_ids": [1, 2, 3]
}
```

### POST /api/feeds/discover

Discover feeds from a single URL.

**Request Body:**

```json
{
  "url": "https://example.com"
}
```

### POST /api/feeds/discover-all

Discover feeds from multiple URLs.

**Request Body:**

```json
{
  "urls": ["https://example.com", "https://blog.example.com"]
}
```

### POST /api/feeds/discover/start

Start background feed discovery for single URL.

**Request Body:**

```json
{
  "url": "https://example.com"
}
```

### GET /api/feeds/discover/progress

Get discovery progress for single URL.

**Response:**

```json
{
  "is_running": true,
  "current": 5,
  "total": 10,
  "feeds": [...]
}
```

### POST /api/feeds/discover/clear

Clear discovery results for single URL.

### POST /api/feeds/discover-all/start

Start background feed discovery for multiple URLs.

**Request Body:**

```json
{
  "urls": ["https://example.com", "https://blog.example.com"]
}
```

### GET /api/feeds/discover-all/progress

Get discovery progress for multiple URLs.

**Response:**

```json
{
  "is_running": true,
  "current": 5,
  "total": 10,
  "feeds": [...]
}
```

### POST /api/feeds/discover-all/clear

Clear discovery results for multiple URLs.

### POST /api/refresh

Refresh all feeds.

---

## Articles API

### GET /api/articles

Get articles with optional filtering.

**Query Parameters:**

- `feed_id` - Filter by feed ID
- `is_read` - Filter by read status (true/false)
- `is_favorite` - Filter by favorite status (true/false)
- `limit` - Maximum number of articles (default: 50)
- `offset` - Pagination offset

**Response:**

```json
[
  {
    "id": 1,
    "feed_id": 1,
    "title": "Article Title",
    "url": "https://example.com/article",
    "image_url": "https://example.com/image.jpg",
    "content": "<p>Article content...</p>",
    "published_at": "2024-01-01T12:00:00Z",
    "is_read": false,
    "is_favorite": false,
    "is_hidden": false,
    "translated_title": null
  }
]
```

### GET /api/articles/images

Get articles with images (for gallery view).

### GET /api/articles/filter

Get filtered articles based on complex criteria.

### POST /api/articles/read

Mark articles as read/unread.

**Request Body:**

```json
{
  "ids": [1, 2, 3],
  "is_read": true
}
```

### POST /api/articles/favorite

Toggle favorite status.

**Request Body:**

```json
{
  "id": 1
}
```

### POST /api/articles/toggle-hide

Toggle article visibility (hide/unhide).

**Request Body:**

```json
{
  "id": 1
}
```

### POST /api/articles/toggle-read-later

Toggle read later status.

**Request Body:**

```json
{
  "id": 1
}
```

### GET /api/articles/content

Get full article content.

**Query Parameters:**

- `id` - Article ID

### POST /api/articles/fetch-full

Fetch full article content from original URL.

**Request Body:**

```json
{
  "id": 1
}
```

### GET /api/articles/unread-counts

Get unread article counts by feed.

**Response:**

```json
{
  "feeds": {
    "1": 5,
    "2": 12
  },
  "total": 17
}
```

### POST /api/articles/mark-all-read

Mark all articles as read.

**Request Body:**

```json
{
  "feed_id": 1  // Optional: mark all in specific feed
}
```

### POST /api/articles/clear-read-later

Clear read later status for all articles.

### POST /api/articles/cleanup

Clean up old articles.

**Request Body:**

```json
{
  "max_age_days": 30
}
```

### POST /api/articles/cleanup-content

Clean up cached article content.

### GET /api/articles/content-cache-info

Get article content cache information.

### POST /api/articles/translate

Translate article content.

**Request Body:**

```json
{
  "id": 1,
  "target_language": "zh"
}
```

### POST /api/articles/translate-text

Translate arbitrary text.

**Request Body:**

```json
{
  "text": "Hello world",
  "target_language": "zh"
}
```

### POST /api/articles/clear-translations

Clear cached translations.

### POST /api/articles/summarize

Generate article summary.

**Request Body:**

```json
{
  "id": 1
}
```

### POST /api/articles/export/obsidian

Export articles to Obsidian format.

**Request Body:**

```json
{
  "article_ids": [1, 2, 3],
  "vault_path": "/path/to/obsidian/vault"
}
```

---

## Discovery API

### POST /api/feeds/discover

Discover feeds from a single URL.

**Request Body:**

```json
{
  "url": "https://example.com"
}
```

### POST /api/feeds/discover/start

Start background feed discovery.

**Request Body:**

```json
{
  "url": "https://example.com"
}
```

### GET /api/feeds/discover/progress

Get discovery progress.

**Response:**

```json
{
  "is_running": true,
  "current": 5,
  "total": 10,
  "feeds": [...]
}
```

### POST /api/feeds/discover/clear

Clear discovery results.

---

## Settings API

### GET /api/settings

Get all application settings.

**Response:**

```json
{
  "update_interval": "30",
  "refresh_mode": "background",
  "translation_enabled": "true",
  "target_language": "zh",
  "theme": "dark",
  "language": "en",
  // ... more settings
}
```

### POST /api/settings

Update application settings.

**Request Body:**

```json
{
  "update_interval": "60",
  "theme": "light"
}
```

---

## OPML API

### POST /api/opml/import

Import feeds from OPML file.

**Request Body:** (multipart/form-data)

- `file` - OPML file

### GET /api/opml/export

Export feeds to OPML format.

**Response:** OPML XML content

### POST /api/opml/import-dialog

**Note:** Not available in server mode (returns 501)

### POST /api/opml/export-dialog

**Note:** Not available in server mode (returns 501)

---

## Media API

### GET /api/media/proxy

Proxy media content through the server.

**Query Parameters:**

- `url` - Media URL to proxy

### GET /api/webpage/proxy

Proxy webpage content through the server.

**Query Parameters:**

- `url` - Webpage URL to proxy

### POST /api/media/cleanup

Clean up old cached media.

### GET /api/media/info

Get media cache information.

---

## Custom CSS API

### POST /api/custom-css/upload-dialog

**Note:** Not available in server mode (returns 501)

### POST /api/custom-css/upload

Upload custom CSS file.

**Request Body:** (multipart/form-data)

- `file` - CSS file

### GET /api/custom-css

Get current custom CSS content.

### POST /api/custom-css/delete

Delete custom CSS.

---

## FreshRSS Integration API

### POST /api/articles/translate-text

Translate arbitrary text.

**Request Body:**

```json
{
  "text": "Hello world",
  "target_language": "zh"
}
```

### POST /api/articles/clear-translations

Clear cached translations.

---

## AI Features API

### GET /api/ai-usage

Get AI usage statistics.

**Response:**

```json
{
  "total_tokens": 15000,
  "monthly_usage": 5000,
  "last_reset": "2024-01-01T00:00:00Z"
}
```

### POST /api/ai-usage/reset

Reset AI usage counters.

### POST /api/ai-chat

Send message to AI chat.

**Request Body:**

```json
{
  "message": "Hello",
  "context": "article_id:123"
}
```

### POST /api/ai/test

Test AI configuration.

**Request Body:**

```json
{
  "provider": "openai",
  "api_key": "sk-...",
  "model": "gpt-3.5-turbo"
}
```

### GET /api/ai/test/info

Get AI test information.

---

## System API

### GET /api/version

Get application version information.

**Response:**

```json
{
  "version": "1.0.0",
  "build_time": "2024-01-01T12:00:00Z",
  "go_version": "go1.24.0"
}
```

### GET /api/progress

Get background operation progress.

### POST /api/check-updates

Check for application updates.

### POST /api/download-update

Download application update.

### POST /api/install-update

Install downloaded update.

---

## Network API

### POST /api/network/detect

Detect network connectivity and speed.

### GET /api/network/info

Get network information.

**Response:**

```json
{
  "speed_level": "fast",
  "bandwidth_mbps": 50.5,
  "detection_success": true
}
```

---

## Browser API

### POST /api/browser/open

**Note:** Not available in server mode (returns appropriate error)

---

## FreshRSS Integration API

### POST /api/freshrss/sync

Sync with FreshRSS instance.

**Request Body:**

```json
{
  "server_url": "https://freshrss.example.com",
  "username": "user",
  "password": "pass",
  "api_key": "abc123"
}
```

### POST /api/freshrss/test-connection

Test FreshRSS connection.

**Request Body:**

```json
{
  "server_url": "https://freshrss.example.com",
  "username": "user",
  "password": "pass",
  "api_key": "abc123"
}
```

---

## Rules API

### POST /api/rules/apply

Apply filtering rules to articles.

---

## Scripts API

### GET /api/scripts/dir

Get scripts directory path.

**Response:**

```json
{
  "path": "/app/data/scripts"
}
```

### POST /api/scripts/open

**Note:** Not available in server mode (returns 501)

### GET /api/scripts/list

List available scripts.

**Response:**

```json
[
  {
    "name": "custom-feed.py",
    "path": "/app/data/scripts/custom-feed.py",
    "type": "python"
  }
]
```

---

## Window API

### GET /api/window/state

Get saved window state.

### POST /api/window/save

Save window state.

---

## Error Responses

### Standard Error Format

```json
{
  "error": "Error message",
  "code": 500
}
```

### Common HTTP Status Codes

- `200` - Success
- `400` - Bad Request
- `404` - Not Found
- `500` - Internal Server Error
- `501` - Not Implemented (GUI features in server mode)

## Deployment Examples

### Docker Compose with Nginx Reverse Proxy

```yaml
version: '3.8'

services:
  mrrss:
    build:
      context: .
      dockerfile: Dockerfile.server
    expose:
      - "1234"
    volumes:
      - ./data:/app/data
    environment:
      - MRRSS_DEBUG=false
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - mrrss
```

### Systemd Service

Create `/etc/systemd/system/mrrss.service`:

```ini
[Unit]
Description=MrRSS Server
After=network.target

[Service]
Type=simple
User=mrrss
WorkingDirectory=/opt/mrrss
ExecStart=/opt/mrrss/mrrss-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mrrss
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mrrss
  template:
    metadata:
      labels:
        app: mrrss
    spec:
      containers:
      - name: mrrss
        image: mrrss-server:latest
        ports:
        - containerPort: 1234
        volumeMounts:
        - name: data
          mountPath: /app/data
        env:
        - name: MRRSS_DEBUG
          value: "false"
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: mrrss-data
```

## Monitoring and Health Checks

### Health Check Endpoint

```bash
curl http://localhost:1234/api/version
```

### Docker Health Check

The Docker Compose file includes automatic health checks that verify the API is responding.

### Log Monitoring

Server logs are output to stdout/stderr and can be captured by Docker or systemd.

## Security Considerations

### Production Deployment

- Add authentication middleware
- Use HTTPS with proper certificates
- Configure firewall rules
- Run as non-root user
- Regularly update dependencies

### API Security

- Implement rate limiting
- Add input validation
- Use HTTPS in production
- Consider API key authentication

## Troubleshooting

### Common Issues

**Port already in use:**

```bash
# Find process using port 1234
netstat -tulpn | grep :1234
# Kill the process or change port
```

**Database permission errors:**

```bash
# Ensure data directory is writable
chmod 755 ./data
```

**Docker build failures:**

```bash
# Clear Docker cache
docker system prune -a
# Rebuild without cache
docker build --no-cache -f Dockerfile.server .
```

### Debug Mode

Enable debug logging:

```bash
export MRRSS_DEBUG=true
./mrrss-server
```

### Log Analysis

```bash
# View recent logs
docker-compose logs -f mrrss-server

# Search for errors
docker-compose logs mrrss-server | grep ERROR
```

## Contributing

When adding new API endpoints:

1. Add the route to `main-core.go`
2. Implement the handler in appropriate package
3. Update this documentation
4. Test with both desktop and server builds

## License

This documentation is part of MrRSS, licensed under the same terms as the main project.
