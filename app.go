package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)


var apiKey = os.Getenv("API_KEY")

var myLog = log.New(os.Stderr, "app: ", log.LstdFlags | log.Lshortfile)

type PredictResp struct {
	Status struct{
		Code int `json:"code"`
		Description string `json:"description"`
	} `json:"status"`
	Outputs []struct {
		Input struct {
			Data struct {
				Image struct {
					Url string `json:"url"`
				} `json:"image"`
			} `json:"data"`
		} `json:"input"`
		Data struct {
			Concepts []struct {
				Id string `json:"id"`
				Name string `json:"name"`
				Value float64 `json:"value"`
			} `json:"concepts"`
		} `json:"data"`
	} `json:"outputs"`
}

type Pair struct {
	Key string
	Value float64
}

type PairList []Pair

type TagMap map[string]PairList

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }


func readMapFromJson(fileName string) TagMap { // read the built tagMap from the json file
	jsonFile, err := os.Open(fileName)
	if err != nil {
		log.Println(errors.New("build the tagMap first with the command: -build=true [path to the image_file]"))
		log.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var tagMapFromJson TagMap
	json.Unmarshal(byteValue, &tagMapFromJson)

	return tagMapFromJson
}


func buildMap(imageFilePath string) { // build the tagMap with imageFile, save as json file
	imagesData, err := ioutil.ReadFile(imageFilePath)
	if err != nil {
		panic(err)
	}
	images := strings.Split(string(imagesData), "\n")

	m := make(map[string]map[string]float64)

	predictChan := make(chan *PredictResp, 100)

	var wg sync.WaitGroup
	for _,s := range images {
		wg.Add(1)
		go predict(apiKey, s, predictChan, &wg)
	}
	go func() {
		defer close(predictChan)
		wg.Wait()
	}()

	for prediction := range predictChan {
		pred := prediction.Outputs[0].Data.Concepts
		url := prediction.Outputs[0].Input.Data.Image.Url
		for _,t := range pred {
			if _, ok := m[t.Name]; ok {
				m[t.Name][url] = t.Value
			} else {
				m[t.Name] = make(map[string]float64)
				m[t.Name][url] = t.Value
			}
		}
	}

	// check m
	a := 0
	for k := range m {
		if len(m[k]) > 10 {
			a++
		}
	}
	fmt.Println(a)

	//

	newMap := make(map[string][]Pair)

	for tag, url := range m {

		pl := make(PairList, len(m[tag]))
		i := 0
		for k, v := range url {
			pl[i] = Pair{k, v}
			i++
		}
		sort.Sort(sort.Reverse(pl))

		newMap[tag] = pl
	}

	b, err := json.Marshal(newMap)
	if err != nil {
		log.Print(err)
	}

	jsonFile, err := os.Create("./tagMap.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	jsonFile.Write(b)

}

func searchKeyword(keyword string, tagMap TagMap) []string {
	var temp []string
	if val, ok := tagMap[keyword]; ok {
		bound := 10

		if len(val) < bound {
			bound = len(val)
		}
		for i := 0; i < bound; i++ {
			temp = append(temp, val[i].Key)
		}
	}
	return temp
}


func predict(api_key string, photo_url string, c chan *PredictResp, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{}


	api_url := "https://api.clarifai.com/v2/models/aaa03c23b3724a16a56b629203edc62c/outputs" // Use the general model

	data_body := `{"inputs": [{"data": {"image": {"url":"` + photo_url + `"}}}]}`

	req, err := http.NewRequest("POST", api_url, strings.NewReader(data_body))
	if err != nil {
		myLog.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key " + api_key)

	resp, err := client.Do(req)
	if err != nil {
		myLog.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		myLog.Fatal(errors.New(resp.Status))
	}

	var rb *PredictResp

	if err := json.NewDecoder(resp.Body).Decode(&rb); err != nil {
		myLog.Fatal(err)
	}

	c <- rb
}


func main() {
	start := time.Now()

	buildPtr  := flag.Bool("build", false, "build the image tag-map with command -build [path to the image_url .txt file")


	searchPtr := flag.String("search", "nature", "keyword")
	flag.Parse()
	imageFilePathInput := flag.Args()


	if *buildPtr {
		imageFilePath := "imagest.txt"
		if len(imageFilePathInput) > 0 {
			imageFilePath = imageFilePathInput[0]
		}
		buildMap(imageFilePath)
	}

	tagMapFromJson := readMapFromJson("tagMap.json")

	urlList := searchKeyword(*searchPtr, tagMapFromJson)


	fmt.Println(urlList)

	elapsed := time.Since(start)
	fmt.Println(elapsed)
}
