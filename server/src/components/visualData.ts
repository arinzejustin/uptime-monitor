import { supabase, REPORTS_TABLE } from '../config/config.js';
import type { QueryInput } from '../../types.js';

export async function fetchVisualization(query: QueryInput) {
    let supabaseQuery = supabase
        .from(REPORTS_TABLE)
        .select('*', { count: 'exact' });

    if (query.environment) {
        supabaseQuery = supabaseQuery.eq('environment', query.environment);
    }

    if (query.from) {
        supabaseQuery = supabaseQuery.gte('timestamp', query.from);
    }

    if (query.to) {
        supabaseQuery = supabaseQuery.lte('timestamp', query.to);
    }

    // Apply sorting
    const sortBy = query.sortBy || 'timestamp';
    const sortOrder = query.sortOrder || 'desc';
    supabaseQuery = supabaseQuery.order(sortBy, { ascending: sortOrder === 'asc' });

    // Apply pagination
    const limit = Math.min(query.limit || 50, 100); // Max 100
    const offset = query.offset || 0;
    supabaseQuery = supabaseQuery.range(offset, offset + limit - 1);

    const { data, error, count } = await supabaseQuery;

    if (error) {
        console.error('Supabase error:', error);
        throw new Error(`Database error: ${error.message}`);
    }

    return {
        data: data || [],
        pagination: {
            total: count || 0,
            limit,
            offset,
            hasMore: (count || 0) > offset + limit
        }
    };
}

export async function getStats(environment?: string) {
    let query = supabase
        .from(REPORTS_TABLE)
        .select('uptime_percent, average_latency_ms, downtime_count');

    if (environment) {
        query = query.eq('environment', environment);
    }

    const { data, error } = await query;

    if (error) throw new Error(error.message);

    if (!data || data.length === 0) {
        return {
            avgUptime: 0,
            avgLatency: 0,
            totalDowntimeIncidents: 0,
            reportCount: 0
        };
    }

    const avgUptime = data.reduce((sum, r) => sum + r.uptime_percent, 0) / data.length;
    const avgLatency = data.reduce((sum, r) => sum + r.average_latency_ms, 0) / data.length;
    const totalDowntime = data.reduce((sum, r) => sum + r.downtime_count, 0);

    return {
        avgUptime: Math.round(avgUptime * 100) / 100,
        avgLatency: Math.round(avgLatency * 100) / 100,
        totalDowntimeIncidents: totalDowntime,
        reportCount: data.length
    };
}