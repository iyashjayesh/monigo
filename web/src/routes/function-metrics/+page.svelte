<script lang="ts">
	import { onMount } from 'svelte';
	import { fetchFunctionTrace, fetchFunctionDetails } from '$lib/api/monigo.js';

	let functions = $state<Record<string, { function_last_ran_at: string }>>({});
	let selectedFunc = $state<string | null>(null);
	let funcDetails = $state<string | null>(null);
	let loading = $state(true);
	let detailsLoading = $state(false);
	let error = $state<string | null>(null);

	function load() {
		loading = true;
		error = null;
		fetchFunctionTrace()
			.then((data) => { functions = data; })
			.catch((e) => (error = e.message))
			.finally(() => (loading = false));
	}

	function viewDetails(name: string) {
		selectedFunc = name;
		funcDetails = null;
		detailsLoading = true;
		fetchFunctionDetails(name)
			.then((data) => (funcDetails = typeof data === 'string' ? data : JSON.stringify(data, null, 2)))
			.catch((e) => (funcDetails = `Error: ${e.message}`))
			.finally(() => (detailsLoading = false));
	}

	onMount(load);
</script>

<svelte:head><title>Function Metrics - MoniGo</title></svelte:head>

<div class="p-4 md:p-6 space-y-4">
	<div class="flex items-center justify-between">
		<div>
			<div class="hud-label mb-1">Performance</div>
			<div class="hud-value-lg">Function Metrics</div>
		</div>
		<button class="hud-button" onclick={load} disabled={loading}>Refresh</button>
	</div>

	<hr class="hud-divider" />

	{#if error}
		<div class="hud-error-panel p-4">
			<div class="hud-label mb-2 text-hud-error">Error</div>
			<div class="hud-value-sm">{error}</div>
		</div>
	{:else if loading}
		<div class="hud-panel p-4">
			<div class="hud-skeleton h-24 w-full"></div>
		</div>
	{:else if Object.keys(functions).length === 0}
		<div class="hud-panel p-4">
			<div class="hud-value-sm text-hud-text-dim">
				No function metrics available. Instrument functions with monigo.TraceFunction() to see metrics.
			</div>
		</div>
	{:else}
		<div class="flex items-center gap-2 mb-2">
			<span class="hud-label">Total Functions</span>
			<span class="hud-value-md text-hud-cyan">{Object.keys(functions).length}</span>
		</div>

		<div class="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
			{#each Object.entries(functions) as [name, data]}
				<div class="hud-panel p-4">
					<div class="hud-value-sm mb-2 truncate text-hud-text-bright" title={name}>{name}</div>
					<div class="hud-label mb-3">Last ran: {data.function_last_ran_at}</div>
					<button class="hud-button" onclick={() => viewDetails(name)}>Details</button>
				</div>
			{/each}
		</div>

		{#if selectedFunc}
			<div class="hud-panel p-4 mt-4">
				<div class="flex items-center justify-between mb-3">
					<div>
						<div class="hud-label mb-1">Details</div>
						<div class="hud-value-sm text-hud-text-bright">{selectedFunc}</div>
					</div>
					<button class="hud-button" onclick={() => (selectedFunc = null)}>Close</button>
				</div>
				{#if detailsLoading}
					<div class="hud-skeleton h-32 w-full"></div>
				{:else if funcDetails}
					<pre class="hud-code max-h-96">{funcDetails}</pre>
				{/if}
			</div>
		{/if}
	{/if}
</div>
