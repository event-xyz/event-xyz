package db

import (
	"github.com/couchbase/gocb/v2"
)

// CreateDocument inserts a document into the specified collection.
// Returns the document ID and any error encountered.
func CreateDocument(collection *gocb.Collection, collectionName string, docID string, doc any) (*gocb.MutationResult, error) {
	mutRes, err := collection.Upsert(docID, doc, nil)
	return mutRes, err
}
