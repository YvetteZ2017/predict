package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)


var apiKey = os.Getenv("API_KEY")

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


func main() {
	start := time.Now()

	imagesData, err := ioutil.ReadFile("images.txt")
	if err != nil {
		panic(err)
	}
	images := strings.Split(string(imagesData), "\n")

	m := make(map[string]map[string]float64)

	predictChan := make(chan *PredictResp, 500)

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

	var a int
	for k := range m {
		if len(m[k]) > 10 {
			a++
		}
	}
	fmt.Println(m)
	fmt.Println(a)

	elapsed := time.Since(start)
	fmt.Println(elapsed)
}

func predict(api_key string, photo_url string, c chan *PredictResp, wg *sync.WaitGroup) (error){
	defer wg.Done()
	client := &http.Client{}


	api_url := "https://api.clarifai.com/v2/models/aaa03c23b3724a16a56b629203edc62c/outputs" // Use the general model

	data_body := `{"inputs": [{"data": {"image": {"url":"` + photo_url + `"}}}]}`

	req, err := http.NewRequest("POST", api_url, strings.NewReader(data_body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key " + api_key)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	var rb *PredictResp

	if err := json.NewDecoder(resp.Body).Decode(&rb); err != nil {
		return err
	}

	c <- rb

	return nil
}