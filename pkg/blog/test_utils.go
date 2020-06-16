package blog

var testConfigStr = `
blog:
  title: monkinetic.blog
  subhead: Since 1999, XVI Edition
  url: "http://monkinetic.blog"
  author:
    name: Steve Ivy
    email: steveivy@gmail.com
    image: "http://monkinetic.blog/static/images/sivy_avatar_256.png"
	timezone: "America/Phoenix"
`
var TEST_CONFIG = LoadConfigStr(testConfigStr)

/*
Do-nothing file saver
*/
type NullPostsRepo struct{}

func (npr *NullPostsRepo) ListPostFiles() []string {
	return []string{}
}
func (npr *NullPostsRepo) SavePostFile(post *Post) error {
	return nil
}
func (npr *NullPostsRepo) DeletePostFile(post *Post) error {
	return nil
}
