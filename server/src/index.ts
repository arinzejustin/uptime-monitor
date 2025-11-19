// server.ts
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { secureHeaders } from 'hono/secure-headers';
import { handle } from '@hono/node-server/vercel'
import { timing } from 'hono/timing';
import { z } from 'zod';
import { zValidator } from '@hono/zod-validator';
import { generateApiKey, getApiKeyInfo } from './components/apiKeyGenerator.ts.ts';
import { validateApiKey } from './components/apiKeyValidator.ts';
import { submitReport } from './components/submitEndpoint.ts';
import { fetchVisualization } from './components/visualData.ts';
import { fetchReports } from './components/fetchReports.ts';
import { currentStatus } from './components/currentStatus.ts';

const NODE_ENV = process.env.NODE_ENV || 'development';
const ALLOWED_USER_AGENT = process.env.ALLOWED_USER_AGENT || 'MonitoringClient/1.0';

const rateLimiter = new Map<string, { count: number; resetAt: number }>();
const RATE_LIMIT = 50;
const RATE_WINDOW = 15 * 60 * 1000;

const rateLimitMiddleware = async (c: any, next: () => Promise<void>) => {
  const forwarded = c.req.header('x-forwarded-for');
  const ip = forwarded
    ? forwarded.split(',')[0].trim()
    : c.req.header('cf-connecting-ip') || 'unknown';

  const now = Date.now();
  const record = rateLimiter.get(ip);

  if (record && now < record.resetAt) {
    if (record.count >= RATE_LIMIT) {
      return c.json({ error: 'Rate limit exceeded' }, 429);
    }
    record.count++;
  } else {
    rateLimiter.set(ip, { count: 1, resetAt: now + RATE_WINDOW });
  }

  if (rateLimiter.size > 10_000 && Math.random() < 0.05) {
    for (const [key, val] of rateLimiter) {
      if (val.resetAt < now) rateLimiter.delete(key);
    }
  }

  await next();
};

const app = new Hono();

app.use('*', logger());
app.use('*', timing());
app.use('*', secureHeaders());
app.use(
  '*',
  cors({
    origin: '*',
    allowMethods: ['GET', 'POST', 'OPTIONS'],
    allowHeaders: ['Content-Type', 'Authorization', 'User-Agent', 'X-Key-Header'],
    exposeHeaders: ['Content-Length'],
    maxAge: 0,
    credentials: true,
  })
);


// Health Check
app.get('/health', (c) => {
  return c.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    version: '1.0.0',
    env: NODE_ENV,
  });
});

// === API Key Generation ===
const generateKeySchema = z.object({
  name: z.string().min(1).max(10000),
  description: z.string().optional(),
});

app.post(
  '/api/v1/keys/generate',
  zValidator('json', generateKeySchema),
  async (c) => {
    if (NODE_ENV !== 'development') {
      return c.json({ error: 'Unauthorized. Can\'t generate API key in production mode' }, 401);
    }
    try {
      const { name, description } = c.req.valid('json');
      const result = await generateApiKey(name, description || '');

      return c.json(
        {
          success: true,
          message: 'API key generated successfully',
          data: result,
        },
        201
      );
    } catch (error: any) {
      console.error('Generate API Key Error:', error.stack || error);
      return c.json(
        { error: 'Failed to generate API key', message: error.message },
        500
      );
    }
  }
);

// === Get API Key Info ===
app.get('/api/v1/keys/:keyId', async (c) => {
  try {
    const keyId = c.req.param('keyId');
    if (!keyId || keyId.length < 8) {
      return c.json({ error: 'Invalid key ID' }, 400);
    }

    const keyInfo = await getApiKeyInfo(keyId);
    if (!keyInfo) {
      return c.json({ error: 'API key not found' }, 404);
    }

    return c.json({ success: true, data: keyInfo });
  } catch (error: any) {
    console.error('Fetch API Key Error:', error.stack || error);
    return c.json(
      { error: 'Failed to fetch API key info', message: error.message },
      500
    );
  }
});

// === Submit Monitoring Report ===
app.post(
  '/api/v1/monitoring/reports',
  rateLimitMiddleware,
  validateApiKey,
  async (c) => {
    try {
      const userAgent = c.req.header('User-Agent');

      // Optimized & safer
      if (!userAgent?.startsWith(ALLOWED_USER_AGENT)) {
        return c.json(
          { error: 'Forbidden', message: 'Invalid User-Agent header' },
          403
        );
      }

      const body = await c.req.json();
      const result = await submitReport(body);

      return c.json(
        {
          success: true,
          message: 'Report stored successfully',
          data: result,
        },
        201
      );
    } catch (error: any) {
      if (error instanceof z.ZodError) {
        return c.json(
          { error: 'Validation failed', message: error.issues },
          400
        );
      }
      return c.json(
        { error: 'Failed to store report', message: error.message },
        500
      );
    }
  }
);

// === Query Reports ===
const querySchema = z.object({
  environment: z.string().optional(),
  url: z.string().url().optional(),
  status: z.enum(['up', 'down']).optional(),
  from: z.string().optional(),
  to: z.string().optional(),
  limit: z.coerce.number().int().min(1).max(1000).default(50),
  offset: z.coerce.number().int().min(0).default(0),
  sortBy: z.string().optional(),
  sortOrder: z.enum(['asc', 'desc']).optional(),
});

app.post(
  '/api/v1/reports/query/visualization',
  validateApiKey,
  zValidator('json', querySchema),
  async (c) => {
    try {
      const query = c.req.valid('json');
      const result = await fetchVisualization(query);

      return c.json({
        success: true,
        data: result.data,
        pagination: result.pagination,
      });
    } catch (error: any) {
      console.error('Query Reports Error:', error.stack || error);
      return c.json(
        { error: 'Failed to fetch reports', message: error.message },
        500
      );
    }
  }
);

const summaryQuerySchema = z.object({
  domains: z.array(z.string()),
  limit: z.coerce.number().int().min(1).max(1000).default(1000),
  days: z.coerce.number().int().min(1).max(60).default(60),
  useCache: z.boolean().default(true),
});

app.post(
  '/api/v1/reports/query/summary',
  validateApiKey,
  rateLimitMiddleware,
  zValidator('json', summaryQuerySchema),
  async (c) => {
    try {
      const query = c.req.valid('json');
      const result = await fetchReports({
        domains: query.domains,
        days: query.days,
        useCache: query.useCache,
        limit: query.limit,
      });

      return c.json({
        success: true,
        message: 'Fetched summary successfully',
        data: result,
      });
    } catch (error: any) {
      return c.json(
        { error: 'Failed to fetch summary', message: error.message },
        500
      );
    }
  }
);

const concurrentQuerySchema = z.object({
  domains: z.string().min(1),
});

app.post(
  '/api/v1/concurrent/status',
  validateApiKey,
  zValidator('json', concurrentQuerySchema),
  async (c) => {
    try {
      const query = c.req.valid('json');

      const result = await currentStatus({
        domains: query.domains
          .split(',')
          .map(d => d.trim())
          .filter(Boolean),
      });

      return c.json({
        success: true,
        message: result.message,
        data: result.statusResults,
      });
    } catch (error: any) {
      return c.json(
        { error: 'Failed to fetch concurrent status', message: error.message },
        500
      );
    }
  }
);

// Dashboard Data
const dashboardQuerySchema = z.object({
  environment: z.string().optional(),
  limit: z.coerce.number().int().min(1).max(100).default(50),
});

app.post(
  '/api/v1/dashboard/data',
  validateApiKey,
  zValidator('query', dashboardQuerySchema),
  async (c) => {
    try {
      const query = c.req.valid('query');
      const result = await fetchVisualization({
        environment: query.environment,
        limit: query.limit,
        offset: 0,
      });

      return c.json(result);
    } catch (error: any) {
      console.error('Dashboard Data Error:', error.stack || error);
      return c.json({ error: 'Failed to fetch dashboard data' }, 500);
    }
  }
);

// 404 Handler
app.notFound((c) => {
  return c.json(
    {
      error: 'Not Found',
      message: `Route ${c.req.method} ${c.req.url} not found`,
    },
    404
  );
});

// Global Error Handler
app.onError((err, c) => {
  console.error('Unhandled Error:', err.stack || err);
  return c.json(
    {
      error: 'Internal Server Error',
      message: NODE_ENV === 'development' ? err.message : 'Something went wrong',
    },
    500
  );
});

export default handle(app);
