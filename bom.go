package packit

import (
	"fmt"
	"strings"
	"time"
)

// BOMEntry contains a bill of materials entry.
type BOMEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`

	// Metadata is the metadata of the entry.  Optional.
	// Metadata map[string]interface{} `toml:"metadata,omitempty"`
	Metadata BOMMetadata `toml:"metadata,omitempty"`
}

type BOMMetadata struct {
	Architecture    string      `toml:"arch,omitempty"`
	CPE             string      `toml:"cpe,omitempty"`
	DeprecationDate time.Time   `toml:"deprecation-date,omitempty"`
	Licenses        []string    `toml:"licenses,omitempty"`
	PURL            string      `toml:"purl,omitempty"`
	Checksum        BOMChecksum `toml:"checksum,omitempty"`
	Summary         string      `toml:"summary,omitempty"`
	URI             string      `toml:"uri,omitempty"`
	Version         string      `toml:"version,omitempty"`
	Source          BOMSource   `toml:"source,omitempty"`
}

// The Algorithm type is private checksumAlgorithm instead of a string to prevent a
// non-supported algorithm string from being used.
type BOMChecksum struct {
	Algorithm checksumAlgorithm `toml:"algorithm,omitempty"`
	Hash      string            `toml:"hash,omitempty"`
}

type BOMSource struct {
	Name            string      `toml:"name,omitempty"`
	Checksum        BOMChecksum `toml:"checksum,omitempty"`
	UpstreamVersion string      `toml:"upstream-version,omitempty"`
	URI             string      `toml:"uri,omitempty"`
}

type checksumAlgorithm struct {
	name string
}

// GetBOMChecksumAlgorithm takes in an algorithm string, and reasonably tries
// to figure out the equivalent CycloneDX-supported algorithm field name.
// It returns an error if no reasonable supported format is found.
// Supported formats:
// { 'MD5'| 'SHA-1'| 'SHA-256'| 'SHA-384'| 'SHA-512'| 'SHA3-256'| 'SHA3-384'| 'SHA3-512'| 'BLAKE2b-256'| 'BLAKE2b-384'| 'BLAKE2b-512'| 'BLAKE3'}
func GetBOMChecksumAlgorithm(alg string) (checksumAlgorithm, error) {
	for _, a := range []string{SHA256, SHA1, SHA384, SHA512, SHA3256, SHA3384, SHA3512, BLAKE2B256, BLAKE2B384, BLAKE2B512, BLAKE3, MD5} {
		if strings.EqualFold(a, alg) || strings.EqualFold(strings.ReplaceAll(a, "-", ""), alg) {
			return checksumAlgorithm{name: a}, nil
		}
	}

	return checksumAlgorithm{}, fmt.Errorf("failed to get supported BOM checksum algorithm: %s is not valid", alg)
}

func (a checksumAlgorithm) String() string {
	return a.name
}

func (a checksumAlgorithm) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

const (
	SHA256     = "SHA-256"
	SHA1       = "SHA-1"
	SHA384     = "SHA-384"
	SHA512     = "SHA-512"
	SHA3256    = "SHA3-256"
	SHA3384    = "SHA3-384"
	SHA3512    = "SHA3-512"
	BLAKE2B256 = "BLAKE2b-256"
	BLAKE2B384 = "BLAKE2b-384"
	BLAKE2B512 = "BLAKE2b-512"
	BLAKE3     = "BLAKE3"
	MD5        = "MD5"
)

// UnmetEntry contains the name of an unmet dependency from the build process
type UnmetEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`
}
