package testing

import "github.com/ygo-skc/skc-suggestion-engine/model"

var (
	ExpectedSupportCardsMocks = map[string]model.CardSupport{
		"Dark Magician": {
			ReferencedBy: []model.Card{CardMocks["Magicians' Souls"]},
			MaterialFor:  []model.Card{CardMocks["Dark Paladin"]},
		},
		"Hamon, Lord of Striking Thunder": {
			ReferencedBy: []model.Card{},
			MaterialFor:  []model.Card{CardMocks["Armityle the Chaos Phantasm"], CardMocks["Armityle the Chaos Phantasm - Phantom of Fury"]},
		},
	}
)
