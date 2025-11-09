# CCF: Content Creation Framework

**CCF** is a Go-based framework that simplifies building content-driven applications or websites. It provides three main systems:

1. **Font Tool** for downloading and installing Google Fonts, and generating CSS the font face css rules.
2. **Assets System** for serving fingerprinted static files (e.g., CSS, JS) with optional embedding.
3. **Content System** for loading and transforming Markdown content (with frontmatter) into Go data and HTML.
4. **File/Page-Based Routing** for automatically converting `.templ` files into Echo routes.

---

## 1. Font Tool

```bash
$ ccf fonts --help
Usage of ccf fonts:
  -config string
        Path to font configuration file (default "fonts.yaml")
  -debug
        Enable debug logging
  -gfonts-key string
        Google Fonts API key. (default "GFONTS_KEY" env var)
```

The **font tool** downloads and installs Google Fonts, and generates CSS rules for the font face.

### 1.1 Defining Your Font Config

Create a file called `fonts.yaml` in the root of your project. It should look like this:

```yaml
---
dir: ./example/fonts
stylesheet: ./example/css/fonts.css
import: ../fonts/
fonts:
  - family: Quicksand
    variants:
      - "regular"
```

### 1.2 Installing Fonts

Run the following command:

```bash
ccf fonts
```

This will download the specified fonts and generate CSS rules for them:

```bash
ls ./example/fonts/
Quicksand_regular.woff2
```

```bash
ls ./example/css/
fonts.css
```

```css
@font-face {
  font-family: 'Quicksand';
  font-style: normal;
  font-weight: 300 700;
  src: url('../fonts/Quicksand_regular.woff2') format('woff2-variations');
}
```

---

## 2. Assets System

CCF includes an **assets** package that can optionally fingerprint and embed static files, such as Tailwind CSS or other resources. It rewrites file paths to include a hash for cache-busting.

### 2.1 Project Setup

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

### 2.2 Using Assets in Templates

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

## Content Management

This guide walks you through setting up each system, with code samples and references to the included `taskfile.yaml` so you can quickly get started.

---

## Prerequisites

1. A working Go environment (Go 1.22+ recommended).
2. [Task](https://taskfile.dev) (or a similar tool) if you want to run the provided tasks directly.
3. Familiarity with [Echo](https://echo.labstack.com/), though basic usage is covered here.

---

## 3. Content System

The **content system** automatically loads Markdown files with frontmatter, converts them to HTML, and stores them in memory for easy retrieval.

### 3.1 Defining Your Content Struct

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

### 3.2 Creating Markdown Files

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

### 3.3 Generating and Loading Content

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

### 3.4 Using Your Content in Go

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

## 4. File/Page-Based Routing

CCF uses [Templ](https://github.com/a-h/templ) files (`.templ`) to define server-side pages. Each `.templ` file can define zero or more HTTP handlers (e.g., GET, POST). CCF **automatically** generates Echo route handlers so you don’t have to write boilerplate.

### 4.1 Creating a `.templ` Page

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

### 4.2 Generating Routes

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

### 4.3 Using the Routes

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
