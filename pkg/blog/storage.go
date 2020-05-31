package blog

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	_ "github.com/mattn/go-sqlite3"
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
	Year       string
	Month      string
	CountPosts int
	CountNotes int
}

var dateFmts = [...]string{
	"2006-01-02T03:04:05",
	"2006-01-02T03:04:05Z",
	"2006-01-02 15:04:05",
}

func GetPosts(db *sql.DB, opts GetPostOpts) []*Post {

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

	sql := "SELECT id, title, slug, postdate, tags, frontmatter, body FROM posts"
	if len(whereColumns) > 0 {
		sql += " WHERE "
		for i, c := range whereColumns {
			sql += fmt.Sprintf("%s like ?", c)
			if i+1 < len(whereColumns) {
				sql += " OR "
			}
		}
	}

	sql += " ORDER BY datetime(postdate) DESC"
	if opts.Offset > 0 {
		sql += " OFFSET ?"
	}
	if opts.Limit == 0 {
		opts.Limit = 100
	}

	if opts.Limit != -1 {
		sql += " LIMIT ?"
		whereValues = append(
			whereValues, fmt.Sprintf("%d", opts.Limit))
	}

	args := make([]interface{}, len(whereValues))
	for i, id := range whereValues {
		args[i] = id
	}
	if opts.Offset > 0 {
		args = append(args, opts.Offset)
	}

	rows, err := db.Query(sql, args...)

	if err != nil {
		logger.Errorf("Could not load posts: %v", err)
		return posts
	}
	if rows.Err() != nil {
		logger.Errorf("Could not load posts: %v", err)
		return posts
	}

	posts = rowsToPosts(rows)

	return posts
}

func GetTaggedPosts(db *sql.DB, tag string) []*Post {

	var posts = make([]*Post, 0)

	var count int
	row := db.QueryRow("SELECT count(*) FROM posts WHERE tags like '%?%'")
	err := row.Scan(&count)

	if err != nil {
		logger.Error(err)
	}

	rows, err := db.Query(`
		SELECT id, title, slug,
			postdate, tags, frontmatter,
			body
		FROM posts
		WHERE tags like ?
		ORDER BY datetime(postdate) DESC
	`, "%"+tag+"%")

	if err != nil {
		logger.Errorf("Could not load posts: %v", err)
	}
	if rows.Err() != nil {
		logger.Error(rows.Err())
	}
	posts = rowsToPosts(rows)

	return posts
}

func GetPost(db *sql.DB, postID string) (*Post, error) {

	var p Post

	rows, err := db.Query(`
		SELECT id, title, slug,
			postdate, tags, frontmatter,
			body
		FROM posts
		WHERE id = ?`, postID)

	if err != nil {
		logger.Errorf("Could not load post %s: %v", postID, err)
		return &p, err
	}

	var posts = make([]*Post, 1)

	posts = rowsToPosts(rows)

	post := posts[0]

	return post, nil
}

func GetPostBySlug(db *sql.DB, postSlug string) (*Post, error) {

	var p Post

	rows, err := db.Query(`
		SELECT
			id,
			title,
			slug,
			postdate,
			tags,
			frontmatter,
			body
	 	FROM posts
		WHERE slug = ? LIMIT 1
	`, postSlug)

	if err != nil {
		logger.Errorf(
			"Could not load post %s: %v", postSlug, err)
		return &p, err
	}

	var posts = make([]*Post, 1)
	posts = rowsToPosts(rows)
	post := posts[0]
	return post, nil
}

func GetArchiveYearMonths(db *sql.DB) []ArchiveEntry {

	rows, err := db.Query(`
	SELECT
		STRFTIME('%Y', postdate) as postyear,
		STRFTIME('%m', postdate) as postmonth,
		SUM(CASE WHEN title != "" THEN 1 ELSE 0 END) AS postcount,
		SUM(CASE WHEN title = "" THEN 1 ELSE 0 END) AS notecount
	FROM
		posts
	GROUP BY
		postyear,
		postmonth
	ORDER BY
		postyear,
		postmonth;
	`)

	if err != nil {
		logger.Errorf("Could not load post data: %v", err)
	}

	var archiveData []ArchiveEntry

	for rows.Next() {
		// fmt.Printf("%v", row)
		var archiveEntry ArchiveEntry

		err = rows.Scan(
			&archiveEntry.Year, &archiveEntry.Month,
			&archiveEntry.CountPosts, &archiveEntry.CountNotes)

		if err != nil {
			logger.Error(err)
		}
		archiveData = append(archiveData, archiveEntry)
	}
	return archiveData
}

func GetArchiveMonthPosts(db *sql.DB, year string, month string) []*Post {

	rows, err := db.Query(`
		SELECT
			id,
			title,
			slug,
			postdate,
			tags,
			frontmatter,
			body
		FROM posts
		WHERE strftime("%Y", postdate) = ?
		AND strftime("%m", postdate) = ?
		ORDER BY datetime(postdate) DESC;
	`, year, month)

	if err != nil {
		logger.Errorf("Could not load posts: %v", err)
	}

	var posts []*Post

	posts = rowsToPosts(rows)

	return posts
}

func GetArchiveDayPosts(
	db *sql.DB, year string, month string, day string) []*Post {

	rows, err := db.Query(`
		SELECT
			id,
			title,
			slug,
			postdate,
			tags,
			frontmatter,
			body
		FROM posts
		WHERE strftime("%Y", postdate) = ?
		AND strftime("%m", postdate) = ?
		AND strftime("%d", postdate) = ?
		ORDER BY datetime(postdate) DESC;
	`, year, month, day)

	if err != nil {
		logger.Errorf("Could not load posts: %v", err)
	}

	var posts []*Post
	posts = rowsToPosts(rows)

	return posts
}

func CreatePost(db *sql.DB, post *Post) error {
	logger.Infof("-- Create new post: %v", post)
	_, err := db.Exec(`
	INSERT into posts (
		slug,
		title,
		tags,
		postdate,
		frontmatter,
		body
	) VALUES (
		?, ?, ?,
		?, ?, ?
	)
	`, post.Slug,
		post.Title,
		post.TagString(),
		post.PostDate.Format(time.RFC3339),
		post.FrontMatterYAML(),
		post.Body)

	if err != nil {
		logger.Errorf("Could not save post: %v", err)
		return err
	}

	p, _ := GetPostBySlug(db, post.Slug)
	logger.Debugf("created post: %v", p)

	return nil

}

func SavePost(db *sql.DB, post *Post) error {
	logger.Infof("-- Save post: %v", post)
	logger.Debugf("-- frontmatter: %v", post.FrontMatterYAML())
	if post.PostDate.IsZero() {
		post.PostDate = time.Now()
	}

	_, err := db.Exec(`
	UPDATE posts SET
		title=?,
		tags=?,
		frontmatter=?,
		body=?,
		postdate=?
	WHERE id=?
	`, post.Title,
		post.TagString(),
		post.FrontMatterYAML(),
		post.Body,
		post.PostDate.Format(time.RFC3339),
		post.ID)

	if err != nil {
		logger.Errorf("Could not save post: %v", err)
		return err
	}

	logger.Debug("saved post, now load for sanity...")
	p, _ := GetPostBySlug(db, post.Slug)

	logger.Debugf("post: %v", p)

	return nil

}

func DeletePost(db *sql.DB, postID string) error {

	_, err := db.Exec(`
	DELETE FROM posts WHERE id=?
	`, postID)

	if err != nil {
		logger.Errorf("Could not delete post: %v", err)
		return err
	}

	return nil
}

// func paginatePosts(db *sql.DB, )

func initDb(dbFile string) error {
	createSql := `
	CREATE TABLE IF NOT EXISTS posts (
		id integer primary key,
		title varchar(1024) default "",
		slug varchar(256) unique,
		postdate varchar(25),
		tags varchar(1024),
		frontmatter text default "",
		body text default "",
		format varchar(15));
	`
	db, err := GetDb(dbFile)
	if err != nil {
		logger.Fatalf("Could not init db at %s: %v", dbFile, err)
		return err
	}

	res, err := db.Exec(createSql)
	if err != nil {
		logger.Fatalf("Could not init db at %s: %v", dbFile, err)
		return err
	}
	logger.Debug(res)
	return nil
}

func checkDb(dbFile string) bool {
	db, err := GetDb(dbFile)
	if err != nil {
		logger.Fatalf("Could not check db at %s", dbFile)
	}
	_, err = db.Exec(`SELECT count(*) FROM posts`)
	return err == nil
}

func rowsToPosts(rows *sql.Rows) []*Post {
	var posts []*Post

	// id,
	// title,
	// slug,
	// postdate,
	// tags,
	// frontmatter,
	// body

	for rows.Next() {
		// fmt.Printf("%v", row)
		var p Post
		var body string
		var tags string
		var dateStr string
		var fmStr string

		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Slug,
			&dateStr,
			&tags,
			&fmStr,
			&body,
		)
		if err != nil {
			logger.Error(err)
		}

		var date time.Time
		date, err = dateparse.ParseAny(dateStr)

		if err != nil {
			logger.Errorf("Cannot parse date from %s", dateStr)
			p.PostDate = time.Now()
		} else {
			p.PostDate = date
		}

		p.Tags = splitTags(tags)
		// logger.Debugf("rowsToPosts frontmatter string: %v", fmStr)
		p.FrontMatter = GetFrontMatter(fmStr)

		p.Body = body

		posts = append(posts, &p)
	}
	return posts
}
