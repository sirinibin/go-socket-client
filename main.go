package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	socketio_client "github.com/zhouhui8915/go-socket.io-client"
)

var desiredItems map[string]int

//var uri = "https://websocket-prod.us-east-1.elasticbeanstalk.com"

var uri = "https://ws.duelbits.com/"

//var uri = "http://localhost:8000"

var client *socketio_client.Client

func main() {

	loadDesiredItemsFromFile("desiredItems.json")
	opts := &socketio_client.Options{
		Transport: "websocket",
		/*Header: map[string][]string{
			"Origin":                   {"https://duelbits.com"},
			"Host":                     {"ws.duelbits.com"},
			"Accept-Encoding":          {"gzip, deflate, br"},
			"Cache-Control":            {"no-cache"},
			"Accept-Language":          {"en-US,en;q=0.9,ml;q=0.8"},
			"Sec-WebSocket-Extensions": {"permessage-deflate; client_max_window_bits"},
		},*/
	}

	var err error

	client, err = socketio_client.NewClient(uri, opts)
	if err != nil {
		fmt.Println("NewClient error:%v\n", err)
		return
	}

	client.On("connect", func() {
		fmt.Println("[DUELBITS] Connected to duelbits socket.")

		client.Emit("auth:authenticate", map[string]string{
			"access": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6Ijc2NTYxMTk4ODIwNDQwODU0IiwiaWF0IjoxNjM2MzMxOTg5LCJleHAiOjE2MzY5MzY3ODksImF1ZCI6Imh0dHBzOi8vZHVlbGJpdHMuY29tIiwiaXNzIjoiRHVlbGJpdHMiLCJzdWIiOiJhY2Nlc3MifQ.VvKCtL_rFKH8kr-3cF7K1Tr0OHAvHOXxsKx1MgQPjOz7aPqSdVPvoMW3ba6t5chC3DNxKovPmcRnyVwnY_4BYg",
		})

	})

	client.On("pay:p2p:newListing", func(data SocketServerData) {
		var dbPrice int

		if len(data.Items) != 1 {
			err = checkBundle(data)
			if err != nil {
				log.Fatal(err.Error())
			}
			return
		}

		dbPrice, ok := desiredItems[data.Items[0].Name]
		if !ok {
			log.Fatal("item :" + data.Items[0].Name + " not exists in desiredItems")
			return
		}

		fmt.Println("Price checking " + data.Items[0].Name + " with price of " + strconv.Itoa(data.Items[0].Price) + " and buff price of " + strconv.Itoa(dbPrice))

		if dbPrice < data.Items[0].Price {
			return
		}

		client.Emit("pay:p2p:join", map[string]interface{}{
			"tradeUrl": "https://steamcommunity.com/tradeoffer/new/?partner=860175126&token=aMtwYm4N",
			"id":       data.ID,
		}, func(data2 string) {
			if govalidator.IsNull(data2) {
				fmt.Println("[DUELBITS] Bought " + data.Items[0].Name + " for " + strconv.Itoa(data.Items[0].Price))
			} else {
				fmt.Println("[DUELBITS] FAILED: " + data2)
				fmt.Println("[DUELBITS] FAILED item " + data.Items[0].Name + " for " + strconv.Itoa(data.Items[0].Price))
			}
		},
		)

	})

	for {
	}
}

func checkBundle(data SocketServerData) error {
	var totaldbPrice = 0
	var totalSuggestedPrice = 0
	var totalPrice = 0
	var names = ""

	for _, item := range data.Items {
		//prettyLog(item)

		dbPrice, ok := desiredItems[item.Name]
		if !ok {
			return errors.New("item :" + item.Name + " not exists in desiredItems")
		}

		names += item.Name + " and"

		totaldbPrice += dbPrice

		totalSuggestedPrice += item.SuggestedPrice

		totalPrice += item.Price

		fmt.Println("Checked " + item.Name + " with price " + strconv.Itoa(item.Price) + " dbPrice " + strconv.Itoa(dbPrice) + " and suggestedPrice " + strconv.Itoa(item.SuggestedPrice))
	}

	names = strings.TrimSuffix(names, " and")

	fmt.Println("Price checking " + names + " with price of " + strconv.Itoa(totalPrice) + " and dbprice of " + strconv.Itoa(totaldbPrice))

	if (float32(totalSuggestedPrice) * float32(1.05)) > float32(totalPrice) {
		return nil
	}
	if totalPrice < 8000 {
		return nil
	}
	if float32(totaldbPrice) < (float32(totalPrice) * float32(1.05)) {
		return nil
	}

	client.Emit("pay:p2p:join", map[string]interface{}{
		"tradeUrl": "https://steamcommunity.com/tradeoffer/new/?partner=860175126&token=aMtwYm4N",
		"id":       data.ID,
	}, func(data2 string) {
		if govalidator.IsNull(data2) {
			fmt.Println("[DUELBITS] Bought " + names + " for " + strconv.Itoa(totalPrice))
		} else {
			fmt.Println("[DUELBITS] FAILED: " + data2)
			fmt.Println("[DUELBITS] FAILED items " + names + " for " + strconv.Itoa(totalPrice))
		}
	},
	)

	return nil
}

func loadDesiredItemsFromFile(filename string) error {
	plan, _ := ioutil.ReadFile(filename)
	return json.Unmarshal(plan, &desiredItems)
}

func prettyLog(res interface{}) {
	b, _ := json.MarshalIndent(res, "", "    ")
	log.Print(string(b))
}

type Seller struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	JoinSteam int64  `json:"joinSteam"`
	TradeUrl  string `json:"tradeUrl"`
}

type Item struct {
	AssetID        string `json:"assetid"`
	NameHash       string `json:"nameHash"`
	Name           string `json:"name"`
	AppID          int    `json:"appid"`
	Price          int    `json:"price"`
	Icon           string `json:"icon"`
	SuggestedPrice int    `json:"suggestedPrice"`
	Quality        string `json:"quality"`
	Rarity         string `json:"rarity"`
}

type SocketServerData struct {
	ID              string `json:"id"`
	LastChecked     int    `json:"lastChecked"`
	Type            int    `json:"type"`
	Status          int    `json:"status"`
	TradeOfferId    string `json:"tradeOfferId"`
	SellTime        int64  `json:"sellTime"`
	Seller          Seller `json:"seller"`
	WebhookEndpoint string `json:"webhookEndpoint"`
	Items           []Item `json:"items"`
	SellerId        string `json:"sellerId"`
	BuyerId         string `json:"buyerId"`
	Price           int    `json:"price"`
}
