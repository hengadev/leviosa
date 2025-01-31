import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';
import type { ConfigEnv, Plugin } from 'vite';
import { kitRoutes } from 'vite-plugin-kit-routes';
import mkcert from 'vite-plugin-mkcert';

const add_browser_onmount: Plugin = {
    name: 'vite-plugin-onmount',
    config(config) {
        if (process.env.VITEST) {
            if (!config.resolve?.conditions) {
                if (!config.resolve) {
                    config.resolve = {};
                }
                config.resolve.conditions = [];
            }
            config.resolve.conditions.unshift('browser');
        }
    }
};

export default defineConfig(({ command, mode }: ConfigEnv) => {
    const port = (mode === "staging" && command == "build") ? 3001 : 3000;
    return {
        plugins: [sveltekit(), add_browser_onmount, kitRoutes(), mkcert()],
        test: {
            include: ['src/**/*.{test,spec}.{js,ts}', 'tests/**/*.{test,spec}.{js,ts}'],
            environment: 'jsdom',
            setupFiles: ['src/vitest.setup.ts'],
            globals: true
        },
        server: { https: true, port: port },
    }
});
