// This nested go.mod marks test_data as a separate module so it is excluded from the parent module's source archive.
// The parent module cannot be archived/installed because some fixture filenames under test_data contain ':' (invalid in module zip paths).
module github.com/canonical/lscompute/test_data

go 1.26.1
