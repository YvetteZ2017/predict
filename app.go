package main

import (
	"encoding/json"
	//"encoding/json"
	"fmt"
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

func main() {
	prediction, err := predict(apiKey, "https://samples.clarifai.com/metro-north.jpg")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(prediction)
}

func predict(api_key string, photo_url string) (map[string]float64, error){
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

	var rb PredictResp

	if err := json.NewDecoder(resp.Body).Decode(&rb); err != nil {
		return nil, err
	}

	p := rb.Outputs[0].Data.Concepts
	
	fmt.Println(p)
	return nil, nil
}