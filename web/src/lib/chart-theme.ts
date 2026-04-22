import * as echarts from 'echarts';

export function isLightMode(): boolean {
	if (typeof document === 'undefined') return false;
	return document.documentElement.classList.contains('light');
}

export function getTheme() {
	const light = isLightMode();
	return {
		bg: light ? '#ffffff' : '#0d1012',
		text: light ? '#5a6a72' : '#7a9ba8',
		textDim: light ? '#8a9aa2' : '#4a5a62',
		textBright: light ? '#1a2a32' : '#c8dce4',
		line: light ? 'rgba(0,0,0,0.08)' : 'rgba(122, 155, 168, 0.15)',
		gridLine: light ? 'rgba(0,0,0,0.05)' : 'rgba(122, 155, 168, 0.08)',
		cyan: light ? '#2a7aa8' : '#5a9ab8',
		purple: light ? '#6a4aa0' : '#8a6ab8',
		success: light ? '#2d7a4a' : '#4a9868',
		warning: light ? '#9a7a28' : '#b89848',
		error: light ? '#c03030' : '#b84848',
		blue: light ? '#3a5a98' : '#5a7ab8',
		tooltipBg: light ? '#ffffff' : '#0d1012',
		tooltipBorder: light ? 'rgba(0,0,0,0.1)' : 'rgba(122, 155, 168, 0.15)',
	};
}

export function chartColors() {
	const t = getTheme();
	return [t.cyan, t.purple, t.success, t.warning, t.blue];
}

export function baseChartOption() {
	const t = getTheme();
	return {
		backgroundColor: 'transparent',
		textStyle: {
			fontFamily: "'JetBrains Mono', monospace",
			color: t.text,
			fontSize: 10,
		},
		animationDuration: 800,
		animationEasing: 'cubicOut' as const,
	};
}

export function titleStyle(text: string) {
	const t = getTheme();
	return {
		text,
		textStyle: {
			color: t.textDim,
			fontSize: 8,
			fontWeight: 400 as const,
			fontFamily: "'JetBrains Mono', monospace",
			letterSpacing: 2,
		},
		top: 0,
		left: 0,
	};
}

export function tooltipStyle() {
	const t = getTheme();
	const light = isLightMode();
	return {
		backgroundColor: t.tooltipBg,
		borderColor: t.tooltipBorder,
		borderWidth: 0.5,
		textStyle: {
			color: t.text,
			fontSize: 10,
			fontFamily: "'JetBrains Mono', monospace",
		},
		extraCssText: light
			? 'border-radius: 6px; box-shadow: 0 4px 16px rgba(0,0,0,0.08);'
			: 'border-radius: 6px; box-shadow: 0 4px 16px rgba(0,0,0,0.3);',
	};
}

export function axisStyle() {
	const t = getTheme();
	return {
		xAxis: {
			axisLabel: { color: t.textDim, fontSize: 8 },
			axisLine: { lineStyle: { color: t.line } },
			axisTick: { show: false },
		},
		yAxis: {
			axisLabel: { color: t.textDim, fontSize: 8 },
			splitLine: { lineStyle: { color: t.gridLine, type: 'dashed' as const } },
			axisLine: { show: false },
		},
	};
}

export function legendStyle() {
	const t = getTheme();
	return {
		textStyle: {
			color: t.textDim,
			fontSize: 8,
			fontFamily: "'JetBrains Mono', monospace",
		},
	};
}

export function barSeries(data: number[], colorFn?: (val: number) => string) {
	const t = getTheme();
	const defaultColor = t.cyan;
	return {
		type: 'bar' as const,
		data,
		barWidth: '45%',
		itemStyle: {
			borderRadius: [4, 4, 0, 0],
			color: colorFn
				? (p: { value: number }) => {
						const c = colorFn(p.value);
						return new echarts.graphic.LinearGradient(0, 0, 0, 1, [
							{ offset: 0, color: c },
							{ offset: 1, color: c + '60' },
						]);
					}
				: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
						{ offset: 0, color: defaultColor },
						{ offset: 1, color: defaultColor + '60' },
					]),
		},
		emphasis: {
			itemStyle: {
				opacity: 0.85,
			},
		},
	};
}

export function lineSeries(name: string, data: unknown[], color: string, showArea = true) {
	const light = isLightMode();
	return {
		name,
		type: 'line' as const,
		data,
		smooth: true,
		lineStyle: {
			width: 2,
			color,
		},
		itemStyle: { color },
		showSymbol: false,
		...(showArea
			? {
					areaStyle: {
						color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
							{ offset: 0, color: color + (light ? '18' : '25') },
							{ offset: 1, color: color + '00' },
						]),
					},
				}
			: {}),
	};
}

export function pieSeries(data: { value: number; name: string; color: string }[]) {
	return {
		type: 'pie' as const,
		radius: ['38%', '58%'],
		center: ['50%', '48%'],
		label: { show: false },
		padAngle: 2,
		itemStyle: {
			borderRadius: 6,
		},
		emphasis: {
			itemStyle: {
				opacity: 0.85,
			},
		},
		data: data.map((d) => ({
			value: d.value,
			name: d.name,
			itemStyle: {
				color: d.color,
			},
		})),
	};
}
