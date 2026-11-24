import { fileURLToPath } from 'node:url'

export const driftLocales = fileURLToPath(new URL('./packages/vue/src/locales/**', import.meta.url))

const buildsDirectory = new URL('../app/internal/frontend/dist/', import.meta.url)

/**
 * Vite settings every Drift interface shares: assets under /ui/<id>/, the build in the Go
 * module's dist/<id>/ with a manifest, and `@` pointing at the app's src.
 */
export function driftAppConfig(appConfigUrl: string, id: string, entries: Record<string, string>) {
    const appDirectory = new URL('./', appConfigUrl)
    return {
        base: `/ui/${id}/`,
        resolve: {
            alias: { '@': fileURLToPath(new URL('./src', appDirectory)) },
        },
        build: {
            outDir: fileURLToPath(new URL(`./${id}`, buildsDirectory)),
            emptyOutDir: true,
            manifest: true,
            rolldownOptions: {
                input: Object.fromEntries(
                    Object.entries(entries).map(([name, path]) => [
                        name,
                        fileURLToPath(new URL(path, appDirectory)),
                    ]),
                ),
            },
        },
    }
}
