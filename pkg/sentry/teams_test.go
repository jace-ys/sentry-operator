package sentry_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jace-ys/sentry-operator/pkg/sentry"
)

var _ = Describe("TeamsService", func() {
	Describe("List", func() {
		var (
			teams []sentry.Team
			resp  *sentry.Response
			err   error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/teams/list.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/organizations/organization/teams/",
			testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Link", newPaginationLinks())
				w.Write(fixture)
			}),
		)

		JustBeforeEach(func() {
			teams, resp, err = client.Teams.List("organization")
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

			Expect(teams).To(Equal([]sentry.Team{
				{
					Avatar: sentry.Avatar{
						AvatarType: "letter_avatar",
					},
					DateCreated: parseTime("2018-11-06T21:20:08.115Z"),
					HasAccess:   true,
					ID:          "3",
					IsMember:    true,
					IsPending:   false,
					MemberCount: 1,
					Projects:    []sentry.Project{},
					Name:        "Ancient Gabelers",
					Slug:        "ancient-gabelers",
				},
			}))
		})
	})

	Describe("Get", func() {
		var (
			teamSlug string

			team *sentry.Team
			resp *sentry.Response
			err  error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/teams/get.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/teams/organization/valid/",
			testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
				w.Write(fixture)
			}),
		)

		BeforeEach(func() {
			teamSlug = "valid"
		})

		JustBeforeEach(func() {
			team, resp, err = client.Teams.Get("organization", teamSlug)
		})

		It("returns a 200 OK response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusOK))

			Expect(team).To(Equal(&sentry.Team{
				Avatar: sentry.Avatar{
					AvatarType: "letter_avatar",
				},
				DateCreated: parseTime("2018-11-06T21:19:55.114Z"),
				HasAccess:   true,
				ID:          "2",
				IsMember:    true,
				IsPending:   false,
				MemberCount: 1,
				Name:        "Powerful Abolitionist",
				Slug:        "powerful-abolitionist",
			}))
		})

		Context("when team does not exist", func() {
			handler.HandleFunc("/api/0/teams/organization/invalid/",
				testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(newAPIError(sentry.APIError{"detail": "The requested resource does not exist"}))
				}),
			)

			BeforeEach(func() {
				teamSlug = "invalid"
			})

			It("returns a 404 Not Found error", func() {
				Expect(err).To(MatchError(sentry.APIError{"detail": "The requested resource does not exist"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusNotFound))
			})
		})
	})

	Describe("Create", func() {
		var (
			exists bool
			params *sentry.CreateTeamParams

			team *sentry.Team
			resp *sentry.Response
			err  error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/teams/create.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/organizations/organization/teams/",
			testHandler(http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
				if exists {
					w.WriteHeader(http.StatusConflict)
					w.Write(newAPIError(sentry.APIError{"detail": "The requested resource already exists"}))
					return
				}

				var rParams sentry.CreateTeamParams
				err := json.NewDecoder(r.Body).Decode(&rParams)
				Expect(err).ToNot(HaveOccurred())

				apiErr := make(sentry.APIError)
				if rParams.Name == "" {
					apiErr["name"] = "This field is required"
				}

				if len(apiErr) > 0 {
					w.WriteHeader(http.StatusBadRequest)
					w.Write(newAPIError(apiErr))
					return
				}

				w.WriteHeader(http.StatusCreated)
				w.Write(fixture)
			}),
		)

		BeforeEach(func() {
			exists = false
			params = &sentry.CreateTeamParams{
				Name: "test",
				Slug: "test",
			}
		})

		JustBeforeEach(func() {
			team, resp, err = client.Teams.Create("organization", params)
		})

		It("returns a 201 Created response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusCreated))

			Expect(team).To(Equal(&sentry.Team{
				Avatar: sentry.Avatar{
					AvatarType: "letter_avatar",
				},
				DateCreated: parseTime("2018-11-06T21:20:08.115Z"),
				HasAccess:   true,
				ID:          "3",
				IsMember:    true,
				IsPending:   false,
				MemberCount: 1,
				Name:        "Ancient Gabelers",
				Slug:        "ancient-gabelers",
			}))
		})

		Context("when params are invalid", func() {
			BeforeEach(func() {
				params = &sentry.CreateTeamParams{
					Slug: "test",
				}
			})

			It("returns a 400 Bad Request error", func() {
				Expect(err).To(MatchError(sentry.APIError{"name": "This field is required"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusBadRequest))
			})
		})

		Context("when team already exists", func() {
			BeforeEach(func() {
				exists = true
			})

			It("returns a 409 Conflict error", func() {
				Expect(err).To(MatchError(sentry.APIError{"detail": "The requested resource already exists"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusConflict))
			})
		})
	})

	Describe("Update", func() {
		var (
			exists bool
			params *sentry.UpdateTeamParams

			team *sentry.Team
			resp *sentry.Response
			err  error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/teams/update.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/teams/organization/test/",
			testHandler(http.MethodPut, func(w http.ResponseWriter, r *http.Request) {
				if exists {
					w.WriteHeader(http.StatusBadRequest)
					w.Write(newAPIError(sentry.APIError{"slug": "Another team is already using that slug"}))
					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write(fixture)
			}),
		)

		BeforeEach(func() {
			exists = false
			params = &sentry.UpdateTeamParams{
				Name: "test",
				Slug: "test",
			}
		})

		JustBeforeEach(func() {
			team, resp, err = client.Teams.Update("organization", "test", params)
		})

		It("returns a 200 OK response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusOK))

			Expect(team).To(Equal(&sentry.Team{
				Avatar: sentry.Avatar{
					AvatarType: "letter_avatar",
				},
				DateCreated: parseTime("2018-11-06T21:20:22.924Z"),
				HasAccess:   true,
				ID:          "4",
				IsMember:    false,
				IsPending:   false,
				Name:        "The Inflated Philosophers",
				Slug:        "the-obese-philosophers",
			}))
		})

		Context("team already exists", func() {
			BeforeEach(func() {
				exists = true
			})

			It("returns a 400 Bad Request response", func() {
				Expect(err).To(MatchError(sentry.APIError{"slug": "Another team is already using that slug"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusBadRequest))
			})
		})
	})

	Describe("Delete", func() {
		var (
			teamSlug string

			resp *sentry.Response
			err  error
		)

		handler, client := setup()

		handler.HandleFunc("/api/0/teams/organization/valid/",
			testHandler(http.MethodDelete, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
		)

		BeforeEach(func() {
			teamSlug = "valid"
		})

		JustBeforeEach(func() {
			resp, err = client.Teams.Delete("organization", teamSlug)
		})

		It("returns a 204 No Content response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusNoContent))
		})

		Context("when team does not exist", func() {
			handler.HandleFunc("/api/0/teams/organization/invalid/",
				testHandler(http.MethodDelete, func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(newAPIError(sentry.APIError{"detail": "The requested resource does not exist"}))
				}),
			)

			BeforeEach(func() {
				teamSlug = "invalid"
			})

			It("returns a 404 Not Found error", func() {
				Expect(err).To(MatchError(sentry.APIError{"detail": "The requested resource does not exist"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusNotFound))
			})
		})
	})

	Describe("ListProjects", func() {
		var (
			projects []sentry.Project
			resp     *sentry.Response
			err      error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/teams/list-projects.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/teams/organization/team/projects/",
			testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Link", newPaginationLinks())
				w.Write(fixture)
			}),
		)

		JustBeforeEach(func() {
			projects, resp, err = client.Teams.ListProjects("organization", "team")
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
					DateCreated:  parseTime("2018-11-06T21:19:58.536Z"),
					HasAccess:    true,
					ID:           "3",
					IsBookmarked: false,
					IsMember:     true,
					Name:         "Prime Mover",
					Slug:         "prime-mover",
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
				},
			}))
		})
	})

	Describe("CreateProject", func() {
		var (
			exists bool
			params *sentry.CreateProjectParams

			project *sentry.Project
			resp    *sentry.Response
			err     error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/teams/create-project.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/teams/organization/team/projects/",
			testHandler(http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
				if exists {
					w.WriteHeader(http.StatusConflict)
					w.Write(newAPIError(sentry.APIError{"detail": "The requested resource already exists"}))
					return
				}

				var rParams sentry.CreateProjectParams
				err := json.NewDecoder(r.Body).Decode(&rParams)
				Expect(err).ToNot(HaveOccurred())

				apiErr := make(sentry.APIError)
				if rParams.Name == "" {
					apiErr["name"] = "This field is required"
				}

				if len(apiErr) > 0 {
					w.WriteHeader(http.StatusBadRequest)
					w.Write(newAPIError(apiErr))
					return
				}

				w.WriteHeader(http.StatusCreated)
				w.Write(fixture)
			}),
		)

		BeforeEach(func() {
			exists = false
			params = &sentry.CreateProjectParams{
				Name: "test",
				Slug: "test",
			}
		})

		JustBeforeEach(func() {
			project, resp, err = client.Teams.CreateProject("organization", "team", params)
		})

		It("returns a 201 Created response", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.Response).To(HaveHTTPStatus(http.StatusCreated))

			Expect(project).To(Equal(&sentry.Project{
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
			}))
		})

		Context("when params are invalid", func() {
			BeforeEach(func() {
				params = &sentry.CreateProjectParams{
					Slug: "test",
				}
			})

			It("returns a 400 Bad Request error", func() {
				Expect(err).To(MatchError(sentry.APIError{"name": "This field is required"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusBadRequest))
			})
		})

		Context("when project already exists", func() {
			BeforeEach(func() {
				exists = true
			})

			It("returns a 409 Conflict error", func() {
				Expect(err).To(MatchError(sentry.APIError{"detail": "The requested resource already exists"}))
				Expect(resp.Response).To(HaveHTTPStatus(http.StatusConflict))
			})
		})
	})
})
