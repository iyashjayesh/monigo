<script lang="ts">
	import { onMount } from 'svelte';
	import * as echarts from 'echarts';
	import { fetchServiceInfo, fetchMetrics, fetchServiceMetrics } from '$lib/api/monigo.js';
	import {
		getTheme, chartColors, baseChartOption, titleStyle, tooltipStyle,
		axisStyle, legendStyle, barSeries, lineSeries, pieSeries
	} from '$lib/chart-theme.js';

	let serviceInfo = $state<{
		service_name: string;
		service_start_time: string;
		go_version: string;
		process_id: number;
	} | null>(null);
	let metrics = $state<{
		core_statistics: Record<string, unknown>;
		load_statistics: Record<string, unknown>;
		cpu_statistics: Record<string, unknown>;
		memory_statistics: Record<string, unknown>;
		health: { service_health: Record<string, unknown>; system_health: Record<string, unknown> };
	} | null>(null);
	let error = $state<string | null>(null);
	let loading = $state(true);
	let historyMetric = $state('heap');
	let historyTimeRange = $state('5m');
	let historyData = $state<Array<{ time: string; value: Record<string, number> }>>([]);
	let historyLoading = $state(false);
	let historyError = $state<string | null>(null);

	function loadData() {
		loading = true;
		error = null;
		Promise.all([fetchServiceInfo(), fetchMetrics()])
			.then(([info, m]) => { serviceInfo = info; metrics = m; })
			.catch((e) => (error = e.message))
			.finally(() => (loading = false));
	}

	function parsePercent(val: unknown): number {
		if (typeof val === 'number') return val;
		if (typeof val === 'string') return parseFloat(val.replace('%', '')) || 0;
		return 0;
	}

	function parseMemoryToMB(val: unknown): number {
		if (typeof val === 'number') return val / 1024;
		const s = String(val ?? '');
		const num = parseFloat(s) || 0;
		if (s.includes('GB')) return num * 1024;
		if (s.includes('MB')) return num;
		if (s.includes('KB')) return num / 1024;
		return num;
	}

	function toLocalISOString(date: Date) {
		const tzOffset = -date.getTimezoneOffset();
		const diff = tzOffset >= 0 ? '+' : '-';
		const pad = (n: number) => Math.floor(Math.abs(n)).toString().padStart(2, '0');
		return (
			date.getFullYear() + '-' + pad(date.getMonth() + 1) + '-' + pad(date.getDate()) +
			'T' + pad(date.getHours()) + ':' + pad(date.getMinutes()) + ':' + pad(date.getSeconds()) +
			'.000' + diff + pad(tzOffset / 60) + ':' + pad(tzOffset % 60)
		);
	}

	const timeRanges: Record<string, number> = {
		'5m': 5, '15m': 15, '30m': 30, '1h': 60, '6h': 360, '1d': 1440, '3d': 4320, '7d': 10080
	};

	const metricFields: Record<string, string[]> = {
		heap: ['heap_alloc', 'heap_sys', 'heap_inuse', 'heap_idle', 'heap_released'],
		stack: ['stack_inuse', 'stack_sys'],
		gc: ['pause_total_ns', 'num_gc', 'gc_cpu_fraction'],
		misc: ['m_span_inuse', 'm_span_sys', 'm_cache_inuse', 'm_cache_sys', 'buck_hash_sys', 'gc_sys', 'other_sys']
	};

	function loadHistoryChart() {
		historyLoading = true;
		historyError = null;
		const now = new Date();
		const mins = timeRanges[historyTimeRange] ?? 5;
		const start = new Date(now.getTime() - mins * 60000);
		fetchServiceMetrics({
			field_name: metricFields[historyMetric] ?? metricFields.heap,
			timerange: historyTimeRange,
			start_time: toLocalISOString(start),
			end_time: toLocalISOString(now)
		})
			.then((data) => { historyData = Array.isArray(data) ? data : []; })
			.catch((e) => { historyError = e.message; historyData = []; })
			.finally(() => { historyLoading = false; });
	}

	let loadChartEl: HTMLDivElement;
	let cpuChartEl: HTMLDivElement;
	let memChartEl: HTMLDivElement;
	let heapChartEl: HTMLDivElement;
	let historyChartEl: HTMLDivElement;

	function renderAllCharts() {
		renderMetricCharts();
		renderHistoryChart();
	}

	function renderMetricCharts() {
		if (!metrics || !loadChartEl || !cpuChartEl || !memChartEl || !heapChartEl) return;
		const load = metrics.load_statistics;
		const cpu = metrics.cpu_statistics;
		const mem = metrics.memory_statistics;
		if (!load || !cpu || !mem) return;

		const t = getTheme();
		const axis = axisStyle();
		const base = baseChartOption();
		const grid = { top: 30, bottom: 30, left: 50, right: 16 };

		// Load Statistics
		let inst = echarts.getInstanceByDom(loadChartEl);
		if (inst) inst.dispose();
		const loadChart = echarts.init(loadChartEl);
		loadChart.setOption({
			...base,
			title: titleStyle('LOAD STATISTICS'),
			tooltip: { ...tooltipStyle(), trigger: 'axis' },
			grid,
			xAxis: { type: 'category', data: ['SVC CPU', 'SYS CPU', 'TOTAL', 'SVC MEM', 'SYS MEM'], ...axis.xAxis },
			yAxis: { type: 'value', max: 100, ...axis.yAxis },
			series: [barSeries(
				[
					parsePercent(load.service_cpu_load),
					parsePercent(load.system_cpu_load),
					parsePercent(load.total_cpu_load),
					parsePercent(load.service_memory_load),
					parsePercent(load.system_memory_load)
				],
				(v) => v > 90 ? t.error : v > 70 ? t.warning : t.cyan
			)],
		});

		// CPU Distribution
		inst = echarts.getInstanceByDom(cpuChartEl);
		if (inst) inst.dispose();
		const cpuChart = echarts.init(cpuChartEl);
		cpuChart.setOption({
			...base,
			title: titleStyle('CPU DISTRIBUTION'),
			tooltip: { ...tooltipStyle(), trigger: 'item' },
			legend: { bottom: 0, ...legendStyle() },
			series: [pieSeries([
				{ value: cpu.cores_used_by_service as number, name: 'SERVICE', color: t.cyan },
				{ value: cpu.cores_used_by_system as number, name: 'SYSTEM', color: t.purple },
				{ value: cpu.total_cores as number, name: 'TOTAL', color: t.textDim },
			])],
		});

		// Memory Distribution
		inst = echarts.getInstanceByDom(memChartEl);
		if (inst) inst.dispose();
		const memChart = echarts.init(memChartEl);
		memChart.setOption({
			...base,
			title: titleStyle('MEMORY DISTRIBUTION'),
			tooltip: { ...tooltipStyle(), trigger: 'item', formatter: '{b}: {c} MB ({d}%)' },
			legend: { bottom: 0, ...legendStyle() },
			series: [pieSeries([
				{ value: parseMemoryToMB(mem.memory_used_by_service), name: 'SERVICE', color: t.cyan },
				{ value: parseMemoryToMB(mem.memory_used_by_system), name: 'SYSTEM', color: t.warning },
				{ value: parseMemoryToMB(mem.available_memory), name: 'AVAILABLE', color: t.textDim },
			])],
		});

		// Heap Memory
		const heapValues: number[] = [];
		for (const r of (mem.mem_stats_records as Array<{ record_name: string; record_value: unknown }>) || []) {
			if (['HeapAlloc', 'HeapSys', 'HeapIdle', 'HeapInuse', 'HeapReleased'].includes(r.record_name)) {
				heapValues.push(Number(r.record_value) || 0);
			}
		}

		inst = echarts.getInstanceByDom(heapChartEl);
		if (inst) inst.dispose();
		const heapChart = echarts.init(heapChartEl);
		heapChart.setOption({
			...base,
			title: titleStyle('HEAP MEMORY'),
			tooltip: tooltipStyle(),
			grid,
			xAxis: { type: 'category', data: ['ALLOC', 'SYS', 'IDLE', 'INUSE', 'RELEASED'], ...axis.xAxis },
			yAxis: { type: 'value', name: 'MB', nameTextStyle: { color: t.textDim, fontSize: 8 }, ...axis.yAxis },
			series: [barSeries(heapValues)],
		});
	}

	function renderHistoryChart() {
		if (!historyChartEl || historyData.length === 0) return;

		const existing = echarts.getInstanceByDom(historyChartEl);
		if (existing) existing.dispose();

		const colors = chartColors();
		const axis = axisStyle();
		const chart = echarts.init(historyChartEl);

		const seriesData: Record<string, [string, number][]> = {};
		historyData.forEach((dp) => {
			const timeLabel = new Date(dp.time).toLocaleTimeString();
			for (const key of Object.keys(dp.value || {})) {
				if (!seriesData[key]) seriesData[key] = [];
				seriesData[key].push([timeLabel, dp.value[key]]);
			}
		});

		chart.setOption({
			...baseChartOption(),
			title: titleStyle('HISTORY'),
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

	function handleHistoryChange() {
		loadHistoryChart();
	}

	onMount(() => {
		loadData();
		loadHistoryChart();

		const onResize = () => {
			[loadChartEl, cpuChartEl, memChartEl, heapChartEl, historyChartEl].forEach((el) => {
				if (el) {
					const inst = echarts.getInstanceByDom(el);
					if (inst) inst.resize();
				}
			});
		};

		const onThemeChange = () => {
			renderAllCharts();
		};

		window.addEventListener('resize', onResize);
		window.addEventListener('theme-change', onThemeChange);
		return () => {
			window.removeEventListener('resize', onResize);
			window.removeEventListener('theme-change', onThemeChange);
			[loadChartEl, cpuChartEl, memChartEl, heapChartEl, historyChartEl].forEach((el) => {
				if (el) {
					const inst = echarts.getInstanceByDom(el);
					if (inst) inst.dispose();
				}
			});
		};
	});

	$effect(() => {
		if (!metrics || loading) return;
		requestAnimationFrame(() => renderMetricCharts());
	});

	$effect(() => {
		if (historyLoading || historyData.length === 0) return;
		requestAnimationFrame(() => renderHistoryChart());
	});
</script>

<svelte:head><title>MoniGo Dashboard</title></svelte:head>

<div class="p-4 md:p-6 space-y-4">
	<div class="flex items-center justify-between">
		<div>
			<div class="hud-label mb-1">System Monitor</div>
			<div class="hud-value-lg">Dashboard</div>
		</div>
		<button class="hud-button" onclick={loadData} disabled={loading}>Refresh</button>
	</div>

	<hr class="hud-divider" />

	{#if error}
		<div class="hud-error-panel p-4">
			<div class="hud-label mb-2 text-hud-error">Error</div>
			<div class="hud-value-sm">{error}</div>
			<div class="hud-value-sm mt-1 text-hud-text-dim">
				Ensure the MoniGo backend is running and the dashboard is served from it.
			</div>
		</div>
	{/if}

	{#if loading}
		<!-- Service Info skeleton -->
		<div class="grid gap-3 grid-cols-2 md:grid-cols-4">
			{#each ['Service', 'Go Version', 'Process ID', 'Health'] as label}
				<div class="hud-panel p-4">
					<div class="hud-label mb-2">{label}</div>
					<div class="hud-skeleton h-5 w-20"></div>
				</div>
			{/each}
		</div>
		<!-- Metrics skeleton -->
		<div class="grid gap-3 grid-cols-2 md:grid-cols-3 lg:grid-cols-6">
			{#each ['Goroutines', 'Load', 'Cores', 'Memory', 'CPU Usage', 'Uptime'] as label}
				<div class="hud-panel p-4">
					<div class="hud-label mb-2">{label}</div>
					<div class="hud-skeleton h-6 w-16"></div>
				</div>
			{/each}
		</div>
		<!-- Charts skeleton -->
		<div class="grid gap-3 md:grid-cols-2">
			{#each Array(4) as _}
				<div class="hud-panel p-4">
					<div class="hud-skeleton h-56 w-full"></div>
				</div>
			{/each}
		</div>
		<!-- History skeleton -->
		<div class="hud-panel p-4">
			<div class="flex items-center justify-between mb-3">
				<div class="hud-skeleton h-3 w-24"></div>
				<div class="flex gap-2">
					<div class="hud-skeleton h-6 w-16"></div>
					<div class="hud-skeleton h-6 w-12"></div>
				</div>
			</div>
			<div class="hud-skeleton h-56 w-full"></div>
		</div>
	{:else if metrics}
		<!-- Service Info -->
		<div class="grid gap-3 grid-cols-2 md:grid-cols-4">
			{#if serviceInfo}
				<div class="hud-panel p-4">
					<div class="hud-label mb-2">Service</div>
					<div class="hud-value-md text-hud-text-bright">{serviceInfo.service_name}</div>
				</div>
				<div class="hud-panel p-4">
					<div class="hud-label mb-2">Go Version</div>
					<div class="hud-value-md text-hud-text-bright">{serviceInfo.go_version}</div>
				</div>
				<div class="hud-panel p-4">
					<div class="hud-label mb-2">Process ID</div>
					<div class="hud-value-md text-hud-cyan">{serviceInfo.process_id}</div>
				</div>
			{/if}
			<div class="hud-panel p-4">
				<div class="hud-label mb-2">Health</div>
				<div class="hud-value-lg text-hud-success">
					{metrics.health?.service_health?.percent ?? 0}%
				</div>
				<div class="hud-value-sm mt-1 text-hud-text-dim">
					{metrics.health?.service_health?.message ?? '-'}
				</div>
			</div>
		</div>

		<!-- Metrics -->
		<div class="grid gap-3 grid-cols-2 md:grid-cols-3 lg:grid-cols-6">
			<div class="hud-panel p-4">
				<div class="hud-label mb-2">Goroutines</div>
				<div class="hud-value-lg text-hud-cyan">{metrics.core_statistics?.goroutines ?? '-'}</div>
			</div>
			<div class="hud-panel p-4">
				<div class="hud-label mb-2">Load</div>
				<div class="hud-value-lg text-hud-text-bright">{metrics.load_statistics?.overall_load_of_service ?? '-'}</div>
			</div>
			<div class="hud-panel p-4">
				<div class="hud-label mb-2">Cores</div>
				<div class="hud-value-lg text-hud-purple">
					{metrics.cpu_statistics?.cores_used_by_service ?? '-'}<span class="hud-value-sm text-hud-text-dim"> / {metrics.cpu_statistics?.total_cores ?? '-'}</span>
				</div>
			</div>
			<div class="hud-panel p-4">
				<div class="hud-label mb-2">Memory</div>
				<div class="hud-value-md text-hud-text-bright">{metrics.memory_statistics?.memory_used_by_service ?? '-'}</div>
			</div>
			<div class="hud-panel p-4">
				<div class="hud-label mb-2">CPU Usage</div>
				<div class="hud-value-md text-hud-warning">{metrics.cpu_statistics?.cores_used_by_service_in_percent ?? '-'}</div>
			</div>
			<div class="hud-panel p-4">
				<div class="hud-label mb-2">Uptime</div>
				<div class="hud-value-md text-hud-success">{metrics.core_statistics?.uptime ?? '-'}</div>
			</div>
		</div>

		<!-- Charts -->
		<div class="grid gap-3 md:grid-cols-2">
			<div class="hud-panel p-4">
				<div bind:this={loadChartEl} class="h-56 w-full"></div>
			</div>
			<div class="hud-panel p-4">
				<div bind:this={cpuChartEl} class="h-56 w-full"></div>
			</div>
			<div class="hud-panel p-4">
				<div bind:this={memChartEl} class="h-56 w-full"></div>
			</div>
			<div class="hud-panel p-4">
				<div bind:this={heapChartEl} class="h-56 w-full"></div>
			</div>
		</div>

		<!-- History -->
		<div class="hud-panel p-4">
			<div class="flex items-center justify-between mb-3">
				<div class="hud-label">History Statistics</div>
				<div class="flex gap-2">
					<select bind:value={historyMetric} class="hud-select" onchange={handleHistoryChange}>
						<option value="heap">Heap</option>
						<option value="stack">Stack</option>
						<option value="gc">GC</option>
						<option value="misc">Misc</option>
					</select>
					<select bind:value={historyTimeRange} class="hud-select" onchange={handleHistoryChange}>
						<option value="5m">5m</option>
						<option value="15m">15m</option>
						<option value="30m">30m</option>
						<option value="1h">1h</option>
						<option value="6h">6h</option>
						<option value="1d">1d</option>
						<option value="3d">3d</option>
						<option value="7d">7d</option>
					</select>
				</div>
			</div>
			{#if historyLoading}
				<div class="hud-skeleton h-56 w-full"></div>
			{:else if historyError}
				<div class="hud-value-sm text-hud-error">{historyError}</div>
			{:else if historyData.length > 0}
				<div bind:this={historyChartEl} class="h-56 w-full"></div>
			{:else}
				<div class="hud-value-sm text-hud-text-dim">No history data available for this time range.</div>
			{/if}
		</div>
	{/if}
</div>
