package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func main() {
	prometheusURL := "http://localhost:9090"

	endTime := time.Now()
	endTime = endTime.Add(5 * time.Minute)
	startTime := endTime.Add(-20 * 24 * time.Hour)

	queryOpts := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  time.Minute * 5,
	}

	exportedService := "summon-dragon-qa-dragon-qa-web-8000@kubernetes"

	var serviceData []map[string]interface{}

	prometheusQuery := fmt.Sprintf("sum(rate(traefik_service_requests_total{exported_service=\"%v\"}[5m]))", exportedService)

	client, err := api.NewClient(api.Config{Address: prometheusURL})
	if err != nil {
		fmt.Println("Error creating client", err)
		return
	}

	promAPI := v1.NewAPI(client)
	v, _, err := promAPI.QueryRange(context.Background(), prometheusQuery, queryOpts)
	if err != nil {
		fmt.Println("Error querying Prometheus server", err)
		return
	}

	if v == nil || len(v.(model.Matrix)) == 0 {
		fmt.Printf("No data found for %s\n", exportedService)
		return
	}

	for _, sample := range v.(model.Matrix)[0].Values {
		timestamp := time.Unix(int64(sample.Timestamp)/1000, 0).Format("2006-01-02 15:04:05")
		value := fmt.Sprint(sample.Value)

		serviceData = append(serviceData, map[string]interface{}{
			"Timestamp": timestamp,
			"Value":     value,
		})
	}

	jsonBytes, err := json.Marshal(serviceData)
	if err != nil {
		fmt.Println("Error marshaling data to JSON:", err)
		return
	}

	outputFile, err := os.Create("output.json")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	_, err = outputFile.Write(jsonBytes)
	if err != nil {
		fmt.Println("Error writing JSON data:", err)
		return
	}
}
