package config

type CommonConfiguration struct {
	MediaDirectory     string
	TemplatesDirectory string
	OutputDirectory    string
	TessdataDirectory  string
	TmpDirectory       string
	DeleteTempFiles    bool
}
