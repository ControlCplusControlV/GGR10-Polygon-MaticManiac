package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
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
func getERC20(address string) []interface{} {
	return httpget("https://api.polygonscan.com/api?module=account&action=tokentx&address=" + address + "&apikey=" + API_KEY).([]interface{})
}
func getNormaltx(address string) []interface{} {
	return httpget("https://api.polygonscan.com/api?module=account&action=txlist&address=" + address + "&tag=latest&apikey=" + API_KEY).([]interface{})
}
func getERC721(address string) []interface{} {
	return httpget("https://api.polygonscan.com/api?module=account&action=tokentx&address=" + address + "&sort=asc&apikey=" + API_KEY).([]interface{})
}
func getBalance(address string) *big.Int {
	var balance string = httpget("https://api.polygonscan.com/api?module=account&action=balance&address=" + address + "&tag=latest&apikey=" + API_KEY).(string)
	intbalance := new(big.Int)
	intbalance, ok := intbalance.SetString(balance, 10)
	if !ok {
		fmt.Print("Error parsing bal")
	}
	return intbalance
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
	Coinblurb         []string
}

func removeDuplicateValues(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func getscore(address string) achievements {
	var wallet string = strings.ToLower(address)
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
		Coinblurb:         []string{},
	}
	/*
		Declare some tracking variables among other things
	*/
	var stx int = 0
	var rtx int = 0
	var maniacScore int = 0
	var Erc20 []interface{} = getERC20(wallet)
	var txlist []interface{} = getNormaltx(wallet)
	var ERC721 []interface{} = getERC721(wallet)
	var Coinblurb []string
	var NFTcount map[string]bool = make(map[string]bool)
	var uniquetokens map[string]bool = make(map[string]bool)
	// Read through the csv and get scores
	file, err := os.Open("matic_contracts.csv")
	if err != nil {
		fmt.Println(err)
	}
	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()
	// With All Scores collected, its time to sift through the data
	var addrweightmap map[string]int = make(map[string]int)
	var coinblurbmmap map[string]string = make(map[string]string)
	for index := 1; index < 17; index++ {
		ivalue, err := strconv.Atoi(records[index][2])
		if err != nil {
			fmt.Println(err)
		}
		addrweightmap[records[index][1]] = ivalue
		coinblurbmmap[records[index][1]] = records[index][3]
	}
	/*
		Get the Maniac score earned from interacting with ERC20 Tokens
	*/
	if len(Erc20) >= 1 {
		for index := 0; index <= len(Erc20)-1; index++ {
			var currenttx = Erc20[index].(map[string]interface{})
			var weight int = addrweightmap[currenttx["contractAddress"].(string)]
			maniacScore += weight
			uniquetokens["contractAddress"] = true
			if weight > 0 {
				Coinblurb = append(Coinblurb, coinblurbmmap[currenttx["contractAddress"].(string)])
			}
		}
		if len(uniquetokens) > 9 {
			maniacScore += 100
			trophycase.Token_connoisseur = true
		}
	}
	/*
		Next get Maniac Score from just regular transactions
	*/
	if len(txlist) >= 1 {
		maniacScore += 10 // Fresh Start
		trophycase.FreshStart = true
		for index := 0; index <= len(txlist)-1; index++ {
			var currenttx = txlist[index].(map[string]interface{})
			if currenttx["to"].(string) == wallet {
				rtx++
			} else if currenttx["from"].(string) == wallet {
				stx++
			}
		}
	}
	// Use rtx and stx to determine achievement
	if rtx < stx {
		maniacScore = maniacScore + 55 // Giver
		trophycase.Giver = true
	} else if rtx > stx {
		maniacScore = maniacScore + 50 //Reciever
		trophycase.Reciever = true
	} else if rtx == stx && rtx >= 1 {
		maniacScore = maniacScore + 100 //Zen
		trophycase.Zen = true
	}
	/*
		Next get Maniac Score for interacting with ERC721s
	*/
	if len(ERC721) >= 1 {
		maniacScore += 10 // NFT Holder
		trophycase.NFTHolder = true
		for index := 0; index <= len(ERC721)-1; index++ {
			var currenttx = ERC721[index].(map[string]interface{})
			NFTcount[currenttx["contractAddress"].(string)] = true
		}
		if len(NFTcount) > 9 {
			maniacScore += 25 // NFT Collector
			trophycase.NFTCollector = true
		}
	}
	compareval := new(big.Int)
	compareval, ok := compareval.SetString("600000000000000000000", 10)
	if !ok {
		fmt.Println("SetString: error")
	}
	if getBalance(wallet).Cmp(compareval) == 1 {
		trophycase.Toptenk = true
		maniacScore += 100
	}
	Coinblurb = removeDuplicateValues(Coinblurb)
	trophycase.Coinblurb = Coinblurb
	trophycase.TotalScore = maniacScore
	return trophycase
}

func main() {
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", ShowLanding)
	http.HandleFunc("/matic", ShowDashboard)
	http.ListenAndServe(":8080", nil)
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
