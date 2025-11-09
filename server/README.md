# 🖥️ Axiolot Hub Uptime Monitor Server

A high-performance monitoring report API built with Bun and Supabase, designed to receive and store uptime monitoring data from the Go monitoring client.

[![Bun](https://img.shields.io/badge/Bun-Latest-black?style=flat&logo=bun)](https://bun.sh)
[![Supabase](https://img.shields.io/badge/Supabase-Powered-3ECF8E?style=flat&logo=supabase)](https://supabase.com)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/arieduportal/uptime-monitor?style=social)](https://github.com/arieduportal/uptime-monitor)
[![Axiolot Hub logo](https://static.axiolot.com.ng/favicon.ico)](https://axiolot.com.ng)

## ✨ Features

### Core Capabilities
- ⚡ **Lightning Fast** - Built on Bun runtime for maximum performance
- 🗄️ **Supabase Integration** - PostgreSQL database with real-time capabilities
- 🔒 **API Key Authentication** - Secure Bearer token authentication
- 📊 **Historical Storage** - Store and query monitoring reports over time
- 🔍 **Rich Querying** - Filter by environment, status, date range, and more
- 📈 **Analytics Ready** - Structured data for uptime trends and SLA reports

### Reliability & Security
- 🛡️ **Input Validation** - Comprehensive request validation with detailed error messages
- 🔐 **Environment-based Config** - Secure configuration management
- 📝 **Structured Logging** - Production-grade request/response logging
- ⚠️ **Error Handling** - Graceful error handling with meaningful responses
- 🚀 **TypeScript** - Full type safety and IntelliSense support

## 📋 Table of Contents

- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Database Setup](#-database-setup)
- [API Endpoints](#-api-endpoints)
- [Usage Examples](#-usage-examples)
- [Monitoring Client Integration](#-monitoring-client-integration)
- [Querying Historical Data](#-querying-historical-data)
- [Deployment](#-deployment)
- [Development](#-development)
- [Troubleshooting](#-troubleshooting)

## 🚀 Quick Start

```bash
# Navigate to server directory
cd server

# Install dependencies
bun install

# Set up environment variables
cp .env.example .env
# Edit .env with your Supabase credentials

# Run database migrations
bun run migrate

# Start the server
bun run dev
```

## 📦 Installation

### Prerequisites

- Bun (latest version)
- Supabase account and project
- Git

### Local Setup

1. **Navigate to the server directory:**
   ```bash
   cd server
   ```

2. **Install dependencies:**
   ```bash
   bun install
   ```

3. **Configure environment variables:**
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` with your configuration (see [Configuration](#-configuration))

4. **Set up the database:**
   ```bash
   # Run migrations to create tables
   bun run migrate
   ```

5. **Start the development server:**
   ```bash
   bun run dev
   ```

The server will start on `http://localhost:3000` (or your configured port).

## ⚙️ Configuration

### Environment Variables

Create a `.env` file in the `server` directory:

```env
# Server Configuration
PORT=3000
NODE_ENV=production

# API Security
API_KEY=your-secret-api-key-here

# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_KEY=your-supabase-service-role-key

# Optional: CORS Configuration
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# Optional: Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60000
```

### Configuration Details

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `PORT` | No | Server port | `3000` |
| `NODE_ENV` | No | Environment mode | `production`, `development` |
| `API_KEY` | Yes | Secret key for API authentication | `sk_live_abc123...` |
| `SUPABASE_URL` | Yes | Your Supabase project URL | `https://xyz.supabase.co` |
| `SUPABASE_SERVICE_KEY` | Yes | Supabase service role key | `eyJ...` |
| `ALLOWED_ORIGINS` | No | Comma-separated allowed CORS origins | `https://app.com` |
| `RATE_LIMIT_REQUESTS` | No | Max requests per window | `100` |
| `RATE_LIMIT_WINDOW` | No | Rate limit window (ms) | `60000` |

### Getting Supabase Credentials

1. **Create a Supabase Project:**
   - Go to [supabase.com](https://supabase.com)
   - Click "New Project"
   - Note your project URL and keys

2. **Find Your Credentials:**
   - Navigate to Project Settings → API
   - Copy `URL` → use as `SUPABASE_URL`
   - Copy `service_role` key → use as `SUPABASE_SERVICE_KEY`
   - ⚠️ **Never expose the service role key publicly**

## 🗄️ Database Setup

### Database Schema

The server uses a single table to store monitoring reports:

```sql
CREATE TABLE monitoring_reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  service VARCHAR(255) NOT NULL,
  environment VARCHAR(100) NOT NULL,
  total_checks INTEGER NOT NULL,
  uptime_count INTEGER NOT NULL,
  downtime_count INTEGER NOT NULL,
  degraded_count INTEGER NOT NULL,
  uptime_percent DECIMAL(5,2) NOT NULL,
  average_latency_ms DECIMAL(10,2),
  timestamp TIMESTAMPTZ NOT NULL,
  results JSONB NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  
  -- Indexes for common queries
  INDEX idx_environment ON monitoring_reports(environment),
  INDEX idx_timestamp ON monitoring_reports(timestamp),
  INDEX idx_created_at ON monitoring_reports(created_at)
);
```

### Running Migrations

```bash
# Run all migrations
bun run migrate

# Rollback last migration
bun run migrate:rollback

# Check migration status
bun run migrate:status
```

### Manual Setup via Supabase Dashboard

If you prefer to set up the table manually:

1. Go to your Supabase Dashboard
2. Navigate to **SQL Editor**
3. Run the schema SQL above
4. Enable Row Level Security (RLS) if needed:
   ```sql
   ALTER TABLE monitoring_reports ENABLE ROW LEVEL SECURITY;
   
   -- Allow service role to do everything
   CREATE POLICY "Service role has full access"
   ON monitoring_reports
   FOR ALL
   TO service_role
   USING (true)
   WITH CHECK (true);
   ```

## 🔌 API Endpoints

### POST /api/monitoring/reports

Submit a new monitoring report.

**Authentication:** Required (Bearer token)

**Request:**
```http
POST /api/monitoring/reports HTTP/1.1
Host: localhost:3000
Content-Type: application/json
Authorization: Bearer your-api-key

{
  "service": "Uptime Monitor",
  "environment": "production",
  "total_checks": 3,
  "uptime_count": 2,
  "downtime_count": 1,
  "degraded_count": 0,
  "uptime_percent": 66.67,
  "average_latency_ms": 250.5,
  "timestamp": "2025-11-09T10:30:00Z",
  "results": [
    {
      "domain": "example.com",
      "url": "https://example.com",
      "status": "up",
      "status_code": 200,
      "response_time_ms": 150,
      "is_ssl": true,
      "ssl_expiry": "2025-12-31T23:59:59Z",
      "ssl_days_left": 55,
      "content_length": 1024,
      "timestamp": "2025-11-09T10:30:00Z",
      "checked_at": "2025-11-09T10:30:00Z"
    }
  ]
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Report stored successfully",
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-11-09T10:30:00.000Z"
}
```

**Response (Error):**
```json
{
  "success": false,
  "error": "Invalid request: missing required field 'environment'"
}
```

### GET /api/monitoring/reports

Retrieve monitoring reports with optional filters.

**Authentication:** Required (Bearer token)

**Query Parameters:**

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `environment` | string | Filter by environment | `production` |
| `from` | ISO date | Start date filter | `2025-11-01T00:00:00Z` |
| `to` | ISO date | End date filter | `2025-11-09T23:59:59Z` |
| `limit` | number | Max results (default: 100) | `50` |
| `offset` | number | Pagination offset | `0` |

**Request:**
```http
GET /api/monitoring/reports?environment=production&limit=10 HTTP/1.1
Host: localhost:3000
Authorization: Bearer your-api-key
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "service": "Uptime Monitor",
      "environment": "production",
      "total_checks": 3,
      "uptime_count": 2,
      "downtime_count": 1,
      "degraded_count": 0,
      "uptime_percent": 66.67,
      "average_latency_ms": 250.5,
      "timestamp": "2025-11-09T10:30:00Z",
      "results": [...],
      "created_at": "2025-11-09T10:30:01Z"
    }
  ],
  "count": 1,
  "limit": 10,
  "offset": 0
}
```

### GET /api/health

Health check endpoint (no authentication required).

**Request:**
```http
GET /api/health HTTP/1.1
Host: localhost:3000
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-11-09T10:30:00.000Z",
  "version": "1.0.0"
}
```

## 📘 Usage Examples

### Submit a Report (cURL)

```bash
curl -X POST http://localhost:3000/api/monitoring/reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "service": "Uptime Monitor",
    "environment": "production",
    "total_checks": 3,
    "uptime_count": 3,
    "downtime_count": 0,
    "degraded_count": 0,
    "uptime_percent": 100.00,
    "average_latency_ms": 150.0,
    "timestamp": "2025-11-09T10:30:00Z",
    "results": []
  }'
```

### Query Reports (cURL)

```bash
# Get last 10 production reports
curl -X GET "http://localhost:3000/api/monitoring/reports?environment=production&limit=10" \
  -H "Authorization: Bearer your-api-key"

# Get reports from date range
curl -X GET "http://localhost:3000/api/monitoring/reports?from=2025-11-01T00:00:00Z&to=2025-11-09T23:59:59Z" \
  -H "Authorization: Bearer your-api-key"
```

### Submit Report (JavaScript/TypeScript)

```typescript
const report = {
  service: "Uptime Monitor",
  environment: "production",
  total_checks: 3,
  uptime_count: 3,
  downtime_count: 0,
  degraded_count: 0,
  uptime_percent: 100.00,
  average_latency_ms: 150.0,
  timestamp: new Date().toISOString(),
  results: []
};

const response = await fetch('http://localhost:3000/api/monitoring/reports', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${process.env.API_KEY}`
  },
  body: JSON.stringify(report)
});

const data = await response.json();
console.log(data);
```

## 🔗 Monitoring Client Integration

### Configure the Go Monitor

In your main uptime monitor configuration:

```bash
# Set the server URL and API key
export API_URL="http://localhost:3000/api/monitoring/reports"
export API_KEY="your-api-key"
```

Or in GitHub Actions secrets:
- `API_URL`: `https://your-domain.com/api/monitoring/reports`
- `API_KEY`: Your generated API key

### Automatic Report Submission

The Go monitoring client will automatically:
1. Run health checks on configured domains
2. Generate JSON report
3. Submit to your server endpoint
4. Retry on transient failures
5. Log submission status

**Monitor logs will show:**
```json
{
  "level": "info",
  "msg": "Successfully submitted report to API",
  "status_code": 200
}
```

## 📊 Querying Historical Data

### Using Supabase Dashboard

1. Go to your Supabase Dashboard
2. Navigate to **Table Editor**
3. Select `monitoring_reports` table
4. Use filters to query data

### SQL Queries

```sql
-- Get average uptime for last 24 hours
SELECT 
  environment,
  AVG(uptime_percent) as avg_uptime,
  COUNT(*) as check_count
FROM monitoring_reports
WHERE created_at >= NOW() - INTERVAL '24 hours'
GROUP BY environment;

-- Find all downtime incidents
SELECT 
  environment,
  timestamp,
  downtime_count,
  results
FROM monitoring_reports
WHERE downtime_count > 0
ORDER BY timestamp DESC;

-- Calculate SLA compliance (99.9% uptime)
SELECT 
  environment,
  COUNT(*) as total_checks,
  SUM(CASE WHEN uptime_percent >= 99.9 THEN 1 ELSE 0 END) as sla_compliant,
  ROUND(
    (SUM(CASE WHEN uptime_percent >= 99.9 THEN 1 ELSE 0 END)::NUMERIC / COUNT(*)) * 100, 
    2
  ) as sla_compliance_rate
FROM monitoring_reports
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY environment;
```

### Programmatic Querying

```typescript
import { createClient } from '@supabase/supabase-js';

const supabase = createClient(
  process.env.SUPABASE_URL!,
  process.env.SUPABASE_SERVICE_KEY!
);

// Get reports for last 7 days
const { data, error } = await supabase
  .from('monitoring_reports')
  .select('*')
  .gte('created_at', new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString())
  .order('created_at', { ascending: false });

if (error) {
  console.error('Query error:', error);
} else {
  console.log(`Found ${data.length} reports`);
}
```

## 🚀 Deployment

### Deploy to Production

1. **Set Production Environment Variables**
   ```bash
   export NODE_ENV=production
   export API_KEY=your-production-api-key
   export SUPABASE_URL=your-production-supabase-url
   export SUPABASE_SERVICE_KEY=your-production-service-key
   ```

2. **Build the Application**
   ```bash
   bun run build
   ```

3. **Start Production Server**
   ```bash
   bun run start
   ```

### Deploy to Cloud Platforms

#### Vercel / Netlify

```bash
# Install Vercel CLI
npm i -g vercel

# Deploy
vercel --prod
```

Add environment variables in the dashboard.

#### Railway

1. Connect your GitHub repository
2. Add environment variables
3. Deploy automatically on push

#### Fly.io

```bash
# Install Fly CLI
curl -L https://fly.io/install.sh | sh

# Initialize and deploy
fly launch
fly secrets set API_KEY=your-key
fly secrets set SUPABASE_URL=your-url
fly secrets set SUPABASE_SERVICE_KEY=your-key
fly deploy
```

#### Docker

```dockerfile
FROM oven/bun:latest

WORKDIR /app

COPY package.json bun.lockb ./
RUN bun install --frozen-lockfile

COPY . .

EXPOSE 3000

CMD ["bun", "run", "start"]
```

```bash
# Build and run
docker build -t uptime-monitor-server .
docker run -p 3000:3000 --env-file .env uptime-monitor-server
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🛠️ Development

### Project Structure

```
server/
├── src/
│   ├── index.ts              # Application entry point
│   ├── routes/
│   │   └── monitoring.ts     # API routes
│   ├── services/
│   │   └── database.ts       # Supabase client
│   ├── middleware/
│   │   ├── auth.ts          # Authentication
│   │   └── validation.ts    # Request validation
│   └── types/
│       └── monitoring.ts     # TypeScript types
├── migrations/
│   └── 001_create_tables.sql
├── tests/
│   └── api.test.ts
├── package.json
├── tsconfig.json
├── .env.example
└── README.md
```

### Available Scripts

```bash
# Development
bun run dev              # Start with hot reload
bun run build            # Build for production
bun run start            # Start production server

# Database
bun run migrate          # Run migrations
bun run migrate:rollback # Rollback last migration
bun run db:seed          # Seed test data

# Testing
bun test                 # Run all tests
bun test:watch          # Run tests in watch mode
bun test:coverage       # Generate coverage report

# Code Quality
bun run lint            # Lint code
bun run format          # Format code with Prettier
bun run type-check      # TypeScript type checking
```

### Adding New Endpoints

1. Create route handler in `src/routes/`
2. Add validation middleware
3. Update TypeScript types
4. Write tests
5. Update API documentation

### Running Tests

```bash
# Run all tests
bun test

# Run specific test file
bun test api.test.ts

# Watch mode
bun test --watch

# With coverage
bun test --coverage
```

## 🐛 Troubleshooting

### Common Issues

**Problem: "Unauthorized" error**
```bash
# Verify API key is set correctly
echo $API_KEY

# Check Authorization header format
# Should be: Authorization: Bearer your-api-key
```

**Problem: "Database connection failed"**
```bash
# Verify Supabase credentials
echo $SUPABASE_URL
echo $SUPABASE_SERVICE_KEY

# Test connection
curl $SUPABASE_URL/rest/v1/ \
  -H "apikey: $SUPABASE_SERVICE_KEY"
```

**Problem: "Table does not exist"**
```bash
# Run migrations
bun run migrate

# Or create table manually in Supabase Dashboard
```

**Problem: Port already in use**
```bash
# Change port
export PORT=3001

# Or kill process using port 3000
lsof -ti:3000 | xargs kill -9
```

**Problem: CORS errors**
```bash
# Add your frontend domain to ALLOWED_ORIGINS
export ALLOWED_ORIGINS=https://yourdomain.com
```

### Debug Mode

Enable detailed logging:

```bash
export LOG_LEVEL=debug
bun run dev
```

### Health Check

```bash
# Verify server is running
curl http://localhost:3000/api/health

# Expected response:
# {"status":"ok","timestamp":"...","version":"1.0.0"}
```

## 📈 Performance Tips

1. **Database Indexes**: Ensure indexes exist on frequently queried columns
   ```sql
   CREATE INDEX IF NOT EXISTS idx_environment ON monitoring_reports(environment);
   CREATE INDEX IF NOT EXISTS idx_timestamp ON monitoring_reports(timestamp);
   ```

2. **Connection Pooling**: Supabase handles this automatically

3. **Rate Limiting**: Implement rate limiting for public endpoints
   ```typescript
   import rateLimit from 'express-rate-limit';
   
   const limiter = rateLimit({
     windowMs: 15 * 60 * 1000, // 15 minutes
     max: 100 // limit each IP to 100 requests per windowMs
   });
   ```

4. **Caching**: Cache frequent queries using Redis or in-memory cache

5. **Query Optimization**: Use Supabase query hints and explain plans

## 🔒 Security Best Practices

1. **Never commit secrets**: Use `.env` and add to `.gitignore`
2. **Rotate API keys**: Regularly update `API_KEY` values
3. **Use HTTPS**: Always use TLS in production
4. **Rate limiting**: Prevent API abuse
5. **Input validation**: Validate all request data
6. **SQL injection**: Use parameterized queries (Supabase handles this)
7. **CORS**: Restrict to known origins only
8. **Monitoring**: Log all API access and errors

## 📄 License

MIT License - see [LICENSE](../LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Bun](https://bun.sh)
- Database by [Supabase](https://supabase.com)
- Inspired by modern observability platforms

## 📞 Support

- 📧 Open an issue on GitHub
- 💬 Start a discussion
- ⭐ Star the repository if you find it useful!

## 🎯 Roadmap

- [ ] Real-time WebSocket updates
- [ ] Grafana dashboard templates
- [ ] Advanced analytics endpoints
- [ ] Multi-tenant support
- [ ] GraphQL API
- [ ] Automated anomaly detection
- [ ] SLA compliance reports

---

**Part of the Axiolot Hub Uptime Monitor ecosystem** | [Main Monitor](../README.md) | [Server API](./README.md)