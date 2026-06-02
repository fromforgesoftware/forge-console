import { ref } from 'vue';

// Open state is module-level so the sidebar trigger and the ⌘K shortcut and the
// dialog all share one source of truth.
const open = ref(false);

function toggle() {
	open.value = !open.value;
}

export function useCommandPalette() {
	return {
		open,
		commandPaletteOpen: open,
		toggle,
	};
}
