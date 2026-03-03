package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/ostafen/clover/v2/document"
)

func Document2MetaData(value *document.Document) MetaData {
	var categories []string
	if value.Has("Categories") {
		switch v := value.Get("Categories").(type) {
		case []string:
			categories = v
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok {
					categories = append(categories, s)
				}
			}
		case string:
			categories = strings.Split(v, ",")
		}
	} else if value.Has("Category") {
		// Migration path from old "Category" field
		cat, _ := value.Get("Category").(string)
		categories = []string{cat}
	}

	name, _ := value.Get("Name").(string)
	infoHash, _ := value.Get("InfoHash").(string)
	discoveredOn, _ := value.Get("DiscoveredOn").(int64)
	totalSize, _ := value.Get("TotalSize").(uint64)
	files, _ := value.Get("Files").([]any)

	return MetaData{
		Name:         name,
		InfoHash:     infoHash,
		DiscoveredOn: time.Unix(discoveredOn, 0).Format(time.RFC822),
		TotalSize:    totalSize,
		Files:        files,
		Categories:   categories,
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

func HasMatchingFile(fileList []any, searchType, searchInput string) bool {
	for _, item := range fileList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for key, value := range m {
			if key == "Path" {
				val, ok := value.(string)
				if !ok {
					continue
				}
				if MatchString(searchType, val, searchInput) {
					return true
				}
			}
		}
	}
	return false
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
	if key == "All" {
		return Matches(doc, "Name", searchType, searchInput) || Matches(doc, "Files", searchType, searchInput)
	}
	if key == "Path" {
		key = "Files"
	}
	rVal := doc.Has(key)
	if rVal {
		value := doc.Get(key)

		if key == "Files" {
			fList, ok := value.([]any)
			if ok {
				rVal = HasMatchingFile(fList, searchType, searchInput)
			}
		} else if key == "DiscoveredOn" {
			vInt, _ := value.(int64)
			rVal = FoundOnDate(time.Unix(vInt, 0), searchInput)
		} else {
			vStr, _ := value.(string)
			rVal = MatchString(searchType, vStr, searchInput)
		}
	}
	return rVal
}
