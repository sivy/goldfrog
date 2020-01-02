package blog

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func IndexPosts(postsDir string, dbFile string) {
	log.Infof("Indexing posts in %s to %s", postsDir, dbFile)
	log.Debugf("Checking db at %s", dbFile)

	if !checkDb(dbFile) {
		log.Infof("Creating DB at %s", dbFile)
		initDb(dbFile)
	}

	db, err := GetDb(dbFile)
	if err != nil {
		log.Error(err)
		return
	}

	repo := PostsRepo{
		PostsDirectory: postsDir,
	}

	files := repo.ListPostFiles()

	var i = 0
	for _, f := range files {
		indexed := IndexFile(f, db)
		if indexed {
			i++
		}
	}

	log.Debugf("Found %d files", len(files))
	log.Debugf("Indexed %d files", i)
}

func IndexFile(file string, db *sql.DB) bool {
	log.Infof("Indexing file %s", file)
	var post Post
	post, err := ParseFile(file)

	if err != nil {
		log.Errorf("Could not parse file %s", file)
		return false
	}
	log.Debug(post)

	// does it exist?
	rows, err := db.Query(fmt.Sprintf(
		"SELECT id FROM posts WHERE slug = '%s' LIMIT 1", post.Slug))

	if err != nil {
		log.Errorf("Error checking for post with slug %s: %v", post.Slug, err)
	}

	var ID int
	for rows.Next() {
		rows.Scan(&ID)
	}
	log.Debugf("found rows ID: %v for post slug: %s", ID, post.Slug)

	var sql string
	if ID != 0 {
		sql = `
		UPDATE posts
		SET title='?', tags='?', postdate=?, body='?'
		WHERE id=?`

		res, err := db.Exec(
			sql,
			post.Title,
			strings.Join(post.Tags, ", "),
			post.PostDate.Format(time.RFC3339),
			post.Body,
		)
		if err != nil {
			log.Errorf("Could not update post %s", err)
			return false
		}

		i, _ := res.RowsAffected()
		log.Debugf("Updated %d rows", i)

	} else {
		sql = `
		INSERT into posts (
			slug, title, tags, postdate, body
		) values (
			?, ?, ?, ?, ?
		)
		`

		res, err := db.Exec(
			sql,
			post.Slug,
			post.Title,
			strings.Join(post.Tags, ", "),
			post.PostDate.Format(time.RFC3339),
			post.Body,
		)

		if err != nil {
			log.Errorf("Could not add post %s", err)
			return false
		}

		i, _ := res.RowsAffected()
		log.Debugf("Inserted %d rows", i)
	}

	return true
}