package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/zenazn/goji/web"
)

const API_KEY string = "9YFMVZHI4IHDB1B3VSAQAUHTI7GUH5G5YI"

func httpget(_url string) interface{} {
	resp, err := http.Get(_url)
	if err != nil {
		print(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		print(err)
	}
	var result map[string]interface{}
	json.Unmarshal([]byte(string(body)), &result)
	return result["result"].(interface{})
}

func getBalance(address string) string {
	return httpget("https://api.polygonscan.com/api?module=account&action=balance&address=" + address + "&tag=latest&apikey=" + API_KEY).(string)
}
func getNormaltx(address string) []interface{} {
	return httpget("https://api.polygonscan.com/api?module=account&action=txlist&address=" + address + "&tag=latest&apikey=" + API_KEY).([]interface{})
}
func getERC20(address string) []interface{} {
	return httpget("https://api.polygonscan.com/api?module=account&action=tokentx&address=" + address + "&apikey=" + API_KEY).([]interface{})
}
func getERC721(address string) []interface{} {
	return httpget("https://api.polygonscan.com/api?module=account&action=tokentx&address=" + address + "&sort=asc&apikey=" + API_KEY).([]interface{})
}

type achievements struct {
	Toptenk           bool
	Token_connoisseur bool
	NFTHolder         bool
	NFTCollector      bool
	Reciever          bool
	Giver             bool
	Zen               bool
	FreshStart        bool
	TotalScore        int
}

func getscore(address string) achievements {
	trophycase := achievements{
		Toptenk:           false,
		Token_connoisseur: false,
		NFTHolder:         false,
		NFTCollector:      false,
		Reciever:          false,
		Giver:             false,
		Zen:               false,
		FreshStart:        false,
		TotalScore:        0,
	}
	var txlist []interface{} = getNormaltx(address)
	var Erc20 []interface{} = getERC20(address)
	var Erc721 []interface{} = getERC721(address)
	var walletMatic string = getBalance(address)
	// Now that data is collected, its time to load in the csv
	file, err := os.Open("matic_contracts.csv")
	if err != nil {
		fmt.Println(err)
	}
	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()

	/*
		The following reads through the csv adding contract pairs into
		a mapping of an address to a weight
	*/
	var addrweightmap map[string]interface{}
	addrweightmap = make(map[string]interface{})
	for index := 1; index < 15; index++ {
		addrweightmap[records[index][1]] = records[1][2]
	}
	var currentscore int
	var rtx int // Not that rtx recieved transactions
	var stx int
	if len(txlist) > 1 {
		currentscore = currentscore + 10
		trophycase.FreshStart = true
		for index := 0; index < len(txlist)-1; index++ {
			var currenttx = txlist[index].(map[string]interface{})
			if currenttx["to"] == "0xf2236990210083b58091966423349f17FCA486FF" {
				rtx++
				if addrweightmap[currenttx["from"].(string)] != nil {
					i, err := strconv.Atoi(addrweightmap[currenttx["from"].(string)].(string))
					if err != nil {
						fmt.Println("Weird Conversion")
					}
					currentscore = currentscore + i
				}
			} else {
				stx++
				if addrweightmap[currenttx["to"].(string)] != nil {
					i, err := strconv.Atoi(addrweightmap[currenttx["to"].(string)].(string))
					if err != nil {
						fmt.Println(err)
					}
					currentscore = currentscore + i
				}
			}
		}
	}
	var uniquetokens map[string]bool
	uniquetokens = make(map[string]bool)
	/*
		var lastqblock int
		var currentqblock int
		lastqblock = 0
		currentqblock = 0
		var quickswap bool
		quickswap = false
	*/
	for index := 0; index < len(txlist)-1; index++ {
		var currenttx = Erc20[index].(map[string]interface{})
		/*
			if addrweightmap[currenttx["contractAddress"].(string)] == "0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff" {
				currentqblock, err = strconv.Atoi(addrweightmap[currenttx["blockNumber"].(string)].(string))
				if currentqblock - 15

			}
		*/
		if addrweightmap[currenttx["contractAddress"].(string)] != nil {
			uniquetokens[addrweightmap[currenttx["contractAddress"].(string)].(string)] = true
			i, err := strconv.Atoi(addrweightmap[currenttx["contractAddress"].(string)].(string))
			if err != nil {
				fmt.Println("Weird Conversion")
			}
			currentscore = currentscore + i
		}
	}
	if len(uniquetokens) > 9 {
		currentscore = currentscore + 100 // Token Connoisseur
		trophycase.Token_connoisseur = true
	}
	var NFTcount map[string]bool
	NFTcount = make(map[string]bool)
	if len(Erc721) > 0 {
		currentscore = currentscore + 10 // NFT Holder
		trophycase.NFTHolder = true
		for index := 0; index < len(txlist)-1; index++ {
			var currenttx = Erc721[index].(map[string]interface{})
			if addrweightmap[currenttx["contractAddress"].(string)] != nil {
				NFTcount[addrweightmap[currenttx["contractAddress"].(string)].(string)] = true
				i, err := strconv.Atoi(addrweightmap[currenttx["contractAddress"].(string)].(string))
				if err != nil {
					fmt.Println("Weird Conversion")
				}
				currentscore = currentscore + i
			}

		}
	}
	if len(NFTcount) > 10 {
		currentscore = currentscore + 25 // NFT Collector
		trophycase.NFTCollector = true
	}
	var accuratebal int
	accuratebal, err = strconv.Atoi(walletMatic)
	var maticbal float64 = float64(accuratebal) / float64(1000000000000000000)
	if maticbal >= 600 {
		currentscore = currentscore + 100 // Memo - Top 10,000 Matic Holder
		trophycase.Toptenk = true
	}
	if stx > rtx {
		currentscore = currentscore + 55 // Giver
		trophycase.Giver = true
	} else if rtx > stx {
		currentscore = currentscore + 50 //Reciever
		trophycase.Reciever = true
	} else if rtx == stx {
		currentscore = currentscore + 100 //Zen
		trophycase.Zen = true
	}
	trophycase.TotalScore = currentscore
	return trophycase
}
func hostpage() {
	http.HandleFunc("/", ShowLanding)
	http.HandleFunc("/matic", ShowDashboard)
	http.ListenAndServe(":8080", nil)
}
func hello(c web.C, w http.ResponseWriter, r *http.Request) {

	//Call to ParseForm makes form fields available.
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
	}

	name := r.PostFormValue("name")
	fmt.Fprintf(w, "Hello, %s!", name)
}
func ShowDashboard(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["wallet"]

	name := "0xf2236990210083b58091966423349f17FCA486FF"

	if ok {

		name = keys[0]
	}
	walletach := getscore(name)
	fp := path.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		fmt.Println(err)
	}
	tmpl.Execute(w, walletach)
}
func ShowLanding(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("templates", "landing.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		fmt.Println(err)
	}
	tmpl.Execute(w, nil)
}
func main() {
	hostpage()
}
