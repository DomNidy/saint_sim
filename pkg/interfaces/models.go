package interfaces

// this package defines database models

type SimDataGet struct {
	ID        string `json:"id"`
	RequestID string `json:"request_id"`
	SimResult string `json:"sim_result"`
}

type SimDataInsert struct {
	RequestID string
	SimResult string
}
