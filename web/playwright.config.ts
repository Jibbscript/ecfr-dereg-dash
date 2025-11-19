import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
    testDir: './tests/e2e',
    fullyParallel: true,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 2 : 0,
    workers: process.env.CI ? 1 : undefined,
    reporter: 'html',
    use: {
        baseURL: 'http://localhost:3000',
        trace: 'on-first-retry',
    },
    projects: [
        {
            name: 'chromium',
            use: { ...devices['Desktop Chrome'] },
        },
    ],
    webServer: [
        {
            command: 'npm run preview',
            url: 'http://localhost:3000',
            reuseExistingServer: !process.env.CI,
            timeout: 120000,
        },
        {
            command: 'cd .. && go run cmd/api/main.go',
            url: 'http://localhost:8080/api/agencies', // Check agencies endpoint
            reuseExistingServer: !process.env.CI,
            timeout: 120000,
            env: {
                PORT: '8080',
                DATA_DIR: './web/test_data', // Relative to project root where go run is executed
                SKIP_VERTEX: 'true',
            }
        }
    ],
});
