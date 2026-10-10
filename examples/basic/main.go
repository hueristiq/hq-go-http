package main

import (
	"context"
	"fmt"
	"log"
	"net/http/httputil"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"
	"github.com/hueristiq/hq-lib-http-go/header"
	"github.com/hueristiq/hq-lib-http-go/method"
)

func main() {
	client, err := hqgohttp.NewClient()
	if err != nil {
		log.Fatalln(err)
	}

	res, err := client.NewRequest().
		Method(method.GET).
		URL("https://google.com").
		SetHeader(header.UserAgent.String(), "test").
		Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	defer res.Body.Close()

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
