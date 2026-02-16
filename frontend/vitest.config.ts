/// <reference types="vitest" />
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
    plugins: [react()],
    test: {
        globals: true,
        environment: 'jsdom',
        setupFiles: ['./vitest.setup.ts'],
        include: ['**/__tests__/**/*.test.{ts,tsx}', '**/*.test.{ts,tsx}'],
        exclude: ['node_modules', '.next', 'dist', 'tests/**/*.spec.ts'],
        coverage: {
            provider: 'v8',
            reporter: ['text', 'json', 'json-summary', 'lcov', 'html'],
            include: ['lib/**/*.ts', 'hooks/**/*.ts', 'components/**/*.tsx'],
            exclude: [
                'node_modules',
                '.next',
                '**/*.d.ts',
                '**/*.test.{ts,tsx}',
                '**/index.ts',
            ],
        },
        alias: {
            '@': path.resolve(__dirname, './'),
        },
    },
})
