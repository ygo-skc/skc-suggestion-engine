package model

type TaggedText struct {
	Text string   `bson:"text" json:"text"`
	Tags []string `bson:"tags" json:"tags"`
}

type Effect struct {
	Text      string       `bson:"text" json:"text"`
	Zone      string       `bson:"zone,omitempty" json:"zone,omitempty"`
	Tags      []string     `bson:"tags" json:"tags"`
	Frequency string       `bson:"frequency,omitempty" json:"frequency,omitempty"`
	Counters  []string     `bson:"counters" json:"counters"`
	Modes     []TaggedText `bson:"modes" json:"modes"`
	Choice    string       `bson:"choice,omitempty" json:"choice,omitempty"`
}

// CardMechanic is a card's text parsed into distinct fields, precomputed offline and stored in the
// cardMechanic collection.
//
// MetaTags, Flags, DoesAll, DoesAny, Counters and Gates are card-level rollups of the per-Effect
// data. They are precomputed and served so clients don't have to aggregate Effects themselves - the
// apparent duplication with Effect is intentional.
type CardMechanic struct {
	ID               string       `bson:"id" json:"id"`
	Name             string       `bson:"name" json:"name"`
	Type             string       `bson:"type" json:"type"`
	Materials        string       `bson:"materials,omitempty" json:"materials,omitempty"`
	SummonConditions []TaggedText `bson:"summonConditions" json:"summonConditions"`
	Effects          []Effect     `bson:"effects" json:"effects"`
	MetaTags         []string     `bson:"metaTags" json:"metaTags"`
	Flags            []string     `bson:"flags" json:"flags"`
	Flavor           bool         `bson:"flavor" json:"flavor"`
	DoesAll          []string     `bson:"doesAll" json:"doesAll"`
	DoesAny          []string     `bson:"doesAny" json:"doesAny"`
	Counters         []string     `bson:"counters" json:"counters"`
	Gates            []string     `bson:"gates" json:"gates"`
}
