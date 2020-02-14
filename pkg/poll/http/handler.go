package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	rt "github.com/readr-media/readr-restful/internal/router"
	"github.com/readr-media/readr-restful/pkg/poll"
	"github.com/readr-media/readr-restful/pkg/poll/mysql"
)

type Handler struct{}

// GetPolls returns poll list to clients
func (h *Handler) GetPolls(c *gin.Context) {

	// Create a PollsFilter, set defaults, and bind to gin.Context
	filter, err := NewPollsFilter(
		func(f *PollsFilter) error {
			f.MaxResult = 20
			f.Page = 1
			f.Sort = "-created_at"
			// var defaultActive = poll.PollDefaultActive
			// f.Active = &defaultActive

			return nil
		}, func(c *gin.Context) func(*PollsFilter) error {
			// Wrap filter's Bind function
			return func(f *PollsFilter) (err error) {
				if err = f.Bind(c); err != nil {
					return err
				}
				return nil
			}
		}(c),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	fmt.Printf("GetPolls filter:%v\n", filter)

	var results struct {
		Items []poll.PollResponse `json:"_items"`
		Meta  *rt.ResponseMeta    `json:"_meta,omitempty"`
	}
	results.Items, err = mysql.PollAPI.Get(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// SetRoutes is used to register APIs in routes package
func (h *Handler) SetRoutes(router *gin.Engine) {

	pollRouter := router.Group("/polls")
	{
		pollRouter.GET("", h.GetPolls)
	}
}

// Router is the explicit interface for router registering
var Router Handler
