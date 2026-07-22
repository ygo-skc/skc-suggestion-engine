package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	cModel "github.com/ygo-skc/skc-go/common/v3/model"
	"github.com/ygo-skc/skc-go/common/v3/parser"
	cUtil "github.com/ygo-skc/skc-go/common/v3/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	cardSupportOp = "Card Support"
)

func getCardSupportHandler(res http.ResponseWriter, req *http.Request) {
	cardID := chi.URLParam(req, "cardID")

	logger, ctx := cUtil.InitRequest(req.Context(), apiName, cardSupportOp, slog.String("card_id", cardID))
	logger.Info("Getting support cards")

	cardProto, err := downstream.YGO.CardService.GetCardByIDProto(ctx, cardID)
	if err != nil {
		err.HandleServerResponse(res)
		return
	}
	subject := cModel.YGOCardRESTFromProto(cardProto)

	cardName := (subject).GetName()
	support := model.CardSupport{Card: subject, ReferencedBy: make([]model.CardReference, 0), MaterialFor: make([]model.CardReference, 0)}

	cr, err := downstream.YGO.CardService.GetCardsReferencingNameInEffectProto(ctx, []string{cardName})
	if err != nil {
		err.HandleServerResponse(res)
		return
	}
	cardRefsProto := cModel.YGOCardListRESTFromProto(cr)

	support.ReferencedBy, support.MaterialFor = determineSupportCards(support.Card, cardRefsProto)
	numNamedReferences, numMaterialReferences := len(support.ReferencedBy), len(support.MaterialFor)
	if numNamedReferences == 0 && numMaterialReferences == 0 {
		logger.Warn("Card has no support")
	} else {
		logger.Info("Card support generated", "referenced_by_count", numNamedReferences, "material_for_count", numMaterialReferences)
	}

	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(support); err != nil {
		logger.Error("Could not encode card support response", "err", err)
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

		if materialString := cModel.GetPotentialMaterialsAsString(reference); materialString != "" && parser.TextContainsSubStr(materialString, subject.GetName()) {
			materialFor = append(materialFor, model.CardReference{Occurrences: 1, Card: reference})
			effect = strings.ReplaceAll(effect, materialString, "") // effect without materials
		}

		if parser.TextContainsSubStr(effect, subject.GetName()) {
			referencedBy = append(referencedBy, model.CardReference{Occurrences: 1, Card: reference})
		}
	}

	return referencedBy, materialFor
}
