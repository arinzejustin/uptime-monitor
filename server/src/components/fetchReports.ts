import { supabase, REPORTS_TABLE } from "../config/config.js";
import { summarizeDomainData } from "../utils/summarizer.js";
import { getCachedSummaries, saveSummariesToCache } from "../utils/cacheSummary.js";

export async function fetchReports(query: {
    domains?: string[];
    days?: number;
    limit?: number;
    useCache?: boolean;
}) {
    const days = query.days || 60;
    const domains = query.domains || [];
    console.log(query.useCache, query.limit, query.days, JSON.stringify(query.domains))

    if (query.useCache && domains.length > 0) {
        console.log('using cache')
        const cached = await getCachedSummaries(domains, days);
        if (cached.length > 0) {
            return groupByDomain(cached);
        }
    }

    const sinceDate = new Date();
    sinceDate.setDate(sinceDate.getDate() - days);

    let supabaseQuery = supabase
        .from(REPORTS_TABLE)
        .select("timestamp, results")
        .gte("timestamp", sinceDate.toISOString())
        .order("timestamp", { ascending: true });

    if (domains.length > 0) {
        domains.map(domain => ({
            domain: domain
        }));

        supabaseQuery = supabaseQuery.or(
            domains.map(domain =>
                `results.cs.${JSON.stringify([{ domain }])}`
            ).join(',')
        );
    }

    const { data, error } = await supabaseQuery.limit(query.limit || 1000);

    if (error) {
        console.log(error)
        throw new Error(`Failed to fetch reports: ${error.message}`);
    }

    const summaries = summarizeDomainData(data, domains);

    if (query.useCache && domains.length > 0) {
        await saveSummariesToCache(summaries);
    }

    return summaries;
}

function groupByDomain(records: any[]) {
    const grouped: Record<string, any[]> = {};
    for (const rec of records) {
        if (!grouped[rec.domain]) grouped[rec.domain] = [];
        grouped[rec.domain]!.push({
            date: rec.date,
            status: rec.status,
            title: rec.title,
            description: rec.description,
            time_down: rec.time_down
        });
    }
    return grouped;
}