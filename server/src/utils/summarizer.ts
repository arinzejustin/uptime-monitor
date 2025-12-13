import { format, parseISO } from "date-fns";

export function summarizeDomainData(records: any[], domains?: string[]) {
    const domainMap: Record<string, Record<string, any[]>> = {};

    for (const rec of records) {
        if (!rec.results) continue;

        for (const result of rec.results) {
            const domain = result.domain;

            if (domains && !domains.includes(domain)) continue;

            const dateKey = format(parseISO(rec.timestamp), "yyyy-MM-dd");

            if (!domainMap[domain]) domainMap[domain] = {};
            if (!domainMap[domain][dateKey]) domainMap[domain][dateKey] = [];

            domainMap[domain][dateKey].push(result);
        }
    }

    const summaries: Record<string, any[]> = {};

    for (const [domain, dailyData] of Object.entries(domainMap)) {
        const dailySummaries = [];

        for (const [dateKey, checks] of Object.entries(dailyData)) {
            const downChecks = checks.filter((c) => c.status === "down");
            const downCount = downChecks.length;
            const avgIntervalMinutes = 10; // Depending on monitoring frequency
            const downMinutes = downCount * avgIntervalMinutes;
            const hours = Math.floor(downMinutes / 60);
            const minutes = downMinutes % 60;
            const timeDown = hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`;

            const displayDate = format(parseISO(dateKey), "MMM dd yyyy");

            let status = "ok";
            let title = "Operational";
            let description = "No issues recorded today";

            if (downCount >= 10) {
                status = "error";
                title = "Major Outage";
                description = `${domain} experienced extended downtime (${timeDown}).`;
            } else if (downCount > 0) {
                status = "warning";
                title = "Partial Outage";
                description = `${domain} had intermittent downtime (${timeDown}).`;
            }

            dailySummaries.push({
                date: displayDate,
                status,
                title,
                description,
                time_down: timeDown,
            });
        }

        dailySummaries.sort((a, b) => (a.date < b.date ? 1 : -1));
        summaries[domain] = dailySummaries;
    }

    console.log(summaries)

    return summaries;
}
