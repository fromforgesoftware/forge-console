<template>
	<CommandPaletteDialog v-model:open="commandPaletteOpen">
		<CommandPalette
			v-slot="{ search, recentValues, removeSelection, clearSelections }"
			v-model:search="searchText"
			:filter="filter"
		>
			<CommandPaletteInput placeholder="Type a command or search..." />

			<CommandPaletteList>
				<CommandPaletteEmpty />

				<!-- Empty state: recents + all shortcuts -->
				<template v-if="!search">
					<CommandPaletteGroup v-if="recentItems(recentValues).length">
						<template #heading>
							<span class="flex-1">Recent</span>
							<button
								type="button"
								class="rounded px-1 py-0.5 text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
								@click.stop="clearSelections()"
							>
								Clear
							</button>
						</template>
						<CommandPaletteItem
							v-for="item in recentItems(recentValues)"
							:key="item.id"
							:value="item.id"
							@select="run(item)"
						>
							<component :is="item.icon" class="size-4 shrink-0 text-muted-foreground" />
							<span v-if="item.parent" class="text-muted-foreground">{{ item.parent }} /</span>
							<span class="flex-1">{{ item.title }}</span>
							<button
								type="button"
								class="shrink-0 rounded p-0.5 text-muted-foreground hover:text-foreground cursor-pointer transition-colors"
								@click.stop="removeSelection(item.id)"
							>
								<X class="size-3" />
							</button>
						</CommandPaletteItem>
					</CommandPaletteGroup>

					<CommandPaletteSeparator v-if="recentItems(recentValues).length" />

					<CommandPaletteGroup heading="Shortcuts">
						<CommandPaletteItem
							v-for="cmd in navCommands"
							:key="cmd.id"
							:value="cmd.id"
							:keywords="cmd.keywords"
							@select="run(cmd)"
						>
							<component :is="cmd.icon" class="size-4 shrink-0 text-muted-foreground" />
							<span v-if="cmd.parent" class="text-muted-foreground">{{ cmd.parent }} /</span>
							<span>{{ cmd.title }}</span>
						</CommandPaletteItem>
					</CommandPaletteGroup>
				</template>

				<!-- Search state: pages + actions -->
				<template v-else>
					<CommandPaletteGroup heading="Pages">
						<CommandPaletteItem
							v-for="cmd in navCommands"
							:key="cmd.id"
							:value="cmd.id"
							:keywords="cmd.keywords"
							@select="run(cmd)"
						>
							<component :is="cmd.icon" class="size-4 shrink-0 text-muted-foreground" />
							<span v-if="cmd.parent" class="text-muted-foreground">{{ cmd.parent }} /</span>
							<span>{{ cmd.title }}</span>
						</CommandPaletteItem>
					</CommandPaletteGroup>

					<CommandPaletteGroup heading="Actions">
						<CommandPaletteItem
							v-for="cmd in actionCommands"
							:key="cmd.id"
							:value="cmd.id"
							:keywords="cmd.keywords"
							@select="run(cmd)"
						>
							<component :is="cmd.icon" class="size-4 shrink-0 text-muted-foreground" />
							<span>{{ cmd.title }}</span>
						</CommandPaletteItem>
					</CommandPaletteGroup>
				</template>
			</CommandPaletteList>

			<CommandPaletteFooter />
		</CommandPalette>
	</CommandPaletteDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
	CommandPaletteDialog,
	CommandPalette,
	CommandPaletteInput,
	CommandPaletteList,
	CommandPaletteEmpty,
	CommandPaletteGroup,
	CommandPaletteItem,
	CommandPaletteSeparator,
	CommandPaletteFooter,
} from '@fromforgesoftware/vue-kit';
import { X } from '@lucide/vue';
import { useAppsStore } from '@/app/features/console/stores/apps';
import { useCommandPalette } from './useCommandPalette';
import { buildCommands } from './commands';
import type { Command } from './types';

const router = useRouter();
const route = useRoute();
const apps = useAppsStore();
const { commandPaletteOpen } = useCommandPalette();

const searchText = ref('');

const allCommands = computed<Command[]>(() => buildCommands(apps));
const navCommands = computed(() => allCommands.value.filter((c) => c.group === 'navigation'));
const actionCommands = computed(() => allCommands.value.filter((c) => c.group === 'action'));

const byId = computed<Record<string, Command>>(() =>
	Object.fromEntries(allCommands.value.map((c) => [c.id, c])),
);

function recentItems(values: string[]): Command[] {
	return values.map((v) => byId.value[v]).filter((c): c is Command => Boolean(c));
}

// Dependency-free substring filter over title + parent + keywords. The command
// set is small, so fuzzy ranking isn't needed.
function filter(value: string, search: string): boolean {
	if (!search) return true;
	const cmd = byId.value[value];
	if (!cmd) return false;
	const haystack = [cmd.title, cmd.parent ?? '', ...(cmd.keywords ?? [])].join(' ').toLowerCase();
	return search
		.toLowerCase()
		.split(/\s+/)
		.every((term) => haystack.includes(term));
}

function run(cmd: Command) {
	commandPaletteOpen.value = false;
	void cmd.handler({ route, router });
}

function onKey(e: KeyboardEvent) {
	if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
		e.preventDefault();
		commandPaletteOpen.value = !commandPaletteOpen.value;
	}
}

onMounted(() => window.addEventListener('keydown', onKey));
onUnmounted(() => window.removeEventListener('keydown', onKey));
</script>
