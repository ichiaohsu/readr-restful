package routes

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/config"
	"github.com/readr-media/readr-restful/models"
	"github.com/readr-media/readr-restful/pkg/mail"
	"github.com/readr-media/readr-restful/utils"
)

type reportHandler struct {
}

func (r *reportHandler) bindQuery(c *gin.Context, args *models.GetReportArgs) (err error) {

	// Start parsing rest of request arguments
	if c.Query("active") != "" {
		if err = json.Unmarshal([]byte(c.Query("active")), &args.Active); err != nil {
			log.Println(err.Error())
			return err
		} else if err == nil {
			// if err = models.ValidateActive(args.Active, models.ReportActive); err != nil {
			if err = models.ValidateActive(args.Active, config.Config.Models.Reports); err != nil {
				return err
			}
		}
	}
	if c.Query("report_publish_status") != "" {
		if err = json.Unmarshal([]byte(c.Query("report_publish_status")), &args.ReportPublishStatus); err != nil {
			log.Println(err.Error())
			return err
		} else if err == nil {
			// if err = models.ValidateActive(args.PublishStatus, models.ProjectPublishStatus); err != nil {
			if err = models.ValidateActive(args.ReportPublishStatus, config.Config.Models.ReportsPublishStatus); err != nil {
				return err
			}
		}
	}
	if c.Query("project_publish_status") != "" {
		if err = json.Unmarshal([]byte(c.Query("project_publish_status")), &args.ProjectPublishStatus); err != nil {
			log.Println(err.Error())
			return err
		} else if err == nil {
			// if err = models.ValidateActive(args.PublishStatus, models.ProjectPublishStatus); err != nil {
			if err = models.ValidateActive(args.ProjectPublishStatus, config.Config.Models.ProjectsPublishStatus); err != nil {
				return err
			}
		}
	}
	if c.Query("ids") != "" {
		if err = json.Unmarshal([]byte(c.Query("ids")), &args.IDs); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if c.Query("report_slugs") != "" {
		if err = json.Unmarshal([]byte(c.Query("report_slugs")), &args.Slugs); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if c.Query("project_slugs") != "" {
		if err = json.Unmarshal([]byte(c.Query("project_slugs")), &args.ProjectSlugs); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if c.Query("project_id") != "" {
		if err = json.Unmarshal([]byte(c.Query("project_id")), &args.Project); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if c.Query("max_result") != "" {
		if err = json.Unmarshal([]byte(c.Query("max_result")), &args.MaxResult); err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if c.Query("page") != "" {
		if err = json.Unmarshal([]byte(c.Query("page")), &args.Page); err != nil {
			log.Println(err.Error())
			return err
		}
	}

	if c.Query("sort") != "" && r.validateReportSorting(c.Query("sort")) {
		args.Sorting = c.Query("sort")
	}

	if c.Query("keyword") != "" {
		args.Keyword = c.Query("keyword")
	}
	if c.Query("fields") != "" {
		if err = json.Unmarshal([]byte(c.Query("fields")), &args.Fields); err != nil {
			log.Println(err.Error())
			return err
		}
		for _, field := range args.Fields {
			if !r.validate(field, fmt.Sprintf("^(%s)$", strings.Join(args.FullAuthorTags(), "|"))) {
				return errors.New("Invalid Fields")
			}
		}
	} else {
		switch c.Query("mode") {
		case "full":
			args.Fields = args.FullAuthorTags()
		default:
			args.Fields = []string{"nickname"}
		}
	}
	return nil
}

func (r *reportHandler) Count(c *gin.Context) {
	var args = models.GetReportArgs{}
	args.Default()
	if err := r.bindQuery(c, &args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	count, err := models.ReportAPI.CountReports(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	resp := map[string]int{"total": count}
	c.JSON(http.StatusOK, gin.H{"_meta": resp})
}

func (r *reportHandler) Get(c *gin.Context) {
	var args = models.GetReportArgs{}
	args.Default()

	if err := r.bindQuery(c, &args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if args.Active == nil {
		args.DefaultActive()
	}
	reports, err := models.ReportAPI.GetReports(args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"_items": reports})
}

func (r *reportHandler) Post(c *gin.Context) {

	report := models.Report{}
	err := c.ShouldBind(&report)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report"})
		return
	}

	// Pre-request test
	if report.Title.Valid == false {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report"})
		return
	}

	if report.PublishStatus.Valid == true && report.PublishStatus.Int == int64(config.Config.Models.ReportsPublishStatus["publish"]) && report.Slug.Valid == false {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Must Have Slug Before Publish"})
		return
	}

	if !report.CreatedAt.Valid {
		report.CreatedAt = models.NullTime{time.Now(), true}
	}
	report.UpdatedAt = models.NullTime{time.Now(), true}
	report.Active = models.NullInt{int64(config.Config.Models.Reports["active"]), true}

	lastID, err := models.ReportAPI.InsertReport(report)
	if err != nil {
		switch err.Error() {
		case "Duplicate entry":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Report Already Existed"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}

	if report.PublishStatus.Valid && report.PublishStatus.Int == int64(config.Config.Models.ReportsPublishStatus["publish"]) {
		r.PublishHandler([]int{lastID})
		if report.UpdatedBy.Valid {
			r.UpdateHandler([]int{lastID}, report.UpdatedBy.Int)
		} else {
			r.UpdateHandler([]int{lastID})
		}
	}

	resp := map[string]int{"last_id": lastID}
	c.JSON(http.StatusOK, gin.H{"_items": resp})
}

func (r *reportHandler) Put(c *gin.Context) {

	report := models.Report{}
	err := c.ShouldBind(&report)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report Data"})
		return
	}

	if report.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report Data"})
		return
	}

	if report.Active.Valid == true && !r.validateReportStatus(report.Active.Int) {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameter"})
		return
	}

	// if report.PublishStatus.Valid == true && (report.PublishStatus.Int == int64(models.ReportPublishStatus["publish"].(float64)) || report.PublishStatus.Int == int64(models.ReportPublishStatus["schedule"].(float64))) {
	if report.PublishStatus.Valid == true && (report.PublishStatus.Int == int64(config.Config.Models.ReportsPublishStatus["publish"]) || report.PublishStatus.Int == int64(config.Config.Models.ReportsPublishStatus["schedule"])) {
		p, err := models.ReportAPI.GetReport(report)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Report Not Found"})
			return
		} else if p.Slug.Valid == false {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Must Have Slug Before Publish"})
			return
		}

		switch p.PublishStatus.Int {
		// case int64(models.ReportPublishStatus["schedule"].(float64)):
		case int64(config.Config.Models.ReportsPublishStatus["schedule"]):
			if !report.PublishedAt.Valid && !p.PublishedAt.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Publish Time"})
				return
			}
			fallthrough
		// case int64(models.ReportPublishStatus["publish"].(float64)):
		case int64(config.Config.Models.ReportsPublishStatus["publish"]):
			if !report.Title.Valid && !p.Title.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report Title"})
				return
			}
			if !report.Slug.Valid && !p.Slug.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Report Content"})
				return
			}
			if !report.PublishedAt.Valid {
				report.PublishedAt = models.NullTime{Time: time.Now(), Valid: true}
			}
			break
		}
	}

	if report.CreatedAt.Valid {
		report.CreatedAt.Valid = false
	}
	report.UpdatedAt = models.NullTime{time.Now(), true}

	err = models.ReportAPI.UpdateReport(report)
	if err != nil {
		switch err.Error() {
		case "Report Not Found":
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Report Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}

	if (report.PublishStatus.Valid && report.PublishStatus.Int != int64(config.Config.Models.ReportsPublishStatus["publish"])) ||
		(report.Active.Valid && report.Active.Int != int64(config.Config.Models.Reports["active"])) {
		// Case: Set a report to unpublished state, Delete the report from cache/searcher
		go models.Algolia.DeleteReport([]int{report.ID})
	} else if report.PublishStatus.Valid || report.Active.Valid {
		// Case: Publish a report or update a report.
		if report.PublishStatus.Int == int64(config.Config.Models.ReportsPublishStatus["publish"]) ||
			report.Active.Int == int64(config.Config.Models.Reports["active"]) {
			r.PublishHandler([]int{report.ID})
			if report.UpdatedBy.Valid {
				r.UpdateHandler([]int{report.ID}, report.UpdatedBy.Int)
			} else {
				r.UpdateHandler([]int{report.ID})
			}
		}
	} else {
		r.UpdateHandler([]int{report.ID})
	}
	c.Status(http.StatusOK)
}

func (r *reportHandler) Delete(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "ID Must Be Integer"})
		return
	}
	input := models.Report{ID: id}
	err = models.ReportAPI.DeleteReport(input)

	if err != nil {
		switch err.Error() {
		case "Report Not Found":
			c.JSON(http.StatusNotFound, gin.H{"Error": "Report Not Found"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal Server Error"})
			return
		}
	}

	go models.Algolia.DeleteReport([]int{id})

	c.Status(http.StatusOK)
}

// func (r *projectHandler) GetAuthors(c *gin.Context) {
// 	//project/authors?ids=[1000010,1000013]&mode=[full]&fields=["id","member_id"]
// 	args := models.GetProjectArgs{}
// 	if err := r.bindQuery(c, &args); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
// 		return
// 	}
// 	fmt.Println(args)
// 	authors, err := models.ProjectAPI.GetAuthors(args)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"_items": authors})
// }

func (r *reportHandler) PostAuthors(c *gin.Context) {
	params := struct {
		ReportID  *int  `json:"report_id"`
		AuthorIDs []int `json:"author_ids"`
	}{}
	err := c.ShouldBind(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameters"})
		return
	}

	if params.ReportID == nil || params.AuthorIDs == nil || len(params.AuthorIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Insufficient Parameters"})
		return
	}
	if err := models.ReportAPI.InsertAuthors(*params.ReportID, params.AuthorIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
}

func (r *reportHandler) PutAuthors(c *gin.Context) {
	params := struct {
		ReportID  *int  `json:"report_id"`
		AuthorIDs []int `json:"author_ids"`
	}{}
	err := c.ShouldBind(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid Parameters"})
		return
	}

	if params.ReportID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Insufficient Parameters"})
		return
	}
	if err := models.ReportAPI.UpdateAuthors(*params.ReportID, params.AuthorIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
}

func (r *reportHandler) PublishHandler(ids []int) error {
	// Redis notification
	// Mail notification

	if len(ids) == 0 {
		return nil
	}

	args := models.GetReportArgs{
		IDs:                 ids,
		Active:              map[string][]int{"IN": []int{int(config.Config.Models.Reports["active"])}},
		ReportPublishStatus: map[string][]int{"IN": []int{int(config.Config.Models.ReportsPublishStatus["publish"])}},
	}
	args.Fields = args.FullAuthorTags()

	reports, err := models.ReportAPI.GetReports(args)
	if err != nil {
		log.Println("Getting reports info fail when running publish handling process", err)
		return err
	}
	if len(reports) == 0 {
		return nil
	}

	for _, report := range reports {
		p := models.Project{ID: report.ProjectID, UpdatedAt: models.NullTime{Time: time.Now(), Valid: true}}
		if report.UpdatedBy.Valid {
			p.UpdatedBy = report.UpdatedBy
		}
		err := models.ProjectAPI.UpdateProjects(p)
		if err != nil {
			return err
		}
	}

	go models.Algolia.InsertReport(reports)
	for _, report := range reports {
		go models.NotificationGen.GenerateProjectNotifications(report, "report")
		go mail.MailAPI.SendReportPublishMail(report)
		go r.insertReportPost(report)
	}

	return nil
}

func (r *reportHandler) UpdateHandler(ids []int, params ...int64) error {
	// update update time for projects

	if len(ids) == 0 {
		return nil
	}

	args := models.GetReportArgs{
		IDs:                 ids,
		Active:              map[string][]int{"IN": []int{int(config.Config.Models.Reports["active"])}},
		ReportPublishStatus: map[string][]int{"IN": []int{int(config.Config.Models.ReportsPublishStatus["publish"])}},
	}
	args.Fields = args.FullAuthorTags()

	reports, err := models.ReportAPI.GetReports(args)
	if err != nil {
		log.Println("Getting reports info fail when running publish handling process", err)
		return err
	}
	if len(reports) == 0 {
		return nil
	}

	for _, report := range reports {
		p := models.Project{ID: report.Project.ID, UpdatedAt: models.NullTime{Time: time.Now(), Valid: true}}
		if len(params) > 0 {
			p.UpdatedBy = models.NullInt{Int: params[0], Valid: true}
		}
		go models.ProjectAPI.UpdateProjects(p)
		go r.updateReportPost(report)
	}
	return nil
}

func (r *reportHandler) insertReportPost(report models.ReportAuthors) {
	linkUrl := utils.GenerateResourceInfo("report", report.ID, report.Slug.String)
	posts, err := r.getPostByReportUrl(linkUrl)
	if err != nil {
		fmt.Printf("Fail to fetching post by link %s, error occored: %v", linkUrl, err.Error())
		return
	}
	if len(posts) > 0 {
		fmt.Printf("Fail to insert a report post: %s, duplicated", linkUrl)
		return
	}
	postID, err := models.PostAPI.InsertPost(models.Post{
		Type:          models.NullInt{int64(config.Config.Models.PostType["report"]), true},
		Title:         report.Report.Title,
		Content:       report.Report.Description,
		Link:          models.NullString{linkUrl, true},
		LinkTitle:     report.Report.Title,
		LinkImage:     report.Report.HeroImage,
		OgTitle:       report.Report.OgTitle,
		OgDescription: report.Report.OgDescription,
		OgImage:       report.Report.OgImage,
		Active:        models.NullInt{int64(config.Config.Models.Posts["active"]), true},
		PublishStatus: models.NullInt{int64(config.Config.Models.PostPublishStatus["publish"]), true},
		Author:        models.NullInt{int64(config.Config.ReadrID), true},
		CreatedAt:     models.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt:     models.NullTime{Time: time.Now(), Valid: true},
		PublishedAt:   models.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		fmt.Printf("Fail to isnert post for new report %s , error: %v", linkUrl, err.Error())
		return
	}
	go PostHandler.PublishPipeline([]uint32{uint32(postID)})
}

func (r *reportHandler) updateReportPost(report models.ReportAuthors) {
	linkUrl := utils.GenerateResourceInfo("report", report.ID, report.Slug.String)
	posts, err := r.getPostByReportUrl(linkUrl)
	if err != nil {
		fmt.Printf("Fail to fetching post by link %s, error occored: %v", linkUrl, err.Error())
		return
	}
	if len(posts) == 0 {
		fmt.Printf("Fail to update a report post: %s, post not exist", linkUrl)
		return
	}
	err = models.PostAPI.UpdatePost(models.Post{
		ID:            uint32(posts[0].Post.ID),
		Title:         report.Report.Title,
		Content:       report.Report.Description,
		LinkImage:     report.Report.HeroImage,
		OgTitle:       report.Report.OgTitle,
		OgDescription: report.Report.OgDescription,
		OgImage:       report.Report.OgImage,
		Active:        report.Report.Active,
		PublishStatus: report.Report.PublishStatus,
		UpdatedAt:     models.NullTime{Time: time.Now(), Valid: true},
		UpdatedBy:     models.NullInt{int64(config.Config.ReadrID), true},
	})
	if err != nil {
		fmt.Printf("Fail to update post for report %s , error: %v", linkUrl, err.Error())
		return
	}
}

func (r *reportHandler) getPostByReportUrl(linkUrl string) ([]models.TaggedPostMember, error) {
	return models.PostAPI.GetPosts(&models.PostArgs{
		Type:      map[string][]int{"in": []int{config.Config.Models.PostType["report"]}},
		Sorting:   "updated_at",
		Page:      1,
		MaxResult: 1,
		Filter: models.Filter{
			Field:     "link",
			Operator:  "=",
			Condition: linkUrl,
		},
	})
}

func (r *reportHandler) SetRoutes(router *gin.Engine) {
	reportRouter := router.Group("/report")
	{
		reportRouter.GET("/count", r.Count)
		reportRouter.GET("/list", r.Get)
		reportRouter.POST("", r.Post)
		reportRouter.PUT("", r.Put)
		reportRouter.DELETE("/:id", r.Delete)

		authorRouter := reportRouter.Group("/author")
		{
			authorRouter.POST("", r.PostAuthors)
			authorRouter.PUT("", r.PutAuthors)
		}
	}
}

func (r *reportHandler) validateReportStatus(i int64) bool {
	// for _, v := range models.ReportActive {
	for _, v := range config.Config.Models.Reports {
		// if i == int64(v.(float64)) {
		if i == int64(v) {
			return true
		}
	}
	return false
}
func (r *reportHandler) validateReportSorting(sort string) bool {
	for _, v := range strings.Split(sort, ",") {
		if matched, err := regexp.MatchString("-?(updated_at|published_at|id|slug|views|comment_amount)", v); err != nil || !matched {
			return false
		}
	}
	return true
}

func (r *reportHandler) validate(target string, paradigm string) bool {
	if matched, err := regexp.MatchString(paradigm, target); err != nil || !matched {
		return false
	}
	return true
}

var ReportHandler reportHandler
