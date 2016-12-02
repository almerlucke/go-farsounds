package farsounds

import "strings"

// Address is a container for the full path and path components
type Address struct {
	// Full path for reference
	Path string

	// Separate path components
	Components []string
}

// Message is an alias for interface{}
type Message interface{}

// NewAddress from path
func NewAddress(path string) *Address {
	cleanPath := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/")
	components := strings.Split(cleanPath, "/")

	return &Address{
		Path:       cleanPath,
		Components: components,
	}
}

// IsValid checks if the address is valid
func (address *Address) IsValid() bool {
	return len(address.Components) > 0
}

// IsResolved checks if there is only one path component left
func (address *Address) IsResolved() bool {
	return len(address.Components) == 1
}

// CurrentIdentifier returns the first address component
func (address *Address) CurrentIdentifier() string {
	if len(address.Components) > 0 {
		return address.Components[0]
	}

	return ""
}

// Next resolves the address to the next component
func (address *Address) Next() {
	if len(address.Components) > 1 {
		address.Components = address.Components[1:]
	}
}
