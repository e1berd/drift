import { defineConfig } from 'kubb/config'
import { adapterOas } from '@kubb/adapter-oas'
import { pluginClient } from '@kubb/plugin-client'
import { pluginTs } from '@kubb/plugin-ts'
import { definePageParam, pluginVueQuery } from '@kubb/plugin-vue-query'

const transportOnlyTags = [
    { type: 'tag', pattern: 'events' },
    { type: 'tag', pattern: 'terminal' },
    { type: 'tag', pattern: 'uploads' },
] as const

export default defineConfig({
    input: '../app/api/openapi.yaml',
    output: { path: './packages', clean: false, barrel: false },
    adapter: adapterOas({ integerType: 'number' }),
    plugins: [
        pluginTs({ output: { path: 'sdk/src/api/types', barrel: { type: 'named' } } }),
        pluginClient({
            output: { path: 'sdk/src/api/clients', barrel: { type: 'named' } },
            importPath: '../../http/client',
            exclude: [...transportOnlyTags],
        }),
        pluginVueQuery({
            output: { path: 'vue/src/api/queries', barrel: { type: 'named' } },
            hooks: true,
            exclude: [...transportOnlyTags],
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
