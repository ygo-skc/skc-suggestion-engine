package testing

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
		"Gem-Knight Master Diamond": {
			CardID:    "39512984",
			CardColor: "Fusion",
			CardName:  "Gem-Knight Master Diamond",
			CardEffect: `3 'Gem-Knight' monsters
Must first be Fusion Summoned. This card gains 100 ATK for each 'Gem-' monster in your Graveyard. Once per turn: You can banish 1 Level 7 or lower 'Gem-Knight' Fusion Monster from your Graveyard; until the End Phase, this card's name becomes that monster's, and replace this effect with that monster's original effects.`,
		},
		"A-to-Z-Dragon Buster Cannon": {
			CardID:    "65172015",
			CardColor: "Fusion",
			CardName:  "A-to-Z-Dragon Buster Cannon",
			CardEffect: `"ABC-Dragon Buster" + "XYZ-Dragon Cannon"
Must be Special Summoned (from your Extra Deck) by banishing cards you control with the above original names, and cannot be Special Summoned by other ways. (You do not use "Polymerization".) During either player's turn, when your opponent activates a Spell/Trap Card, or monster effect: You can discard 1 card; negate the activation, and if you do, destroy that card. During either player's turn: You can banish this card, then target 1 each of your banished "ABC-Dragon Buster", and "XYZ-Dragon Cannon"; Special Summon them.`,
		},
		"ABC-Dragon Buster": {
			CardID:    "01561110",
			CardColor: "Fusion",
			CardName:  "ABC-Dragon Buster",
			CardEffect: `'A-Assault Core' + 'B-Buster Drake' + 'C-Crush Wyvern'
Must first be Special Summoned (from your Extra Deck) by banishing the above cards you control and/or from your GY. (You do not use 'Polymerization'.) Once per turn (Quick Effect): You can discard 1 card, then target 1 card on the field; banish it. During your opponent's turn (Quick Effect): You can Tribute this card, then target 3 of your banished LIGHT Machine Union monsters with different names; Special Summon them.`,
		},
		"XYZ-Dragon Cannon": {
			CardID:    "91998119",
			CardColor: "Fusion",
			CardName:  "XYZ-Dragon Cannon",
			CardEffect: `"X-Head Cannon" + "Y-Dragon Head" + "Z-Metal Tank"
Must first be Special Summoned (from your Extra Deck) by banishing the above cards you control. (You do not use "Polymerization".) Cannot be Special Summoned from the GY. You can discard 1 card, then target 1 card your opponent controls; destroy that target.`,
		},
		"Polymerization": {
			CardID:     "24094653",
			CardColor:  "Spell",
			CardName:   "Polymerization",
			CardEffect: `Fusion Summon 1 Fusion Monster from your Extra Deck, using monsters from your hand or field as Fusion Material.`,
		},
		"The Legendary Fisherman II": {
			CardID:     "19801646",
			CardColor:  "Effect",
			CardName:   "The Legendary Fisherman II",
			CardEffect: `This card's name becomes "The Legendary Fisherman" while on the field or in the GY. While "Umi" is on the field, this card is unaffected by other monsters' effects. If this face-up card is destroyed by battle, or leaves the field because of an opponent's card effect while its owner controls it: You can add 1 Level 7 WATER monster from your Deck to your hand.`,
		},
		"The Legendary Fisherman": {
			CardID:     "03643300",
			CardColor:  "Effect",
			CardName:   "The Legendary Fisherman",
			CardEffect: `While "Umi" is on the field, this card is unaffected by Spell effects and cannot be targeted for attacks, but does not prevent your opponent from attacking you directly.`,
		},
		"Umi": {
			CardID:     "22702055",
			CardColor:  "Spell",
			CardName:   "Umi",
			CardEffect: `All Fish, Sea Serpent, Thunder, and Aqua monsters on the field gain 200 ATK/DEF, also all Machine and Pyro monsters on the field lose 200 ATK/DEF.`,
		},
		"Armityle the Chaos Phantasm": {
			CardID:    "43378048",
			CardColor: "Fusion",
			CardName:  "Armityle the Chaos Phantasm",
			CardEffect: `"Uria, Lord of Searing Flames" + "Hamon, Lord of Striking Thunder" + "Raviel, Lord of Phantasms"
Must first be Special Summoned (from your Extra Deck) by banishing the above cards you control. (You do not use "Polymerization".) Cannot be destroyed by battle. Gains 10,000 ATK during your turn only.`,
		},
		"Armityle the Chaos Phantasm - Phantom of Fury": {
			CardID:    "60110982",
			CardColor: "Fusion",
			CardName:  "Armityle the Chaos Phantasm - Phantom of Fury",
			CardEffect: `"Uria, Lord of Searing Flames" + "Hamon, Lord of Striking Thunder" + "Raviel, Lord of Phantasms"
This card's name becomes "Armityle the Chaos Phantasm" while on the field. Once per turn, during your Main Phase: You can give control of this card to your opponent. Once per turn, during the End Phase, if this card's control was changed this turn: Banish all cards you control, then the owner of this card can Special Summon 1 "Armityle the Chaos Phantasm" from their Extra Deck, ignoring its Summoning conditions.`,
		},
		"Hamon, Lord of Striking Thunder": {
			CardID:     "32491822",
			CardColor:  "Effect",
			CardName:   "Hamon, Lord of Striking Thunder",
			CardEffect: `Cannot be Normal Summoned/Set. Must be Special Summoned (from your hand) by sending 3 face-up Continuous Spells you control to the GY. If this card destroys an opponent's monster by battle and sends it to the GY: Inflict 1000 damage to your opponent. While this card is in face-up Defense Position, monsters your opponent controls cannot target monsters for attacks, except this one.`,
		},
		"Raviel, Lord of Phantasms": {
			CardID:     "69890967",
			CardColor:  "Effect",
			CardName:   "Raviel, Lord of Phantasms",
			CardEffect: `Cannot be Normal Summoned/Set. Must be Special Summoned (from your hand) by Tributing 3 Fiend monsters. Each time your opponent Normal Summons a monster: Special Summon 1 "Phantasm Token" (Fiend/DARK/Level 1/ATK 1000/DEF 1000), but it cannot declare an attack. Once per turn: You can Tribute 1 monster; this card gains ATK equal to the Tributed monster's original ATK, until the end of this turn.`,
		},
		"Uria, Lord of Searing Flames": {
			CardID:     "06007213",
			CardColor:  "Effect",
			CardName:   "Uria, Lord of Searing Flames",
			CardEffect: `Cannot be Normal Summoned/Set. Must be Special Summoned (from your hand) by sending 3 face-up Traps you control to the GY. This card gains 1000 ATK for each Continuous Trap in your GY. Once per turn: You can target 1 Set Spell/Trap your opponent controls; destroy that target. Neither player can activate Spell/Trap Cards in response to this effect's activation.`,
		},
		"King Dragun": {
			CardID:    "13756293",
			CardColor: "Fusion",
			CardName:  "King Dragun",
			CardEffect: `"Lord of D." + "Divine Dragon Ragnarok"
Your opponent cannot target Dragon monsters with card effects. Once per turn: You can Special Summon 1 Dragon monster from your hand.`,
		},
		"Divine Dragon Ragnarok": {
			CardID:     "62113340",
			CardColor:  "Dragon/Normal",
			CardName:   "Divine Dragon Ragnarok",
			CardEffect: `A legendary dragon sent by the gods as their instrument. Legends say that if provoked, the whole world will sink beneath the sea.`,
		},
		"Lord of D.": {
			CardID:     "17985575",
			CardColor:  "Effect",
			CardName:   "Lord of D.",
			CardEffect: `Neither player can target Dragon monsters on the field with card effects.`,
		},
		"Great Mammoth of Goldfine": {
			CardID:     "54622031",
			CardColor:  "Fusion",
			CardName:   "Great Mammoth of Goldfine",
			CardEffect: `"The Snake Hair" + "Dragon Zombie"`,
		},
		"Dragon Zombie": {
			CardID:     "66672569",
			CardColor:  "Normal",
			CardName:   "Dragon Zombie",
			CardEffect: `"A dragon revived by sorcery. Its breath is highly corrosive."`,
		},
		"The Snake Hair": {
			CardID:     "29491031",
			CardColor:  "Normal",
			CardName:   "The Snake Hair",
			CardEffect: `"A monster with a head of poison snakes. One look from this monster can turn an opponent to stone."`,
		},
	}
)
