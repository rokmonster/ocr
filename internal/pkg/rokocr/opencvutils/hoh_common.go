package opencvutils

type Coordinates struct {
	X, Y, W, H int
}

type HOHResult struct {
	T4         int               `json:"t4"`
	T5         int               `json:"t5"`
	RawResults map[string]string `json:"_raw"`
}
