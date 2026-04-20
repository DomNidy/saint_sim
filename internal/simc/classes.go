package simc

import api "github.com/DomNidy/saint_sim/internal/api"

func parseClassIdentifier(value string) (api.CharacterClass, bool) {
	switch api.CharacterClass(value) {
	case api.Warrior,
		api.Hunter,
		api.Monk,
		api.Paladin,
		api.Rogue,
		api.Shaman,
		api.Mage,
		api.Warlock,
		api.Druid,
		api.Deathknight,
		api.Priest,
		api.Demonhunter,
		api.Evoker:
		return api.CharacterClass(value), true
	default:
		return "", false
	}
}
