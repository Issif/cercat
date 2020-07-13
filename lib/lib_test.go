package lib_test

import (
	"cercat/config"
	"cercat/lib"
	"cercat/pkg/homoglyph"
	"cercat/pkg/model"
	"cercat/pkg/slack"
	"io/ioutil"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	config := &config.Configuration{
		Homoglyph:     homoglyph.GetHomoglyphMap(),
		SlackUsername: "test",
		SlackIconURL:  "http://test",
	}
	reg, _ := regexp.Compile(".*test.*")
	Describe("isMatchingCert", func() {
		Describe("If certificate matches", func() {
			cert := &model.Result{Domain: "www.test.com"}
			It("should return true", func() {
				result := lib.IsMatchingCert(&config.Homoglyph, cert, reg)
				Expect(result).To(BeTrue())
			})
		})
		Describe("If alternative subject matches", func() {
			cert := &model.Result{Domain: "www.tset.net", SAN: []string{"www.test.com"}}
			It("should return true", func() {
				result := lib.IsMatchingCert(&config.Homoglyph, cert, reg)
				Expect(result).To(BeTrue())
			})
		})
		Describe("If domain is IDN", func() {
			cert := &model.Result{Domain: "www.xn--tst-rdd.com"}
			It("should return true", func() {
				result := lib.IsMatchingCert(&config.Homoglyph, cert, reg)
				Expect(result).To(BeTrue())
				Expect(cert.IDN).To(Equal("www.tеst.com")) // e is cyrillic
			})
		})
	})
	Describe("postToSlack", func() {
		msg, _ := ioutil.ReadFile("../res/cert.json")
		It("should return a valid payload", func() {
			result, err := lib.ParseResultCertificate(msg)
			slackPayload := slack.NewPayload(config, result)
			Expect(slackPayload.Text).Should(Equal("A certificate for baden-mueller.de has been issued"))
			Expect(slackPayload.Username).Should(Equal("test"))
			Expect(slackPayload.IconURL).Should(Equal("http://test"))
			Expect(err).ToNot(HaveOccurred())
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
			msg, _ := ioutil.ReadFile("../res/heartbeat.json")
			It("should return nil", func() {
				result, err := lib.ParseResultCertificate(msg)
				Expect(result).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Describe("If message is regular", func() {
			msg, _ := ioutil.ReadFile("../res/cert.json")
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
		Describe("If message is for IDN", func() {
			msg, _ := ioutil.ReadFile("../res/cert_idn.json")
			It("should return valid infos", func() {
				result, err := lib.ParseResultCertificate(msg)
				lib.IsMatchingCert(&config.Homoglyph, result, reg)
				Expect(result.Domain).Should(Equal("xn--badn-mullr-msiec.de"))
				Expect(result.IDN).Should(Equal("badеn-muеllеr.de")) // e is cyrillic
				Expect(result.SAN).Should(Equal([]string{"xn--badn-mullr-msiec.de", "www.baden-mueller.de"}))
				Expect(result.Issuer).Should(Equal("Let's Encrypt"))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
