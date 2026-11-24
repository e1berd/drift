import { readdir, readFile, writeFile } from 'node:fs/promises'
import { join } from 'node:path'

const queriesDirectory = new URL('../packages/vue/src/api/queries/', import.meta.url).pathname
const relativeSdkImport = /'(?:\.\.\/)+sdk\/src\//g

async function importSdkByPackageName(fileName) {
    const path = join(queriesDirectory, fileName)
    const source = await readFile(path, 'utf8')
    const linked = source.replaceAll(relativeSdkImport, "'@drift/sdk/")
    if (linked !== source) await writeFile(path, linked)
}

const fileNames = await readdir(queriesDirectory)
await Promise.all(fileNames.map(importSdkByPackageName))
