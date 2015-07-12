package main

import (
	"os"
	"log"
	"encoding/json"
	"bitbucket.org/ckvist/twilio/twiml"
	"bitbucket.org/ckvist/twilio/twirest"
	"net/http"
	"database/sql"
	_ "github.com/lib/pq"
)

var (
	db *sql.DB = nil
)

func helloMonkey(w http.ResponseWriter, r *http.Request) {
	callers := map[string]string{"+15005550001": "Langur"}

	resp := twiml.NewResponse()

	r.ParseForm()
	from := r.Form.Get("From")
	caller, ok := callers[from]

	msg := "Hello monkey"
	if ok {
		msg = "Hello " + caller
	}

	resp.Action(
		twiml.Say{Text: msg},
		twiml.Play{Url: "http://demo.twilio.com/hellomonkey/monkey.mp3"})
	resp.Send(w)
}

func main() {
	var errd error

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	db, errd = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if errd != nil {
		log.Fatalf("Error opening database: %q", errd)
	}

	http.HandleFunc("/", helloMonkey)
	http.HandleFunc("/summon", Summon)
	http.ListenAndServe(":" + port, nil)
}

func SlackReply(w http.ResponseWriter, msg string) {
	if msg == "" {
		return
	}

	res := map[string]string { "text": msg }

	b, err := json.Marshal(res)
	if err != nil {
		log.Println("Error encoding slack response: " + err.Error())
		w.Write([]byte("{'text': 'error'}"))
		return
	}

	w.Write([]byte(b))
}

//  INITIATION
//////////////

func Summon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8") 

	stoken := os.Getenv("SLACK_TOKEN")

	if stoken != r.FormValue("token") {
		log.Println("Got request with invalid slack auth token")
		SlackReply(w, "Invalid auth token")
		return
	}

	target := r.FormValue("text")
	from := r.FormValue("user_name")

	if target == "" {
		log.Println("got request with no text")
		SlackReply(w, "Error: no destination specified")
		return
	}

	// look up person in DB
	rows, err := db.Query("SELECT p.id,phone FROM person p FULL OUTER JOIN alias a on a.person=p.id WHERE p.phone IS NOT NULL AND " +
		"( a.name=$1 OR p.name=$2 )", target, target)
    if err != nil {
        log.Println("Error querying person: " + err.Error())
        return
    }
    defer rows.Close()
	var phone string
	var personID int
    for rows.Next() {
    	err = rows.Scan(&personID, &phone)
    	if err != nil {
	        log.Println("Error querying person: " + err.Error())
    		return
    	}
    	if phone != "" {
    		log.Println("found phone: " + phone)
    		break
    	}
    }

	if phone == "" {
		SlackReply(w, "Sorry, I don't have contact information for " + target)
		return
	}

	db.Exec("UPDATE person SET last_summon=NOW() WHERE id=$1", personID)

	Initiate(phone, from)
	SlackReply(w, "Summoning " + target)
}

func Initiate(num string, fromName string) {
	accountSid := os.Getenv("ACCT_SID")
	authToken := os.Getenv("AUTH_TOKEN")
	fromNum := os.Getenv("FROM_NUM")

	client := twirest.NewClient(accountSid, authToken)

	msg := twirest.SendMessage{
			Text: "You have been summoned to chat by " + fromName,
			To:   num,
			From: fromNum}

	resp, err := client.Request(msg)
	if err != nil {
		log.Println(err)
		return
	}

	log.Print(resp.Message.Status)
}