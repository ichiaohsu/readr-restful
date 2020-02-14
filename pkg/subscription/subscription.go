package subscription

import (
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/pkg/payment"
)

const (
	// StatusInit means subscriptions is init. It's different from zero-value of Go
	StatusInit = iota + 1
	// StatusOK means it is a functioning subscription
	StatusOK
	// StatusInactive indicates this is not an active subscription now
	StatusInactive
	// StatusInitPayFail denotes fail when pay for the first time
	StatusInitPayFail
	// StatusRoutinePayFail indicates failure when pays with token
	StatusRoutinePayFail
)

type Payer interface {
	Pay() (err error)
}

type Invoicer interface {
	Create() (resp map[string]interface{}, err error)
	Validate() error
}

// Subscriber provides the interface for different db backend
//go:generate mockgen -package=mock -destination=test/mock/mock.go github.com/readr-media/readr-restful/pkg/subscription Subscriber
type Subscriber interface {
	GetSubscriptions(f ListFilter) (results []Subscription, err error)
	CreateSubscription(s Subscription) error
	UpdateSubscriptions(s Subscription) error
	RoutinePay(subscribers []Subscription) error
}

// Subscription is the model for unmarshalling JSON, and serialized to database
type SubscriptionInfos struct {
	ID             uint64         `json:"id,omitempty" db:"id"`                                 // Subscription id
	MemberID       rrsql.NullInt  `json:"member_id,omitempty" db:"member_id"`                   // Member who subscribed
	Email          string         `json:"email,omitempty" db:"email" binding:"required"`        // Email for failure handle, invoice create
	Amount         int            `json:"amount,omitempty" db:"amount" binding:"required,gt=0"` // Amount to pay
	CreatedAt      rrsql.NullTime `json:"created_at,omitempty" db:"created_at"`                 // The time when first created
	UpdatedAt      rrsql.NullTime `json:"updated_at,omitempty" db:"updated_at"`                 // The time when renewal
	LastPaidAt     rrsql.NullTime `json:"last_paid_at,omitempty" db:"last_paid_at"`             // Last time paid
	PaymentService string         `json:"payment_service,omitempty" db:"payment_service"`       // Payment service name
	InvoiceService string         `json:"invoice_service,omitempty" db:"invoice_service"`       // Invoice service name
	Status         int            `json:"status,omitempty" db:"status"`
}

type Subscription struct {
	SubscriptionInfos
	PaymentInfos payment.Provider `json:"payment_infos,omitempty" db:"payment_infos"`
	InvoiceInfos Invoicer         `json:"invoice_infos,omitempty" db:"invoice_infos"`
}

type ListFilter interface {
	Select() (string, []interface{}, error)
}
