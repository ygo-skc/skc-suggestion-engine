package model

import (
	cModel "github.com/ygo-skc/skc-go/common/model"
)

type CardReference struct {
	Occurrences int         `json:"occurrences"`
	Card        cModel.Card `json:"card"`
}

type CardSuggestions struct {
	Card                 cModel.Card     `json:"card"`
	HasSelfReference     bool            `json:"hasSelfReference"`
	NamedMaterials       []CardReference `json:"namedMaterials"`
	NamedReferences      []CardReference `json:"namedReferences"`
	MaterialArchetypes   []string        `json:"materialArchetypes"`
	ReferencedArchetypes []string        `json:"referencedArchetypes"`
}

type BatchCardSuggestions[IS cModel.IdentifierSlice] struct {
	NamedMaterials       []CardReference `json:"namedMaterials"`
	NamedReferences      []CardReference `json:"namedReferences"`
	MaterialArchetypes   []string        `json:"materialArchetypes"`
	ReferencedArchetypes []string        `json:"referencedArchetypes"`
	UnknownResources     IS              `json:"unknownResources"`
	FalsePositives       IS              `json:"falsePositives"`
}

type CardSupport struct {
	Card         cModel.Card     `json:"card"`
	ReferencedBy []CardReference `json:"referencedBy"`
	MaterialFor  []CardReference `json:"materialFor"`
}

type BatchCardSupport[IS cModel.IdentifierSlice] struct {
	ReferencedBy     []CardReference `json:"referencedBy"`
	MaterialFor      []CardReference `json:"materialFor"`
	UnknownResources IS              `json:"unknownResources"`
	FalsePositives   IS              `json:"falsePositives"`
}

type ProductSuggestions[IS cModel.IdentifierSlice] struct {
	Suggestions BatchCardSuggestions[IS] `json:"suggestions"`
	Support     BatchCardSupport[IS]     `json:"support"`
}

type ArchetypalSuggestions struct {
	Total      int           `json:"total"`
	UsingName  []cModel.Card `json:"usingName"`
	UsingText  []cModel.Card `json:"usingText"`
	Exclusions []cModel.Card `json:"exclusions"`
}
