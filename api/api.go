package api

import (
	"encoding/json"
	"log"
	"net/http"

	monigodb "github.com/iyashjayesh/monigo/monigoDb"
)

func GetServiceInfoAPI(w http.ResponseWriter, r *http.Request) {

	dbObj := monigodb.GetDbInstance()
	serviceInfo := dbObj.GetServiceDetails()

	
	serviceInfo, err := dbObj.GetServiceInfo(serviceInfo.ServiceName)
	if err != nil {
		log.Println("Error getting service info:", err)
	}

	jsonServiceInfo, err := json.Marshal(serviceInfo)
	if err != nil {
		log.Println("Error marshalling service info:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonServiceInfo)
}
