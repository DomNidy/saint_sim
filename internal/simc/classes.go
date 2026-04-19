package simc

// tciClassIdentifier represents a string that TCI interprets as
// a character class.
type tciClassIdentifier string

// TODO: Move this up into openapi spec & autogenerate it
const (
	Warrior     tciClassIdentifier = "warrior"
	Hunter      tciClassIdentifier = "hunter"
	Monk        tciClassIdentifier = "monk"
	Paladin     tciClassIdentifier = "paladin"
	Rogue       tciClassIdentifier = "rogue"
	Shaman      tciClassIdentifier = "shaman"
	Mage        tciClassIdentifier = "mage"
	Warlock     tciClassIdentifier = "warlock"
	Druid       tciClassIdentifier = "druid"
	DeathKnight tciClassIdentifier = "deathknight"
	Priest      tciClassIdentifier = "priest"
	DemonHunter tciClassIdentifier = "demonhunter"
	Evoker      tciClassIdentifier = "evoker"
)
