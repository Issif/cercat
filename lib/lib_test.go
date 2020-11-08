package lib_test

import (
	"cercat/config"
	"cercat/lib"
	"cercat/pkg/homoglyph"
	"cercat/pkg/model"
	"cercat/pkg/slack"
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/idna"
)

var _ = Describe("Handler", func() {
	config := &config.Configuration{
		Domains:           []string{"test.com"},
		HomoglyphPatterns: homoglyph.GetHomoglyphMap(),
		OmissionPatterns: map[string][]string{
			"test.com": []string{"tst", "tet", "tes"},
		},
		TranspositionPatterns: map[string][]string{
			"test.com": []string{"etst", "tset", "tets"},
		},
		RepetitionPatterns: map[string][]string{
			"test.com": []string{"teest", "tesst"},
		},
		InclusionPatterns: map[string][]string{
			"test.com": []string{"test"},
		},
		VowelSwapPatterns: map[string][]string{
			"test.com": []string{"tast", "tust", "tist", "tast"},
		},
		BitsquattingPatterns: map[string][]string{
			"test.com": []string{"tfst"},
		},
		SlackUsername: "test",
		SlackIconURL:  "http://test",
	}
	Describe("isMatchingCert", func() {
		Describe("If certificate matches (inclusion)", func() {
			cert := &model.Result{Domain: "www.montest.com"}
			It("should return true", func() {
				result, attack := lib.IsMatchingCert(cert, config)
				Expect(result).To(BeTrue())
				Expect(attack).To(Equal("inclusion"))
			})
		})
		Describe("If alternative subject matches (inclusion)", func() {
			cert := &model.Result{Domain: "www.else.net", SAN: []string{"www.montest.com"}}
			It("should return true", func() {
				result, attack := lib.IsMatchingCert(cert, config)
				Expect(result).To(BeTrue())
				Expect(attack).To(Equal("inclusion"))
			})
		})
		Describe("If domain is IDN (inclusion)", func() {
			cert := &model.Result{Domain: "www.xn--tst-rdd.com"}
			cert.IDN, _ = idna.ToUnicode(cert.Domain)
			cert.UnicodeIDN = homoglyph.ReplaceHomoglyph(cert.IDN, config.HomoglyphPatterns)
			It("should return true", func() {
				result, attack := lib.IsMatchingCert(cert, config)
				Expect(result).To(BeTrue())
				Expect(cert.IDN).To(Equal("www.tеst.com")) // e is cyrillic
				Expect(attack).To(Equal("homoglyph + inclusion"))
			})
		})
		Describe("If certificate matches (transposition)", func() {
			cert := &model.Result{Domain: "www.tset.com"}
			It("should return true", func() {
				result, attack := lib.IsMatchingCert(cert, config)
				Expect(result).To(BeTrue())
				Expect(attack).To(Equal("transposition"))
			})
		})
		Describe("If certificate matches (omission)", func() {
			cert := &model.Result{Domain: "www.tet.com"}
			It("should return true", func() {
				result, attack := lib.IsMatchingCert(cert, config)
				Expect(result).To(BeTrue())
				Expect(attack).To(Equal("omission"))
			})
		})
		Describe("If certificate matches (vowelswap)", func() {
			cert := &model.Result{Domain: "www.tist.com"}
			It("should return true", func() {
				result, attack := lib.IsMatchingCert(cert, config)
				Expect(result).To(BeTrue())
				Expect(attack).To(Equal("vowelswap"))
			})
		})
		Describe("If certificate matches (repetition)", func() {
			cert := &model.Result{Domain: "www.tesst.com"}
			It("should return true", func() {
				result, attack := lib.IsMatchingCert(cert, config)
				Expect(result).To(BeTrue())
				Expect(attack).To(Equal("repetition"))
			})
		})
		Describe("If certificate matches (bitsquatting)", func() {
			cert := &model.Result{Domain: "www.tfst.com"}
			It("should return true", func() {
				fmt.Println(config.BitsquattingPatterns)
				result, attack := lib.IsMatchingCert(cert, config)
				Expect(result).To(BeTrue())
				Expect(attack).To(Equal("bitsquatting"))
			})
		})
	})
	Describe("postToSlack", func() {
		msg, _ := ioutil.ReadFile("../res/cert.json")
		It("should return a valid payload", func() {
			result, err := lib.ParseResultCertificate(msg, &config.HomoglyphPatterns)
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
				result, err := lib.ParseResultCertificate(msg, &config.HomoglyphPatterns)
				Expect(result).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
		Describe("If message is heartbeat", func() {
			msg, _ := ioutil.ReadFile("../res/heartbeat.json")
			It("should return nil", func() {
				result, err := lib.ParseResultCertificate(msg, &config.HomoglyphPatterns)
				Expect(result).To(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Describe("If message is regular", func() {
			msg, _ := ioutil.ReadFile("../res/cert.json")
			It("should return valid infos", func() {
				result, err := lib.ParseResultCertificate(msg, &config.HomoglyphPatterns)
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
				result, err := lib.ParseResultCertificate(msg, &config.HomoglyphPatterns)
				lib.IsMatchingCert(result, config)
				Expect(result.Domain).Should(Equal("xn--badn-mullr-msiec.de"))
				Expect(result.IDN).Should(Equal("badеn-muеllеr.de")) // e is cyrillic
				Expect(result.SAN).Should(Equal([]string{"xn--badn-mullr-msiec.de", "www.baden-mueller.de"}))
				Expect(result.Issuer).Should(Equal("Let's Encrypt"))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
