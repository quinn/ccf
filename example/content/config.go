package content

type Post struct {
	// Frontmatter
	Title string `yaml:"title"`
	Date  string `yaml:"date"`
	// Magic props
	Slug     string
	Rendered string
}
