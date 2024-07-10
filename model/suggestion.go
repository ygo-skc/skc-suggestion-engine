package model

type CardReference struct {
	Occurrences int  `json:"occurrences"`
	Card        Card `json:"card"`
}

type CardSuggestions struct {
	Card                 Card            `json:"card"`
	HasSelfReference     bool            `json:"hasSelfReference"`
	NamedMaterials       []CardReference `json:"namedMaterials"`
	NamedReferences      []CardReference `json:"namedReferences"`
	MaterialArchetypes   []string        `json:"materialArchetypes"`
	ReferencedArchetypes []string        `json:"referencedArchetypes"`
}

type BatchCardSuggestions[IS IdentifierSlice] struct {
	NamedMaterials       []CardReference `json:"namedMaterials"`
	NamedReferences      []CardReference `json:"namedReferences"`
	MaterialArchetypes   []string        `json:"materialArchetypes"`
	ReferencedArchetypes []string        `json:"referencedArchetypes"`
	UnknownResources     IS              `json:"unknownResources"`
	FalsePositives       IS              `json:"falsePositives"`
}

type CardSupport struct {
	Card         Card            `json:"card"`
	ReferencedBy []CardReference `json:"referencedBy"`
	MaterialFor  []CardReference `json:"materialFor"`
}

type BatchCardSupport[IS IdentifierSlice] struct {
	ReferencedBy     []CardReference `json:"referencedBy"`
	MaterialFor      []CardReference `json:"materialFor"`
	UnknownResources IS              `json:"unknownResources"`
	FalsePositives   IS              `json:"falsePositives"`
}

type ArchetypalSuggestions struct {
	Total      int    `json:"total"`
	UsingName  []Card `json:"usingName"`
	UsingText  []Card `json:"usingText"`
	Exclusions []Card `json:"exclusions"`
}
