export type Report = {
    timestamp: string;
    environment: string;
    uptime_percent: number;
    average_latency_ms: number;
    uptime_count: number;
    downtime_count: number;
    degraded_count: number;
};