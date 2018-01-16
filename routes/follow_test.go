package routes

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/readr-media/readr-restful/models"
)

func initFollowTest() {
	mockMemberDSBack = mockMemberDS

	for _, params := range []models.Member{
		models.Member{ID: "followtest0@mirrormedia.mg", Active: models.NullInt{1, true}},
		models.Member{ID: "followtest1@mirrormedia.mg", Active: models.NullInt{1, true}},
		models.Member{ID: "followtest2@mirrormedia.mg", Active: models.NullInt{1, true}},
	} {
		err := models.MemberAPI.InsertMember(params)
		if err != nil {
			log.Fatalf("Init test case fail, aborted. Error: %v", err)
			return
		}
	}

	for _, params := range []models.Post{
		models.Post{ID: 42, Active: 1},
		models.Post{ID: 84, Active: 1},
	} {
		err := models.PostAPI.InsertPost(params)
		if err != nil {
			log.Fatalf("Init test case fail, aborted. Error: %v", err)
			return
		}
	}

	for _, params := range []models.Project{
		models.Project{ID: 420, PostID: 42, Active: models.NullInt{1, true}},
		models.Project{ID: 840, PostID: 84, Active: models.NullInt{1, true}},
	} {
		err := models.ProjectAPI.InsertProject(params)
		if err != nil {
			log.Fatalf("Init test case fail, aborted. Error: %v", err)
			return
		}
	}

	for _, params := range []map[string]string{
		map[string]string{"resource": "member", "subject": "followtest1@mirrormedia.mg", "object": "followtest2@mirrormedia.mg"},
		map[string]string{"resource": "post", "subject": "followtest1@mirrormedia.mg", "object": "42"},
		map[string]string{"resource": "project", "subject": "followtest1@mirrormedia.mg", "object": "420"},
	} {
		err := models.FollowingAPI.AddFollowing(params)
		if err != nil {
			log.Fatalf("Init test case fail, aborted. Error: %v", err)
			return
		}
	}
}

func clearFollowTest() {
	//restore the backuped data
	mockMemberDS = mockMemberDSBack
}

type mockFollowingAPI struct{}

type followDS struct {
	ID     string
	Object string
}

var mockFollowingDS = map[string][]followDS{
	"post":    []followDS{},
	"member":  []followDS{},
	"project": []followDS{},
}

func (a *mockFollowingAPI) AddFollowing(params map[string]string) error {
	store, ok := mockFollowingDS[params["resource"]]
	if !ok {
		log.Fatalln("unexpected error")
	}

	store = append(store, followDS{ID: params["subject"], Object: params["object"]})
	return nil
}

func (a *mockFollowingAPI) DeleteFollowing(params map[string]string) error {
	store, ok := mockFollowingDS[params["resource"]]
	if !ok {
		log.Fatalln("unexpected error")
	}

	for index, follow := range store {
		if follow.ID == params["subject"] && follow.Object == params["object"] {
			store = append(store[:index], store[index+1:]...)
		}
	}
	return nil
}

func (a *mockFollowingAPI) GetFollowing(params map[string]string) (interface{}, error) {
	switch {
	case params["subject"] == "unknown@user.who":
		return nil, errors.New("Not Found")
	case params["resource"] == "member":
		return []models.Member{
			models.Member{ID: "followtest2@mirrormedia.mg", Active: models.NullInt{1, true}},
		}, nil
	case params["resource"] == "post":
		return []models.Post{
			models.Post{ID: 42, Active: 1},
		}, nil
	case params["resource"] == "project":
		return []models.Project{
			models.Project{ID: 420, PostID: 42, Active: models.NullInt{1, true}},
		}, nil
	default:
		return nil, nil
	}
}

func TestFollowingGet(t *testing.T) {

	initFollowTest()

	type CaseIn struct {
		Resource string `json:resource,omitempty`
		Subject  string `json:subject,omitempty`
	}

	type CaseOut struct {
		httpcode int
		resp     string
	}

	var TestRouteName = "/following"
	var TestRouteMethod = "GET"

	var TestFollowingGetCases = []struct {
		name string
		in   CaseIn
		out  CaseOut
	}{
		{"GetFollowingPostOK", CaseIn{"post", "followtest1@mirrormedia.mg"}, CaseOut{http.StatusOK, "[{\"id\":42,\"author\":null,\"created_at\":null,\"liked\":0,\"comment_amount\":0,\"title\":null,\"content\":null,\"link\":null,\"og_title\":null,\"og_description\":null,\"og_image\":null,\"active\":1,\"updated_at\":null,\"updated_by\":null,\"published_at\":null}]"}},
		{"GetFollowingMemberOK", CaseIn{"member", "followtest1@mirrormedia.mg"}, CaseOut{http.StatusOK, "[{\"id\":\"followtest2@mirrormedia.mg\",\"name\":null,\"nickname\":null,\"birthday\":null,\"gender\":null,\"occupation\":null,\"mail\":null,\"register_mode\":null,\"social_id\":null,\"created_at\":null,\"updated_at\":null,\"updated_by\":null,\"description\":null,\"profile_image\":null,\"identity\":null,\"role\":null,\"active\":1,\"custom_editor\":null,\"hide_profile\":null,\"profile_push\":null,\"post_push\":null,\"comment_push\":null}]"}},
		{"GetFollowingProjectOK", CaseIn{"project", "followtest1@mirrormedia.mg"}, CaseOut{http.StatusOK, "[{\"id\":420,\"created_at\":null,\"updated_at\":null,\"updated_by\":null,\"published_at\":null,\"post_id\":42,\"like_amount\":null,\"comment_amount\":null,\"active\":1,\"hero_image\":null,\"title\":null,\"description\":null,\"author\":null,\"og_title\":null,\"og_description\":null,\"og_image\":null}]"}},
		{"GetFollowingFollowerNotExist", CaseIn{"project", "unknown@user.who"}, CaseOut{http.StatusNotFound, `{"Error":"Not Found"}`}},
	}

	for _, testcase := range TestFollowingGetCases {

		req, _ := http.NewRequest(TestRouteMethod, TestRouteName+"/"+testcase.in.Subject+"/"+testcase.in.Resource, nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != testcase.out.httpcode {
			t.Errorf("Want %d but get %d, testcase %s", testcase.out.httpcode, w.Code, testcase.name)
			t.Fail()
		}

		if w.Body.String() != testcase.out.resp {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.resp, w.Body.String(), testcase.name)
			t.Fail()
		}

	}
	clearFollowTest()
}

func TestFollowingAddDelete(t *testing.T) {

	type PubsubWrapperMessage struct {
		ID   string   `json:"messageId"`
		Attr []string `json:"attributes"`
		Body []byte   `json:"data"`
	}

	type PubsubWrapper struct {
		Subscription string               `json:"subscription"`
		Message      PubsubWrapperMessage `json:"message"`
	}

	type CaseIn struct {
		Action   string `json:action,omitempty`
		Resource string `json:resource,omitempty`
		Subject  string `json:subject,omitempty`
		Object   string `json:object,omitempty`
	}

	type CaseOut struct {
		httpcode int
		Error    string
	}

	var TestRouteName = "/api/pubsub"
	var TestRouteMethod = "POST"

	var TestFollowingGetCases = []struct {
		name string
		in   CaseIn
		out  CaseOut
	}{
		{"AddFollowingPostOK", CaseIn{"follow", "post", "followtest0@mirrormedia.mg", "84"}, CaseOut{http.StatusOK, ""}},
		{"AddFollowingMemberOK", CaseIn{"follow", "member", "followtest0@mirrormedia.mg", "followtest2@mirrormedia.mg"}, CaseOut{http.StatusOK, ""}},
		{"AddFollowingProjectOK", CaseIn{"follow", "project", "followtest0@mirrormedia.mg", "840"}, CaseOut{http.StatusOK, ""}},
		{"AddFollowingMissingResource", CaseIn{"follow", "", "followtest0@mirrormedia.mg", "followtest2@mirrormedia.mg"}, CaseOut{http.StatusOK, `{"Error":"Bad Request"}`}},
		{"AddFollowingMissingAction", CaseIn{"", "member", "followtest0@mirrormedia.mg", "followtest2@mirrormedia.mg"}, CaseOut{http.StatusOK, `{"Error":"Bad Request"}`}},
		{"AddFollowingWrongIDForPost", CaseIn{"follow", "post", "followtest0@mirrormedia.mg", "zexal"}, CaseOut{http.StatusOK, `{"Error":"Bad Request"}`}},
		{"DeleteFollowingPostOK", CaseIn{"unfollow", "post", "followtest0@mirrormedia.mg", "84"}, CaseOut{http.StatusOK, ""}},
		{"DeleteFollowingMemberOK", CaseIn{"unfollow", "member", "followtest0@mirrormedia.mg", "followtest2@mirrormedia.mg"}, CaseOut{http.StatusOK, ""}},
		{"DeleteFollowingProjectOK", CaseIn{"unfollow", "project", "followtest0@mirrormedia.mg", "840"}, CaseOut{http.StatusOK, ""}},
	}

	for _, testcase := range TestFollowingGetCases {
		bodyJsonStr, err := json.Marshal(&testcase.in)
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}

		jsonStr, err := json.Marshal(&PubsubWrapper{"subs", PubsubWrapperMessage{"1", []string{"1"}, bodyJsonStr}})
		if err != nil {
			t.Errorf("Fail to marshal input objects, error: %v", err.Error())
			t.Fail()
		}

		req, _ := http.NewRequest(TestRouteMethod, TestRouteName, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Body.String() != testcase.out.Error {
			t.Errorf("Expect get error message %v but get %v, testcase %s", testcase.out.Error, w.Body.String(), testcase.name)
			t.Fail()
		}
	}

	clearFollowTest()
}
