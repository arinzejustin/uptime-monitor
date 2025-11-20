import { supabase, REPORTS_TABLE } from '../config/config.js';
import type { MonitorReport } from '../../types.js';

export async function submitReport(report: MonitorReport) {
    if (!report.service || !report.environment) {
        throw new Error('Missing required fields: service, environment');
    }

    if (report.uptime_percent < 0 || report.uptime_percent > 100) {
        throw new Error('uptime_percent must be betwe.jsen 0 and 100');
    }

    if (!Array.isArray(report.results) || report.results.length === 0) {
        throw new Error('results array must not be empty');
    }

    // const domains = report.results.map((r) => r.domain).filter(Boolean);

    const payload = {
        service: report.service,
        environment: report.environment || "production",
        total_checks: report.total_checks,
        uptime_count: report.uptime_count,
        downtime_count: report.downtime_count,
        degraded_count: report.degraded_count,
        uptime_percent: Number(report.uptime_percent),
        average_latency_ms: report.average_latency_ms,
        timestamp: report.timestamp || new Date().toISOString(),
        results: report.results,
    };

    const { data, error } = await supabase
        .from(REPORTS_TABLE)
        .insert(payload)
        .select('id, created_at')
        .single();

    if (error) {
        console.error('Supabase error:', error);
        throw new Error(`Database error: ${error.message}`);
    }

    return {
        id: data.id,
        created_at: data.created_at,
    };
}
