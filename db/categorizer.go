package db

import (
	dhtcclient "dhtc/dhtc-client"
	"path/filepath"
	"strings"
)

func Categorize(md dhtcclient.Metadata) []string {
	if len(md.Files) == 0 {
		return []string{"Unknown"}
	}

	extensionCounts := make(map[string]int)
	for _, f := range md.Files {
		ext := strings.ToLower(filepath.Ext(f.Path))
		if ext != "" {
			extensionCounts[ext]++
		}
	}

	if len(extensionCounts) == 0 {
		return []string{"Unknown"}
	}

	// Define categories and their extensions
	categories := map[string][]string{
		"Video":    {".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpg", ".mpeg"},
		"Audio":    {".mp3", ".flac", ".wav", ".m4a", ".ogg", ".aac", ".wma"},
		"Software": {".exe", ".msi", ".dmg", ".iso", ".pkg", ".deb", ".rpm", ".sh", ".bat"},
		"Document": {".pdf", ".epub", ".mobi", ".txt", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"},
		"Picture":  {".jpg", ".jpeg", ".png", ".gif", ".svg", ".bmp", ".tiff"},
		"Archive":  {".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz"},
	}

	foundCategories := make(map[string]bool)
	for ext := range extensionCounts {
		for cat, exts := range categories {
			for _, catExt := range exts {
				if ext == catExt {
					foundCategories[cat] = true
				}
			}
		}
	}

	if len(foundCategories) == 0 {
		return []string{"Other"}
	}

	var res []string
	for cat := range foundCategories {
		res = append(res, cat)
	}

	return res
}
