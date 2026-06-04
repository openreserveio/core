package workflows

type WorkflowDetails struct {
	WorkflowID    string
	WorkflowRunID string
}

type ResultOfReview struct {
	Result string `json:"result"`
}
