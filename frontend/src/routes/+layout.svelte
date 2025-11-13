<script lang="ts">
	import { mode, ModeWatcher } from "mode-watcher";
	import { Toaster } from "$lib/components/ui/sonner/index.js";
	import "../app.css";

	let { children } = $props();
</script>

<ModeWatcher />
<main data-sveltekit-preload-code>
	{@render children()}
</main>
<Toaster richColors closeButton theme={mode.current} position="top-right" />
