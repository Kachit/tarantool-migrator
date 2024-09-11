package tarantool_migrator

// Options define options for all migrations.
type Options struct {
	// SpaceName is the migrations space.
	SpaceName string
	Instances []string

	UseTransaction bool
	WithoutConfirm bool
}

var DefaultOptions = &Options{
	SpaceName: "migrations",
}
