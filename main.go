package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Frameit struct {
	Type   string `json:"type"`
	Button string `json:"button"`
	Text   string `json:"text"`
}

func mdfunc(w http.ResponseWriter, r *http.Request) {
	htmlText, btns := MD2HTMLButtonsV2(r.FormValue("rtext"))
	formatedtrext := Frameit{
		Type:   "ok",
		Button: fmt.Sprintf("%v", btns),
		Text:   htmlText,
	}
	if err := json.NewEncoder(w).Encode(formatedtrext); err != nil {
		log.Print(err.Error())
	}
}

func homePage(w http.ResponseWriter, _ *http.Request) {
	if _, err := fmt.Fprint(w, "Ok!"); err != nil {
		log.Println(err.Error())
	}
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	port := os.Getenv("PORT")
	if port == "" {
		port = "9002"
	}
	server := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	router.HandleFunc("/", homePage)
	router.HandleFunc("/get/", mdfunc).Methods("POST")
	log.Println("Started API!")
	log.Fatal(server.ListenAndServe())
}
