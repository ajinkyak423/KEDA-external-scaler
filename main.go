package main

import (
	"context"
	"fmt"
	"log"
	pb "my-external-scaler/externalscaler"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExternalScaler struct {
	pb.ExternalScalerServer
}

type USGSResponse struct {
	Features []USGSFeature `json:"features"`
}

type USGSFeature struct {
	Properties USGSProperties `json:"properties"`
}

type USGSProperties struct {
	Mag float64 `json:"mag"`
}

func getData(prometheusAddress string, query string) (string, error) {

	endTime := time.Now()
	endTime = endTime.Add(5 * time.Minute)
	startTime := endTime.Add(-20 * 24 * time.Hour)

	queryOpts := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  time.Minute * 5,
	}

	var serviceData []map[string]interface{}

	prometheusQuery := query

	client, err := api.NewClient(api.Config{
		Address: prometheusAddress,
	})
	if err != nil {
		fmt.Println("Error authenticating Prometheus server", err)
		return "", err
	}

	promAPI := v1.NewAPI(client)
	v, _, err := promAPI.QueryRange(context.Background(), prometheusQuery, queryOpts)
	if err != nil {
		fmt.Println("Error querying Prometheus server", err)
		return "", err
	}

	var lastValue string

	for _, sample := range v.(model.Matrix)[0].Values {
		timestamp := time.Unix(int64(sample.Timestamp)/1000, 0).Format("2006-01-02 15:04:05")
		value := fmt.Sprint(sample.Value)

		serviceData = append(serviceData, map[string]interface{}{
			"Timestamp": timestamp,
			"Value":     value,
		})
		lastValue = value
	}

	return lastValue, nil
}

func (e *ExternalScaler) IsActive(ctx context.Context, scaledObject *pb.ScaledObjectRef) (*pb.IsActiveResponse, error) {
	// prometheusAddress := scaledObject.ScalerMetadata["prometheusAddress"]
	// query := scaledObject.ScalerMetadata["query"]

	// prometheusAddress := "http://prometheus-infra.prometheus.svc.cluster.local:9090"
	// tanent := "summon-dragon-qa-dragon-qa-web-8000@kubernetes"
	// query := fmt.Sprintf("sum(rate(traefik_service_requests_total{exported_service=\"%v\"}[5m]))", tanent)

	// if len(longitude) == 0 || len(latitude) == 0 {
	// 	return nil, status.Error(codes.InvalidArgument, "longitude and latitude must be specified")
	// }

	// startTime := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	// endTime := time.Now().Format("2006-01-02")
	// radiusKM := 500
	// query := fmt.Sprintf("format=geojson&starttime=%s&endtime=%s&longitude=%s&latitude=%s&maxradiuskm=%d", startTime, endTime, longitude, latitude, radiusKM)

	// resp, err := http.Get(fmt.Sprintf("https://earthquake.usgs.gov/fdsnws/event/1/query?%s", query))
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// defer resp.Body.Close()
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// payload := USGSResponse{}
	// err = json.Unmarshal(body, &payload)
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// count := 0
	// for _, f := range payload.Features {
	// 	if f.Properties.Mag > 1.0 {
	// 		count++
	// 	}
	// }

	// return &pb.IsActiveResponse{
	// 	Result: count > 2,
	// }, nil

	// serviceData, err := getData(prometheusAddress, query)
	serviceData := "4"
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	serviceDataInt, err := strconv.ParseFloat(serviceData, 64)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.IsActiveResponse{
		Result: serviceDataInt > -1,
	}, nil
}

func (e *ExternalScaler) GetMetricSpec(context.Context, *pb.ScaledObjectRef) (*pb.GetMetricSpecResponse, error) {
	return &pb.GetMetricSpecResponse{
		MetricSpecs: []*pb.MetricSpec{{
			MetricName: "eqThreshold",
			TargetSize: 10,
		}},
	}, nil
}

func (e *ExternalScaler) GetMetrics(_ context.Context, metricRequest *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	// prometheusAddress := metricRequest.ScaledObjectRef.ScalerMetadata["prometheusAddress"]
	// query := metricRequest.ScaledObjectRef.ScalerMetadata["query"]

	// if len(longitude) == 0 || len(latitude) == 0 {
	// 	return nil, status.Error(codes.InvalidArgument, "longitude and latitude must be specified")
	// }

	// prometheusAddress := "http://prometheus-infra.prometheus.svc.cluster.local:9090"
	// tanent := "summon-dragon-qa-dragon-qa-web-8000@kubernetes"
	// query := fmt.Sprintf("sum(rate(traefik_service_requests_total{exported_service=\"%v\"}[5m]))", tanent)

	// serviceData, err := getData(prometheusAddress, query)
	serviceData := "4"
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	serviceDataInt, err := strconv.ParseFloat(serviceData, 64)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// earthquakeCount, err := getEarthQuakeCount(longitude, latitude, 1.0)
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	return &pb.GetMetricsResponse{
		MetricValues: []*pb.MetricValue{{
			MetricName:  "earthquakeThreshold",
			MetricValue: int64(serviceDataInt),
		}},
	}, nil
}

func getEarthQuakeCount(longitude, latitude string, magThreshold float64) (int, error) {
	return 0, nil
}

func (e *ExternalScaler) StreamIsActive(scaledObject *pb.ScaledObjectRef, epsServer pb.ExternalScaler_StreamIsActiveServer) error {
	longitude := scaledObject.ScalerMetadata["longitude"]
	latitude := scaledObject.ScalerMetadata["latitude"]

	if len(longitude) == 0 || len(latitude) == 0 {
		return status.Error(codes.InvalidArgument, "longitude and latitude must be specified")
	}

	for {
		select {
		case <-epsServer.Context().Done():
			// call cancelled
			return nil
		case <-time.Tick(time.Hour * 1):
			earthquakeCount, err := getEarthQuakeCount(longitude, latitude, 1.0)
			if err != nil {
				// log error
			} else if earthquakeCount > 2 {
				err = epsServer.Send(&pb.IsActiveResponse{
					Result: true,
				})
			}
		}
	}
}

func main() {
	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", ":6000")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	pb.RegisterExternalScalerServer(grpcServer, &ExternalScaler{})

	fmt.Println("Listening on :6000")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
