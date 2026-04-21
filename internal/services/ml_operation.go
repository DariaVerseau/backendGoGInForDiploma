package services

type MLOperation interface {
	Endpoint() string // Например, "/upscale", "/enhance"
	GetTitle() string
	GetStyle() string
	NeedsStyle() bool
}