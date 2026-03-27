#!/bin/bash
# cover.sh - Runs tests and calculates coverage excluding generated/internal boilerplate

# Run tests and generate profile
# Run all tests in a single command to aggregate coverage correctly
go list ./... | grep "tests" | xargs go test -v -coverprofile=coverage.out -coverpkg=./...

# Filter out generated code and main entry points from coverage.out
grep -v ".pb.go" coverage.out | grep -v "_mock.go" | grep -v "cmd/api" | grep -v "internal/grpc/aegis" > coverage_filtered.out

# Show total coverage of filtered results
go tool cover -func=coverage_filtered.out | grep total
