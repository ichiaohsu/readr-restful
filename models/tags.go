package models

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Tag struct {
	ID             int      `json:"id" db:"tag_id"`
	Text           string   `json:"text" db:"tag_content"`
	CreateAt       NullTime `json:"created_at" db:"created_at"`
	UpdateAt       NullTime `json:"updated_at" db:"updated_at"`
	Active         NullInt  `json:"active" db:"active"`
	RelatedReviews NullInt  `json:"related_reviews" db:"related_reviews"`
	RelatedNews    NullInt  `json:"related_news" db:"related_news"`
}

type TagInterface interface {
	ToggleTags(ids []int, active string) error
	GetTags(args GetTagsArgs) ([]Tag, error)
	InsertTag(text string) (int, error)
	UpdateTag(tag Tag) error
	UpdatePostTags(postId int, tag_ids []int) error
	CountTags() (int, error)
}

type GetTagsArgs struct {
	MaxResult uint8  `form:"max_result" json:"max_result"`
	Page      uint16 `form:"page" json:"page"`
	Sorting   string `form:"sort" json:"sort"`
	Keyword   string `form:"keyword" json:"keyword"`
	ShowStats bool   `form:"stats" json:"stats"`
}

func DefaultGetTagsArgs() GetTagsArgs {
	return GetTagsArgs{
		MaxResult: 50,
		Page:      1,
		Sorting:   "-updated_at",
		ShowStats: false,
	}
}

type tagApi struct{}

func (t *tagApi) inCondition(isIn bool) string {
	if isIn {
		return "IN"
	} else {
		return "NOT IN"
	}
}

func (t *tagApi) ToggleTags(ids []int, active string) error {

	query := fmt.Sprintf("UPDATE tags SET active=%s WHERE tag_id IN (?);", strconv.FormatFloat(PostStatus[active].(float64), 'g', 1, 64))
	query, args, err := sqlx.In(query, ids)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	query = DB.Rebind(query)

	_, err = DB.Exec(query, args...)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (t *tagApi) GetTags(args GetTagsArgs) (tags []Tag, err error) {

	var query bytes.Buffer

	if args.ShowStats {
		query.WriteString(fmt.Sprintf(`
			SELECT ta.*, pt.related_reviews, pt.related_news FROM tags as ta 
			LEFT JOIN (SELECT t.tag_id as tag_id,
				COUNT(CASE WHEN p.type=%d THEN p.post_id END) as related_reviews,
				COUNT(CASE WHEN p.type=%d THEN p.post_id END) as related_news 
				FROM post_tags as t LEFT JOIN posts as p ON t.post_id=p.post_id GROUP BY t.tag_id ) as pt 
			ON ta.tag_id = pt.tag_id WHERE ta.active=%d
			`, int(PostType["review"].(float64)), int(PostType["news"].(float64)), int(TagStatus["active"].(float64))))
	} else {
		query.WriteString(fmt.Sprintf(`SELECT ta.* FROM tags as ta WHERE ta.active=%d `, int(TagStatus["active"].(float64))))
	}

	if args.Keyword != "" {
		query.WriteString(` AND ta.tag_content LIKE :keyword`)
		args.Keyword = args.Keyword + "%"
	}

	args.Page = (args.Page - 1) * uint16(args.MaxResult)
	query.WriteString(fmt.Sprintf(` ORDER BY %s LIMIT :maxresult OFFSET :page;`, orderByHelper(args.Sorting)))

	rows, err := DB.NamedQuery(query.String(), args)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	tags = []Tag{}
	for rows.Next() {
		var singleTag Tag
		err = rows.StructScan(&singleTag)
		if err != nil {
			tags = []Tag{}
			log.Println(err.Error())
			return tags, err
		}
		if args.ShowStats {
			if !singleTag.RelatedNews.Valid {
				singleTag.RelatedNews = NullInt{0, true}
			}
			if !singleTag.RelatedReviews.Valid {
				singleTag.RelatedReviews = NullInt{0, true}
			}
		}
		tags = append(tags, singleTag)
	}

	return tags, nil
}

func (t *tagApi) InsertTag(text string) (int, error) {

	query := fmt.Sprintf(`INSERT INTO tags (tag_content) VALUES (?);`)

	result, err := DB.Exec(query, text)
	if err != nil {
		sqlerr, ok := err.(*mysql.MySQLError)
		if ok && sqlerr.Number == 1062 {
			return 0, DuplicateError
		} else {
			return 0, err
		}
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Fail to get last insert ID when insert a tag: %v", err)
	}

	return int(lastID), nil
}

func (t *tagApi) UpdateTag(tag Tag) error {

	dbTags := getStructDBTags("partial", tag)
	fields := makeFieldString("update", `%s = :%s`, dbTags)
	query := fmt.Sprintf(`UPDATE tags SET %s WHERE tag_id = :tag_id`,
		strings.Join(fields, ", "))

	result, err := DB.NamedExec(query, tag)
	if err != nil {
		sqlerr, ok := err.(*mysql.MySQLError)
		if ok && sqlerr.Number == 1062 {
			return DuplicateError
		} else {
			return err
		}
	}

	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return MultipleRowAffectedError
	} else if rowCnt == 0 {
		return ItemNotFoundError
	}

	return nil
}

func (t *tagApi) UpdatePostTags(post_id int, tag_ids []int) error {
	//To add new tags and eliminate unwanted tags, we need to perfom two sql queries
	//The update is success only if all query succeed, to make sure this, we use transaction.

	tx, err := DB.Beginx()
	if err != nil {
		log.Printf("Fail to get sql connection: %v", err)
		return err
	}

	delquery, args, err := sqlx.In(fmt.Sprintf("DELETE FROM post_tags WHERE post_id=%d AND tag_id NOT IN (?);", post_id), tag_ids)
	if err != nil {
		log.Printf("Fail to generate query: %v", err)
		return err
	}

	delquery = DB.Rebind(delquery)

	_ = tx.MustExec(delquery, args...)

	var insqueryBuffer bytes.Buffer
	var insargs []interface{}
	insqueryBuffer.WriteString("INSERT IGNORE INTO post_tags VALUES ")
	for index, tag_id := range tag_ids {
		insqueryBuffer.WriteString("( ? ,? )")
		insargs = append(insargs, post_id, tag_id)
		if index < len(tag_ids)-1 {
			insqueryBuffer.WriteString(",")
		} else {
			insqueryBuffer.WriteString(";")
		}
	}
	_ = tx.MustExec(insqueryBuffer.String(), insargs...)

	tx.Commit()

	return nil
}

func (a *tagApi) CountTags() (result int, err error) {

	err = DB.Get(&result, `SELECT COUNT(*) FROM tags`)
	if err != nil {
		return 0, err
	}

	return result, err
}

var TagStatus map[string]interface{}
var TagAPI TagInterface = new(tagApi)