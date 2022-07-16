package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ip2location/ip2location-go/v9"
	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

var (
	ipDB *ip2location.DB
)

func init() {
	var err error
	ipDB, err = ip2location.OpenDB("./data/IPv4-DB.BIN")

	if err != nil {
		log.Fatalln("Could not load IP DB file...")
	}
}

func submitNewTrafficData(res http.ResponseWriter, req *http.Request) {
	log.Println("Adding new traffic record...")
	var body model.TrafficAnalysisInput

	if b, err := ioutil.ReadAll(req.Body); err != nil {
		log.Println("Error occurred while reading the request body.")
	} else {
		json.Unmarshal(b, &body)
	}

	ipData, _ := ipDB.Get_all(body.IP)

	location := model.Location{Zip: ipData.Zipcode, City: ipData.City, Country: ipData.Country_short}
	userData := model.UserData{Location: location, IP: body.IP}
	trafficAnalysis := model.TrafficAnalysis{Timestamp: time.Now(), UserData: userData, ResourceUtilized: body.ResourceUtilized}

	db.XXX(trafficAnalysis)
}
