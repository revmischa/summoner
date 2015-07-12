package main

import (
    "os"
    "log"
    "bitbucket.org/ckvist/twilio/twiml"
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
    http.ListenAndServe(":" + port, nil)
}
