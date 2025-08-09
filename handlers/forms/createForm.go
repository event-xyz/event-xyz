package forms

import (
	"log"
	"net/http"

	"github.com/eventloop-testbed/backend/db"
	"github.com/eventloop-testbed/backend/db/utils"
	"github.com/gin-gonic/gin"
)

type Form struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Label    string   `json:"label"`
	Type     string   `json:"type"`
	Options  []string `json:"options"`
	Required bool     `json:"required"`
}

func SaveForm(c *gin.Context) {
	var form Form
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	docID := utils.GenerateDocID(form.Title)

	// Insert the form into Couchbase
	form.ID = docID
	collection := db.InitialiseBucket().Scope("eventloop").Collection("forms")
	_, err := collection.Upsert(docID, form, nil)
	if err != nil {
		log.Printf("Failed to save form: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save form"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Form saved successfully", "form_id": docID})
}
