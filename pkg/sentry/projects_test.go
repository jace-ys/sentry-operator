package sentry_test

import (
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jace-ys/sentry-operator/pkg/sentry"
)

var _ = Describe("ProjectsService", func() {
	Describe("List", func() {
		var (
			projects []sentry.Project
			resp     *sentry.Response
			err      error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/projects/list.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/projects/",
			testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Link", newPaginationLinks())
				w.Write(fixture)
			}),
		)

		JustBeforeEach(func() {
			projects, resp, err = client.Projects.List()
		})

		It("returns a 200 OK response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusOK))

			Expect(resp.NextPage).To(Equal(&sentry.Page{
				URL:     "https://sentry.io/api/0/next",
				Results: false,
			}))
			Expect(resp.PrevPage).To(Equal(&sentry.Page{
				URL:     "https://sentry.io/api/0/previous",
				Results: true,
			}))

			Expect(projects).To(Equal([]sentry.Project{
				{
					Avatar: sentry.Avatar{
						AvatarType: "letter_avatar",
					},
					Color:        "#bf6e3f",
					DateCreated:  parseTime("2018-11-06T21:20:08.064Z"),
					Features:     []string{"servicehooks", "sample-events", "data-forwarding", "rate-limits", "minidump"},
					HasAccess:    true,
					ID:           "4",
					IsBookmarked: false,
					IsInternal:   false,
					IsMember:     true,
					IsPublic:     false,
					Name:         "The Spoiled Yoghurt",
					Organization: sentry.Organization{
						Avatar: sentry.Avatar{
							AvatarType: "letter_avatar",
						},
						DateCreated:    parseTime("2018-11-06T21:19:55.101Z"),
						ID:             "2",
						IsEarlyAdopter: false,
						Name:           "The Interstellar Jurisdiction",
						Require2FA:     false,
						Slug:           "the-interstellar-jurisdiction",
						Status: sentry.OrganizationStatus{
							ID:   "active",
							Name: "active",
						},
					},
					Slug:   "the-spoiled-yoghurt",
					Status: "active",
				},
			}))
		})
	})

	Describe("Get", func() {
		var (
			projectSlug string

			project *sentry.Project
			resp    *sentry.Response
			err     error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/projects/get.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/projects/organization/valid/",
			testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
				w.Write(fixture)
			}),
		)

		BeforeEach(func() {
			projectSlug = "valid"
		})

		JustBeforeEach(func() {
			project, resp, err = client.Projects.Get("organization", projectSlug)
		})

		It("returns a 200 OK response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusOK))

			Expect(project).To(Equal(&sentry.Project{
				Avatar: sentry.Avatar{
					AvatarType: "letter_avatar",
				},
				Color:        "#3fbf7f",
				DateCreated:  parseTime("2018-11-06T21:19:55.121Z"),
				Features:     []string{"releases", "sample-events", "minidump", "servicehooks", "rate-limits", "data-forwarding"},
				HasAccess:    true,
				ID:           "2",
				IsBookmarked: false,
				IsInternal:   false,
				IsMember:     true,
				IsPublic:     false,
				Name:         "Pump Station",
				Organization: sentry.Organization{
					Avatar: sentry.Avatar{
						AvatarType: "letter_avatar",
					},
					DateCreated:    parseTime("2018-11-06T21:19:55.101Z"),
					ID:             "2",
					IsEarlyAdopter: false,
					Name:           "The Interstellar Jurisdiction",
					Require2FA:     false,
					Slug:           "the-interstellar-jurisdiction",
					Status: sentry.OrganizationStatus{
						ID:   "active",
						Name: "active",
					},
				},
				Slug:   "pump-station",
				Status: "active",
				Team: sentry.Team{
					ID:   "2",
					Name: "Powerful Abolitionist",
					Slug: "powerful-abolitionist",
				},
				Teams: []sentry.Team{
					{
						ID:   "2",
						Name: "Powerful Abolitionist",
						Slug: "powerful-abolitionist",
					},
				},
			}))
		})

		Context("when project does not exist", func() {
			handler.HandleFunc("/api/0/projects/organization/invalid/",
				testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(newAPIError(sentry.APIError{"detail": "The requested resource does not exist"}))
				}),
			)

			BeforeEach(func() {
				projectSlug = "invalid"
			})

			It("returns a 404 Not Found error", func() {
				Expect(err).To(MatchError(sentry.APIError{"detail": "The requested resource does not exist"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusNotFound))
			})
		})
	})

	Describe("Update", func() {
		var (
			exists bool
			params *sentry.UpdateProjectParams

			project *sentry.Project
			resp    *sentry.Response
			err     error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/projects/update.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/projects/organization/test/",
			testHandler(http.MethodPut, func(w http.ResponseWriter, r *http.Request) {
				if exists {
					w.WriteHeader(http.StatusBadRequest)
					w.Write(newAPIError(sentry.APIError{"slug": "Another project is already using that slug"}))
					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write(fixture)
			}),
		)

		BeforeEach(func() {
			exists = false
			params = &sentry.UpdateProjectParams{
				Name: "test",
				Slug: "test",
			}
		})

		JustBeforeEach(func() {
			project, resp, err = client.Projects.Update("organization", "test", params)
		})

		It("returns a 200 OK response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusOK))

			Expect(project).To(Equal(&sentry.Project{
				Avatar: sentry.Avatar{
					AvatarType: "letter_avatar",
				},
				Color:        "#bf803f",
				DateCreated:  parseTime("2018-11-06T21:20:19.624Z"),
				Features:     []string{"releases", "sample-events", "minidump", "servicehooks", "rate-limits", "data-forwarding"},
				HasAccess:    true,
				ID:           "5",
				IsBookmarked: false,
				IsInternal:   false,
				IsMember:     true,
				IsPublic:     false,
				Name:         "Plane Proxy",
				Organization: sentry.Organization{
					Avatar: sentry.Avatar{
						AvatarType: "letter_avatar",
					},
					DateCreated:    parseTime("2018-11-06T21:19:55.101Z"),
					ID:             "2",
					IsEarlyAdopter: false,
					Name:           "The Interstellar Jurisdiction",
					Require2FA:     false,
					Slug:           "the-interstellar-jurisdiction",
					Status: sentry.OrganizationStatus{
						ID:   "active",
						Name: "active",
					},
				},
				Platform: "javascript",
				Slug:     "plane-proxy",
				Status:   "active",
				Team: sentry.Team{
					ID:   "2",
					Name: "Powerful Abolitionist",
					Slug: "powerful-abolitionist",
				},
				Teams: []sentry.Team{
					{
						ID:   "2",
						Name: "Powerful Abolitionist",
						Slug: "powerful-abolitionist",
					},
				},
			}))
		})

		Context("project already exists", func() {
			BeforeEach(func() {
				exists = true
			})

			It("returns a 400 Bad Request response", func() {
				Expect(err).To(MatchError(sentry.APIError{"slug": "Another project is already using that slug"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusBadRequest))
			})
		})
	})

	Describe("Delete", func() {
		var (
			projectSlug string

			resp *sentry.Response
			err  error
		)

		handler, client := setup()

		handler.HandleFunc("/api/0/projects/organization/valid/",
			testHandler(http.MethodDelete, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
		)

		BeforeEach(func() {
			projectSlug = "valid"
		})

		JustBeforeEach(func() {
			resp, err = client.Projects.Delete("organization", projectSlug)
		})

		It("returns a 204 No Content response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusNoContent))
		})

		Context("when project does not exist", func() {
			handler.HandleFunc("/api/0/projects/organization/invalid/",
				testHandler(http.MethodDelete, func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(newAPIError(sentry.APIError{"detail": "The requested resource does not exist"}))
				}),
			)

			BeforeEach(func() {
				projectSlug = "invalid"
			})

			It("returns a 404 Not Found error", func() {
				Expect(err).To(MatchError(sentry.APIError{"detail": "The requested resource does not exist"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusNotFound))
			})
		})
	})

	Describe("ListKeys", func() {
		var (
			keys []sentry.ProjectKey
			resp *sentry.Response
			err  error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/project_keys/list.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/projects/organization/project/keys/",
			testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Link", newPaginationLinks())
				w.Write(fixture)
			}),
		)

		JustBeforeEach(func() {
			keys, resp, err = client.Projects.ListKeys("organization", "project")
		})

		It("returns a 200 OK response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusOK))

			Expect(resp.NextPage).To(Equal(&sentry.Page{
				URL:     "https://sentry.io/api/0/next",
				Results: false,
			}))
			Expect(resp.PrevPage).To(Equal(&sentry.Page{
				URL:     "https://sentry.io/api/0/previous",
				Results: true,
			}))

			Expect(keys).To(Equal([]sentry.ProjectKey{
				{
					BrowserSDK: sentry.ProjectKeyBrowserSDK{
						Choices: [][]string{{"latest", "latest"}, {"4.x", "4.x"}},
					},
					BrowserSDKVersion: "4.x",
					DateCreated:       parseTime("2018-11-06T21:20:07.941Z"),
					DSN: sentry.ProjectKeyDSN{
						CDN:      "https://sentry.io/js-sdk-loader/cec9dfceb0b74c1c9a5e3c135585f364.min.js",
						CSP:      "https://sentry.io/api/2/csp-report/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
						Minidump: "https://sentry.io/api/2/minidump/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
						Public:   "https://cec9dfceb0b74c1c9a5e3c135585f364@sentry.io/2",
						Secret:   "https://cec9dfceb0b74c1c9a5e3c135585f364:4f6a592349e249c5906918393766718d@sentry.io/2",
						Security: "https://sentry.io/api/2/security/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
					},
					ID:        "cec9dfceb0b74c1c9a5e3c135585f364",
					IsActive:  true,
					Label:     "Fabulous Key",
					Name:      "Fabulous Key",
					ProjectID: 2,
					Public:    "cec9dfceb0b74c1c9a5e3c135585f364",
					RateLimit: sentry.ProjectKeyRateLimit{
						Window: 0,
						Count:  0,
					},
					Secret: "4f6a592349e249c5906918393766718d",
				},
			}))
		})
	})

	Describe("GetKey", func() {
		var (
			keyID string

			key  *sentry.ProjectKey
			resp *sentry.Response
			err  error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/project_keys/get.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/projects/organization/project/keys/valid/",
			testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
				w.Write(fixture)
			}),
		)

		BeforeEach(func() {
			keyID = "valid"
		})

		JustBeforeEach(func() {
			key, resp, err = client.Projects.GetKey("organization", "project", keyID)
		})

		It("returns a 200 OK response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusOK))

			Expect(key).To(Equal(&sentry.ProjectKey{
				BrowserSDK: sentry.ProjectKeyBrowserSDK{
					Choices: [][]string{{"latest", "latest"}, {"4.x", "4.x"}},
				},
				BrowserSDKVersion: "4.x",
				DateCreated:       parseTime("2018-11-06T21:20:07.941Z"),
				DSN: sentry.ProjectKeyDSN{
					CDN:      "https://sentry.io/js-sdk-loader/cec9dfceb0b74c1c9a5e3c135585f364.min.js",
					CSP:      "https://sentry.io/api/2/csp-report/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
					Minidump: "https://sentry.io/api/2/minidump/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
					Public:   "https://cec9dfceb0b74c1c9a5e3c135585f364@sentry.io/2",
					Secret:   "https://cec9dfceb0b74c1c9a5e3c135585f364:4f6a592349e249c5906918393766718d@sentry.io/2",
					Security: "https://sentry.io/api/2/security/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
				},
				ID:        "cec9dfceb0b74c1c9a5e3c135585f364",
				IsActive:  true,
				Label:     "Fabulous Key",
				Name:      "Fabulous Key",
				ProjectID: 2,
				Public:    "cec9dfceb0b74c1c9a5e3c135585f364",
				RateLimit: sentry.ProjectKeyRateLimit{
					Window: 0,
					Count:  0,
				},
				Secret: "4f6a592349e249c5906918393766718d",
			}))
		})

		Context("when project key does not exist", func() {
			handler.HandleFunc("/api/0/projects/organization/project/keys/invalid/",
				testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(newAPIError(sentry.APIError{"detail": "The requested resource does not exist"}))
				}),
			)

			BeforeEach(func() {
				keyID = "invalid"
			})

			It("returns a 404 Not Found error", func() {
				Expect(err).To(MatchError(sentry.APIError{"detail": "The requested resource does not exist"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusNotFound))
			})
		})
	})

	Describe("CreateKey", func() {
		var (
			params *sentry.CreateProjectKeyParams

			key  *sentry.ProjectKey
			resp *sentry.Response
			err  error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/project_keys/create.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/projects/organization/project/keys/",
			testHandler(http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write(fixture)
			}),
		)

		BeforeEach(func() {
			params = &sentry.CreateProjectKeyParams{
				Name: "test",
			}
		})

		JustBeforeEach(func() {
			key, resp, err = client.Projects.CreateKey("organization", "project", params)
		})

		It("returns a 201 Created response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusCreated))

			Expect(key).To(Equal(&sentry.ProjectKey{
				BrowserSDK: sentry.ProjectKeyBrowserSDK{
					Choices: [][]string{{"latest", "latest"}, {"4.x", "4.x"}},
				},
				BrowserSDKVersion: "4.x",
				DateCreated:       parseTime("2018-11-06T21:20:07.941Z"),
				DSN: sentry.ProjectKeyDSN{
					CDN:      "https://sentry.io/js-sdk-loader/cec9dfceb0b74c1c9a5e3c135585f364.min.js",
					CSP:      "https://sentry.io/api/2/csp-report/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
					Minidump: "https://sentry.io/api/2/minidump/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
					Public:   "https://cec9dfceb0b74c1c9a5e3c135585f364@sentry.io/2",
					Secret:   "https://cec9dfceb0b74c1c9a5e3c135585f364:4f6a592349e249c5906918393766718d@sentry.io/2",
					Security: "https://sentry.io/api/2/security/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
				},
				ID:        "cec9dfceb0b74c1c9a5e3c135585f364",
				IsActive:  true,
				Label:     "Fabulous Key",
				Name:      "Fabulous Key",
				ProjectID: 2,
				Public:    "cec9dfceb0b74c1c9a5e3c135585f364",
				RateLimit: sentry.ProjectKeyRateLimit{
					Window: 0,
					Count:  0,
				},
				Secret: "4f6a592349e249c5906918393766718d",
			}))
		})
	})

	Describe("UpdateKey", func() {
		var (
			params *sentry.UpdateProjectKeyParams

			key  *sentry.ProjectKey
			resp *sentry.Response
			err  error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/project_keys/update.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/projects/organization/project/keys/test/",
			testHandler(http.MethodPut, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(fixture)
			}),
		)

		BeforeEach(func() {
			params = &sentry.UpdateProjectKeyParams{
				Name: "test",
			}
		})

		JustBeforeEach(func() {
			key, resp, err = client.Projects.UpdateKey("organization", "project", "test", params)
		})

		It("returns a 200 OK response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusOK))

			Expect(key).To(Equal(&sentry.ProjectKey{
				BrowserSDK: sentry.ProjectKeyBrowserSDK{
					Choices: [][]string{{"latest", "latest"}, {"4.x", "4.x"}},
				},
				BrowserSDKVersion: "4.x",
				DateCreated:       parseTime("2018-11-06T21:20:07.941Z"),
				DSN: sentry.ProjectKeyDSN{
					CDN:      "https://sentry.io/js-sdk-loader/cec9dfceb0b74c1c9a5e3c135585f364.min.js",
					CSP:      "https://sentry.io/api/2/csp-report/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
					Minidump: "https://sentry.io/api/2/minidump/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
					Public:   "https://cec9dfceb0b74c1c9a5e3c135585f364@sentry.io/2",
					Secret:   "https://cec9dfceb0b74c1c9a5e3c135585f364:4f6a592349e249c5906918393766718d@sentry.io/2",
					Security: "https://sentry.io/api/2/security/?sentry_key=cec9dfceb0b74c1c9a5e3c135585f364",
				},
				ID:        "cec9dfceb0b74c1c9a5e3c135585f364",
				IsActive:  true,
				Label:     "Fabulous Key",
				Name:      "Fabulous Key",
				ProjectID: 2,
				Public:    "cec9dfceb0b74c1c9a5e3c135585f364",
				RateLimit: sentry.ProjectKeyRateLimit{
					Window: 0,
					Count:  0,
				},
				Secret: "4f6a592349e249c5906918393766718d",
			}))
		})
	})

	Describe("DeleteKey", func() {
		var (
			keyID string

			resp *sentry.Response
			err  error
		)

		handler, client := setup()

		handler.HandleFunc("/api/0/projects/organization/project/keys/valid/",
			testHandler(http.MethodDelete, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
		)

		BeforeEach(func() {
			keyID = "valid"
		})

		JustBeforeEach(func() {
			resp, err = client.Projects.DeleteKey("organization", "project", keyID)
		})

		It("returns a 204 No Content response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusNoContent))
		})

		Context("when project key does not exist", func() {
			handler.HandleFunc("/api/0/projects/organization/project/keys/invalid/",
				testHandler(http.MethodDelete, func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(newAPIError(sentry.APIError{"detail": "The requested resource does not exist"}))
				}),
			)

			BeforeEach(func() {
				keyID = "invalid"
			})

			It("returns a 404 Not Found error", func() {
				Expect(err).To(MatchError(sentry.APIError{"detail": "The requested resource does not exist"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusNotFound))
			})
		})
	})
})
