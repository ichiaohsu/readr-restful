package poll

import "github.com/readr-media/readr-restful/models"

const PollDefaultActive int64 = 1

// Poll is the struct mapping to table polls
type Poll struct {
	ID          int64             `json:"id" db:"id"`
	Status      int64             `json:"status" db:"status"`
	Active      int64             `json:"active" db:"active"`
	Title       models.NullString `json:"title" db:"title"`
	Description models.NullString `json:"description" db:"description"`
	TotalVote   int64             `json:"total_vote" db:"total_vote"`
	Frequency   models.NullInt    `json:"frequency" db:"frequency"`
	StartAt     models.NullTime   `json:"start_at" db:"start_at"`
	EndAt       models.NullTime   `json:"end_at" db:"end_at"`
	MaxChoice   int64             `json:"max_choice" db:"max_choice"`
	Changeable  int64             `json:"changeable" db:"changeable"`
	PublishedAt models.NullTime   `json:"published_at" db:"published_at"`
	CreatedAt   models.NullTime   `json:"created_at" db:"created_at"`
	CreatedBy   models.NullInt    `json:"created_by" db:"created_by"`
	UpdatedAt   models.NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy   models.NullInt    `json:"updated_by" db:"updated_by"`
}

// Choice is the struct mapping to table polls_choices
// storing choice data for each poll
type Choice struct {
	ID         int64             `json:"id" db:"id"`
	Choice     models.NullString `json:"choice" db:"choice"`
	TotalVote  models.NullInt    `json:"total_vote" db:"total_vote"`
	PollID     models.NullInt    `json:"poll_id" db:"poll_id"`
	Active     models.NullInt    `json:"active" db:"active"`
	GroupOrder models.NullInt    `json:"group_order" db:"group_order"`
	CreatedAt  models.NullTime   `json:"created_at" db:"created_at"`
	CreatedBy  models.NullInt    `json:"created_by" db:"created_by"`
	UpdatedAt  models.NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy  models.NullInt    `json:"updated_by" db:"updated_by"`
}

// Pick is the mapping struct for table polls_chosen_choice
// to record the choosing history for every users
// type Pick struct {
// 	ID        int64           `json:"id" db:"id"`
// 	MemberID  int64           `json:"member_id" db:"member_id"`
// 	PollID    int64           `json:"poll_id" db:"poll_id"`
// 	ChoiceID  int64           `json:"choice_id" db:"choice_id"`
// 	Active    bool            `json:"active" db:"active"`
// 	CreatedAt models.NullTime `json:"created_at" db:"created_at"`
// }

type PollResponse struct {
	Poll
	CreatedBy models.Stunt `json:"created_by" db:"created_by"`
	Choices   []Choice     `json:"choices,omitempty" db:"choices"`
}

type PollRequest struct {
	Poll
	Choices []Choice `json:"choices,omitempty" db:"choices"`
}

type PollParams interface {
	Parse()
	Select() (string, []interface{}, error)
	Count() (string, []interface{}, error)
}

// PollData provides database interface for dependency injection and testing
//go:generate mockgen -package=mock -destination=mock/mock.go github.com/readr-media/readr-restful/pkg/poll PollData
type PollData interface {
	Get(params PollParams) (results []PollResponse, err error)
	Insert(p PollRequest) (err error)
	Update(p Poll) (err error)
	Count(params PollParams) (count int, err error)
}
