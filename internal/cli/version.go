package cli

// Build-time variables injected via ldflags (see Makefile).
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Version returns the build version string.
func Version() string {
	return version
}

// Commit returns the short commit hash from build time.
func Commit() string {
	return commit
}

// BuildDate returns the build timestamp.
func BuildDate() string {
	return date
}
