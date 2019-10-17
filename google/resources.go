package google

// ResourceType is the type used to define all the Resources
// from the Provider
type ResourceType int

//go:generate enumer -type ResourceType -addprefix google_ -transform snake -linecomment
const (
	NotImplemented ResourceType = iota
)
