package util

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/microcosm-cc/bluemonday"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ParseQuery(collection string, method string, filter interface{}, options interface{}, elapsed time.Duration) {
	filterJSON, _ := json.Marshal(filter)
	// optionsJSON, _ := json.Marshal(options)
	query := ""
	if os.Getenv("DB_LOG") == "true" {
		query = fmt.Sprintf("[Query Database] %s db.%s.%s(%v)", elapsed.String(), collection, method, string(filterJSON))
		fmt.Println(query)
	} else {
		query = fmt.Sprintf("[Query Database] %s db.%s.%s(...)", elapsed.String(), collection, method)
		fmt.Println(query)
	}
}

func StoragePath(elem ...string) string {
	return filepath.Join("public", "uploads", filepath.Join(elem...))
}

func StorageDir(elem ...string) string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, filepath.Join(elem...))
}

func StripTags(content string) string {
	p := bluemonday.StripTagsPolicy()
	return p.Sanitize(content)
}

func RandomHex(length int) string {
	const charset = "0123456789abcdef"
	var now = time.Now().Unix()
	var hex string
	faker := gofakeit.New(now)
	for i := 0; i < length; i++ {
		hex += faker.RandomString(strings.Split(charset, ""))
	}
	return hex
}

func HiddenPhone(phone string) string {
	phone = strings.Replace(phone, "+84", "0", 1)
	if len(phone) > 5 {
		return fmt.Sprintf("%s****", phone[0:5])
	} else {
		return phone
	}
}

func StringInArray(array []string, elm string) bool {
	for _, item := range array {
		if item == elm {
			return true
		}
	}
	return false
}
