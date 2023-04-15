package testing

import "github.com/ygo-skc/skc-suggestion-engine/model"

var (
	ExpectedSupportCardsMocks = map[string]model.CardSupport{
		"Dark Magician": {
			ReferencedBy: []model.Card{CardMocks["Magicians' Souls"], CardMocks["The Dark Magicians"]},
			MaterialFor:  []model.Card{CardMocks["Dark Paladin"], CardMocks["The Dark Magicians"]},
		},
		"Hamon, Lord of Striking Thunder": {
			ReferencedBy: []model.Card{},
			MaterialFor:  []model.Card{CardMocks["Armityle the Chaos Phantasm"], CardMocks["Armityle the Chaos Phantasm - Phantom of Fury"]},
		},
		"Elemental HERO Neos": {
			ReferencedBy: []model.Card{CardMocks["Neos Wiseman"]},
			MaterialFor:  []model.Card{CardMocks["Elemental HERO Air Neos"]},
		},
	}
)
