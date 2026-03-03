package dictloader

import _ "embed"

//go:embed events.zstd.bin
var DictBytes []byte

// LoadDict loads the zstd dictionary from the embedded file
//
// Parameters:
//   - none
//
// Returns:
//   - []byte: the zstd dictionary
func LoadDict() []byte {
	return DictBytes
}
