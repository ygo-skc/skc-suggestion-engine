package testing

import (
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
)

var (
	CardColors = map[string]uint32{
		"Normal":           1,
		"Effect":           2,
		"Fusion":           3,
		"Ritual":           4,
		"Synchro":          5,
		"Xyz":              6,
		"Pendulum-Normal":  7,
		"Pendulum-Effect":  8,
		"Pendulum-Ritual":  9,
		"Pendulum-Fusion":  10,
		"Pendulum-Xyz":     11,
		"Pendulum-Synchro": 12,
		"Link":             13,
		"Spell":            14,
		"Trap":             15,
		"Token":            16,
	}

	CardMocks = map[string]cModel.YGOCardREST{
		"Elemental HERO Sunrise": {
			ID:    "22908820",
			Color: "Fusion",
			Name:  "Elemental HERO Sunrise",
			Effect: `2 'HERO' monsters with different Attributes
Must be Fusion Summoned. Monsters you control gain 200 ATK for each different Attribute you control. You can only use each of the following effects of 'Elemental HERO Sunrise' once per turn.
• If this card is Special Summoned: You can add 1 'Miracle Fusion' from your Deck to your hand.
• When an attack is declared involving another 'HERO' monster you control: You can target 1 card on the field; destroy it.`,
			MonsterType: cUtil.InlineStringPointer("Warrior/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(2500),
			Defense:     cUtil.InlineUInt32Pointer(1200),
		},
		"Miracle Fusion": {
			ID:     "45906428",
			Color:  "Spell",
			Name:   "Miracle Fusion",
			Effect: `Fusion Summon 1 'Elemental HERO' Fusion Monster from your Extra Deck, by banishing Fusion Materials listed on it from your field or your GY.`,
		},
		"Gem-Knight Master Diamond": {
			ID:    "39512984",
			Color: "Fusion",
			Name:  "Gem-Knight Master Diamond",
			Effect: `3 'Gem-Knight' monsters
Must first be Fusion Summoned. This card gains 100 ATK for each 'Gem-' monster in your Graveyard. Once per turn: You can banish 1 Level 7 or lower 'Gem-Knight' Fusion Monster from your Graveyard; until the End Phase, this card's name becomes that monster's, and replace this effect with that monster's original effects.`,
			MonsterType: cUtil.InlineStringPointer("Rock/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(2900),
			Defense:     cUtil.InlineUInt32Pointer(2500),
		},
		"A-to-Z-Dragon Buster Cannon": {
			ID:    "65172015",
			Color: "Fusion",
			Name:  "A-to-Z-Dragon Buster Cannon",
			Effect: `"ABC-Dragon Buster" + "XYZ-Dragon Cannon"
Must be Special Summoned (from your Extra Deck) by banishing cards you control with the above original names, and cannot be Special Summoned by other ways. (You do not use "Polymerization".) During either player's turn, when your opponent activates a Spell/Trap Card, or monster effect: You can discard 1 card; negate the activation, and if you do, destroy that card. During either player's turn: You can banish this card, then target 1 each of your banished "ABC-Dragon Buster", and "XYZ-Dragon Cannon"; Special Summon them.`,
			MonsterType: cUtil.InlineStringPointer("Machine/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(4000),
			Defense:     cUtil.InlineUInt32Pointer(4000),
		},
		"ABC-Dragon Buster": {
			ID:    "01561110",
			Color: "Fusion",
			Name:  "ABC-Dragon Buster",
			Effect: `'A-Assault Core' + 'B-Buster Drake' + 'C-Crush Wyvern'
Must first be Special Summoned (from your Extra Deck) by banishing the above cards you control and/or from your GY. (You do not use 'Polymerization'.) Once per turn (Quick Effect): You can discard 1 card, then target 1 card on the field; banish it. During your opponent's turn (Quick Effect): You can Tribute this card, then target 3 of your banished LIGHT Machine Union monsters with different names; Special Summon them.`,
			MonsterType: cUtil.InlineStringPointer("Machine/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(3000),
			Defense:     cUtil.InlineUInt32Pointer(2800),
		},
		"XYZ-Dragon Cannon": {
			ID:    "91998119",
			Color: "Fusion",
			Name:  "XYZ-Dragon Cannon",
			Effect: `"X-Head Cannon" + "Y-Dragon Head" + "Z-Metal Tank"
Must first be Special Summoned (from your Extra Deck) by banishing the above cards you control. (You do not use "Polymerization".) Cannot be Special Summoned from the GY. You can discard 1 card, then target 1 card your opponent controls; destroy that target.`,
			MonsterType: cUtil.InlineStringPointer("Machine/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(2800),
			Defense:     cUtil.InlineUInt32Pointer(2600),
		},
		"Polymerization": {
			ID:     "24094653",
			Color:  "Spell",
			Name:   "Polymerization",
			Effect: `Fusion Summon 1 Fusion Monster from your Extra Deck, using monsters from your hand or field as Fusion Material.`,
		},
		"The Legendary Fisherman II": {
			ID:          "19801646",
			Color:       "Effect",
			Name:        "The Legendary Fisherman II",
			Effect:      `This card's name becomes "The Legendary Fisherman" while on the field or in the GY. While "Umi" is on the field, this card is unaffected by other monsters' effects. If this face-up card is destroyed by battle, or leaves the field because of an opponent's card effect while its owner controls it: You can add 1 Level 7 WATER monster from your Deck to your hand.`,
			MonsterType: cUtil.InlineStringPointer("Warrior/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(2200),
			Defense:     cUtil.InlineUInt32Pointer(1800),
		},
		"The Legendary Fisherman": {
			ID:          "03643300",
			Color:       "Effect",
			Name:        "The Legendary Fisherman",
			Effect:      `While "Umi" is on the field, this card is unaffected by Spell effects and cannot be targeted for attacks, but does not prevent your opponent from attacking you directly.`,
			MonsterType: cUtil.InlineStringPointer("Warrior/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(1850),
			Defense:     cUtil.InlineUInt32Pointer(1600),
		},
		"Umi": {
			ID:     "22702055",
			Color:  "Spell",
			Name:   "Umi",
			Effect: `All Fish, Sea Serpent, Thunder, and Aqua monsters on the field gain 200 ATK/DEF, also all Machine and Pyro monsters on the field lose 200 ATK/DEF.`,
		},
		"Armityle the Chaos Phantasm": {
			ID:    "43378048",
			Color: "Fusion",
			Name:  "Armityle the Chaos Phantasm",
			Effect: `"Uria, Lord of Searing Flames" + "Hamon, Lord of Striking Thunder" + "Raviel, Lord of Phantasms"
Must first be Special Summoned (from your Extra Deck) by banishing the above cards you control. (You do not use "Polymerization".) Cannot be destroyed by battle. Gains 10,000 ATK during your turn only.`,
			MonsterType: cUtil.InlineStringPointer("Fiend/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(0),
			Defense:     cUtil.InlineUInt32Pointer(0),
		},
		"Armityle the Chaos Phantasm - Phantom of Fury": {
			ID:    "60110982",
			Color: "Fusion",
			Name:  "Armityle the Chaos Phantasm - Phantom of Fury",
			Effect: `"Uria, Lord of Searing Flames" + "Hamon, Lord of Striking Thunder" + "Raviel, Lord of Phantasms"
This card's name becomes "Armityle the Chaos Phantasm" while on the field. Once per turn, during your Main Phase: You can give control of this card to your opponent. Once per turn, during the End Phase, if this card's control was changed this turn: Banish all cards you control, then the owner of this card can Special Summon 1 "Armityle the Chaos Phantasm" from their Extra Deck, ignoring its Summoning conditions.`,
			MonsterType: cUtil.InlineStringPointer("Fiend/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(0),
			Defense:     cUtil.InlineUInt32Pointer(0),
		},
		"Hamon, Lord of Striking Thunder": {
			ID:          "32491822",
			Color:       "Effect",
			Name:        "Hamon, Lord of Striking Thunder",
			Effect:      `Cannot be Normal Summoned/Set. Must be Special Summoned (from your hand) by sending 3 face-up Continuous Spells you control to the GY. If this card destroys an opponent's monster by battle and sends it to the GY: Inflict 1000 damage to your opponent. While this card is in face-up Defense Position, monsters your opponent controls cannot target monsters for attacks, except this one.`,
			MonsterType: cUtil.InlineStringPointer("Thunder/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(4000),
			Defense:     cUtil.InlineUInt32Pointer(4000),
		},
		"Raviel, Lord of Phantasms": {
			ID:          "69890967",
			Color:       "Effect",
			Name:        "Raviel, Lord of Phantasms",
			Effect:      `Cannot be Normal Summoned/Set. Must be Special Summoned (from your hand) by Tributing 3 Fiend monsters. Each time your opponent Normal Summons a monster: Special Summon 1 "Phantasm Token" (Fiend/DARK/Level 1/ATK 1000/DEF 1000), but it cannot declare an attack. Once per turn: You can Tribute 1 monster; this card gains ATK equal to the Tributed monster's original ATK, until the end of this turn.`,
			MonsterType: cUtil.InlineStringPointer("Fiend/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(4000),
			Defense:     cUtil.InlineUInt32Pointer(4000),
		},
		"Uria, Lord of Searing Flames": {
			ID:          "06007213",
			Color:       "Effect",
			Name:        "Uria, Lord of Searing Flames",
			Effect:      `Cannot be Normal Summoned/Set. Must be Special Summoned (from your hand) by sending 3 face-up Traps you control to the GY. This card gains 1000 ATK for each Continuous Trap in your GY. Once per turn: You can target 1 Set Spell/Trap your opponent controls; destroy that target. Neither player can activate Spell/Trap Cards in response to this effect's activation.`,
			MonsterType: cUtil.InlineStringPointer("Thunder/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(0),
			Defense:     cUtil.InlineUInt32Pointer(0),
		},
		"King Dragun": {
			ID:    "13756293",
			Color: "Fusion",
			Name:  "King Dragun",
			Effect: `"Lord of D." + "Divine Dragon Ragnarok"
Your opponent cannot target Dragon monsters with card effects. Once per turn: You can Special Summon 1 Dragon monster from your hand.`,
			MonsterType: cUtil.InlineStringPointer("Dragon/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(2400),
			Defense:     cUtil.InlineUInt32Pointer(1100),
		},
		"Divine Dragon Ragnarok": {
			ID:          "62113340",
			Color:       "Dragon/Normal",
			Name:        "Divine Dragon Ragnarok",
			Effect:      `A legendary dragon sent by the gods as their instrument. Legends say that if provoked, the whole world will sink beneath the sea.`,
			MonsterType: cUtil.InlineStringPointer("Dragon/Normal"),
			Attack:      cUtil.InlineUInt32Pointer(1500),
			Defense:     cUtil.InlineUInt32Pointer(1000),
		},
		"Lord of D.": {
			ID:          "17985575",
			Color:       "Effect",
			Name:        "Lord of D.",
			Effect:      `Neither player can target Dragon monsters on the field with card effects.`,
			MonsterType: cUtil.InlineStringPointer("Spellcaster/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(1200),
			Defense:     cUtil.InlineUInt32Pointer(1100),
		},
		"Great Mammoth of Goldfine": {
			ID:          "54622031",
			Color:       "Fusion",
			Name:        "Great Mammoth of Goldfine",
			Effect:      `"The Snake Hair" + "Dragon Zombie"`,
			MonsterType: cUtil.InlineStringPointer("Zombie/Fusion"),
			Attack:      cUtil.InlineUInt32Pointer(2200),
			Defense:     cUtil.InlineUInt32Pointer(1800),
		},
		"Dragon Zombie": {
			ID:          "66672569",
			Color:       "Normal",
			Name:        "Dragon Zombie",
			Effect:      `"A dragon revived by sorcery. Its breath is highly corrosive."`,
			MonsterType: cUtil.InlineStringPointer("Zombie/Normal"),
			Attack:      cUtil.InlineUInt32Pointer(1600),
			Defense:     cUtil.InlineUInt32Pointer(0),
		},
		"The Snake Hair": {
			ID:          "29491031",
			Color:       "Normal",
			Name:        "The Snake Hair",
			Effect:      `"A monster with a head of poison snakes. One look from this monster can turn an opponent to stone."`,
			MonsterType: cUtil.InlineStringPointer("Zombie/Normal"),
			Attack:      cUtil.InlineUInt32Pointer(1500),
			Defense:     cUtil.InlineUInt32Pointer(1200),
		},
		"Elemental HERO Stratos": {
			ID:    "40044918",
			Color: "Effect",
			Name:  "Elemental HERO Stratos",
			Effect: `When this card is Normal or Special Summoned: You can activate 1 of these effects.
&bull; Destroy Spells/Traps on the field, up to the number of "HERO" monsters you control, except this card.
&bull; Add 1 "HERO" monster from your Deck to your hand.`,
			MonsterType: cUtil.InlineStringPointer("Warrior/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(1800),
			Defense:     cUtil.InlineUInt32Pointer(300),
		},
		"Dark Magician": {
			ID:          "46986414",
			Color:       "Normal",
			Name:        "Dark Magician",
			Effect:      `The ultimate wizard in terms of attack and defense.`,
			MonsterType: cUtil.InlineStringPointer("Spellcaster/Normal"),
			Attack:      cUtil.InlineUInt32Pointer(2500),
			Defense:     cUtil.InlineUInt32Pointer(2100),
		},
		"Magicians' Souls": {
			ID:    "97631303",
			Color: "Effect",
			Name:  "Magicians' Souls",
			Effect: `You can send up to 2 Spells/Traps from your hand and/or field to the GY; draw that many cards. If this card is in your hand: You can send 1 Level 6 or higher Spellcaster monster from your Deck to the GY, then activate 1 of these effects;
&bull; Special Summon this card.
&bull; Send this card to the GY, then, you can Special Summon 1 "Dark Magician" or 1 "Dark Magician Girl" from your GY.
You can only use each effect of "Magicians' Souls" once per turn.`,
			MonsterType: cUtil.InlineStringPointer("Spellcaster/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(0),
			Defense:     cUtil.InlineUInt32Pointer(0),
		},
		"Dark Paladin": {
			ID:    "98502113",
			Color: "Fusion",
			Name:  "Dark Paladin",
			Effect: `"Dark Magician" + "Buster Blader"
Must be Fusion Summoned. When a Spell Card is activated (Quick Effect): You can discard 1 card; negate the activation, and if you do, destroy it. This card must be face-up on the field to activate and to resolve this effect. Gains 500 ATK for each Dragon monster on the field and in the GY.`,
			MonsterType: cUtil.InlineStringPointer("Spellcaster/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(2900),
			Defense:     cUtil.InlineUInt32Pointer(2400),
		},
		"Elemental HERO Air Neos": {
			ID:    "11502550",
			Color: "Fusion",
			Name:  "Elemental HERO Air Neos",
			Effect: `"Elemental HERO Neos" + "Neo-Spacian Air Hummingbird"
This card can only be Special Summoned from your Extra Deck by returning the above cards you control to the Deck. (You do not use "Polymerization".) While your Life Points are lower than your opponent's, this card gains ATK equal to the difference. This card returns to the Extra Deck during the End Phase.`,
			MonsterType: cUtil.InlineStringPointer("Warrior/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(2500),
			Defense:     cUtil.InlineUInt32Pointer(2000),
		},
		"Neos Wiseman": {
			ID:          "05126490",
			Color:       "Effect",
			Name:        "Neos Wiseman",
			Effect:      `Cannot be Normal Summoned or Set. Must be Special Summoned (from your hand) by sending 1 face-up "Elemental HERO Neos" and 1 face-up "Yubel" you control to the Graveyard, and cannot be Special Summoned by other ways. This card cannot be destroyed by card effects. At the end of the Damage Step, if this card battled an opponent's monster: Inflict damage to your opponent equal to the ATK of the monster it battled, and you gain Life Points equal to that monster's DEF.`,
			MonsterType: cUtil.InlineStringPointer("Spellcaster/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(3000),
			Defense:     cUtil.InlineUInt32Pointer(3000),
		},
		"Elemental HERO Neos": {
			ID:          "89943723",
			Color:       "Normal",
			Name:        "Elemental HERO Neos",
			Effect:      `A new Elemental HERO has arrived from Neo-Space! When he initiates a Contact Fusion with a Neo-Spacian his unknown powers are unleashed.`,
			MonsterType: cUtil.InlineStringPointer("Warrior/Normal"),
			Attack:      cUtil.InlineUInt32Pointer(2500),
			Defense:     cUtil.InlineUInt32Pointer(2000),
		},
		"The Dark Magicians": {
			ID:    "50237654",
			Color: "Fusion",
			Name:  "The Dark Magicians",
			Effect: `'Dark Magician' or 'Dark Magician Girl' + 1 Spellcaster monster
Once per turn, if a Spell/Trap Card or effect is activated (except during the Damage Step): You can draw 1 card, then if it was a Spell/Trap, you can Set it, and if it was a Trap or Quick-Play Spell, you can activate it this turn. If this card is destroyed: You can Special Summon both 1 'Dark Magician' and 1 'Dark Magician Girl' from your hand, Deck, and/or GY.`,
			MonsterType: cUtil.InlineStringPointer("Spellcaster/Fusion/Effect"),
			Attack:      cUtil.InlineUInt32Pointer(2800),
			Defense:     cUtil.InlineUInt32Pointer(2000),
		},
	}
)
