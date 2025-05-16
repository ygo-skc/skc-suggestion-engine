package downstream

import (
	"log"

	"github.com/ygo-skc/skc-go/common/client"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-go/common/ygo"
)

var (
	CardServiceClient ygo.CardServiceClient
)

func CreateYGOServiceClients() {
	if client, err := client.CreateCardServiceClient("ygo-service.skc.cards", cUtil.EnvMap["YGO_SERVICE_HOST"]); err != nil {
		log.Fatalf("Failed to connect to ygo-service: %v", err)
	} else {
		CardServiceClient = *client
	}
}
