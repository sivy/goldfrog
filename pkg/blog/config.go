package blog

import (
	"bytes"

	"github.com/spf13/viper"
)

type Config struct {
	Version string `json:"version" yaml:"version"`
	Blog    struct {
		Title   string `json:"title" yaml:"title"`
		Subhead string `json:"subhead" yaml:"subhead"`
		Author  struct {
			Name     string `json:"name" yaml:"name"`
			Email    string `json:"email" yaml:"email"`
			Image    string `json:"image" yaml:"image"`
			TimeZone string `json:"timezone" yaml:"timezone"`
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

	// directory of post Markdown files
	PostsDir string `json:"postsdir" yaml:"postsdir"`
	// directory of data files
	DataDir string `json:"datadir" yaml:"datadir"`
	// directory where templates are stored
	// templates are part of "blog content" and will probably be handled
	// alongside posts
	TemplatesDir string `json:"templatesdir" yaml:"templatesdir"`
	// directory where static assets will be found
	// also part of "blog content"
	StaticDir string `json:"staticdir" yaml:"staticdir"`
	// directory for file uploads
	UploadsDir string `json:"uploadsdir" yaml:"uploadsdir"`

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

func LoadConfig(configPath string) Config {
	viper.AddConfigPath(configPath)
	viper.ReadInConfig()

	var config Config
	viper.Unmarshal(&config)

	return config
}

func LoadConfigStr(configStr string) Config {
	configBytes := []byte(configStr)
	viper.ReadConfig(bytes.NewBuffer(configBytes))

	var config Config
	viper.Unmarshal(&config)

	return config
}
