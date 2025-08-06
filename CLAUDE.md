# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AWS Cost Exporter - Prometheus exporter to collect daily cost usage from AWS Cost Explorer.

Language: Go

## Development Commands

- `go mod init github.com/LOKE/aws-cost-exporter` - Initialize Go module
- `go build` - Build the application
- `go run main.go` - Run the application
- `go test ./...` - Run all tests
- `go mod tidy` - Clean up dependencies

## Architecture

Expected components:
- AWS Cost Explorer API client
- Prometheus metrics server with `/metrics` endpoint
- Configuration via environment variables
- Cost data collection and caching