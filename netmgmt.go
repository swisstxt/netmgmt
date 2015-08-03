package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sontags/env"
	"github.com/sontags/logger"
)

type configuration struct {
	Port         string `json:"port"`
	File         string `json:"file"`
	LockDuration string `json:"lockDuration"`
}

func (c configuration) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

var config configuration

func init() {
	env.Var(&config.Port, "PORT", "8080", "Port to bind to")
	env.Var(&config.File, "FILE", "data/ipranges.json", "Base directories of the repos")
	env.Var(&config.LockDuration, "LOCK_DURATION", "30", "Duration of a lock in minutes")
}

var locker Locker
var networks *[]network

func main() {
	env.Parse("NETMGMT", false)

	duration, err := strconv.Atoi(config.LockDuration)
	if err != nil {
		log.Fatal(err)
	}

	locker.Init(duration)
	locker.Add("10.100.18.28", "temp", "menetda`")

	f, err := os.Open(config.File)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	networks, err = ReadNetworks(b)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/nodes/{node}", GetNodeInfo).Methods("GET")
	router.HandleFunc("/networks", GetNetworks).Methods("GET")
	router.HandleFunc("/networks/{net}", GetNetwork).Methods("GET")
	router.HandleFunc("/networks/{net}/list", GetNetworkIps).Methods("GET")
	router.HandleFunc("/networks/{net}", PostReservation).Methods("POST")

	n := negroni.New(
		negroni.NewRecovery(),
		logger.NewLogger(),
		cors.New(cors.Options{AllowedOrigins: []string{"*"}}),
	)
	n.UseHandler(router)

	http.ListenAndServe(":"+config.Port, n)

}
