<script lang="ts">
	import { onMount } from 'svelte';
	import * as echarts from 'echarts';
	import { fetchGoRoutinesStats, fetchServiceMetrics } from '$lib/api/monigo.js';
	import {
		getTheme, baseChartOption, titleStyle, tooltipStyle, axisStyle, lineSeries
	} from '$lib/chart-theme.js';

	let stats = $state<{ number_of_goroutines: number; stack_view: string[] } | null>(null);
	let historyData = $state<Array<{ time: string; value: Record<string, number> }>>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let chartEl: HTMLDivElement;

	function toLocalISOString(d: Date) {
		const tz = -d.getTimezoneOffset();
		const pad = (n: number) => Math.floor(Math.abs(n)).toString().padStart(2, '0');
		return (
			d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
			'T' + pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds()) +
			'.000' + (tz >= 0 ? '+' : '-') + pad(tz / 60) + ':' + pad(tz % 60)
		);
	}

	function load() {
		loading = true;
		error = null;
		const now = new Date();
		const start = new Date(now.getTime() - 60 * 60000);
		Promise.all([
			fetchGoRoutinesStats(),
			fetchServiceMetrics({
				field_name: ['goroutines'],
				timerange: '1h',
				start_time: toLocalISOString(start),
				end_time: toLocalISOString(now)
			})
		])
			.then(([s, h]) => { stats = s; historyData = Array.isArray(h) ? h : []; })
			.catch((e) => (error = e.message))
			.finally(() => (loading = false));
	}

	function renderChart() {
		if (!chartEl || historyData.length === 0) return;

		const existing = echarts.getInstanceByDom(chartEl);
		if (existing) existing.dispose();

		const t = getTheme();
		const axis = axisStyle();
		const chart = echarts.init(chartEl);

		const seriesData: Record<string, [string, number][]> = {};
		historyData.forEach((dp) => {
			const time = new Date(dp.time).toLocaleTimeString();
			for (const k of Object.keys(dp.value || {})) {
				if (!seriesData[k]) seriesData[k] = [];
				seriesData[k].push([time, dp.value[k]]);
			}
		});

		chart.setOption({
			...baseChartOption(),
			title: titleStyle('GOROUTINES OVER TIME'),
			tooltip: { ...tooltipStyle(), trigger: 'axis' },
			grid: { top: 30, bottom: 20, left: 50, right: 16 },
			xAxis: { type: 'category', boundaryGap: false, ...axis.xAxis },
			yAxis: { type: 'value', ...axis.yAxis },
			series: Object.entries(seriesData).map(([name, data]) =>
				lineSeries(name, data, t.cyan)
			),
		});
	}

	onMount(() => {
		load();

		const onResize = () => {
			if (chartEl) {
				const inst = echarts.getInstanceByDom(chartEl);
				if (inst) inst.resize();
			}
		};

		const onThemeChange = () => renderChart();

		window.addEventListener('resize', onResize);
		window.addEventListener('theme-change', onThemeChange);
		return () => {
			window.removeEventListener('resize', onResize);
			window.removeEventListener('theme-change', onThemeChange);
			if (chartEl) {
				const inst = echarts.getInstanceByDom(chartEl);
				if (inst) inst.dispose();
			}
		};
	});

	$effect(() => {
		if (loading || historyData.length === 0) return;
		requestAnimationFrame(() => renderChart());
	});
</script>

<svelte:head><title>Go Routines Stats - MoniGo</title></svelte:head>

<div class="p-4 md:p-6 space-y-4">
	<div class="flex items-center justify-between">
		<div>
			<div class="hud-label mb-1">Concurrency</div>
			<div class="hud-value-lg">Goroutines</div>
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
	{:else if stats}
		<div class="hud-panel p-4 flex items-center justify-between">
			<div class="hud-label">Active Goroutines</div>
			<div class="hud-value-xl text-hud-cyan">{stats.number_of_goroutines}</div>
		</div>

		<div class="hud-panel p-4">
			{#if historyData.length > 0}
				<div bind:this={chartEl} class="h-56 w-full"></div>
			{:else}
				<div class="hud-value-sm text-hud-text-dim">No history data available.</div>
			{/if}
		</div>

		{#if stats.stack_view?.length}
			<div class="hud-panel p-4">
				<div class="hud-label mb-3">Stack Traces</div>
				<div class="space-y-2">
					{#each stats.stack_view as trace, i}
						<details class="hud-details">
							<summary>Goroutine {i + 1}</summary>
							<pre>{trace}</pre>
						</details>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>
