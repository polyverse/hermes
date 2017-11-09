package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/polyverse-security/hermes"
)

var (
	serverAddr string

	parentUrl     string
	ourName       string
	parentFromEnv bool

	children string
)

func main() {
	flag.StringVar(&serverAddr, "serveraddr", ":9091", "Sets the address at which to expose Hermes status reports")

	flag.StringVar(&parentUrl, "parenturl", "", "Sets the Parent URL to push to.")
	flag.StringVar(&ourName, "ourname", "", "Sets our name for the parent to get updates from.")
	flag.BoolVar(&parentFromEnv, "parentfromenv", false, "Sets our parent parameters from Environment")

	flag.StringVar(&children, "children", "", "Set scraping children with the form: 'childname1=http://childurl1,childname2=http://childurl2'")

	flag.Parse()

	if parentFromEnv {
		hermes.SetParentFromEnv()
	} else if parentUrl != "" && ourName != "" {
		u, err := url.Parse(parentUrl)
		if err != nil {
			fmt.Printf("Unable to parse Parent URL %s, due to error: %s\n", parentUrl, err.Error())
			return
		}

		hermes.SetParent(ourName, *u)
	}

	hermes.SetDefaultPushToParent(false)

	go hermes.PushModelToParentForever(context.Background(), time.Duration(30)*time.Second)
	go generateHermesStatusUpdates()

	if children != "" {
		childs := strings.Split(children, ",")
		for _, child := range childs {
			childparams := strings.Split(child, "=")
			u, err := url.Parse(childparams[1])
			if err != nil {
				fmt.Printf("Error when parsing child url %s: %s\n", childparams[1], err.Error())
				return
			}
			go hermes.ScrapeStatusForever(context.Background(), time.Duration(20)*time.Second, childparams[0], *u)
		}
	}

	fmt.Printf("Listening and serving hermes statuses at address: %s\n", serverAddr)
	err := http.ListenAndServe(serverAddr, hermes.GetHandler())
	fmt.Printf("Error when listening for Hermes updates: %s\n", err.Error())
}

func generateHermesStatusUpdates() {
	hermes.ReportStatus("Allowing Timeout", "This key will expire.", hermes.WithTTL(time.Duration(5)*time.Minute))

	for {
		hermes.ReportStatus("pushedToParentImmediately", fmt.Sprintf("It is now %s", time.Now().String()), hermes.WithPushToParent())
		hermes.ReportStatus("pushedToParentAtInterval", fmt.Sprintf("It is now %s", time.Now().String()))

		time.Sleep(time.Duration(5) * time.Second)
	}
}
