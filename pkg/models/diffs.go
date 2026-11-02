package models

type DiffFormat string

const (
	DiffFormatText       DiffFormat = "text"
	DiffFormatJSON       DiffFormat = "json"
	DiffFormatJSONSchema DiffFormat = "json-schema"
	DiffFormatConfigMap  DiffFormat = "configmap"
)

var validDiffFormats = []DiffFormat{
	DiffFormatText,
	DiffFormatJSON,
	DiffFormatJSONSchema,
	DiffFormatConfigMap,
}

func (f DiffFormat) IsValid() bool {
	for _, format := range validDiffFormats {
		if f == format {
			return true
		}
	}
	return false
}

// DiffConfig selects changed files to render in the PR validation comment
type DiffConfig struct {
	Paths  []string   `yaml:"paths"`
	Format DiffFormat `yaml:"format"`
}

type DiffConfigs []DiffConfig

// DiffSection is one rendered block of the file changes comment, e.g. a single ConfigMap data key
type DiffSection struct {
	Key     string
	Summary string
	// Body is the markdown shown inside the details block, empty for summary-only sections
	Body string
	// Open renders the details block expanded
	Open bool
	// TruncatedRows is the number of rows dropped from the table by the row cap
	TruncatedRows int
}
