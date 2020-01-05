package blog

import (
	"database/sql"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func GetDb(dbFile string) (*sql.DB, error) {
	return sql.Open("sqlite3", dbFile)
}

type GetPostOpts struct {
	Title string
	// PostDate time.Time
	Tags  []string
	Body  string
	Limit int
}

type ArchiveEntry struct {
	Year  string
	Month string
	Count int
}

var dateFmts = [...]string{
	"2006-01-02T03:04:05",
	"2006-01-02T03:04:05Z",
	"2006-01-02 03:04:05",
}

func GetPosts(db *sql.DB, opts GetPostOpts) []Post {
	log := logrus.New()

	var posts = make([]Post, 0)

	rows, err := db.Query(`
		SELECT id, slug, title, tags,
		postdate, body FROM posts
		ORDER BY datetime(postdate) DESC
		LIMIT 20
	`)

	if err != nil {
		log.Errorf("Could not load posts: %v", err)
	}

	for rows.Next() {
		// fmt.Printf("%v", row)
		var p Post
		var body string
		var tags string
		var dateStr string

		err = rows.Scan(&p.ID, &p.Slug, &p.Title, &tags, &dateStr, &body)
		if err != nil {
			log.Error(err)
		}
		var date time.Time
		var err error

		date, err = dateparse.ParseAny(dateStr)

		if err != nil {
			log.Errorf("Cannot parse date from %s", dateStr)
			p.PostDate = time.Now()
		} else {
			p.PostDate = date
		}

		p.Tags = strings.Split(tags, ", ")

		p.Body = body

		posts = append(posts, p)
	}
	return posts
}

func GetPost(db *sql.DB, postID string) (Post, error) {
	log := logrus.New()

	var p Post

	rows, err := db.Query(`
		SELECT id, slug, title, tags,
		postdate, body FROM posts
		WHERE id = ?`, postID)

	if err != nil {
		log.Errorf("Could not load post %s: %v", postID, err)
		return p, err
	}

	for rows.Next() {
		// fmt.Printf("%v", row)
		var body string
		var tags string
		var dateStr string

		err = rows.Scan(&p.ID, &p.Slug, &p.Title, &tags, &dateStr, &body)
		if err != nil {
			log.Error(err)
		}
		var date time.Time
		var err error

		date, err = dateparse.ParseAny(dateStr)

		if err != nil {
			log.Errorf("Cannot parse date from %s", dateStr)
			p.PostDate = time.Now()
		} else {
			p.PostDate = date
		}

		p.Tags = strings.Split(tags, ", ")

		p.Body = body
	}
	return p, nil
}

func GetPostBySlug(db *sql.DB, postSlug string) (Post, error) {
	log := logrus.New()

	var p Post

	rows, err := db.Query(`
		SELECT id, slug, title, tags,
		postdate, body FROM posts
		WHERE slug = ? LIMIT 1
	`, postSlug)

	if err != nil {
		log.Errorf(
			"Could not load post %s: %v", postSlug, err)
		return p, err
	}

	for rows.Next() {
		// fmt.Printf("%v", row)
		var body string
		var tags string
		var dateStr string

		err = rows.Scan(&p.ID, &p.Slug, &p.Title, &tags, &dateStr, &body)
		if err != nil {
			log.Error(err)
		}
		var date time.Time
		var err error

		date, err = dateparse.ParseAny(dateStr)

		if err != nil {
			log.Errorf("Cannot parse date from %s", dateStr)
			p.PostDate = time.Now()
		} else {
			p.PostDate = date
		}

		p.Tags = strings.Split(tags, ", ")

		p.Body = body
	}
	return p, nil
}

func GetArchiveYearMonths(db *sql.DB) []ArchiveEntry {
	log := logrus.New()

	rows, err := db.Query(`
	SELECT
		STRFTIME('%Y', postdate) postyear,
		STRFTIME('%m', postdate) postmonth,
		COUNT(id) postcount
	FROM
		posts
	GROUP BY
		STRFTIME('%Y', postdate),
		STRFTIME('%m', postdate)
	ORDER BY
		postyear,
		postmonth;
	`)

	if err != nil {
		log.Errorf("Could not load post data: %v", err)
	}

	var archiveData []ArchiveEntry

	for rows.Next() {
		// fmt.Printf("%v", row)
		var archiveEntry ArchiveEntry

		err = rows.Scan(
			&archiveEntry.Year, &archiveEntry.Month, &archiveEntry.Count)

		if err != nil {
			log.Error(err)
		}
		archiveData = append(archiveData, archiveEntry)
	}
	return archiveData
}

func GetArchivePosts(db *sql.DB, year string, month string) []Post {
	log := logrus.New()

	rows, err := db.Query(`
		SELECT id, slug, title, tags,
			postdate, body
		FROM posts
		WHERE strftime("%Y", postdate) = ?
		AND strftime("%m", postdate) = ?
		ORDER BY datetime(postdate) DESC;
	`, year, month)

	if err != nil {
		log.Errorf("Could not load posts: %v", err)
	}

	var posts []Post

	for rows.Next() {
		// fmt.Printf("%v", row)
		var p Post
		var body string
		var tags string
		var dateStr string

		err = rows.Scan(&p.ID, &p.Slug, &p.Title, &tags, &dateStr, &body)
		if err != nil {
			log.Error(err)
		}
		var date time.Time
		var err error

		date, err = dateparse.ParseAny(dateStr)

		if err != nil {
			log.Errorf("Cannot parse date from %s", dateStr)
			p.PostDate = time.Now()
		} else {
			p.PostDate = date
		}

		p.Tags = strings.Split(tags, ", ")

		p.Body = body

		posts = append(posts, p)
	}
	return posts
}

func CreatePost(db *sql.DB, post Post) error {
	log := logrus.New()

	_, err := db.Exec(`
	INSERT into posts (
		slug, title, tags,
		postdate, body
	) VALUES (
		?, ?, ?,
		datetime(), ?
	)
	`, post.Slug, post.Title,
		strings.Join(post.Tags, ","),
		post.Body)

	if err != nil {
		log.Errorf("Could not save post: %v", err)
		return err
	}

	return nil

}

func SavePost(db *sql.DB, post Post) error {
	log := logrus.New()

	_, err := db.Exec(`
	UPDATE posts SET
		title=?, tags=?, body=?
	WHERE id=?
	`, post.Title,
		strings.Join(post.Tags, ","),
		post.Body,
		post.ID)

	if err != nil {
		log.Errorf("Could not save post: %v", err)
		return err
	}

	return nil

}

func DeletePost(db *sql.DB, postID string) error {
	log := logrus.New()

	_, err := db.Exec(`
	DELETE FROM posts WHERE id=?
	`, postID)

	if err != nil {
		log.Errorf("Could not delete post: %v", err)
		return err
	}

	return nil
}

func initDb(dbFile string) {
	createSql := `
	CREATE TABLE IF NOT EXISTS posts (
		id integer primary key,
		slug varchar(256) unique,
		title varchar(1024),
		tags varchar(1024),
		postdate varchar(25),
		body text,
		format varchar(15));
	`
	db, err := GetDb(dbFile)
	if err != nil {
		log.Fatalf("Could not init db at %s: %v", dbFile, err)
	}

	res, err := db.Exec(createSql)
	if err != nil {
		log.Fatalf("Could not init db at %s: %v", dbFile, err)
	}
	log.Debug(res)
}

func checkDb(dbFile string) bool {
	db, err := GetDb(dbFile)
	if err != nil {
		log.Fatalf("Could not check db at %s", dbFile)
	}
	_, err = db.Exec(`SELECT count(*) FROM posts`)
	return err == nil
}
