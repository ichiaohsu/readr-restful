package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/readr-media/readr-restful/internal/rrsql"
	"github.com/readr-media/readr-restful/pkg/invoice"
	"github.com/readr-media/readr-restful/pkg/payment"
	"github.com/readr-media/readr-restful/pkg/subscription"
	"github.com/readr-media/readr-restful/pkg/subscription/mysql"
)

// Handler is the object with routing methods and database service
type Handler struct {
	Service subscription.Subscriber
}

// Get handles the list-all request
// func (h *Handler) Get(c *gin.Context) {

// 	var err error
// 	params := NewListRequest()
// 	if err = params.bind(c); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
// 		return
// 	}

// 	// Formatting JSON
// 	var results struct {
// 		Items []subscription.Subscription `json:"_itmes"`
// 	}
// 	results.Items, err = h.Service.GetSubscriptions(params)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, results)
// }

func BindPost(c *gin.Context, sub *subscription.Subscription) error {

	var payload struct {
		subscription.SubscriptionInfos
		PaymentInfos json.RawMessage `json:"payment_infos"`
		InvoiceInfos json.RawMessage `json:"invoice_infos"`
	}
	if err := c.BindJSON(&payload); err != nil {
		return err
	}

	sub.SubscriptionInfos = payload.SubscriptionInfos
	// TODO: overwrite created_at, set status to StatusInit

	if payment, err := payment.NewDisposableProvider(payload.PaymentService); err == nil {
		if err = json.Unmarshal(payload.PaymentInfos, &payment); err != nil {
			return err
		}
		// TODO: inject payment "amount", "remember" = true
		sub.PaymentInfos = payment
	}
	if invoice, err := invoice.NewInvoicer(payload.InvoiceService); err == nil {
		if err = json.Unmarshal(payload.InvoiceInfos, invoice); err != nil {
			return err
		}
		// TODO: inject invoice "TotalAmt"
		if err = invoice.Validate(); err != nil {
			return err
		}
		sub.InvoiceInfos = invoice
	}
	return nil
}

// Post handles create requests
func (h *Handler) Post(c *gin.Context) {

	var sub = subscription.Subscription{}
	if err := BindPost(c, &sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	// if err := c.BindJSON(&sub); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
	// 	return
	// }
	err := h.Service.CreateSubscription(sub)
	if err != nil {
		switch err.Error() {
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		}
		return
	}
	// return 201
	c.Status(http.StatusCreated)
}

// // Put handles update requests
// func (h *Handler) Put(c *gin.Context) {
// 	var sub = subscription.Subscription{}
// 	if err := c.BindJSON(&sub); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
// 		return
// 	}
// 	err := h.Service.UpdateSubscriptions(sub)
// 	if err != nil {
// 		switch err.Error() {
// 		default:
// 			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
// 		}
// 		return
// 	}
// 	// return 204
// 	c.Status(http.StatusNoContent)
// }

// // RecurringPay handles today's interval, get the subscription info list, and Pay with RoutinePay().
// func (h *Handler) RecurringPay(c *gin.Context) {

// 	// Get subscription list on today's interval
// 	start, end, _ := payInterval(time.Now())
// 	params := NewListRequest(func(p *ListRequest) {
// 		p.LastPaidAt = map[string]time.Time{
// 			"$gte": start,
// 			"$lt":  end,
// 		}
// 		p.Status = subscription.StatusOK
// 	})

// 	list, err := h.Service.GetSubscriptions(params)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
// 		return
// 	}
// 	err = h.Service.RoutinePay(list)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
// 		return
// 	}

// 	// Return 202
// 	c.Status(http.StatusAccepted)
// }

// SetRoutes provides a public function to set gin router
func (h *Handler) SetRoutes(router *gin.Engine) {

	// Create a mySQL subscription service, and make handlers use it to process subscriptions if there is no other service
	if h.Service == nil {
		// Set default service to MySQL
		log.Println("Set subscription service to default MySQL")

		s := &mysql.SubscriptionService{DB: rrsql.DB.DB}
		h.Service = s
	}
	// Register subscriptions endpoints
	subscriptionRouter := router.Group("/subscriptions")
	{
		// subscriptionRouter.GET("", h.Get)
		subscriptionRouter.POST("", h.Post)
		// subscriptionRouter.PUT("/:id", h.Put)

		// subscriptionRouter.POST("/recurring", h.RecurringPay)
	}
}

// Router is the instances for routing sets
var Router Handler
