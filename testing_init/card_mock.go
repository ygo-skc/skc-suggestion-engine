package testing_init

import "github.com/ygo-skc/skc-suggestion-engine/model"

var (
	CardMocks = map[string]model.Card{
		"Elemental HERO Sunrise": {
			CardID:    "22908820",
			CardColor: "Fusion",
			CardName:  "Elemental HERO Sunrise",
			CardEffect: `2 'HERO' monsters with different Attributes
Must be Fusion Summoned. Monsters you control gain 200 ATK for each different Attribute you control. You can only use each of the following effects of 'Elemental HERO Sunrise' once per turn.
• If this card is Special Summoned: You can add 1 'Miracle Fusion' from your Deck to your hand.
• When an attack is declared involving another 'HERO' monster you control: You can target 1 card on the field; destroy it.`,
		},
		"Miracle Fusion": {
			CardID:     "45906428",
			CardColor:  "Spell",
			CardName:   "Miracle Fusion",
			CardEffect: `Fusion Summon 1 'Elemental HERO' Fusion Monster from your Extra Deck, by banishing Fusion Materials listed on it from your field or your GY.`,
		},
	}
)
