package goa_test

import (
	. "../goa"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

/*
* This file tests the various examples in the source comment to make sure they actually work...
 */
var _ = Describe("Model", func() {

	Context("example in attribute.go", func() {

		var article = Attribute{
			Description: "An article",
			Type: Composite{
				"title": Attribute{
					Type:        String,
					Description: "Article title",
					MaxLength:   200,
				},
				"author": Attribute{
					Type: Composite{
						"firstName": Attribute{
							Type:        String,
							Description: "Author first name",
						},
						"lastName": Attribute{
							Type:        String,
							Description: "Author last name",
						},
					},
					Required: true,
				},
				"published": Attribute{
					Type:        DateTime,
					Description: "Article publication date",
					Required:    true,
				},
			},
		}

		var document = map[string]interface{}{
			"article": map[string]interface{}{
				"title": "goa, a novel go web application framework",
				"author": map[string]interface{}{
					"firstName": "Leeroy",
					"lastName":  "Jenkins",
				},
				"published": time.Now(),
			},
		}

		var documentType = Composite{
			"article": article,
		}

		It("should coerce valid values", func() {
			val, err := documentType.Load(document)
			Ω(err).Should(BeNil())
			Ω(val).Should(Equal(document))
		})
	})

	Context("example in model.go", func() {

		// Data types corresponding to how we want to manipulate the data internally in our application
		type Address struct {
			Street string `attribute:"street"` // attribute tag is used to map struct field to attribute definition below
			City   string `attribute:"city"`
		}
		type Employee struct {
			Name    string  `attribute:"name"`
			Title   string  `attribute:"title"`
			Address Address `attribute:"address"`
		}

		// External representation of the data (e.g. loaded from JSON)
		var data = map[string]interface{}{
			"name":  "John",
			"title": "Accountant",
			"address": map[string]interface{}{
				"street": "5779 Maley Drive",
				"city":   "Santa Barbara",
			},
		}

		// Model attribute definitions
		// Used to both validate and coerce external representation of the data into internal representation
		var attributes = Attributes{
			"name": Attribute{
				Type:      String,
				MinLength: 1,
				Required:  true,
			},
			"title": Attribute{
				Type:      String,
				MinLength: 1,
				Required:  true,
			},
			"address": Attribute{
				Type: Composite{
					"street": Attribute{
						Type: String,
					},
					"city": Attribute{
						Type: String,
					},
				},
			},
		}

		It("creates models", func() {
			model, err := NewModel(attributes, Employee{})
			Ω(err).ShouldNot(HaveOccurred())
			Ω(model.Attributes).Should(Equal(attributes))
		})

		It("loads data into models", func() {
			model, err := NewModel(attributes, Employee{})
			Ω(err).ShouldNot(HaveOccurred())
			employee, err := model.Load(data)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(*(employee.(*Employee))).Should(Equal(Employee{"John", "Accountant", Address{"5779 Maley Drive", "Santa Barbara"}}))
		})
	})

})
