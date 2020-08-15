package sentry_test

import (
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jace-ys/sentry-operator/pkg/sentry"
)

var _ = Describe("TeamsService", func() {
	Describe("ListProjects", func() {
		var (
			projects []sentry.Project
			resp     *sentry.Response
			err      error
		)

		handler, client := setup()
		fixture, err := ioutil.ReadFile("fixtures/organizations/list-projects.json")
		Expect(err).ToNot(HaveOccurred())

		handler.HandleFunc("/api/0/organizations/organization/projects/",
			testHandler(http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Link", newPaginationLinks())
				w.Write(fixture)
			}),
		)

		JustBeforeEach(func() {
			projects, resp, err = client.Organizations.ListProjects("organization")
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
})
