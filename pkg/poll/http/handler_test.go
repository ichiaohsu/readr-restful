package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
	"github.com/readr-media/readr-restful/pkg/poll/mock"
	"github.com/readr-media/readr-restful/pkg/poll/mysql"
)

func TestGetPolls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockData := mock.NewMockPollData(ctrl)
	mysql.PollAPI = mockData

	gin.SetMode(gin.TestMode)
	r := gin.New()

	Router.SetRoutes(r)

	for _, tc := range []struct {
		name     string
		httpcode int
		path     string
		params   *PollsFilter
		err      string
	}{
		{"default-params", http.StatusOK, `/polls`, &PollsFilter{MaxResult: 20, Page: 1, Sort: "-created_at"}, ``},
		{"all-active", http.StatusOK, `/polls?active=$in:0,1`, &PollsFilter{MaxResult: 20, Page: 1, Sort: "-created_at", Active: map[string][]int{"$in": []int{0, 1}}}, ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.path, nil)

			// expect database service Get to be called once
			// gomock could validate the input argument
			if tc.httpcode == http.StatusOK {
				mockData.EXPECT().Get(tc.params).Times(1)
			}
			// if using count=true, expect Count to be called once
			if tc.name == "count" {
				mockData.EXPECT().Count(tc.params).Times(1)
			}
			r.ServeHTTP(w, req)
			// Check return http status code
			assert.Equal(t, w.Code, tc.httpcode)
			// Check return error if http status is not 200
			if tc.httpcode != http.StatusOK && tc.err != `` {
				assert.Equal(t, w.Body.String(), tc.err)
			}
		})
	}
}
