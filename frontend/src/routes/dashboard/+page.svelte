<script lang="ts">
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";
    import { setMode } from "mode-watcher";
    import {
        ChartSpline,
        RefreshCcwDot,
        CloudDownload,
        Key,
    } from "@lucide/svelte";
    import { PUBLIC_API_V1_URL } from "$env/static/public";
    import type { Report } from "$lib/types";

    let apiKey = $state("");
    let environment = $state("");
    let allData: Report[] = $state([]);
    let loading = $state(true);
    let showModal = $state(false);
    let generating = $state(false);
    let refreshing = $state(false);

    let stats = $state({
        avgUptime: 0,
        avgLatency: 0,
        totalReports: 0,
        incidents: 0,
    });

    const stars = Array.from({ length: 500 }, (_, i) => ({
        id: i,
        left: Math.random() * 100,
        top: Math.random() * 100,
        delay: Math.random() * 3,
        duration: Math.random() * 3 + 2,
    }));

    let uptimeChartData = $derived(
        allData.slice(-20).map((r) => ({
            time: new Date(r.timestamp).toLocaleTimeString(),
            uptime: r.uptime_percent,
        })),
    );
    let totalUp = $derived(allData.reduce((sum, r) => sum + r.uptime_count, 0));
    let totalDown = $derived(
        allData.reduce((sum, r) => sum + r.downtime_count, 0),
    );
    let totalDegraded = $derived(
        allData.reduce((sum, r) => sum + r.degraded_count, 0),
    );

    async function generateKey() {
        const url = `${PUBLIC_API_V1_URL}/keys/generate`;
        generating = true;
    }

    async function loadDashboard(key = apiKey, env = environment) {
        if (!key) {
            toast.error("Please enter your API key", {
                duration: 4000,
            });
            return;
        }

        loading = true;
        refreshing = true;

        const url = `${PUBLIC_API_V1_URL}/dashboard/data`;

        try {
            const response = await fetch(url, {
                headers: { Authorization: `Bearer ${key}` },
                method: "POST",
                body: JSON.stringify({
                    environment: env || "",
                }),
            });

            if (!response.ok) {
                throw new Error("Failed to fetch data. Check your API key.");
            }

            const result = await response.json();
            const data = result.data.data;
            allData = data;

            if (allData.length === 0) {
                toast.info("No data available for the selected environment.", {
                    duration: 4000,
                });
                return;
            }
            updateStats(allData);
            loading = false;
        } catch (error: any) {
            toast.error("Error loading dashboard: " + error.message, {
                duration: 4000,
            });
            loading = true;
        } finally {
            refreshing = false;
        }
    }

    function updateStats(data: Report[]) {
        if (data.length === 0) return;

        const avgUptime =
            data.reduce((sum, r) => sum + r.uptime_percent, 0) / data.length;
        const avgLatency =
            data.reduce((sum, r) => sum + r.average_latency_ms, 0) /
            data.length;
        const incidents = data.reduce((sum, r) => sum + r.downtime_count, 0);

        stats = {
            avgUptime: parseFloat(avgUptime.toFixed(2)),
            avgLatency: parseFloat(avgLatency.toFixed(2)),
            totalReports: data.length,
            incidents,
        };
    }

    function handleApiKeyChange(e: Event) {
        const target = e.target as HTMLInputElement;
        apiKey = target.value;
        localStorage.setItem("apiKey", apiKey);
    }

    function exportData() {
        if (allData.length === 0) {
            toast.error("No data to export", {
                duration: 4000,
            });
            return;
        }
        const csv =
            "Timestamp,Environment,Uptime%,Latency,Up,Down,Degraded\n" +
            allData
                .map(
                    (r) =>
                        `${r.timestamp},${r.environment},${r.uptime_percent},${r.average_latency_ms},${r.uptime_count},${r.downtime_count},${r.degraded_count}`,
                )
                .join("\n");

        const blob = new Blob([csv], { type: "text/csv" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `uptime-report-${new Date().toISOString()}.csv`;
        a.click();
    }

    onMount(() => {
        const saved = localStorage.getItem("apiKey");
        if (saved) {
            apiKey = saved;
            loadDashboard(saved, "");
        }
        setMode("dark");
    });
</script>

<svelte:head>
    <title>Dashboard Monitor</title>
</svelte:head>

<!-- Dashboard -->
<div
    class="min-h-screen bg-[#222] text-white relative overflow-hidden font-code"
>
    <!-- Animated stars -->
    <div class="absolute inset-0 pointer-events-none">
        {#each stars as star (star.id)}
            <div
                class="absolute w-0.5 h-0.5 bg-white rounded-full animate-pulse"
                style="left: {star.left}%; top: {star.top}%; animation-delay: {star.delay}s; animation-duration: {star.duration}s;"
            ></div>
        {/each}
    </div>

    <div class="relative z-10 max-w-7xl mx-auto p-6">
        <!-- Header -->
        <div
            class="text-center mb-10 border p-10 rounded-[60%_40%_30%_70%/60%_30%_70%_40%] bg-neutral-900 hover:shadow-lg"
        >
            <h1 class="text-4xl font-bold mb-2 inline-flex items-center gap-2">
                <ChartSpline class="size-10 text-purple-600" />
                UPTIME MONITOR
            </h1>
            <p class="text-gray-200 text-sm">
                REAL-TIME MONITORING - AXIOLOT HUB
            </p>
        </div>

        <!-- API Key -->
        <div
            class="mb-6 border space-y-4 border-yellow-400 p-4 bg-neutral-900 rounded-[40%_60%_70%_30%/50%_60%_30%_60%] py-8"
        >
            <!-- svelte-ignore a11y_label_has_associated_control -->
            <label class="text-yellow-400 pb-2 text-lg">⚠️ API KEY</label>
            <input
                type="password"
                class="w-full bg-neutral-950 border border-slate-600 md: ml-4 md:w-3/4 mx-auto text-white p-3 rounded-xl outline-none focus:border-gray-500"
                bind:value={apiKey}
                oninput={handleApiKeyChange}
                placeholder="axh_..."
            />
        </div>

        <!-- Controls -->
        <div class="flex flex-wrap gap-4 mb-6">
            <select
                bind:value={environment}
                class="bg-neutral-900 border border-gray-700 p-3 rounded-xl text-white focus:border-gray-500"
            >
                <option value="">ALL ENVIRONMENTS</option>
                <option value="production">PRODUCTION</option>
                <option value="staging">STAGING</option>
                <option value="development">DEVELOPMENT</option>
            </select>
            <button
                disabled={refreshing}
                onclick={() => loadDashboard()}
                class="bg-neutral-900 border border-slate-700 px-6 py-3 rounded-xl hover:-translate-y-0.5 transition hover:shadow-xl cursor-pointer inline-flex items-center gap-x-2 disabled:pointer-events-none disabled:opacity-60"
            >
                <RefreshCcwDot
                    class="size-6 text-teal-400 {refreshing
                        ? 'animate-spin'
                        : ''}"
                /> REFRESH
            </button>
            <button
                onclick={exportData}
                class="bg-neutral-900 border border-slate-700 px-6 py-3 rounded-xl hover:-translate-y-0.5 transition hover:shadow-xl cursor-pointer inline-flex items-center gap-x-2"
            >
                <CloudDownload class="size-6 text-amber-500" /> EXPORT CSV
            </button>
            <button
                disabled={generating}
                onclick={() => (showModal = true)}
                class="bg-neutral-900 border border-slate-700 px-6 py-3 rounded-xl hover:-translate-y-0.5 transition hover:shadow-xl cursor-pointer inline-flex items-center gap-x-2 disabled:pointer-events-none disabled:opacity-50"
            >
                <Key class="size-6 text-green-500" />
                {generating ? "CREATING KEY" : "CREATE KEY"}
            </button>
        </div>

        <!-- Stats Grid -->
        <div class="grid md:grid-cols-4 sm:grid-cols-2 gap-4 mb-8">
            <div
                class="border border-slate-700 p-6 bg-neutral-900 rounded-[60%_40%_30%_70%/60%_30%_70%_40%] text-center transition hover:shadow-xl hover:-translate-y-0.5"
            >
                <div class="text-gray-400 text-xs mb-1">AVG. UPTIME</div>
                <div class="text-3xl font-bold">{stats.avgUptime}%</div>
            </div>
            <div
                class="border border-slate-700 p-6 bg-neutral-900 rounded-[40%_60%_70%_30%/50%_60%_30%_60%] text-center"
            >
                <div class="text-gray-400 text-xs mb-1">AVG. LATENCY</div>
                <div class="text-2xl font-bold">
                    {stats.avgLatency}<span class="text-lg">ms</span>
                </div>
            </div>
            <div
                class="border border-gray-700 p-6 bg-neutral-900 rounded-[48%_52%_68%_32%/42%_28%_72%_58%] text-center"
            >
                <div class="text-gray-400 text-xs mb-1">TOTAL REPORTS</div>
                <div class="text-3xl font-bold">{stats.totalReports}</div>
            </div>
            <div
                class="border border-gray-700 p-6 bg-neutral-900 rounded-[30%_70%_70%_30%/30%_30%_70%_70%] text-center"
            >
                <div class="text-gray-400 text-xs mb-1">INCIDENTS</div>
                <div class="text-3xl font-bold text-red-400">
                    {stats.incidents}
                </div>
            </div>
        </div>

        <!-- Charts -->
        <div class="grid md:grid-cols-2 gap-6 mb-8">
            <div
                class="border border-gray-700 p-6 bg-neutral-900 rounded-[60%_40%_30%_70%/60%_30%_70%_40%]"
            >
                <h2 class="text-gray-300 text-sm mb-4">UPTIME TREND</h2>
                <div
                    class="h-64 bg-linear-to-tr from-neutral-800 to-neutral-900 rounded-md flex items-center justify-center text-gray-300 border border-slate-700 hover:shadow-xl"
                >
                    Chart requires Recharts<br />({uptimeChartData.length} points)
                </div>
            </div>
            <div
                class="border border-gray-700 p-6 bg-neutral-900 rounded-[40%_60%_70%_30%/50%_60%_30%_60%]"
            >
                <h2 class="text-gray-300 text-sm mb-4">STATUS DISTRIBUTION</h2>
                <div
                    class="h-64 bg-linear-to-bl from-neutral-800 to-neutral-900 rounded-md flex items-center justify-center text-gray-300 border border-slate-700 hover:shadow-xl"
                >
                    Up: {totalUp} | Down: {totalDown} | Degraded: {totalDegraded}
                </div>
            </div>
        </div>

        <div
            class="border border-slate-700 p-6 bg-neutral-900 rounded-[48%_52%_68%_32%/42%_28%_72%_58%]"
        >
            <h2 class="text-gray-300 text-sm mb-4">RECENT REPORTS</h2>

            {#if !loading}
                <div class="text-center py-12 text-gray-300 animate-pulse">
                    LOADING DATA...
                </div>
            {:else}
                <div class="overflow-x-auto">
                    <table
                        class="w-full text-sm bg-black h-full rounded-t-md shadow-xl"
                    >
                        <thead class="border-b border-slate-700 text-gray-300">
                            <tr class="px-3">
                                <th class="text-left py-2 pl-4">TIMESTAMP</th>
                                <th class="text-left py-2">ENV</th>
                                <th class="text-left py-2">UPTIME_%</th>
                                <th class="text-left py-2">LATENCY</th>
                                <th class="text-left py-2">STATUS</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each allData.slice(0, 20) as report, idx (idx)}
                                <tr
                                    class="border-b border-neutral-800 hover:bg-neutral-950"
                                >
                                    <td class="py-2"
                                        >{new Date(
                                            report.timestamp,
                                        ).toLocaleString()}</td
                                    >
                                    <td class="text-gray-400 py-2"
                                        >{report.environment}</td
                                    >
                                    <td class="py-2"
                                        >{report.uptime_percent.toFixed(2)}%</td
                                    >
                                    <td class="py-2"
                                        >{report.average_latency_ms.toFixed(
                                            2,
                                        )}ms</td
                                    >
                                    <td class="py-2">
                                        <span
                                            class="bg-green-900 text-green-300 text-xs px-2 py-1 rounded mr-1"
                                            >{report.uptime_count}↑</span
                                        >
                                        <span
                                            class="bg-red-900 text-red-300 text-xs px-2 py-1 rounded mr-1"
                                            >{report.downtime_count}↓</span
                                        >
                                        <span
                                            class="bg-yellow-900 text-yellow-300 text-xs px-2 py-1 rounded"
                                            >{report.degraded_count}~</span
                                        >
                                    </td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            {/if}
        </div>
    </div>
</div>
