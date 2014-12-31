package goa_test

import (
	. "../goa"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Attribute", func() {

	Describe("String", func() {

		Context("with a valid string", func() {
			var raw = "foo"

			It("should coerce", func() {
				Ω(String.Load(raw)).Should(Equal(raw))
			})

			It("should have the right kind", func() {
				Ω(String.GetKind()).Should(Equal(TString))
			})
		})

		Context("with an invalid value", func() {
			var raw = 42

			It("should not coerce", func() {
				_, err := String.Load(raw)
				Ω(err).Should(HaveOccurred())
			})

			It("should provide an informative error on coerce", func() {
				_, err := String.Load(raw)
				message := err.Error()
				Ω(message).Should(ContainSubstring("int"))
				Ω(message).Should(ContainSubstring("String"))
				Ω(message).Should(ContainSubstring("42"))
			})
		})

	})

	Describe("Composite", func() {

		Context("with a simple map", func() {
			composite := Composite(map[string]Attribute{"foo": Attribute{Type: Integer}})
			raw := map[string]interface{}{"foo": "1"}

			It("coerces", func() {
				Ω(composite.Load(raw)).Should(Equal(map[string]interface{}{"foo": int(1)}))
			})

			It("has the right kind", func() {
				Ω(composite.GetKind()).Should(Equal(TComposite))
			})
		})

		Context("with a recursive map", func() {
			composite := Composite(map[string]Attribute{"foo": Attribute{Type: Composite{"bar": Attribute{Type: String}}}})
			raw := map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}

			It("coerces", func() {
				Ω(composite.Load(raw)).Should(Equal(raw))
			})

			It("has the right kind", func() {
				Ω(composite.GetKind()).Should(Equal(TComposite))
			})
		})
	})

})
