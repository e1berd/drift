import { defineConfig } from 'kubb/config'
import { adapterOas } from '@kubb/adapter-oas'
import { pluginClient } from '@kubb/plugin-client'
import { pluginTs } from '@kubb/plugin-ts'
import { definePageParam, pluginVueQuery } from '@kubb/plugin-vue-query'

export default defineConfig({
    input: '../app/api/openapi.yaml',
    output: { path: './src/gen', clean: true, barrel: { type: 'named' } },
    adapter: adapterOas({ integerType: 'number' }),
    plugins: [
        pluginTs({ output: { path: 'types', barrel: { type: 'named' } } }),
        pluginClient({
            output: { path: 'clients', barrel: { type: 'named' } },
            importPath: '../../shared/http/client',
            exclude: [
                { type: 'tag', pattern: 'events' },
                { type: 'tag', pattern: 'terminal' },
                { type: 'tag', pattern: 'uploads' },
            ],
        }),
        pluginVueQuery({
            output: { path: 'queries', barrel: { type: 'named' } },
            hooks: true,
            exclude: [
                { type: 'tag', pattern: 'events' },
                { type: 'tag', pattern: 'terminal' },
                { type: 'tag', pattern: 'uploads' },
            ],
            infinite: {
                queryParam: 'page_token',
                initialPageParam: '',
                getNextPageParam: definePageParam<{ nextPageToken?: string }, string>(
                    (lastPage) => lastPage.nextPageToken || undefined,
                ),
            },
        }),
    ],
})
