package dtos

type (
	Hashes []Hash

	Hash struct {
		Algorithm string `json:"alg,omitzero"`
		Value     string `json:"value,omitzero"`
	}
)
