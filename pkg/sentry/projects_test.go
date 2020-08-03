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
		Expect(err).NotTo(HaveOccurred())

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
			Expect(err).NotTo(HaveOccurred())
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
					Slug:         "the-spoiled-yoghurt",
					Status:       "active",
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
		Expect(err).NotTo(HaveOccurred())

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
			Expect(err).NotTo(HaveOccurred())
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
				Slug:         "pump-station",
				Status:       "active",
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
		Expect(err).NotTo(HaveOccurred())

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
			Expect(err).NotTo(HaveOccurred())
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
				Platform:     "javascript",
				Slug:         "plane-proxy",
				Status:       "active",
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
			Expect(err).NotTo(HaveOccurred())
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
})
