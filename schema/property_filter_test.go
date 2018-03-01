package schema

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Property filter tests", func() {
	var filterFactory = &FilterFactory{}

	Describe("FilterFactory tests", func() {
		It("Should return error for filter with both visible and hidden properties", func() {
			_, err := filterFactory.CreateFilterFromProperties([]string{"a"}, []string{"b"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Filter tests", func() {
		var (
			filter     *Filter
			inputMap   = map[string]interface{}{"a": nil, "b": nil}
			inputSlice = []string{"a", "b"}
			key        = "a"
		)

		Context("includeAllFilter tests", func() {
			BeforeEach(func() {
				var err error
				filter, err = filterFactory.CreateFilterFromProperties(nil, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should filter map", func() {
				actual := filter.RemoveHiddenKeysFromMap(inputMap)

				expected := map[string]interface{}{"a": nil, "b": nil}
				Expect(expected).To(Equal(actual))
			})

			It("Should filter slice", func() {
				actual := filter.RemoveHiddenKeysFromSlice(inputSlice)

				expected := []string{"a", "b"}
				Expect(expected).To(Equal(actual))
			})

			It("Should filter key", func() {
				Expect(filter.IsForbidden(key)).To(BeFalse())
			})

			It("Should allow all", func() {
				Expect(filter.AllowsAll()).To(BeTrue())
			})
		})

		Context("excludeAllFilter tests", func() {
			BeforeEach(func() {
				filter = CreateExcludeAllFilter()
			})

			It("Should filter map", func() {
				actual := filter.RemoveHiddenKeysFromMap(inputMap)

				expected := map[string]interface{}{}
				Expect(expected).To(Equal(actual))
			})

			It("Should filter slice", func() {
				actual := filter.RemoveHiddenKeysFromSlice(inputSlice)

				expected := []string{}
				Expect(expected).To(Equal(actual))
			})

			It("Should filter key", func() {
				Expect(filter.IsForbidden(key)).To(BeTrue())
			})

			It("Should not allow all", func() {
				Expect(filter.AllowsAll()).To(BeFalse())
			})
		})

		Context("visibleFilter tests", func() {
			BeforeEach(func() {
				var err error
				filter, err = filterFactory.CreateFilterFromProperties([]string{key}, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should filter map", func() {
				actual := filter.RemoveHiddenKeysFromMap(inputMap)

				expected := map[string]interface{}{"a": nil}
				Expect(expected).To(Equal(actual))
			})

			It("Should filter slice", func() {
				actual := filter.RemoveHiddenKeysFromSlice(inputSlice)

				expected := []string{"a"}
				Expect(expected).To(Equal(actual))
			})

			It("Should filter key", func() {
				Expect(filter.IsForbidden(key)).To(BeFalse())
			})

			It("Should not allow all", func() {
				Expect(filter.AllowsAll()).To(BeFalse())
			})
		})

		Context("hiddenFilter tests", func() {
			BeforeEach(func() {
				var err error
				filter, err = filterFactory.CreateFilterFromProperties(nil, []string{key})
				Expect(err).ToNot(HaveOccurred())
			})

			It("Should filter map", func() {
				actual := filter.RemoveHiddenKeysFromMap(inputMap)

				expected := map[string]interface{}{"b": nil}
				Expect(expected).To(Equal(actual))
			})

			It("Should filter slice", func() {
				actual := filter.RemoveHiddenKeysFromSlice(inputSlice)

				expected := []string{"b"}
				Expect(expected).To(Equal(actual))
			})

			It("Should filter key", func() {
				Expect(filter.IsForbidden(key)).To(BeTrue())
			})

			It("Should not allow all", func() {
				Expect(filter.AllowsAll()).To(BeFalse())
			})
		})
	})
})
