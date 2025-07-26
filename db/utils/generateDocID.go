package utils

import "time"

// generate unique doc id using timestamp and identifier string
func GenerateDocID(docName string) string {
	return docName + "_" + time.Now().Format("20060102150405.000000")
}
