package glmodel

var (
	MD_KEY_TRANSACTION_TYPE           = "transaction_type"
	MD_TRANSACTION_TYPE_JOURNAL_ENTRY = "journal_entry"
	MD_TRANSACTION_TYPE_PAYMENT       = "payment"

	MD_KEY_TRANSACTION_REFERENCE = "transaction_reference"

	MD_KEY_PAYMENT_CHANNEL       = "payment_channel"
	MD_PAYMENT_CHANNEL_US_FEDNOW = "us-fednow"
	MD_PAYMENT_CHANNEL_US_ACH    = "us-ach"

	MD_KEY_PAYMENT_LIFECYCLE                      = "payment_lifecycle"
	MD_PAYMENT_LIFECYCLE_INITIAL_POSTING          = "initial_posting"
	MD_PAYMENT_LIFECYCLE_CLEARING_POSTING         = "clearing_posting"
	MD_PAYMENT_LIFECYCLE_CUSTOMER_ACCOUNT_POSTING = "customer_account_posting"

	MD_KEY_ACCOUNT_TYPE          = "account_type"
	MD_ACCOUNT_TYPE_REGULAR_GL   = "regular_gl"
	MD_ACCOUNT_TYPE_FBO          = "fbo"
	MD_ACCOUNT_TYPE_FBO_CUSTOMER = "fbo_customer"
	MD_ACCOUNT_TYPE_PAYMENT      = "payment"
	MD_ACCOUNT_TYPE_OTHER        = "other"

	MD_KEY_TAGS = "tags"

	MD_KEY_NOTE    = "note"
	MD_KEY_PURPOSE = "purpose"
)
