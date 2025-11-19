import type { Context, Next } from 'hono';
import { supabase, API_KEY_TABLE } from '../config/config.ts';
import { createHash } from 'crypto';
import { updateKeyUsage } from './apiKeyGenerator.ts.ts';


export async function validateApiKey(c: Context, next: Next) {
    const authHeader = c.req.header('Authorization');

    if (!authHeader) {
        return c.json({
            error: 'Unauthorized',
            message: 'Missing Authorization header'
        }, 401);
    }

    const parts = authHeader.split(' ');
    if (parts.length !== 2 || parts[0] !== 'Bearer') {
        return c.json({
            error: 'Unauthorized',
            message: 'Invalid Authorization header format. Use: Bearer <api_key>'
        }, 401);
    }

    const apiKey = parts[1] as string;

    // Validate key format
    if (!apiKey.startsWith('axh_')) {
        return c.json({
            error: 'Unauthorized',
            message: 'Invalid API key format'
        }, 401);
    }

    // Hash the provided key
    const hashedKey = createHash('sha256').update(apiKey).digest('hex');

    // Check database
    const { data, error } = await supabase
        .from(API_KEY_TABLE)
        .select('id, name, is_active, rate_limit')
        .eq('key_hash', hashedKey)
        .single();

    if (error || !data) {
        return c.json({
            error: 'Unauthorized',
            message: 'Invalid API key'
        }, 401);
    }

    if (!data.is_active) {
        return c.json({
            error: 'Unauthorized',
            message: 'API key has been deactivated'
        }, 401);
    }

    updateKeyUsage(hashedKey).catch(console.error);

    c.set('apiKeyId', data.id);
    c.set('apiKeyName', data.name);

    await next();
}