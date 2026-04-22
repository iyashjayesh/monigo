<script lang="ts">
	import { onMount } from 'svelte';
	import * as echarts from 'echarts';
	import { fetchReports } from '$lib/api/monigo.js';
	import {
		chartColors, baseChartOption, titleStyle, tooltipStyle, axisStyle, legendStyle, lineSeries
	} from '$lib/chart-theme.js';

	let reports = $state<Array<{ time: string; value: Record<string, number> }>>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let timeframe = $state('1h');
	let chartEl: HTMLDivElement;

	const timeRanges: Record<string, number> = {
		'5m': 5, '15m': 15, '30m': 30, '1h': 60, '6h': 360, '1d': 1440, '3d': 4320, '7d': 10080
	};

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
		const mins = timeRanges[timeframe] ?? 60;
		const start = new Date(now.getTime() - mins * 60000);
		fetchReports({
			topic: 'LoadStatistics',
			start_time: toLocalISOString(start),
			end_time: toLocalISOString(now),
			time_frame: timeframe
		})
			.then((data) => { reports = Array.isArray(data) ? data : []; })
			.catch((e) => { error = e.message; reports = []; })
			.finally(() => { loading = false; });
	}

	function renderChart() {
		if (!chartEl || reports.length === 0) return;

		const existing = echarts.getInstanceByDom(chartEl);
		if (existing) existing.dispose();

		const colors = chartColors();
		const axis = axisStyle();
		const chart = echarts.init(chartEl);

		const seriesData: Record<string, [string, number][]> = {};
		reports.forEach((dp) => {
			const t = new Date(dp.time).toLocaleTimeString();
			for (const k of Object.keys(dp.value || {})) {
				if (!seriesData[k]) seriesData[k] = [];
				seriesData[k].push([t, dp.value[k]]);
			}
		});

		chart.setOption({
			...baseChartOption(),
			title: titleStyle('LOAD REPORT'),
			tooltip: { ...tooltipStyle(), trigger: 'axis' },
			legend: { top: 0, right: 0, ...legendStyle() },
			grid: { top: 30, bottom: 20, left: 50, right: 16 },
			xAxis: { type: 'category', boundaryGap: false, ...axis.xAxis },
			yAxis: { type: 'value', ...axis.yAxis },
			series: Object.entries(seriesData).map(([name, data], i) =>
				lineSeries(name, data, colors[i % colors.length])
			),
		});
	}

	function handleTimeframeChange() {
		load();
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
		if (loading || reports.length === 0) return;
		requestAnimationFrame(() => renderChart());
	});
</script>

<svelte:head><title>Reports - MoniGo</title></svelte:head>

<div class="p-4 md:p-6 space-y-4">
	<div class="flex items-center justify-between">
		<div>
			<div class="hud-label mb-1">Analytics</div>
			<div class="hud-value-lg">Reports</div>
		</div>
		<div class="flex gap-2">
			<select bind:value={timeframe} class="hud-select" onchange={handleTimeframeChange}>
				<option value="5m">5m</option>
				<option value="15m">15m</option>
				<option value="30m">30m</option>
				<option value="1h">1h</option>
				<option value="6h">6h</option>
				<option value="1d">1d</option>
				<option value="3d">3d</option>
				<option value="7d">7d</option>
			</select>
			<button class="hud-button" onclick={() => load()} disabled={loading}>Refresh</button>
		</div>
	</div>

	<hr class="hud-divider" />

	{#if error}
		<div class="hud-error-panel p-4">
			<div class="hud-label mb-2 text-hud-error">Error</div>
			<div class="hud-value-sm">{error}</div>
		</div>
	{:else}
		<div class="hud-panel p-4">
			<div class="hud-label mb-3">Load Metrics Over Time</div>
			{#if loading}
				<div class="hud-skeleton h-56 w-full"></div>
			{:else if reports.length > 0}
				<div bind:this={chartEl} class="h-56 w-full"></div>
			{:else}
				<div class="hud-value-sm text-hud-text-dim">No report data available for this time range.</div>
			{/if}
		</div>
	{/if}
</div>
