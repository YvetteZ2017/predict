package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)


var apiKey = os.Getenv("API_KEY")

type PredictResp struct {
	Status struct{
		Code int `json:"code"`
		Description string `json:"description"`
	} `json:"status"`
	Outputs []struct {
		Data struct {
			Concepts []struct {
				Id string `json:"id"`
				Name string `json:"name"`
				Value float64 `json:"value"`
			} `json:"concepts"`
		} `json:"data"`
	} `json:"outputs"`
}

type PredictData struct {
	Name string `json:"name"`
	Value float64 `json:"value"`
}

type TagMap struct {
	TagData struct {
		Url struct {
			Value float64
		}
	}
}


func check(e error) {
	if e != nil {
		panic(e)
	}
}


func main() {

	imagesData, err := ioutil.ReadFile("images.txt")
	check(err)
	images := strings.Split(string(imagesData), "\n")

	m := make(map[string]map[string]float64)

	for _,s := range images {
		//predictC := make(chan *PredictResp)
		prediction, err := predict(apiKey, s)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(prediction.Outputs[0].Data.Concepts)
		pred := prediction.Outputs[0].Data.Concepts
		for _,t := range pred {
			m[t.Name] = make(map[string]float64)
			m[t.Name][s] = t.Value
		}

	}
	fmt.Println(m)

}

func predict(api_key string, photo_url string) (*PredictResp, error){
	client := &http.Client{}


	api_url := "https://api.clarifai.com/v2/models/aaa03c23b3724a16a56b629203edc62c/outputs" // Use the general model

	data_body := `{"inputs": [{"data": {"image": {"url":"` + photo_url + `"}}}]}`

	req, err := http.NewRequest("POST", api_url, strings.NewReader(data_body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key " + api_key)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var rb *PredictResp

	if err := json.NewDecoder(resp.Body).Decode(&rb); err != nil {
		return nil, err
	}

	return rb, nil
}