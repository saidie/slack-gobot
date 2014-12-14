package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
	"regexp"
)

var (
	IncomingUrl string = ""
	OutcomingToken string = ""
)

type Slack struct {
	Text     string `json:"text"`
	Username string `json:"username"`
}

func post(text string) bool {
	params, _ := json.Marshal(Slack {
		text,
		"gobot",
	})

	resp, _ := http.PostForm(
		IncomingUrl,
		url.Values{"payload": {string(params)}},
	)

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if string(body) != "ok" {
		os.Stderr.Write(body)
		return false
	}
	return true
}

func postCode(code string) bool {
	return post("```\n" + code + "\n```")
}

var goClient = exec.Command("gnugo", "--mode", "gtp")
var gocInput, _ = goClient.StdinPipe()
var gocOutput, _ = goClient.StdoutPipe()

func gocSend(cmd string) string {
	gocInput.Write([]byte(cmd + "\n"))

	buf := make([]byte, 4096)

	time.Sleep(2 * 100000000) // sec

	var n, err = gocOutput.Read(buf)
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	if n == 4096 {
		os.Stderr.WriteString("buffer size may be too small")
	}

	return string(buf)
}

func handle(w http.ResponseWriter, r *http.Request) {
	println((*r.URL).String())

	r.ParseForm()
	params := r.Form

	if params.Get("token") != OutcomingToken {
		return
	}

	text := params.Get("text")

	reNew, _ := regexp.Compile("^gobot *new *$")
	reShow, _ := regexp.Compile("^gobot *show *$")
	reHand, _ := regexp.Compile("^gobot *([A-T][1-9][0-9]?) *$")

	switch {
	case reNew.MatchString(text):
		println("new")
		gocSend("clear_board")

	case reShow.MatchString(text):
		println("show")

	case reHand.MatchString(text):
		println("play")
		pos := reHand.FindStringSubmatch(text)[1]
		gocSend("play black " + pos)
		gocSend("genmove_white")
	}

	postCode(gocSend("showboard"))
}

func main() {
	goClient.Stderr = os.Stderr
	goClient.Start()

	http.HandleFunc("/", handle)
	http.ListenAndServe(":8080", nil)
}
