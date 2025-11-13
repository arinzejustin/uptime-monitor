// server.ts
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { secureHeaders } from 'hono/secure-headers';
import { timing } from 'hono/timing';
import { z } from 'zod';
import { zValidator } from '@hono/zod-validator';
import { generateApiKey, getApiKeyInfo } from './src/apiKeyGenerator.ts';
import { validateApiKey } from './src/apiKeyValidator';
import { submitReport } from './src/submitEndpoint';
import { fetchVisualization } from './src/visualData';
import { visualizationPage } from './src/visualization';
import { fetchReports } from './src/fetchReports.ts';

// Environment
const NODE_ENV = Bun.env.NODE_ENV || 'development';
const ALLOWED_ORIGINS = (Bun.env.ALLOWED_ORIGINS || '*').split(',').map(s => s.trim());
const ALLOWED_USER_AGENT = Bun.env.ALLOWED_USER_AGENT || 'MonitoringClient/1.0';

const rateLimiter = new Map<string, { count: number; resetAt: number }>();
const RATE_LIMIT = 50; // requests per 15 minutes
const RATE_WINDOW = 15 * 60 * 1000;

const rateLimitMiddleware = async (c: any, next: () => Promise<void>) => {
  const ip = c.req.header('x-forwarded-for') || 'unknown';
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

  if (rateLimiter.size > 10_000) {
    for (const [key, val] of rateLimiter.entries()) {
      if (val.resetAt < now) rateLimiter.delete(key);
    }
  }

  await next();
};

const app = new Hono();

// Global Middlewares
app.use('*', logger());
app.use('*', timing());
app.use('*', secureHeaders());
app.use(
  '*',
  cors({
    origin: ALLOWED_ORIGINS,
    allowMethods: ['GET', 'POST', 'OPTIONS'],
    allowHeaders: ['Content-Type', 'Authorization', 'User-Agent'],
    exposeHeaders: ['Content-Length'],
    maxAge: 600,
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
  name: z.string().min(1).max(100),
  description: z.string().optional(),
});

app.post(
  '/api/v1/keys/generate',
  zValidator('json', generateKeySchema),
  async (c) => {
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
      console.error('Generate API Key Error:', error);
      return c.json(
        { error: 'Failed to generate API key', details: error.message },
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
    console.error('Fetch API Key Error:', error);
    return c.json(
      { error: 'Failed to fetch API key info', details: error.message },
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
      if (userAgent !== ALLOWED_USER_AGENT) {
        return c.json(
          {
            error: 'Forbidden',
            message: 'Invalid User-Agent header'
          },
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
          { error: 'Validation failed', details: error.issues },
          400
        );
      }
      return c.json(
        { error: 'Failed to store report', details: error.message },
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
  from: z.coerce.string().optional(),
  to: z.coerce.string().optional(),
  limit: z.coerce.number().int().min(1).max(1000).default(50),
  offset: z.coerce.number().int().min(0).default(0),
  sortBy: z.string().optional(),
  sortOrder: z.enum(['asc', 'desc']).optional()
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
      console.error('Query Reports Error:', error);
      return c.json(
        { error: 'Failed to fetch reports', details: error.message },
        500
      );
    }
  }
);

const summaryQuerySchema = z.object({
  domains: z.array(z.string().url()).optional(),
  limit: z.coerce.number().int().min(1).max(1000).default(100).optional(),
  days: z.coerce.number().int().min(1).max(31).default(30),
  useCache: z.boolean().default(true),
});

app.post(
  '/api/v1/reports/query/summary',
  validateApiKey, zValidator('json', summaryQuerySchema),
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
        data: result,
      });
    } catch (error: any) {
      console.error('Fetch Summary Error:', error);
      return c.json(
        { error: 'Failed to fetch summary', details: error.message },
        500
      );
    }
  }
);

// === Dashboard Page ===
app.get('/dashboard', async (c) => {
  try {
    const html = await visualizationPage();
    return c.html(html);
  } catch (error: any) {
    console.error('Dashboard Load Error:', error);
    return c.html('<h1>Dashboard Unavailable</h1><p>Please try again later.</p>', 500);
  }
});

// === Dashboard Data (Protected) ===
const dashboardQuerySchema = z.object({
  environment: z.string().optional(),
  limit: z.coerce.number().int().min(1).max(100).default(50),
});

app.get(
  '/dashboard/data',
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
      console.error('Dashboard Data Error:', error);
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
  console.error('Unhandled Error:', err);
  return c.json(
    {
      error: 'Internal Server Error',
      message: NODE_ENV === 'development' ? err.message : 'Something went wrong',
    },
    500
  );
});


export default app;