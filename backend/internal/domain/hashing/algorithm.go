package hashing

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
		return "crc32"
	case HashAlgMD5:
		return "md5"
	case HashAlgSHA1:
		return "sha1"
	case HashAlgED2K:
		return "ed2k"
	case HashAlgSHA256:
		return "sha256"
	default:
		return "unknown"
	}
}

func (a HashAlgorithm) IsValid() bool {
	return a > HashAlgUnknown && a < hashAlgSentinel
}
