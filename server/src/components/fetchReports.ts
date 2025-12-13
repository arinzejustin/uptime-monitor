import { supabase, REPORTS_TABLE } from "../config/config.js";
import { summarizeDomainData } from "../utils/summarizer.js";
import {
  getCachedSummaries,
  saveSummariesToCache,
} from "../utils/cacheSummary.js";
import type { HealthCheckResult } from "../../types.d.js";

export async function fetchReports(query: {
  domains?: string[];
  days?: number;
  useCache?: boolean;
}) {
  let allData: any[] = [];
  let from = 0;
  const pageSize = 1000;

  const days = query.days ?? 60;
  const domains = query.domains ?? [];

  if (query.useCache && domains.length > 0) {
    const cached = await getCachedSummaries(domains, days);
    if (cached.length > 0) {
      return groupByDomain(cached);
    }
  }

  const sinceDate = new Date();
  sinceDate.setDate(sinceDate.getDate() - days);

  while (true) {
    const { data, error } = await supabase
      .from(REPORTS_TABLE)
      .select("timestamp, results")
      .gte("timestamp", sinceDate.toISOString())
      .order("timestamp", { ascending: true })
      .range(from, from + pageSize - 1);

    if (error) throw new Error(error.message);

    allData = allData.concat(data ?? []);
    if (!data || data.length < pageSize) break;

    from += pageSize;
  }

  const filteredData = allData.map((row: any) => ({
    ...row,
    results: row.results.filter(
      (r: HealthCheckResult) =>
        domains.length === 0 || domains.includes(r.domain),
    ),
  }));

  const summaries = summarizeDomainData(filteredData, domains);

  if (query.useCache && domains.length > 0) {
    await saveSummariesToCache(summaries);
  }
  console.log("sinceDate:", sinceDate.toISOString());
  console.log("rows fetched:", allData.length);
  console.log("first row timestamp:", allData[0]?.timestamp);
  console.log("last row timestamp:", allData[allData.length - 1]?.timestamp);

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
