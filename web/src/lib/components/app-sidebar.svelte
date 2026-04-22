<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { HouseIcon, CalendarIcon, SearchIcon, SettingsIcon } from 'lucide-svelte';
	import monigoLogo from '$lib/assets/monigo-icon.png';
    import * as Card from "$lib/components/ui/card/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
	const items = [
		{
			title: 'Dashboard',
			url: '/',
			icon: HouseIcon
		},
        {
			title: 'Function Metrics',
			url: '/function-metrics',
			icon: SearchIcon
		},
		{
			title: 'Go Routines Stats',
			url: '/go-routines-stats',
			icon: CalendarIcon
		},
        {
            title: 'Reports',
            url: '/reports',
            icon: SettingsIcon
        }
	];
</script>

<Sidebar.Root>
	<Sidebar.Header />
	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupLabel class="flex items-center justify-center gap-2 text-lg font-bold mb-4">
				<img src={monigoLogo} alt="MoniGo" class="h-auto w-1/2" />
			</Sidebar.GroupLabel>
			<Sidebar.Menu>
				{#each items as item (item.title)}
					<Sidebar.MenuItem>
						<Sidebar.MenuButton>
							{#snippet child({ props })}
								<a href={item.url} {...props}>
									<item.icon />
									<span>{item.title}</span>
								</a>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				{/each}
			</Sidebar.Menu>
		</Sidebar.Group>
	</Sidebar.Content>
	<Sidebar.Footer>
		<Card.Root>
            <Card.Content>
                <p class="text-sm text-muted-foreground">
                    Found a bug or have a feature request? Please report it on
                    GitHub. If you find this project useful, we would appreciate
                    a star on GitHub.
                </p>
                <Button 
                    variant="outline"
                    size="sm"
                    class="mt-4"
                >
                    <a href="https://github.com/iyashjayesh/monigo" target="_blank">
                        Support the project
                    </a>
                </Button>
            </Card.Content>
		</Card.Root>
	</Sidebar.Footer>
</Sidebar.Root>
