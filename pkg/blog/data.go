/*
Code to manage data files for the blog, especially per-post webmention data
*/

package blog

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

func GetPostData(dataDir string, post *Post) (map[string]string, error) {
	path := fmt.Sprintf("%s/%s.yaml", dataDir, post.Slug)
	dataContent, err := ioutil.ReadFile(path)

	if err != nil {
		logger.Errorf("Error reading post data: %s", err)
		return nil, err
	}

	var postData = make(map[string]string)

	yaml.Unmarshal([]byte(dataContent), &postData)

	return postData, nil

}
