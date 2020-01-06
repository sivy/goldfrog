package blog

func init() {

}

type CrossPoster interface {
	PostContent(post Post) bool
}

type CrossPosterFields struct {
	Config Config
}

type MastodonCrossPoster struct {
	config Config
}

func (xp *MastodonCrossPoster) PostContent(post Post) bool {
	return false
}

func NewMastodonCrossPoster(config Config) MastodonCrossPoster {

	return MastodonCrossPoster{
		config: config,
	}
}
