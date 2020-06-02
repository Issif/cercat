package lib_test

import (
	"cercat/lib"
	"io/ioutil"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	config := &lib.Configuration{
		Homoglyph: lib.GetHomoglyphMap(),
	}
	reg, _ := regexp.Compile(".*test.*")
	Describe("isMatchingCert", func() {
		Describe("If certificate matches", func() {
			cert := &lib.Result{Domain: "www.test.com"}
			It("should return true", func() {
				result := lib.IsMatchingCert(config, cert, reg)
				Expect(result).To(BeTrue())
			})
		})
		Describe("If alternative subject matches", func() {
			cert := &lib.Result{Domain: "www.test.net", SAN: []string{"www.test.com"}}
			It("should return true", func() {
				result := lib.IsMatchingCert(config, cert, reg)
				Expect(result).To(BeTrue())
			})
		})
		Describe("If domain is IDN", func() {
			cert := &lib.Result{Domain: "xn--tst-rdd.com"}
			It("should return true", func() {
				result := lib.IsMatchingCert(config, cert, reg)
				Expect(result).To(BeTrue())
				Expect(cert.IDN).To(Equal("tеst.com")) // e is cyrillic
			})
		})
	})
	Describe("parseResultCertificate", func() {
		Describe("If cannot marshall message", func() {
			msg := []byte("")
			It("should return nil and error", func() {
				result, err := lib.ParseResultCertificate(msg)
				Expect(result).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
		Describe("If message is heartbeat", func() {
			msg, _ := ioutil.ReadFile("./res/heartbeat.json")
			It("should return nil", func() {
				result, err := lib.ParseResultCertificate(msg)
				Expect(result).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Describe("If message is regular", func() {
			msg, _ := ioutil.ReadFile("./res/cert.json")
			It("should return valid infos", func() {
				result, err := lib.ParseResultCertificate(msg)
				Expect(result.Domain).Should(Equal("baden-mueller.de"))
				Expect(result.IDN).Should(Equal(""))
				Expect(result.SAN).Should(Equal([]string{"baden-mueller.de", "www.baden-mueller.de"}))
				Expect(result.Issuer).Should(Equal("Let's Encrypt"))
				Expect(result.Addresses).Should(Equal([]string{"23.236.62.147"}))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
