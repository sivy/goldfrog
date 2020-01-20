package blog

import (
	"database/sql"
	"fmt"
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
	Tags   []string
	Body   string
	Offset int
	Limit  int
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

func GetPosts(db *sql.DB, opts GetPostOpts) []*Post {
	log := logrus.New()

	var whereClauses = make(map[string]string)

	if opts.Title != "" {
		whereClauses["title"] = "%" + opts.Title + "%"
	}
	if opts.Body != "" {
		whereClauses["body"] = "%" + opts.Body + "%"
	}

	var whereColumns []string
	var whereValues []string

	for column, value := range whereClauses {
		whereColumns = append(whereColumns, column)
		whereValues = append(whereValues, value)
	}

	var posts = make([]*Post, 0)

	sql := "SELECT id, slug, title, tags, postdate, body FROM posts"
	if len(whereColumns) > 0 {
		sql += " WHERE "
		for i, c := range whereColumns {
			sql += fmt.Sprintf("%s like ?", c)
			if i+1 < len(whereColumns) {
				sql += " OR "
			}
		}
	}
	sql += " ORDER BY datetime(postdate) DESC LIMIT ?"
	if opts.Offset > 0 {
		sql += " OFFSET ?"
	}
	if opts.Limit == 0 {
		opts.Limit = 100
	}

	whereValues = append(
		whereValues, fmt.Sprintf("%d", opts.Limit))

	args := make([]interface{}, len(whereValues))
	for i, id := range whereValues {
		args[i] = id
	}
	if opts.Offset > 0 {
		args = append(args, opts.Offset)
	}

	rows, err := db.Query(sql, args...)

	if err != nil {
		log.Errorf("Could not load posts: %v", err)
		return posts
	}
	if rows.Err() != nil {
		log.Errorf("Could not load posts: %v", err)
		return posts
	}

	posts = rowsToPosts(rows)

	return posts
}

func GetTaggedPosts(db *sql.DB, tag string) []*Post {
	log := logrus.New()
	var posts = make([]*Post, 0)

	var count int
	row := db.QueryRow("SELECT count(*) FROM posts WHERE tags like '%?%'")
	err := row.Scan(&count)

	if err != nil {
		log.Error(err)
	}

	rows, err := db.Query(`
		SELECT id, slug, title, tags, postdate, body FROM posts
		WHERE tags like ?
		ORDER BY datetime(postdate) DESC
	`, "%"+tag+"%")

	if err != nil {
		log.Errorf("Could not load posts: %v", err)
	}
	if rows.Err() != nil {
		log.Error(rows.Err())
	}
	posts = rowsToPosts(rows)

	return posts
}

func GetPost(db *sql.DB, postID string) (*Post, error) {
	log := logrus.New()

	var p Post

	rows, err := db.Query(`
		SELECT id, slug, title, tags,
		postdate, body FROM posts
		WHERE id = ?`, postID)

	if err != nil {
		log.Errorf("Could not load post %s: %v", postID, err)
		return &p, err
	}

	var posts = make([]*Post, 1)

	posts = rowsToPosts(rows)

	post := posts[0]

	return post, nil
}

func GetPostBySlug(db *sql.DB, postSlug string) (*Post, error) {
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
		return &p, err
	}

	var posts = make([]*Post, 1)
	posts = rowsToPosts(rows)
	post := posts[0]
	return post, nil
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

func GetArchiveMonthPosts(db *sql.DB, year string, month string) []*Post {
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

	var posts []*Post

	posts = rowsToPosts(rows)

	return posts
}

func GetArchiveDayPosts(db *sql.DB, year string, month string, day string) []*Post {
	log := logrus.New()

	rows, err := db.Query(`
		SELECT id, slug, title, tags,
			postdate, body
		FROM posts
		WHERE strftime("%Y", postdate) = ?
		AND strftime("%m", postdate) = ?
		AND strftime("%d", postdate) = ?
		ORDER BY datetime(postdate) DESC;
	`, year, month, day)

	if err != nil {
		log.Errorf("Could not load posts: %v", err)
	}

	var posts []*Post
	posts = rowsToPosts(rows)

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

func SavePost(db *sql.DB, post *Post) error {
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

// func paginatePosts(db *sql.DB, )

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

func rowsToPosts(rows *sql.Rows) []*Post {
	var posts []*Post

	for rows.Next() {
		// fmt.Printf("%v", row)
		var p Post
		var body string
		var tags string
		var dateStr string

		err := rows.Scan(&p.ID, &p.Slug, &p.Title, &tags, &dateStr, &body)
		if err != nil {
			log.Error(err)
		}
		var date time.Time

		date, err = dateparse.ParseAny(dateStr)

		if err != nil {
			log.Errorf("Cannot parse date from %s", dateStr)
			p.PostDate = time.Now()
		} else {
			p.PostDate = date
		}

		p.Tags = splitTags(tags)

		p.Body = body

		posts = append(posts, &p)
	}
	return posts
}
