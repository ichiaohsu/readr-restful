package mysql

import "github.com/readr-media/readr-restful/pkg/poll"

type pollAPI struct{}

var PollAPI poll.PollData = new(pollAPI)

func (p *pollAPI) Get(params poll.PollParams) (results []poll.PollResponse, err error) {
	panic("not implemented")
}

func (p *pollAPI) Insert(poll poll.PollRequest) (err error) {
	panic("not implemented")
}

func (p *pollAPI) Update(poll poll.Poll) (err error) {
	panic("not implemented")
}

func (p *pollAPI) Count(params poll.PollParams) (count int, err error) {
	panic("not implemented")
}
