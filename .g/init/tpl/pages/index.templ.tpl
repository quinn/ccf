package pages

import (
	"github.com/labstack/echo/v4"
	"go.quinn.io/ccf/assets"
	"{{ .mod }}/content"
)

func IndexHandler(c echo.Context) ([]content.PostItem, error) {
	// Get items for a specific type
	return content.GetPosts()
}

templ Index(posts []content.PostItem) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="UTF-8"/>
			<title>Go Astro</title>
			<link rel="stylesheet" href={ assets.Path("styles.css") }/>
		</head>
		<body>
			<h1>Welcome to Go Astro</h1>
			<p>A Go implementation of Astro-like functionality</p>
			<nav>
				<ul>
					for _, post := range posts {
						<li>
							<a href={ templ.URL("/blog/" + post.Slug) }>
								{ post.Meta.Title }
							</a>
						</li>
					}
				</ul>
			</nav>
		</body>
	</html>
}
