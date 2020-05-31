package blog

const (
	syndicationConfigStr string = `
twitter:
  userid: foo
  linkformat: "%s/%s"

mastodon:
  linkformat: "%s"
`
)

// func TestTootLinker(t *testing.T) {
// 	var c Config

// 	yaml.Unmarshal([]byte(syndicationConfigStr), &c)

// 	out := string(tootLinker(c, "123"))
// 	assert.NotNil(t, out)
// 	assert.Equal(t, "123", out)
// }

// func TestTweetLinker(t *testing.T) {
// 	var c Config

// 	yaml.Unmarshal([]byte(syndicationConfigStr), &c)

// 	out := string(tweetLinker(c, "123"))
// 	assert.NotNil(t, out)
// 	assert.Equal(t, "foo/123", out)
// }
