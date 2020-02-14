package payment

type Provider interface {
	Pay() (err error)
	Token() (Provider, error)
}

func NewRecurringProvider(name string) (p Provider, err error) {
	switch name {
	case "tappay":
		fallthrough
	default:
		// default using tappay
		p = &PayByCardToken{}
		err = nil
	}
	return p, err
}

func NewDisposableProvider(name string) (p Provider, err error) {
	switch name {
	case "tappay":
		fallthrough
	default:
		// default using tappay
		p = &PayByPrime{}
		err = nil
	}
	return p, err
}
