package blog

import (
	"database/sql"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func IndexPosts(postsDir string, dbFile string, verbose bool) {
	logger := logrus.New()

	logger.Debugf("Checking db at %s", dbFile)

	if !checkDb(dbFile) {
		logger.Infof("Creating DB at %s", dbFile)
		initDb(dbFile)
	}

	db, err := GetDb(dbFile)
	if err != nil {
		logger.Error(err)
		return
	}

	repo := PostsRepo{
		PostsDirectory: postsDir,
	}

	files := repo.ListPostFiles()

	var i = 0
	for _, f := range files {
		indexed := IndexFile(f, db, verbose)
		if indexed > 0 {
			i += indexed
		}
	}

	logger.Debugf("Found %d files", len(files))
	logger.Debugf("Indexed %d files", i)
}

func IndexFile(file string, db *sql.DB, verbose bool) int {
	logger := logrus.New()

	if verbose {
		logger.Debugf("=== Indexing file %s", file)
	}
	var post Post
	post, err := ParseFile(file)

	if err != nil {
		logger.Errorf("Could not parse file %s", file)
		return 0
	}
	if verbose {
		logger.Debugf("loaded post: %s", post.Title)
	}

	var sql = `
		INSERT INTO posts (
			slug, title, tags, postdate, body, format
		) VALUES (
			?, ?, ?, ?, ?, 'markdown'
		) ON CONFLICT(slug) DO UPDATE
		SET
			title=excluded.title,
			tags=excluded.tags,
			postdate=excluded.postdate,
			body=excluded.body;
	`
	logger.Infof("Insert/Update post %s", post.Slug)

	tx, err := db.Begin()
	res, err := tx.Exec(
		sql,
		post.Slug,
		post.Title,
		strings.Join(post.Tags, ", "),
		post.PostDate.Format(time.RFC3339),
		post.Body,
	)

	if err != nil {
		logger.Errorf("Could not add post: %s", err)
		err := tx.Rollback()
		if err != nil {
			logger.Errorf("Could not rollback tx: %s", err)
		}
		return 0
	}

	err = tx.Commit()
	if err != nil {
		logger.Errorf("Could not commit post: %s", err)
		err := tx.Rollback()
		if err != nil {
			logger.Errorf("Could not rollback tx: %s", err)
		}
		return 0
	}

	rowCount, _ := res.RowsAffected()
	lastId, _ := res.LastInsertId()

	logger.Infof(
		"Inserted %d records, ID %d",
		rowCount, lastId)

	return int(rowCount)
}
