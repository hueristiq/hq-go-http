package main

import (
	"fmt"
	"log"
	"net/http/httputil"

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/hq-go-http/header"
)

func main() {
	client := hqgohttp.DefaultClient

	res, err := client.Get("https://google.com", &hqgohttp.RequestConfiguration{
		Headers: []hqgohttp.Header{
			hqgohttp.NewSetHeader(header.UserAgent.String(), "test"),
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	rawReq, err := httputil.DumpRequestOut(res.Request, true)
	if err != nil {
		log.Fatalf("Error dumping request: %v", err)
	}

	fmt.Printf("----- RAW HTTP REQUEST -----\n%s\n", string(rawReq))

	rawResp, err := httputil.DumpResponse(res, true)
	if err != nil {
		log.Fatalf("Error dumping response: %v", err)
	}

	fmt.Printf("----- RAW HTTP RESPONSE -----\n%s\n", string(rawResp))
}
