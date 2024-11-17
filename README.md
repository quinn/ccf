### Go Astro

Reproduce some functionality of the Astro framework using Go.

## Features

### Page-based Routing
- Pages go in a folder called `pages/`
- File names define routes using dot notation:
  - `pages/blog.[slug].templ` -> `/blog/:slug`
  - `pages/posts.[id].edit.templ` -> `/posts/:id/edit`
  - `pages/index.templ` -> `/`
- Route handlers are automatically generated from page templates
- Generated handlers are attached to the server
- templ names must follow a pattern e.g. posts.[id].edit.templ is PostsIDEdit or BlogSlug
- Handlers are exported with Handler appended e.g. BlogSlugHandler
- Handlers receive the echo context and any url params as args


## Development

### Code Generation
The router code is generated from page templates. To generate the router:

```bash
go run cmd/generate/main.go -output example/router/router.go -package router
```

This will:
1. Scan the `pages/` directory for `.templ` files
2. Generate route handlers based on file names
3. Create a router package with Echo handlers

### Running the Example
To start the example server:

```bash
cd example && go run main.go
```

The server will be available at http://localhost:3000.

## Done Features
- Page-based routing with code generation
  - Automatic route generation from page templates
  - Support for dynamic route parameters (e.g., [slug])
  - Integration with Echo web framework
  - Path generation from dot-notation filenames
