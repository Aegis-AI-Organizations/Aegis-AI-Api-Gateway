#!/bin/bash
# cover.sh - Runs tests and calculates coverage excluding generated/internal boilerplate

# Run tests and generate profile
go test -coverprofile=coverage.out ./...

# Filter out generated code and main entry points from coverage.out
grep -v "_pb.go" coverage.out | grep -v "cmd/api" | grep -v "internal/grpc/aegis/v1" | grep -v "_mock.go" > coverage_filtered.out

# Show total coverage of filtered results
go tool cover -func=coverage_filtered.out | grep total
