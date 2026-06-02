/// <reference types="vitest/config" />
import { fileURLToPath, URL } from 'node:url';
import { resolve } from 'node:path';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';

const tsKit = (sub: string) => resolve(__dirname, `../../ts-kit/src/${sub}`);

export default defineConfig({
	plugins: [vue(), tailwindcss()],
	resolve: {
		alias: {
			'@': fileURLToPath(new URL('./src', import.meta.url)),
			'@fromforgesoftware/ts-kit/jsonapi-client': tsKit('jsonapi-client/index.ts'),
			'@fromforgesoftware/ts-kit/jsonapi': tsKit('jsonapi/index.ts'),
			'@fromforgesoftware/ts-kit/resource-state': tsKit('resource-state/index.ts'),
			'@fromforgesoftware/ts-kit/http': tsKit('http/index.ts'),
			'@fromforgesoftware/ts-kit/errors': tsKit('errors/index.ts'),
			'@fromforgesoftware/ts-kit/date': tsKit('date/index.ts'),
			'@fromforgesoftware/ts-kit': tsKit('index.ts'),
			'@fromforgesoftware/vue-kit': resolve(__dirname, '../../vue-kit/src/index.ts'),
		},
	},
	test: {
		name: 'unit',
		include: ['src/**/*.test.ts'],
		environment: 'jsdom',
		globals: true,
	},
});
