package util

import (
	"encoding/hex"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"time"
)

const KEY = "C4EFF1ED2DD62F41551163E7687EA24C"

type LicenseInfo struct {
	Time        string
	Allow       bool
	Expired     bool
	Forever     bool
	LimitDevice int64
}

func CheckLicense(license string) bool {
	/* */
	id, err := machineid.ID()
	if err != nil {
		return false
	} else {
		fmt.Println("ID", MD5Hash(id))
	}

	/* Chuyển license từ hex sang chuỗi gốc */
	licenseOrigin, err := hex.DecodeString(license)
	if err != nil {
		return false
	}
	/* Giải mã chuỗi license */
	licenseDecrypted, err := AESDecryptWithIV(KEY, MD5Hash(id), string(licenseOrigin))
	if err != nil {
		return false
	}
	if licenseDecrypted == "0000-00-00" {
		return true
	}
	if expired, err := time.Parse("2006-01-02", licenseDecrypted); err != nil {
		return false
	} else if time.Now().After(expired) {
		return false
	} else {
		return true
	}
}
