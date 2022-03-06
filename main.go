package main

import (
	"io/ioutil"

	"github.com/cvhariharan/gemini-crawler/gemini"
	"github.com/cvhariharan/gemini-crawler/gemtext"
)

func main() {
	client := gemini.NewClient(gemini.ClientOptions{Insecure: true})
	resp, _ := client.Fetch("gemini://gemini.circumlunar.space/")
	txt, _ := ioutil.ReadAll(resp.Body)
	gemtext.Parse(string(txt), "gemini://gemini.circumlunar.space/")
}
