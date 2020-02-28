module github.com/sivy/goldfrog

go 1.13

require (
	github.com/PuerkitoBio/goquery v1.5.1
	github.com/araddon/dateparse v0.0.0-20190622164848-0fb0a474d195
	github.com/dghubble/go-twitter v0.0.0-20190719072343-39e5462e111f // indirect
	github.com/dghubble/oauth1 v0.6.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-chi/chi v4.0.3+incompatible
	github.com/gomarkdown/markdown v0.0.0-20200127000047-1813ea067497
	github.com/leekchan/gtf v0.0.0-20190214083521-5fba33c5b00b
	github.com/mattn/go-mastodon v0.0.4
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/opentracing/opentracing-go v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/sivy/go-twitter v0.0.0-20200228143626-89362039a5e4
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.5.1
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/dghubble/go-twitter v0.0.0-20190719072343-39e5462e111f => github.com/sivy/go-twitter v0.0.0-20200228143626-89362039a5e4
