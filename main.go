package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	awsCostGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aws_daily_cost_usd",
			Help: "Daily AWS cost in USD",
		},
		[]string{"service", "region"},
	)
)

type CostExporter struct {
	client *costexplorer.Client
}

func init() {
	prometheus.MustRegister(awsCostGauge)
}

func NewCostExporter(client *costexplorer.Client) *CostExporter {
	return &CostExporter{client: client}
}

func (e *CostExporter) updateMetrics(ctx context.Context) error {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	
	start := yesterday.Format("2006-01-02")
	end := now.Format("2006-01-02")


	costInput := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &start,
			End:   &end,
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  &[]string{string(types.DimensionService)}[0],
			},
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  &[]string{string(types.DimensionRegion)}[0],
			},
		},
	}

	result, err := e.client.GetCostAndUsage(ctx, costInput)
	if err != nil {
		return fmt.Errorf("failed to get cost and usage: %w", err)
	}

	awsCostGauge.Reset()

	for _, resultByTime := range result.ResultsByTime {
		for _, group := range resultByTime.Groups {
			if len(group.Keys) >= 2 {
				service := group.Keys[0]
				region := group.Keys[1]
				if cost, ok := group.Metrics["UnblendedCost"]; ok && cost.Amount != nil {
					amount, err := strconv.ParseFloat(*cost.Amount, 64)
					if err == nil {
						awsCostGauge.WithLabelValues(service, region).Set(amount)
					}
				}
			}
		}
	}

	return nil
}

func (e *CostExporter) startMetricsUpdater(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	e.updateMetrics(ctx)

	for {
		select {
		case <-ticker.C:
			if err := e.updateMetrics(ctx); err != nil {
				log.Printf("Error updating metrics: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	client := costexplorer.NewFromConfig(cfg)
	exporter := NewCostExporter(client)

	go exporter.startMetricsUpdater(ctx)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting AWS Cost Exporter on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}