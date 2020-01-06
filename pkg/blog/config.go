package blog

type Config struct {
	Blog struct {
		Title   string `json:"title" yaml:"title"`
		Subhead string `json:"subhead" yaml:"subhead"`
		Author  struct {
			Name  string `json:"name" yaml:"name"`
			Email string `json:"email" yaml:"email"`
		} `json:"author" yaml:"author"`
		Url  string            `json:"url" yaml:"url"`
		Meta map[string]string `json:"meta" yaml:"meta"`
	} `json:"blog" yaml:"blog"`

	Services []map[string]string `json:"services" yaml:"services"`

	PostsDir     string `json:"postsdir" yaml:"postsdir"`
	TemplatesDir string `json:"templatesdir" yaml:"templatesdir"`
	StaticDir    string `json:"staticdir" yaml:"staticdir"`
	UploadsDir   string `json:"uploadsdir" yaml:"uploadsdir"`

	Signin struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"signin"`

	Server struct {
		Location string `yaml:"location"`
		Port     string `yaml:"port"`
	} `yaml:"server"`
}
