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
	
	awsMonthlyCostGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aws_monthly_cost_usd",
			Help: "Monthly AWS cost in USD",
		},
		[]string{"service", "region"},
	)
	
	awsPreviousDayCostGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aws_previous_day_cost_usd",
			Help: "Previous day AWS cost in USD (stable metric)",
		},
		[]string{"service", "region"},
	)
	
	awsPreviousMonthCostGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aws_previous_month_cost_usd",
			Help: "Previous month AWS cost in USD (stable metric)",
		},
		[]string{"service", "region"},
	)
)

type CostExporter struct {
	client *costexplorer.Client
}

func init() {
	prometheus.MustRegister(awsCostGauge)
	prometheus.MustRegister(awsMonthlyCostGauge)
	prometheus.MustRegister(awsPreviousDayCostGauge)
	prometheus.MustRegister(awsPreviousMonthCostGauge)
}

func NewCostExporter(client *costexplorer.Client) *CostExporter {
	return &CostExporter{client: client}
}

func (e *CostExporter) updateDailyMetrics(ctx context.Context) error {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	
	start := yesterday.Format("2006-01-02")
	end := now.Format("2006-01-02")

	log.Printf("Fetching daily cost data from %s to %s", start, end)

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
		log.Printf("Failed to fetch daily cost data from AWS Cost Explorer: %v", err)
		return fmt.Errorf("failed to get daily cost and usage: %w", err)
	}

	log.Printf("Received %d daily result periods from AWS Cost Explorer", len(result.ResultsByTime))

	awsCostGauge.Reset()

	metricsCount := 0
	for _, resultByTime := range result.ResultsByTime {
		for _, group := range resultByTime.Groups {
			if len(group.Keys) >= 2 {
				service := group.Keys[0]
				region := group.Keys[1]
				if cost, ok := group.Metrics["UnblendedCost"]; ok && cost.Amount != nil {
					amount, err := strconv.ParseFloat(*cost.Amount, 64)
					if err == nil {
						awsCostGauge.WithLabelValues(service, region).Set(amount)
						metricsCount++
					}
				}
			}
		}
	}

	log.Printf("Updated %d daily cost metrics for period %s to %s", metricsCount, start, end)
	return nil
}

func (e *CostExporter) updateMonthlyMetrics(ctx context.Context) error {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	start := startOfMonth.Format("2006-01-02")
	end := now.Format("2006-01-02")

	log.Printf("Fetching monthly cost data from %s to %s", start, end)

	costInput := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &start,
			End:   &end,
		},
		Granularity: types.GranularityMonthly,
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
		log.Printf("Failed to fetch monthly cost data from AWS Cost Explorer: %v", err)
		return fmt.Errorf("failed to get monthly cost and usage: %w", err)
	}

	log.Printf("Received %d monthly result periods from AWS Cost Explorer", len(result.ResultsByTime))

	awsMonthlyCostGauge.Reset()

	metricsCount := 0
	for _, resultByTime := range result.ResultsByTime {
		for _, group := range resultByTime.Groups {
			if len(group.Keys) >= 2 {
				service := group.Keys[0]
				region := group.Keys[1]
				if cost, ok := group.Metrics["UnblendedCost"]; ok && cost.Amount != nil {
					amount, err := strconv.ParseFloat(*cost.Amount, 64)
					if err == nil {
						awsMonthlyCostGauge.WithLabelValues(service, region).Set(amount)
						metricsCount++
					}
				}
			}
		}
	}

	log.Printf("Updated %d monthly cost metrics for period %s to %s", metricsCount, start, end)
	return nil
}

func (e *CostExporter) updatePreviousDayMetrics(ctx context.Context) error {
	now := time.Now()
	twoDaysAgo := now.AddDate(0, 0, -2)
	yesterday := now.AddDate(0, 0, -1)
	
	start := twoDaysAgo.Format("2006-01-02")
	end := yesterday.Format("2006-01-02")

	log.Printf("Fetching previous day cost data from %s to %s", start, end)

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
		log.Printf("Failed to fetch previous day cost data from AWS Cost Explorer: %v", err)
		return fmt.Errorf("failed to get previous day cost and usage: %w", err)
	}

	log.Printf("Received %d previous day result periods from AWS Cost Explorer", len(result.ResultsByTime))

	awsPreviousDayCostGauge.Reset()

	metricsCount := 0
	for _, resultByTime := range result.ResultsByTime {
		for _, group := range resultByTime.Groups {
			if len(group.Keys) >= 2 {
				service := group.Keys[0]
				region := group.Keys[1]
				if cost, ok := group.Metrics["UnblendedCost"]; ok && cost.Amount != nil {
					amount, err := strconv.ParseFloat(*cost.Amount, 64)
					if err == nil {
						awsPreviousDayCostGauge.WithLabelValues(service, region).Set(amount)
						metricsCount++
					}
				}
			}
		}
	}

	log.Printf("Updated %d previous day cost metrics for period %s to %s", metricsCount, start, end)
	return nil
}

func (e *CostExporter) updatePreviousMonthMetrics(ctx context.Context) error {
	now := time.Now()
	// Get the first day of the previous month
	firstDayOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	firstDayOfPreviousMonth := firstDayOfCurrentMonth.AddDate(0, -1, 0)
	
	start := firstDayOfPreviousMonth.Format("2006-01-02")
	end := firstDayOfCurrentMonth.Format("2006-01-02")

	log.Printf("Fetching previous month cost data from %s to %s", start, end)

	costInput := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &start,
			End:   &end,
		},
		Granularity: types.GranularityMonthly,
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
		log.Printf("Failed to fetch previous month cost data from AWS Cost Explorer: %v", err)
		return fmt.Errorf("failed to get previous month cost and usage: %w", err)
	}

	log.Printf("Received %d previous month result periods from AWS Cost Explorer", len(result.ResultsByTime))

	awsPreviousMonthCostGauge.Reset()

	metricsCount := 0
	for _, resultByTime := range result.ResultsByTime {
		for _, group := range resultByTime.Groups {
			if len(group.Keys) >= 2 {
				service := group.Keys[0]
				region := group.Keys[1]
				if cost, ok := group.Metrics["UnblendedCost"]; ok && cost.Amount != nil {
					amount, err := strconv.ParseFloat(*cost.Amount, 64)
					if err == nil {
						awsPreviousMonthCostGauge.WithLabelValues(service, region).Set(amount)
						metricsCount++
					}
				}
			}
		}
	}

	log.Printf("Updated %d previous month cost metrics for period %s to %s", metricsCount, start, end)
	return nil
}

func (e *CostExporter) updateMetrics(ctx context.Context) error {
	if err := e.updateDailyMetrics(ctx); err != nil {
		return err
	}
	
	if err := e.updateMonthlyMetrics(ctx); err != nil {
		return err
	}
	
	if err := e.updatePreviousDayMetrics(ctx); err != nil {
		return err
	}
	
	if err := e.updatePreviousMonthMetrics(ctx); err != nil {
		return err
	}
	
	return nil
}

func (e *CostExporter) startMetricsUpdater(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour)
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

	// Load AWS config with region fallback
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1" // Default region
		log.Printf("AWS_REGION not set, using default region: %s", region)
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	client := costexplorer.NewFromConfig(cfg)
	exporter := NewCostExporter(client)

	// Update metrics immediately on startup
	log.Printf("Updating metrics on startup...")
	if err := exporter.updateMetrics(ctx); err != nil {
		log.Printf("Warning: Failed to update metrics on startup: %v", err)
	}

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