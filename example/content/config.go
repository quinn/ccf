package content

// type Post will load a folder called posts
type Post struct {
	// Frontmatter
	Title       string `yaml:"title"`
	Date        string `yaml:"date"`
	Description string `yaml:"description"`
}
