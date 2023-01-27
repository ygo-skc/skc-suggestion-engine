package testing

import "github.com/ygo-skc/skc-suggestion-engine/model"

var (
	ExpectedReferences = InitSuggestionMocks()
)

func InitSuggestionMocks() map[string]model.CardSuggestions {
	return map[string]model.CardSuggestions{
		"Elemental HERO Sunrise": {
			NamedMaterials:       &[]model.CardReference{},
			MaterialArchetypes:   &[]string{"HERO"},
			NamedReferences:      &[]model.CardReference{{Occurrences: 1, Card: CardMocks["Elemental HERO Sunrise"]}, {Occurrences: 1, Card: CardMocks["Miracle Fusion"]}},
			ReferencedArchetypes: &[]string{"HERO"},
		},
		"Gem-Knight Master Diamond": {
			NamedMaterials:       &[]model.CardReference{},
			MaterialArchetypes:   &[]string{"Gem-Knight"},
			NamedReferences:      &[]model.CardReference{},
			ReferencedArchetypes: &[]string{"Gem-", "Gem-Knight"},
		},
		"A-to-Z-Dragon Buster Cannon": {
			NamedMaterials:       &[]model.CardReference{{Occurrences: 1, Card: CardMocks["ABC-Dragon Buster"]}, {Occurrences: 1, Card: CardMocks["XYZ-Dragon Cannon"]}},
			MaterialArchetypes:   &[]string{},
			NamedReferences:      &[]model.CardReference{{Occurrences: 1, Card: CardMocks["ABC-Dragon Buster"]}, {Occurrences: 1, Card: CardMocks["Polymerization"]}, {Occurrences: 1, Card: CardMocks["XYZ-Dragon Cannon"]}},
			ReferencedArchetypes: &[]string{},
		},
		"The Legendary Fisherman II": {
			NamedMaterials:       &[]model.CardReference{},
			MaterialArchetypes:   &[]string{},
			NamedReferences:      &[]model.CardReference{{Occurrences: 1, Card: CardMocks["The Legendary Fisherman"]}, {Occurrences: 1, Card: CardMocks["Umi"]}},
			ReferencedArchetypes: &[]string{},
		},
		"Armityle the Chaos Phantasm": {
			NamedMaterials: &[]model.CardReference{
				{Occurrences: 1, Card: CardMocks["Hamon, Lord of Striking Thunder"]},
				{Occurrences: 1, Card: CardMocks["Raviel, Lord of Phantasms"]},
				{Occurrences: 1, Card: CardMocks["Uria, Lord of Searing Flames"]},
			},
			MaterialArchetypes: &[]string{},
			NamedReferences: &[]model.CardReference{
				{Occurrences: 1, Card: CardMocks["Polymerization"]},
			},
			ReferencedArchetypes: &[]string{},
		},
		"Armityle the Chaos Phantasm - Phantom of Fury": {
			NamedMaterials: &[]model.CardReference{
				{Occurrences: 1, Card: CardMocks["Hamon, Lord of Striking Thunder"]},
				{Occurrences: 1, Card: CardMocks["Raviel, Lord of Phantasms"]},
				{Occurrences: 1, Card: CardMocks["Uria, Lord of Searing Flames"]},
			},
			MaterialArchetypes: &[]string{},
			NamedReferences: &[]model.CardReference{
				{Occurrences: 2, Card: CardMocks["Armityle the Chaos Phantasm"]},
			},
			ReferencedArchetypes: &[]string{},
		},
		"King Dragun": {
			NamedMaterials: &[]model.CardReference{
				{Occurrences: 1, Card: CardMocks["Divine Dragon Ragnarok"]},
				{Occurrences: 1, Card: CardMocks["Lord of D."]},
			},
			MaterialArchetypes:   &[]string{},
			NamedReferences:      &[]model.CardReference{},
			ReferencedArchetypes: &[]string{},
		},
		"Great Mammoth of Goldfine": {
			NamedMaterials: &[]model.CardReference{
				{Occurrences: 1, Card: CardMocks["Dragon Zombie"]},
				{Occurrences: 1, Card: CardMocks["The Snake Hair"]},
			},
			MaterialArchetypes:   &[]string{},
			NamedReferences:      &[]model.CardReference{},
			ReferencedArchetypes: &[]string{},
		},
		"Elemental HERO Stratos": {
			NamedMaterials:     &[]model.CardReference{},
			MaterialArchetypes: &[]string{},
			NamedReferences:    &[]model.CardReference{},
			ReferencedArchetypes: &[]string{
				"HERO", "HERO",
			},
		},
	}
}
