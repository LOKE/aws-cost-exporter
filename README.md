# AWS Cost Exporter

A Prometheus exporter that collects AWS cost and usage data from AWS Cost Explorer, providing both daily and monthly cost metrics grouped by service and region.

## Features

- **Daily Cost Metrics**: Today's AWS costs up until now by service and region
- **Monthly Cost Metrics**: Month-to-date AWS costs by service and region
- **Previous Day Cost Metrics**: Complete previous day AWS costs (stable data) 
- **Previous Month Cost Metrics**: Complete previous month AWS costs (stable data)
- **Prometheus Integration**: Exposes metrics at `/metrics` endpoint
- **Health Check**: Health endpoint at `/health`
- **Automatic Updates**: Metrics refresh every 6 hours
- **Comprehensive Logging**: Detailed logging for monitoring and debugging

## Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `aws_daily_cost_usd` | Gauge | Today's AWS cost in USD up until now (current data) | `service`, `region` |
| `aws_monthly_cost_usd` | Gauge | Month-to-date AWS cost in USD (current data) | `service`, `region` |
| `aws_previous_day_cost_usd` | Gauge | Complete previous day AWS cost in USD (stable data) | `service`, `region` |
| `aws_previous_month_cost_usd` | Gauge | Complete previous month AWS cost in USD (stable data) | `service`, `region` |

## Prerequisites

- AWS credentials configured (via AWS CLI, IAM roles, or environment variables)
- AWS Cost Explorer enabled in your AWS account (may take 24 hours after first enabling)
- Go 1.21+ (for building from source)

## Installation

### Using Docker

```bash
docker run -p 8080:8080 \
  -e AWS_ACCESS_KEY_ID=your_access_key \
  -e AWS_SECRET_ACCESS_KEY=your_secret_key \
  -e AWS_REGION=us-east-1 \
  ghcr.io/loke/aws-cost-exporter:latest
```

### Using Pre-built Binaries

Download the latest binary from the [releases page](https://github.com/LOKE/aws-cost-exporter/releases) and run:

```bash
./aws-cost-exporter
```

### Building from Source

```bash
git clone https://github.com/LOKE/aws-cost-exporter.git
cd aws-cost-exporter
go build -o aws-cost-exporter .
./aws-cost-exporter
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `AWS_REGION` | AWS region | Uses AWS default config |
| `AWS_ACCESS_KEY_ID` | AWS access key | Uses AWS default config |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | Uses AWS default config |

### AWS Credentials

The exporter supports all standard AWS credential methods:
- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- AWS credentials file (`~/.aws/credentials`)
- IAM roles (for EC2, ECS, Lambda)
- AWS SSO

Required AWS permissions:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ce:GetCostAndUsage"
      ],
      "Resource": "*"
    }
  ]
}
```

## Usage

1. Start the exporter:
   ```bash
   ./aws-cost-exporter
   ```

2. Check health:
   ```bash
   curl http://localhost:8080/health
   ```

3. View metrics:
   ```bash
   curl http://localhost:8080/metrics
   ```

4. Configure Prometheus to scrape the metrics:
   ```yaml
   scrape_configs:
     - job_name: 'aws-cost-exporter'
       static_configs:
         - targets: ['localhost:8080']
   ```

## Development

### Make Targets

```bash
make build    # Build the application
make run      # Run the application
make test     # Run tests
make tidy     # Clean up dependencies
make clean    # Clean build artifacts
make help     # Show available targets
```

### Project Structure

- `main.go` - Main application code
- `Makefile` - Build automation
- `.github/workflows/release.yml` - CI/CD pipeline
- `CLAUDE.md` - Development instructions for AI assistants

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Support

For issues and questions, please use the [GitHub Issues](https://github.com/LOKE/aws-cost-exporter/issues) page.
