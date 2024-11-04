function config({ outDir }) {
    let name = outDir.split('/').pop()

    if (name === '' && outDir.length > 0) {
        name = outDir.pop()
    }

    return { name }
}
