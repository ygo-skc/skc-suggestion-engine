package downstream

import (
	"log"

	"github.com/ygo-skc/skc-go/common/client"
	cUtil "github.com/ygo-skc/skc-go/common/util"
)

var (
	YGOClient client.YGOClientImp
)

func ConnectToYGOService() {
	if c, err := client.CreateCardServiceClient("ygo-service.skc.cards", cUtil.EnvMap["YGO_SERVICE_HOST"]); err != nil {
		log.Fatalf("Failed to connect to ygo-service: %v", err)
	} else {
		YGOClient = client.NewYGOClientImpV1(*c)
	}
}
