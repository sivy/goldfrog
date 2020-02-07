package syndication

import "time"

type PostData struct {
	Title       string
	Slug        string
	PostDate    time.Time
	Tags        []string
	Body        string
	FrontMatter map[string]string
	PermaLink   string
	ShortID     string
}

type TwitterOpts struct {
	ClientKey    string `yaml:"clientkey"`
	ClientSecret string `yaml:"clientsecret"`
	AccessKey    string `yaml:"acceskey"`
	AccessSecret string `yaml:"accessecret"`
	UserID       string `yaml:"userid"`
	LinkFormat   string `yaml:"linkformat"`
}

type MastodonOpts struct {
	Site         string `yaml:"site"`
	ClientID     string `yaml:"clientid"`
	ClientSecret string `yaml:"clientsecret"`
	AccessToken  string `yaml:"accesstoken"`
	LinkFormat   string `yaml:"linkformat"`
}

type WebmentionOpts struct {
}
