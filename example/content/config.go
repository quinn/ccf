package content

type Post struct {
	// Frontmatter
	Title string `yaml:"title"`
	Date  string `yaml:"date"`
}
