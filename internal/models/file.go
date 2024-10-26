package models

import "time"

type ExcelFile struct {
	ID          string    `json:"id"`
	Created     time.Time `json:"created"`
	SourceImage string    `json:"source_image"`
	ExcelFile   string    `json:"excel_file"`
}

type ImageFile struct {
	ID        string    `json:"id"`
	Created   time.Time `json:"created"`
	ImageFile string    `json:"image_file"`
}
