package main

import (
    "os"
    "log"
    "encoding/json"
    "bitbucket.org/ckvist/twilio/twiml"
    "bitbucket.org/ckvist/twilio/twirest"
    "net/http"
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
    port := os.Getenv("PORT")

    if port == "" {
        log.Fatal("$PORT must be set")
    }

    http.HandleFunc("/", helloMonkey)
    http.HandleFunc("/summon", Summon)
    http.ListenAndServe(":" + port, nil)
}

func SlackReply(w http.ResponseWriter, msg string) {
    type SlackResponse struct {
        text string
    }

    if msg == "" {
        return
    }

    res := SlackResponse{ text: msg }

    b, err := json.Marshal(res)
    if err != nil {
        log.Fatal("Error encoding slack response: " + err.Error())
        w.Write([]byte("{'text': 'error'}"))
        return
    }

    w.Write([]byte(b))
}

//  INITIATION
//////////////

func Summon(w http.ResponseWriter, r *http.Request) {
    stoken := os.Getenv("SLACK_TOKEN")

    if stoken != r.FormValue("token") {
        log.Fatal("Got request with invalid slack auth token")
        SlackReply(w, "Invalid auth token")
        return
    }

    target := r.FormValue("text")
    from := r.FormValue("user_name")

    if target == "" {
        log.Fatal("got request with no text")
        SlackReply(w, "Error: no destination specified")
        return
    }

    num := "+1" + target
    Initiate(num, from)
    SlackReply(w, "Summoning " + num)
}

func Initiate(num string, fromName string) {
    accountSid := os.Getenv("ACCT_SID")
    authToken := os.Getenv("AUTH_TOKEN")
    fromNum := os.Getenv("FROM_NUM")

    client := twirest.NewClient(accountSid, authToken)

    msg := twirest.SendMessage{
            Text: "You have been summoned by " + fromName,
            To:   num,
            From: fromNum}

    resp, err := client.Request(msg)
    if err != nil {
        log.Fatal(err)
        return
    }

    log.Print(resp.Message.Status)
}