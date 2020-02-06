package syndication

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
	UserEmail    string `yaml:"useremail"`
	UserPassword string `yaml:"userpassword"`
	LinkFormat   string `yaml:"linkformat"`
}

type WebmentionOpts struct {
}
