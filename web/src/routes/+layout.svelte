<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { page } from '$app/state';
	import {
		LayoutDashboard,
		Zap,
		GitBranch,
		FileBarChart,
		Github,
		Menu,
		X,
		Sun,
		Moon
	} from 'lucide-svelte';
	import { onMount } from 'svelte';

	let { children } = $props();
	let sidebarOpen = $state(false);
	let lightMode = $state(false);

	const navItems = [
		{ title: 'Dashboard', url: '/', icon: LayoutDashboard },
		{ title: 'Functions', url: '/function-metrics', icon: Zap },
		{ title: 'Goroutines', url: '/go-routines-stats', icon: GitBranch },
		{ title: 'Reports', url: '/reports', icon: FileBarChart }
	];

	function toggleTheme() {
		lightMode = !lightMode;
		document.documentElement.classList.toggle('light', lightMode);
		localStorage.setItem('monigo-theme', lightMode ? 'light' : 'dark');
		window.dispatchEvent(new CustomEvent('theme-change'));
	}

	onMount(() => {
		const saved = localStorage.getItem('monigo-theme');
		if (saved === 'light') {
			lightMode = true;
			document.documentElement.classList.add('light');
		}
	});
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

<div class="flex min-h-screen bg-hud-bg">
	<!-- Mobile menu toggle -->
	<button
		class="hud-button fixed top-3 left-3 z-50 p-2 md:hidden"
		onclick={() => (sidebarOpen = !sidebarOpen)}
	>
		{#if sidebarOpen}<X size={14} />{:else}<Menu size={14} />{/if}
	</button>

	<!-- Sidebar -->
	<aside
		class="hud-sidebar fixed inset-y-0 left-0 z-40 flex w-[180px] flex-col bg-hud-bg border-r border-hud-line transition-transform duration-200 md:relative md:translate-x-0"
		class:translate-x-0={sidebarOpen}
		class:-translate-x-full={!sidebarOpen}
	>
		<!-- Logo -->
		<div class="px-4 py-5 border-b border-hud-line">
			<span class="text-[16px] font-medium tracking-[0.04em] text-hud-text-bright">MoniGo</span>
		</div>

		<!-- Nav -->
		<nav class="flex-1 py-4">
			<div class="hud-label px-4 mb-3">Navigation</div>
			{#each navItems as item}
				<a
					href={item.url}
					class="hud-nav-item"
					class:active={page.url.pathname === item.url}
					onclick={() => (sidebarOpen = false)}
				>
					<item.icon size={12} />
					<span>{item.title}</span>
				</a>
			{/each}
		</nav>

		<!-- Footer -->
		<div class="px-4 py-4 border-t border-hud-line">
			<div class="flex items-center justify-between mb-3">
				<a
					href="https://github.com/iyashjayesh/monigo"
					target="_blank"
					rel="noopener noreferrer"
					class="hud-nav-item !px-0 !border-l-0 !py-0"
				>
					<Github size={12} />
					<span>Source</span>
				</a>
				<button class="theme-toggle" onclick={toggleTheme} title={lightMode ? 'Switch to dark mode' : 'Switch to light mode'}>
					{#if lightMode}<Moon size={12} />{:else}<Sun size={12} />{/if}
				</button>
			</div>
			<div class="hud-label">MoniGo Monitor</div>
		</div>
	</aside>

	<!-- Overlay for mobile -->
	{#if sidebarOpen}
		<button
			class="fixed inset-0 z-30 bg-black/50 md:hidden"
			onclick={() => (sidebarOpen = false)}
			aria-label="Close menu"
		></button>
	{/if}

	<!-- Main content -->
	<main class="flex-1 min-w-0 overflow-auto">
		{@render children()}
	</main>
</div>
