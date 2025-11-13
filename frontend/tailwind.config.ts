import type { Config } from "tailwindcss";

const config: Config = {
    content: ["./src/**/*.{html,js,svelte,ts,md,json}"],
    theme: {
        extend: {
            fontFamily: {
                open: ['"Open Sans"', "sans-serif"],
                code: ["Source Code Pro", "monospace"]
            },
        },
    },
    important: true,
    darkMode: 'class'
}

export default config;
