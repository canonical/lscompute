// This file is added here to exclude the test_data directory from being treated as part of the main go module.
// It needs to be excluded as the test data contains the `:` character which is not allowed in module paths and will cause `go mod tidy` to fail.
module github.com/canonical/lscompute/test_data

go 1.26.1
