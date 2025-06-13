package model

import (
	cModel "github.com/ygo-skc/skc-go/common/model"
)

type CardReference struct {
	Occurrences int            `json:"occurrences"`
	Card        cModel.YGOCard `json:"card"`
}

type CardSuggestions struct {
	Card                 cModel.YGOCard  `json:"card"`
	HasSelfReference     bool            `json:"hasSelfReference"`
	NamedMaterials       []CardReference `json:"namedMaterials"`
	NamedReferences      []CardReference `json:"namedReferences"`
	MaterialArchetypes   []string        `json:"materialArchetypes"`
	ReferencedArchetypes []string        `json:"referencedArchetypes"`
}

type BatchCardSuggestions[RK cModel.YGOResourceKey] struct {
	NamedMaterials       []CardReference `json:"namedMaterials"`
	NamedReferences      []CardReference `json:"namedReferences"`
	MaterialArchetypes   []string        `json:"materialArchetypes"`
	ReferencedArchetypes []string        `json:"referencedArchetypes"`
	UnknownResources     RK              `json:"unknownResources"`
	FalsePositives       RK              `json:"falsePositives"`
}

type CardSupport struct {
	Card         cModel.YGOCard  `json:"card"`
	ReferencedBy []CardReference `json:"referencedBy"`
	MaterialFor  []CardReference `json:"materialFor"`
}

type BatchCardSupport[RK cModel.YGOResourceKey] struct {
	ReferencedBy     []CardReference `json:"referencedBy"`
	MaterialFor      []CardReference `json:"materialFor"`
	UnknownResources RK              `json:"unknownResources"`
	FalsePositives   RK              `json:"falsePositives"`
}

type ProductSuggestions[RK cModel.YGOResourceKey] struct {
	Suggestions BatchCardSuggestions[RK] `json:"suggestions"`
	Support     BatchCardSupport[RK]     `json:"support"`
}

type ArchetypalSuggestions struct {
	Total      int              `json:"total"`
	UsingName  []cModel.YGOCard `json:"usingName"`
	UsingText  []cModel.YGOCard `json:"usingText"`
	Exclusions []cModel.YGOCard `json:"exclusions"`
}

// looks for a self reference, if a self reference is found it is removed from original slice
// this method returns true if a self reference is found
func RemoveSelfReference(self string, cr *[]CardReference) bool {
	hasSelfRef := false

	if cr != nil {
		x := 0
		for _, ref := range *cr {
			if ref.Card.GetName() != self {
				(*cr)[x] = ref
				x++
			} else {
				hasSelfRef = true
			}
		}

		*cr = (*cr)[:x]
		return hasSelfRef
	} else {
		return hasSelfRef
	}
}
