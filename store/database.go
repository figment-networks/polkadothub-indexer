package store

type Database interface {
	GetTotalSize() (*GetTotalSizeResult, error)
}

type GetTotalSizeResult struct {
	Size float64 `json:"size"`
}
