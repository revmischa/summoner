package main

import (
	"bitbucket.org/ckvist/twilio/twiml"
	"bitbucket.org/ckvist/twilio/twirest"
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	db *sql.DB = nil
)

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

	http.HandleFunc("/summonCallback", summonCallback)
	http.HandleFunc("/summon", Summon)
	http.ListenAndServe(":"+port, nil)
}

func SlackReply(w http.ResponseWriter, msg string) {
	if msg == "" {
		return
	}

	res := map[string]string{"text": msg}

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
	trigger := r.FormValue("trigger_word")

	// remove trigger from target
	target = strings.Replace(target, trigger, "", 1)
	target = strings.TrimSpace(target)
	target = strings.ToLower(target)

	if target == "" {
		log.Println("got request with no text")
		SlackReply(w, "Error: no destination specified")
		return
	}

	// look up person in DB
	rows, err := db.Query("SELECT p.id,phone FROM person p FULL OUTER JOIN alias a on a.person=p.id WHERE p.phone IS NOT NULL AND "+
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
			break
		}
	}

	if phone == "" {
		SlackReply(w, "Sorry, I don't have contact information for "+target)
		return
	}

	db.Exec("UPDATE person SET last_summon=NOW() WHERE id=$1", personID)

	Initiate(phone)
	SMS(phone, from)
	SlackReply(w, "Summoning "+target)
}

func SMS(num string, fromName string) {
	client := TwiREST()
	fromNum := os.Getenv("SMS_FROM_NUM")

	msg := twirest.SendMessage{
		Text: "You have been summoned to chat by " + fromName,
		To:   num,
		From: fromNum}

	_, err := client.Request(msg)
	if err != nil {
		log.Println(err)
		return
	}
}

func Initiate(num string) {
	client := TwiREST()
	fromNum := os.Getenv("CALL_FROM_NUM")
	callSid := os.Getenv("CALL_APPLICATION_SID")

	msg := twirest.MakeCall{
		To:   num,
		From: fromNum,
		ApplicationSid: callSid}

	_, err := client.Request(msg)
	if err != nil {
		log.Println(err)
		return
	}
}

/* when caller picks up */
func summonCallback(w http.ResponseWriter, r *http.Request) {
	resp := twiml.NewResponse()

	r.ParseForm()
	resp.Action(
		twiml.Play{Url: "https://s3-us-west-2.amazonaws.com/hard.chat/summoner/summon1.mp3"})
	resp.Send(w)
}

func TwiREST() *twirest.TwilioClient {
	accountSid := os.Getenv("ACCT_SID")
	authToken := os.Getenv("AUTH_TOKEN")	
	client := twirest.NewClient(accountSid, authToken)
	return client
}