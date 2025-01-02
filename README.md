# CCF: Content Creation Framework

**CCF** is a Go-based framework that simplifies building content-driven applications. It provides three main systems:

1. **Content System** for loading and transforming Markdown content (with frontmatter) into Go data and HTML.
2. **File/Page-Based Routing** for automatically converting `.templ` files into Echo routes.
3. **Assets System** for serving fingerprinted static files (e.g., CSS, JS) with optional embedding.

This guide walks you through setting up each system, with code samples and references to the included `taskfile.yaml` so you can quickly get started.

---

## Prerequisites

1. A working Go environment (Go 1.22+ recommended).
2. [Task](https://taskfile.dev) (or a similar tool) if you want to run the provided tasks directly.
3. Familiarity with [Echo](https://echo.labstack.com/), though basic usage is covered here.

---

## 1. Content System

The **content system** automatically loads Markdown files with frontmatter, converts them to HTML, and stores them in memory for easy retrieval.

### 1.1 Defining Your Content Struct

In your `content/config.go` (or similar file within your `content` folder), define a struct for your frontmatter fields:

```go
package content

// Example Post struct
type Post struct {
    Title       string `yaml:"title"`
    Date        string `yaml:"date"`
    Description string `yaml:"description"`
}
```

### 1.2 Creating Markdown Files

Place markdown files under `content/<typeName>/`. For example, if your struct is `Post`, put them in `content/posts/`:

```
content/
  posts/
    my-first-post.md
    2025/some-other-post.md
```

Each markdown file should have frontmatter at the top:

```markdown
---
title: "My First Post"
date: "2025-01-01"
description: "This is a sample post."
---

# Hello World

This is my first blog post using **CCF**.
```

### 1.3 Generating and Loading Content

CCF’s code generator reads your `content/config.go`, finds structs, locates matching directories, and generates a `fs.go` file (or similar) to embed or read those files at runtime.

In the **example project**, there is a `taskfile.yaml` target called `gen-content` that invokes the generator:

```yaml
# From example/taskfile.yaml

tasks:
  gen-content:
    cmds:
      - |
        source ../scripts/ccff.sh
        ccff generate/content \
          -content content
```

You can run:
```bash
task gen-content
```
This will:

1. Parse your `content/config.go`
2. Embed or reference all Markdown files in `content/`
3. Create/update a `fs.go` (or similar) file that your code can import

### 1.4 Using Your Content in Go

After generation, CCF provides helpers like `GetPosts()` (if your struct is called `Post`) or a more generic `GetItems[T]()`. For instance:

```go
import "myproject/content"

func main() {
    // If you’re using an embedded FS approach, initialize it:
    // content.Initialize(echoInstance)
    // or manually load items:

    // load posts if not using the generated Initialize function
    err := content.LoadItems[Post](os.DirFS("content"), "posts")
    if err != nil {
        panic(err)
    }

    posts, err := content.GetItems[Post]()
    if err != nil {
        panic(err)
    }

    for _, p := range posts {
        fmt.Println("Title:", p.Meta.Title)
        fmt.Println("Slug:", p.Slug)
        fmt.Println("HTML:", p.HTML)
    }
}
```

The system stores both the **raw Markdown** and the **rendered HTML** (with code highlighting, relative image rewriting, etc.), making it convenient to display in your templates.

---
Below is an updated **Section 2** discussing **automatically generated POST routes** alongside GET routes.

---

## 2. File/Page-Based Routing

CCF uses [Templ](https://github.com/a-h/templ) files (`.templ`) to define server-side pages. Each `.templ` file can define zero or more HTTP handlers (e.g., GET, POST). CCF **automatically** generates Echo route handlers so you don’t have to write boilerplate.

### 2.1 Creating a `.templ` Page

Inside your `pages/` directory, create a file such as `pages/blog.[slug].templ`:

```go
package pages

import "github.com/labstack/echo/v4"

// The "GET" function that runs before rendering the GET template
func BlogSlugGET(c echo.Context, slug string) (string, error) {
    // Return the data that the template needs
    return slug, nil
}

// Optionally, you can define a POST handler in the same file.
// The generator will pick it up and create a POST route automatically.
func BlogSlugPOST(c echo.Context, slug string) error {
    // Add your logic for creating/updating data here
    // e.g. read form values, save to a DB, etc.
    return nil
}

// The Templ syntax defines how the GET template is rendered:
templ BlogSlug(slug string) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8"/>
        <title>Blog Post: { slug }</title>
    </head>
    <body>
        <h1>Blog Post Slug: { slug }</h1>
        <p>This is a dynamic page for blog slug "{ slug }".</p>
    </body>
    </html>
}
```

By naming the file `blog.[slug].templ`, you automatically get **two** routes:
- **GET** `/blog/:slug` (via `BlogSlugGET`)
- **POST** `/blog/:slug` (via `BlogSlugPOST`, if defined)

If you omit the `BlogSlugPOST` function, then no POST route is generated.

### 2.2 Generating Routes

In the **example** `Taskfile.yaml`, there is a `gen-pages` target that runs a script to generate your router code:

```yaml
tasks:
  gen-pages:
    cmds:
      - |
        source ../scripts/ccff.sh
        ccff generate/pages \
          -pages pages \
          -output internal/router/router.go \
          -package router
      - task gen-templ
      - goimports -w internal/router/router.go
```

Running:
```bash
task gen-pages
```
will scan the `pages/` directory, find `.templ` files (and their handlers), and generate a file such as `internal/router/router.go`. This file contains **both** GET and POST routes if you’ve defined them in the `.templ`:

```go
// RegisterRoutes adds all page routes to the Echo instance
func RegisterRoutes(e *echo.Echo) {
    e.GET("/blog/:slug", BlogSlugGET)      // from BlogSlugGET
    e.POST("/blog/:slug", BlogSlugPOST)    // from BlogSlugPOST
    // ... additional routes
}
```

### 2.3 Using the Routes

In your main server code, just call the generated registration function:

```go
import (
    "github.com/labstack/echo/v4"
    "myproject/internal/router"
)

func main() {
    e := echo.New()
    router.RegisterRoutes(e)
    e.Logger.Fatal(e.Start(":3000"))
}
```

Now requests to:
- **GET** `/blog/my-article` → calls `BlogSlugGET`
- **POST** `/blog/my-article` → calls `BlogSlugPOST`

Depending on which handler is defined in your `.templ` file.

---

**Tip**: If your `.templ` file does not define a `POST` function (e.g., `SomethingPOST`), CCF will **not** generate the corresponding POST route. This makes it easy to keep everything in one place while only creating routes you actually need.

---

## 3. Assets System

CCF includes an **assets** package that can optionally fingerprint and embed static files, such as Tailwind CSS or other resources. It rewrites file paths to include a hash for cache-busting.

### 3.1 Project Setup

A typical structure might include:

```
internal/
  web/
    public/
      styles.css
      script.js
```

You can embed or read these files by referencing the directory in Go:

```go
package web

import (
    "embed"
    "github.com/labstack/echo/v4"
    "go.quinn.io/ccf/assets"
    "log"
    "os"
)

//go:embed public
var assetsFS embed.FS

func Run() {
    e := echo.New()

    // Attach the fingerprinted assets.
    assets.Attach(
        e,
        "public",               // URL prefix -> /public
        "internal/web/public",  // local directory path
        assetsFS,               // embedded FS
        os.Getenv("USE_EMBEDDED_ASSETS") == "true",
    )

    log.Fatal(e.Start(":3000"))
}
```

### 3.2 Using Assets in Templates

Inside your `.templ` files, you can reference `assets.Path("filename.css")` to get a fingerprinted path:

```go
templ Index() {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8"/>
        <link rel="stylesheet" href={ assets.Path("styles.css") }/>
    </head>
    <body>
        <h1>Hello World!</h1>
    </body>
    </html>
}
```

If `styles.css` is fingerprinted to `styles.a1b2c3.css`, the above call automatically produces `<link rel="stylesheet" href="/public/styles.a1b2c3.css" />` at runtime.

### 3.3 Tailwind Example

If you use Tailwind, you might have a `tailwind.config.js` and a top-level `tailwind.css` that you compile to `internal/web/public/styles.css`. In the example `taskfile.yaml`, a `gen-tailwind` target shows how to do it:

```yaml
tasks:
  gen-tailwind:
    cmds:
      - |
        tailwindcss \
          -i ./tailwind.css \
          -o ./internal/web/public/styles.css
```

Then you can run:
```bash
task gen-tailwind
```
to compile your CSS. Combined with the fingerprinting approach, your references in templates will always point to the latest hashed file.

---

## Putting It All Together

### 1. Create or clone the structure

Make sure you have these folders (at least):

```
content/    # Markdown content
pages/      # .templ page files
internal/   # Where router is generated
cmd/        # Main entrypoints (like cmd/main.go)
```

### 2. Add/Update Markdown & Templates

Create or modify your Markdown in `content/` and your `.templ` pages in `pages/`.

### 3. Generate

Use the provided tasks (or adapt them) in your `taskfile.yaml`:

```bash
# Generate content
task gen-content

# Generate router (page-based routing)
task gen-pages

# Format & generate templ (optional)
task gen-templ
```

### 4. Run the Server

```bash
go run cmd/main.go
```
Now open [http://localhost:3000/](http://localhost:3000/) to see your pages.

---

## Example Taskfile for Reference

Below is an excerpt from the **example/taskfile.yaml** that you can adapt:

```yaml
version: '3'

tasks:
  gen-templ:
    cmds:
      - go run github.com/a-h/templ/cmd/templ@latest fmt .
      - go run github.com/a-h/templ/cmd/templ@latest generate

  gen-pages:
    cmds:
      - |
        source ../scripts/ccff.sh
        ccff generate/pages \
          -pages pages \
          -output internal/router/router.go \
          -package router
      - task gen-templ
      - goimports -w internal/router/router.go

  gen-content:
    cmds:
      - |
        source ../scripts/ccff.sh
        ccff generate/content \
          -content content

  gen-tailwind:
    cmds:
      - |
        tailwindcss \
          -i ./tailwind.css \
          -o ./internal/web/public/styles.css

  build:
    cmds:
      - task gen-templ
      - task gen-pages
      - task gen-content
      - task gen-tailwind
      - go build -o ./tmp/main cmd/main.go
```

You can then run:
```bash
task build
```
to execute all generation steps and build your server binary.

---

## Conclusion

**CCF** (Content Creation Framework) provides:

1. **Content System**: Parse frontmatter from Markdown, generate Go data and HTML.
2. **File-Based Routing**: Turn `.templ` files into Echo routes with zero extra code.
3. **Assets System**: Serve and fingerprint static files, easily referenced in Templ.

With these features, you can rapidly build content-centric Go apps without repetitive boilerplate. For more advanced examples and usage patterns, explore the `example` directory and the `taskfile.yaml` in this repository. Happy coding!
