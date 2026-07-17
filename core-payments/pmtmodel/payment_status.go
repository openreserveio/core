package pmtmodel

var (
	PAYMENT_STATUS_NEW_BATCH_RECEIVED    = "NEW_BATCH_RECEIVED"
	PAYMENT_STATUS_ERROR                 = "ERROR"
	PAYMENT_STATUS_INSTRUCTION_RECEIVED  = "RECEIVED"
	PAYMENT_STATUS_PROCESSING            = "PROCESSING"
	PAYMENT_STATUS_PROCESSED             = "PROCESSED"
	PAYMENT_STATUS_AWAITING_TRANSMISSION = "AWAITING_TRANSMISSION"
	PAYMENT_STATUS_TRANSMITTED           = "TRANSMITTED"
	PAYMENT_STATUS_COMPLETE              = "COMPLETE"
	PAYMENT_STATUS_RETURNED              = "RETURNED"
	PAYMENT_STATUS_OFAC_IN_PROGRESS      = "OFAC_IN_PROGRESS"
	PAYMENT_STATUS_IN_REVIEW             = "IN_REVIEW"
)

type PaymentStatusUpdateNotification struct {
	PaymentID        string `json:"payment_id"`
	NotificationDate string `json:"notification_date"`
	PreviousStatus   string `json:"previous_status"`
	CurrentStatus    string `json:"current_status"`
	AdditionalInfo   string `json:"additional_info"`
}
