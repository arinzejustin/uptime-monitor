import { supabase, SUMMARY_TABLE } from "../config/config.js";
import type { DailySummary } from "../../types.js";

export async function getCachedSummaries(domains: string[], days: number) {
    const sinceDate = new Date();
    sinceDate.setDate(sinceDate.getDate() - days);

    const { data, error } = await supabase
        .from(SUMMARY_TABLE)
        .select("*")
        .in("domain", domains)
        .gte("date", sinceDate.toISOString().slice(0, 10));


    if (error) {
        console.log(error)
        return [];
    }

    console.log(data)

    return (data || []).map(item => ({
        ...item,
        time_down: item.time_down || '0m'
    }));
}

export async function saveSummariesToCache(summaries: Record<string, DailySummary[]>) {
    const toInsert = [];
    for (const [domain, records] of Object.entries(summaries)) {
        for (const rec of records) {
            const timeDownInterval = convertToPostgresInterval(rec.time_down);

            toInsert.push({
                domain,
                date: rec.date,
                status: rec.status,
                title: rec.title,
                description: rec.description,
                time_down: timeDownInterval,
                total_downtime: calculateTotalMinutes(rec.time_down),
                updated_at: new Date().toISOString()
            });
        }
    }

    const { error } = await supabase
        .from(SUMMARY_TABLE)
        .upsert(toInsert, { onConflict: "domain,date" });

    if (error) {
        console.error("Cache save error:", error);
    } else {
        console.log(`Cached ${toInsert.length} summaries to Supabase`);
    }
}

function convertToPostgresInterval(timeDown: string): string {
    const hourMatch = timeDown.match(/(\d+)h/);
    const minMatch = timeDown.match(/(\d+)m/);

    const hours = hourMatch ? parseInt(hourMatch[1] || "0") : 0;
    const minutes = minMatch ? parseInt(minMatch[1] || "0") : 0;

    if (hours > 0 && minutes > 0) {
        return `${hours} hours ${minutes} minutes`;
    } else if (hours > 0) {
        return `${hours} hours`;
    } else {
        return `${minutes} minutes`;
    }
}

function calculateTotalMinutes(timeDown: string): number {
    const hourMatch = timeDown.match(/(\d+)h/);
    const minMatch = timeDown.match(/(\d+)m/);

    const hours = hourMatch ? parseInt(hourMatch[1] || "0") : 0;
    const minutes = minMatch ? parseInt(minMatch[1] || "0") : 0;

    return (hours * 60) + minutes;
}