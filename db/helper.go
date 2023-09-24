package db

import (
	"fmt"
	"github.com/ostafen/clover/v2/document"
	"strings"
	"time"
)

func Document2MetaData(value *document.Document) MetaData {
	return MetaData{
		value.Get("Name").(string),
		value.Get("InfoHash").(string),
		time.Unix(value.Get("DiscoveredOn").(int64), 0).Format(time.RFC822),
		value.Get("TotalSize").(uint64),
		value.Get("Files").([]interface{}),
	}
}

func Documents2MetaData(values []*document.Document) []MetaData {
	rVal := make([]MetaData, len(values))
	for i, value := range values {
		rVal[i] = Document2MetaData(value)
	}
	return rVal
}

func MatchString(searchType string, x string, y string) bool {
	rVal := false
	switch searchType {
	case "contains":
		rVal = strings.Contains(strings.ToLower(x), strings.ToLower(y))
	case "equals":
		rVal = strings.ToLower(x) == strings.ToLower(y)
	case "startswith":
		rVal = strings.HasPrefix(strings.ToLower(x), strings.ToLower(y))
	case "endswith":
		rVal = strings.HasSuffix(strings.ToLower(x), strings.ToLower(y))
	default:
		rVal = false
	}
	return rVal
}

func HasMatchingFile(fileList []interface{}, searchType string, searchInput string) bool {
	rVal := false
	for _, item := range fileList {
		for key, value := range item.(map[string]interface{}) {
			if key == "path" {
				rVal = MatchString(searchType, value.(string), searchInput)
			}
		}
	}
	return rVal
}

func FoundOnDate(date time.Time, searchInput string) bool {
	parsedInput, err := time.Parse("2006.01.02", searchInput)
	if err != nil {
		fmt.Println(err)
		return false
	}
	endDate := parsedInput.AddDate(0, 0, 1)
	return date.After(parsedInput) && date.Before(endDate)
}

func Matches(doc *document.Document, key string, searchType string, searchInput string) bool {
	rVal := doc.Has(key)
	if rVal {
		value := doc.Get(key)

		if key == "Files" {
			rVal = HasMatchingFile(value.([]interface{}), searchType, searchInput)
		} else if key == "DiscoveredOn" {
			rVal = FoundOnDate(time.Unix(int64(value.(float64)), 0), searchInput)
		} else {
			rVal = MatchString(searchType, value.(string), searchInput)
		}
	}
	return rVal
}
