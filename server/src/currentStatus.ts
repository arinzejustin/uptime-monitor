export const currentStatus = async (params: { domains: string[] }) => {
    const { domains } = params;
    const statusResults: Record<string, "Up" | "Down"> = {};
    const messages: string[] = [];

    for (const domain of domains) {
        let url = domain;

        if (!url.startsWith("http://") && !url.startsWith("https://")) {
            url = "https://" + url;
        }

        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 5000);

            const response = await fetch(url, {
                method: "HEAD",
                signal: controller.signal
            });

            clearTimeout(timeoutId);

            if (!response.ok) {
                statusResults[domain] = "Down";
                messages.push(`${domain} is down`);
                continue;
            }

            statusResults[domain] = "Up";
        } catch (error) {
            statusResults[domain] = "Down";
            messages.push(`${domain} is down`);
        }
    }

    return {
        statusResults,
        message: messages.length ? messages.join(", ") : ""
    };
};
