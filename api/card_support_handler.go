package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	json "github.com/goccy/go-json"
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	cardSupportOp = "Card Support"
)

func getCardSupportHandler(res http.ResponseWriter, req *http.Request) {
	cardID := chi.URLParam(req, "cardID")

	logger, ctx := cUtil.InitRequest(context.Background(), apiName, cardSupportOp, slog.String("card_id", cardID))
	logger.Info("Getting support cards")

	if subject, err := downstream.YGO.CardService.GetCardByID(ctx, cardID); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		cardName := (*subject).GetName()
		support := model.CardSupport{Card: *subject, ReferencedBy: make([]model.CardReference, 0), MaterialFor: make([]model.CardReference, 0)}

		if cardRefs, err := downstream.YGO.CardService.GetCardsReferencingNameInEffect(ctx, []string{cardName}); err != nil {
			err.HandleServerResponse(res)
			return
		} else {
			support.ReferencedBy, support.MaterialFor = determineSupportCards(support.Card, cardRefs)
			numNamedReferences, numMaterialReferences := len(support.ReferencedBy), len(support.MaterialFor)
			if numNamedReferences == 0 && numMaterialReferences == 0 {
				logger.Warn("Card has no support")
			} else {
				logger.Info(fmt.Sprintf("Referenced by %d non-ED cards. Referenced by %d ED cards", numNamedReferences, numMaterialReferences))
			}
			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(support)
			return
		}
	}
}

// Iterates over a list of support cards and attempts to determine if subject is found in material clause or within the body of the reference.
// If the name is found in the material clause, we can assume the subject is a required or optional summoning material - otherwise its a support card.
func determineSupportCards(subject cModel.YGOCard, references []cModel.YGOCard) ([]model.CardReference, []model.CardReference) {
	referencedBy := []model.CardReference{}
	materialFor := []model.CardReference{}

	for _, reference := range references {
		if reference.GetName() == subject.GetName() {
			continue
		}

		effect := reference.GetEffect()

		if materialString := cModel.GetPotentialMaterialsAsString(reference); materialString != "" && cardRefContainsName(materialString, subject.GetName()) {
			materialFor = append(materialFor, model.CardReference{Occurrences: 1, Card: reference})
			effect = strings.ReplaceAll(effect, materialString, "") // effect without materials
		}

		if cardRefContainsName(effect, subject.GetName()) {
			referencedBy = append(referencedBy, model.CardReference{Occurrences: 1, Card: reference})
		}
	}

	return referencedBy, materialFor
}

func cardRefContainsName(cardText, cardName string) bool {
	runes := []rune(cardText)
	nameRunes := []rune(cardName)
	textLen := len(runes)
	nameLen := len(nameRunes)

	for i := range textLen {
		if runes[i] == '"' || runes[i] == '\'' {
			start := i + 1
			end := start + nameLen

			if end >= textLen {
				break
			}

			if runes[end] != runes[i] {
				continue
			}

			if string(runes[start:end]) == cardName {
				return true
			}
		}
	}
	return false
}
