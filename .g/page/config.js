function config({ path }) {
    let rpath = path

    // remove first char of path if it is '/'
    if (path.startsWith('/')) {
        rpath = path.slice(1)
    }

    let args = []

    const parts = rpath.split('/')
    let filename = parts.map((part) => {
        if (part.startsWith(':')) {
            let arg = part.slice(1)
            args.push(arg)
            return '[' + arg + ']'
        }

        return part
    }).join('.')

    let templName = parts.map((part) => {
        return convertCase('pascal', part.startsWith(':') ? part.slice(1) : part)
    }).join('')

    let templParams = args.length > 0
        ? ', ' + args.join(', ') + ' string'
        : ''

    return { filename, templName, templParams }
}
