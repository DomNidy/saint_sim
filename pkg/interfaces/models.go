package interfaces

// this package defines database models

type SimDataGet struct {
	ID          string `json:"id"`
	FromRequest string `json:"from_request"`
	SimResult   string `json:"sim_result"`
}

type SimDataInsert struct {
	FromRequest string
	SimResult   string
}
