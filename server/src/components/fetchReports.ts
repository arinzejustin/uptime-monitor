import { supabase, REPORTS_TABLE } from "../config/config.js";
import { summarizeDomainData } from "../utils/summarizer.js";
import {
  getCachedSummaries,
  saveSummariesToCache,
} from "../utils/cacheSummary.js";

export async function fetchReports(query: {
  domains?: string[];
  days?: number;
  useCache?: boolean;
}) {
  const days = query.days || 60;
  const domains = query.domains || [];
  console.log(
    query.useCache,
    query.days,
    JSON.stringify(query.domains),
  );

  if (query.useCache && domains.length > 0) {
    console.log("using cache");
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
    const conditions = domains
      .map((d) => `results@>>[{"domain":"${d}"}]`)
      .join(",");
    console.log("Filter condition:", conditions);
    supabaseQuery = supabaseQuery.or(conditions);
  }

  const { data, error } = await supabaseQuery;

  console.log(`Fetched ${data?.length} records from Supabase`);
  if (data && data.length > 0) {
    console.log(
      `Date range: ${data[0]?.timestamp} to ${data[data.length - 1]?.timestamp}`,
    );
  }

  if (error) {
    console.log(error);
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
      time_down: rec.time_down,
    });
  }
  return grouped;
}
