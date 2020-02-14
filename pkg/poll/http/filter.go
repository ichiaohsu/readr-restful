package http

import (
	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/pkg/poll/mysql"
)

// PollsFilter is used to bind query parameters for poll listing
type PollsFilter struct {
	MaxResult int    `form:"max_result"`
	Page      int    `form:"page"`
	Sort      string `form:"sort"`
	Embed     string `form:"embed"`

	Status  string `form:"status"`
	StartAt string `form:"start_at"`
	IDS     string `form:"ids"`
	// Active *int64 `form:"active"`

	Active map[string][]int
	o      *mysql.SQLO
}

func SetPollsFilter(f *PollsFilter, options ...func(*PollsFilter) (err error)) (err error) {

	for _, option := range options {
		if err := option(f); err != nil {
			return err
		}
	}
	return nil
}

func NewPollsFilter(options ...func(*PollsFilter) (err error)) (*PollsFilter, error) {

	params := PollsFilter{}
	err := SetPollsFilter(&params, options...)
	return &params, err
}

func (f *PollsFilter) Parse() {
	panic("not implemented")
}

func (f *PollsFilter) Select() (string, []interface{}, error) {
	panic("not implemented")
}

func (f *PollsFilter) Count() (string, []interface{}, error) {
	panic("not implemented")
}

func (f *PollsFilter) Bind(c *gin.Context) error {

	if err := c.ShouldBindQuery(f); err != nil {
		return err
	}
	return nil
}
