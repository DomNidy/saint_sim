package interfaces

// Character represents a candidate World of Warcraft character
// We will retrieve the Character's gear using the Armory API
// Note the usage of candidate, as we will need to validate that
// this Character struct contains a valid character name, realm, etc.
type Character struct {
	Region string `json:"region"` // region the character is in
	Realm  string `json:"realm"`  // realm which the character is on
	Name   string `json:"name"`   // name of the character
}

// Validates that the character metadata points to a character that actually exists
func (c *Character) Validate() error {
	return nil
}
