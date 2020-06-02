package lib_test

import (
	"cercat/lib"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {
	CertCache := lib.GetNewCache(2)

	Describe("storeCache", func() {
		Describe("If cache under limit", func() {
			It("should store element", func() {
				CertCache.StoreCache("test")
				CertCache.StoreCache("test2")
				Expect(len(CertCache.Slab)).To(Equal(2))
				Expect(len(CertCache.List)).To(Equal(2))
				Expect(CertCache.Counter).To(Equal(2))
			})
		})
		Describe("If cache exceeds limit", func() {
			It("should remove an element", func() {
				CertCache.Reset()
				CertCache.StoreCache("test")
				CertCache.StoreCache("test2")
				CertCache.StoreCache("test3")
				Expect(len(CertCache.Slab)).To(Equal(2))
				Expect(len(CertCache.List)).To(Equal(2))
				Expect(CertCache.List[0]).To(Equal("test2"))
			})
		})
	})
	Describe("InCache", func() {
		Describe("If element in cache", func() {
			It("should return true", func() {
				CertCache.StoreCache("test")
				Expect(CertCache.InCache("test")).To(BeTrue())
			})
		})
		Describe("If element not in cache", func() {
			It("should return false", func() {
				CertCache.StoreCache("test")
				Expect(CertCache.InCache("inexistant")).ToNot(BeTrue())
			})
		})
	})
})
