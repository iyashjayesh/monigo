const API_BASE = '/monigo/api/v1';

function getAuthHeaders(): Record<string, string> {
	if (typeof window === 'undefined') return {};
	const urlParams = new URLSearchParams(window.location.search);
	const apiKey = urlParams.get('api_key');
	const secret = urlParams.get('secret');
	const headers: Record<string, string> = { 'User-Agent': 'MoniGo-Admin/1.0' };
	if (apiKey) return headers;
	if (secret === 'monigo-admin-secret') return headers;
	headers['X-User-Role'] = 'admin';
	return headers;
}

function getUrl(path: string): string {
	const urlParams = new URLSearchParams(window.location.search);
	const apiKey = urlParams.get('api_key');
	const secret = urlParams.get('secret');
	const base = `${API_BASE}${path}`;
	if (apiKey) return `${base}${base.includes('?') ? '&' : '?'}api_key=${encodeURIComponent(apiKey)}`;
	if (secret === 'monigo-admin-secret') return `${base}${base.includes('?') ? '&' : '?'}secret=${encodeURIComponent(secret)}`;
	return base;
}

export async function fetchServiceInfo() {
	const res = await fetch(getUrl('/service-info'), { headers: getAuthHeaders() });
	if (!res.ok) throw new Error(`Fetch failed: ${res.status}`);
	return res.json();
}

export async function fetchMetrics() {
	const res = await fetch(getUrl('/metrics'), { headers: getAuthHeaders() });
	if (!res.ok) throw new Error(`Fetch failed: ${res.status}`);
	return res.json();
}

export async function fetchServiceMetrics(data: {
	field_name: string[];
	timerange: string;
	start_time: string;
	end_time: string;
}) {
	const res = await fetch(getUrl('/service-metrics'), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
		body: JSON.stringify(data)
	});
	if (!res.ok) throw new Error(`Fetch failed: ${res.status}`);
	return res.json();
}

export async function fetchGoRoutinesStats() {
	const res = await fetch(getUrl('/go-routines-stats'), { headers: getAuthHeaders() });
	if (!res.ok) throw new Error(`Fetch failed: ${res.status}`);
	return res.json();
}

export async function fetchReports(data: {
	topic: string;
	start_time: string;
	end_time: string;
	time_frame: string;
}) {
	const res = await fetch(getUrl('/reports'), {
		method: 'POST',
		headers: { 'Content-Type': 'application/json', ...getAuthHeaders() },
		body: JSON.stringify(data)
	});
	if (!res.ok) throw new Error(`Fetch failed: ${res.status}`);
	return res.json();
}

export async function fetchFunctionTrace() {
	const res = await fetch(getUrl('/function'), { headers: getAuthHeaders() });
	if (!res.ok) throw new Error(`Fetch failed: ${res.status}`);
	return res.json();
}

export async function fetchFunctionDetails(name: string, reportType = 'text') {
	const res = await fetch(getUrl(`/function-details?name=${encodeURIComponent(name)}&reportType=${reportType}`), {
		headers: getAuthHeaders()
	});
	if (!res.ok) throw new Error(`Fetch failed: ${res.status}`);
	return res.json();
}
