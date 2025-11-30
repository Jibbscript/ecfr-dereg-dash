import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
    plugins: [vue()],
    test: {
        environment: 'happy-dom',
        include: ['tests/**/*.test.ts'],
        coverage: {
            provider: 'v8',
            reporter: ['text', 'html', 'lcov'],
            lines: 91,
            statements: 91,
            branches: 85,
            functions: 90,
            exclude: [
                'tests/**',
                '.nuxt/**',
                'coverage/**',
                'playwright.config.*',
                'playwright.docker.config.*',
                'vitest.config.*',
                'nuxt.config.*',
                '**/*.d.ts',
                '.output/**',
            ],
        }
    }
})
