package blog

type Config struct {
	Version string `json:"version" yaml:"version"`
	Blog    struct {
		Title   string `json:"title" yaml:"title"`
		Subhead string `json:"subhead" yaml:"subhead"`
		Author  struct {
			Name  string `json:"name" yaml:"name"`
			Email string `json:"email" yaml:"email"`
			Image string `json:"image" yaml:"image"`
		} `json:"author" yaml:"author"`
		Url  string            `json:"url" yaml:"url"`
		Meta map[string]string `json:"meta" yaml:"meta"`
	} `json:"blog" yaml:"blog"`

	Services []struct {
		ID    string `yaml:"id"`
		Name  string `yaml:"name"`
		Link  string `yaml:"link"`
		Class string `yaml:"class"`
		Icon  string `yaml:"icon"`
	} `json:"services" yaml:"services"`

	Links []struct {
		Name  string `yaml:"name"`
		Link  string `yaml:"link"`
		Class string `yaml:"class"`
	} `json:"links" yaml:"links"`

	PostsDir     string `json:"postsdir" yaml:"postsdir"`
	TemplatesDir string `json:"templatesdir" yaml:"templatesdir"`
	StaticDir    string `json:"staticdir" yaml:"staticdir"`
	UploadsDir   string `json:"uploadsdir" yaml:"uploadsdir"`

	WebMentionEnabled bool `json:"webmentionenabled" yaml:"webmentionenabled"`

	Signin struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"signin"`

	Server struct {
		Location string `yaml:"location"`
		Port     string `yaml:"port"`
	} `yaml:"server"`

	Twitter struct {
		ClientKey    string `yaml:"clientkey"`
		ClientSecret string `yaml:"clientsecret"`
		AccessKey    string `yaml:"acceskey"`
		AccessSecret string `yaml:"accessecret"`
		UserID       string `yaml:"userid"`
		LinkFormat   string `yaml:"linkformat"`
	} `yaml:"twitter"`

	Mastodon struct {
		Site         string `yaml:"site"`
		ClientID     string `yaml:"clientid"`
		ClientSecret string `yaml:"clientsecret"`
		AccessToken  string `yaml:"accesstoken"`
		LinkFormat   string `yaml:"linkformat"`
	} `yaml:"mastodon"`
}
