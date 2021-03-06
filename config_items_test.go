package main_test

import (
	. "github.com/alphagov/cloudflare-configure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("ConfigItems", func() {
	Describe("CompareConfigItemsForUpdate()", func() {
		settingValAlwaysOnline := "on"
		settingValBrowserCache := 123

		It("should return nothing when local and remote are identical", func() {
			config, err := CompareConfigItemsForUpdate(
				ConfigItems{
					"always_online":     settingValAlwaysOnline,
					"browser_cache_ttl": settingValBrowserCache,
				},
				ConfigItems{
					"always_online":     settingValAlwaysOnline,
					"browser_cache_ttl": settingValBrowserCache,
				},
			)

			Expect(config).To(Equal(ConfigItemsForUpdate{}))
			Expect(err).To(BeNil())
		})

		It("should return all items in local when remote is empty", func() {
			config, err := CompareConfigItemsForUpdate(
				ConfigItems{},
				ConfigItems{
					"always_online":     settingValAlwaysOnline,
					"browser_cache_ttl": settingValBrowserCache,
				},
			)

			Expect(config).To(Equal(ConfigItemsForUpdate{
				"always_online": ConfigItemForUpdate{
					Current:  nil,
					Expected: settingValAlwaysOnline,
				},
				"browser_cache_ttl": ConfigItemForUpdate{
					Current:  nil,
					Expected: settingValBrowserCache,
				},
			}))
			Expect(err).To(BeNil())
		})

		It("should return one item in local overwriting always_online", func() {
			config, err := CompareConfigItemsForUpdate(
				ConfigItems{
					"always_online":     "off",
					"browser_cache_ttl": settingValBrowserCache,
				},
				ConfigItems{
					"always_online":     settingValAlwaysOnline,
					"browser_cache_ttl": settingValBrowserCache,
				},
			)

			Expect(config).To(Equal(ConfigItemsForUpdate{
				"always_online": ConfigItemForUpdate{
					Current:  "off",
					Expected: settingValAlwaysOnline,
				},
			}))
			Expect(err).To(BeNil())
		})

		It("should return a public error when item is missing in local", func() {
			config, err := CompareConfigItemsForUpdate(
				ConfigItems{
					"always_online":     settingValAlwaysOnline,
					"browser_cache_ttl": settingValBrowserCache,
				},
				ConfigItems{
					"browser_cache_ttl": settingValBrowserCache,
				},
			)

			Expect(config).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError(
				ConfigMismatch{Missing: ConfigItems{"always_online": settingValAlwaysOnline}}))
		})
	})

	Describe("Difference", func() {
		It("returns the difference of two ConfigItems objects", func() {
			Expect(DifferenceConfigItems(
				ConfigItems{
					"always_online": "off",
					"browser_check": "off",
				},
				ConfigItems{
					"always_online":     "on",
					"browser_check":     "off",
					"browser_cache_ttl": 14400,
				},
			)).To(Equal(
				ConfigItems{
					"always_online":     "on",
					"browser_cache_ttl": 14400,
				},
			))
		})
	})

	Describe("Union", func() {
		It("merges two ConfigItems objects overwriting values of the latter with the former", func() {
			Expect(UnionConfigItems(
				ConfigItems{
					"always_online": "off",
					"browser_check": "off",
				},
				ConfigItems{
					"always_online":     "on",
					"browser_cache_ttl": 14400,
				},
			)).To(Equal(
				ConfigItems{
					"always_online":     "on",
					"browser_check":     "off",
					"browser_cache_ttl": 14400,
				},
			))
		})
	})

	Describe("file handling", func() {
		var (
			tempDir  string
			tempFile string
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "cloudflare-configure")
			Expect(err).To(BeNil())

			tempFile = filepath.Join(tempDir, "cloudflare-configure.json")
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).To(BeNil())
		})

		configObject := ConfigItems{
			"always_online":      "off",
			"browswer_cache_ttl": float64(14400),
			"mobile_redirect": map[string]interface{}{
				"mobile_subdomain": nil,
				"status":           "off",
				"strip_uri":        false,
			},
		}
		configJSON := `{
			"always_online": "off",
			"browswer_cache_ttl": 14400,
			"mobile_redirect": {
				"mobile_subdomain": null,
				"status": "off",
				"strip_uri": false
			}
		}`

		Describe("SaveConfigItems()", func() {
			It("should save ConfigItems to a file as pretty-formatted JSON", func() {
				err := SaveConfigItems(configObject, tempFile)
				Expect(err).To(BeNil())

				out, err := ioutil.ReadFile(tempFile)
				Expect(out).To(MatchJSON(configJSON))
				Expect(err).To(BeNil())
			})
		})

		Describe("LoadConfigItems()", func() {
			It("should read ConfigItems from a JSON file", func() {
				err := ioutil.WriteFile(tempFile, []byte(configJSON), 0644)
				Expect(err).To(BeNil())

				out, err := LoadConfigItems(tempFile)
				Expect(out).To(Equal(configObject))
				Expect(err).To(BeNil())
			})
		})
	})
})
