package downstream

import (
	"log"

	"github.com/ygo-skc/skc-go/common/client"
	"github.com/ygo-skc/skc-go/common/service"
	cUtil "github.com/ygo-skc/skc-go/common/util"
)

var (
	YGOService service.YGOService
)

func CreateYGOServiceClients() {
	if client, err := client.CreateCardServiceClient("ygo-service.skc.cards", cUtil.EnvMap["YGO_SERVICE_HOST"]); err != nil {
		log.Fatalf("Failed to connect to ygo-service: %v", err)
	} else {
		YGOService = service.NewYGOServiceV1(*client)
	}
}
