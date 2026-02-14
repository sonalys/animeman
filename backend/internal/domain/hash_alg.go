package domain

type HashAlgorithm uint

const (
	HashAlgUnknown HashAlgorithm = iota
	HashAlgCRC32
	HashAlgMD5
	HashAlgSHA1
	HashAlgED2K
	HashAlgSHA256
	hashAlgSentinel
)

func (a HashAlgorithm) String() string {
	switch a {
	case HashAlgCRC32:
		return "CRC32"
	case HashAlgMD5:
		return "MD5"
	case HashAlgSHA1:
		return "SHA1"
	case HashAlgED2K:
		return "ED2K"
	case HashAlgSHA256:
		return "SHA256"
	default:
		return "UNKNOWN"
	}
}

func (a HashAlgorithm) IsValid() bool {
	return a > HashAlgUnknown && a < hashAlgSentinel
}
