package blog

type Config struct {
	Blog struct {
		Title   string            `json:"title" yaml:"title"`
		Subhead string            `json:"subhead" yaml:"subhead"`
		Author  string            `json:"author" yaml:"author"`
		Meta    map[string]string `json:"meta" yaml:"meta"`
	} `json:"blog" yaml:"blog"`

	Services []map[string]string `json:"services" yaml:"services"`

	PostsDir     string `json:"posts" yaml:"posts"`
	TemplatesDir string `json:"templates" yaml:"templates"`
	StaticDir    string `json:"static" yaml:"static"`

	Signin struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"signin"`

	Server struct {
		Location string `yaml:"location"`
		Port     string `yaml:"port"`
	} `yaml:"server"`
}
