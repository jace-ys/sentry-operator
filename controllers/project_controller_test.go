package controllers_test

import (
	"context"
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	sentryv1alpha1 "github.com/jace-ys/sentry-operator/api/v1alpha1"
	"github.com/jace-ys/sentry-operator/controllers"
	"github.com/jace-ys/sentry-operator/pkg/sentry"
)

var _ = Describe("ProjectReconciler", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250

		projectName      = "test-project"
		projectNamespace = "test-project-namespace"
	)

	var (
		lookupKey types.NamespacedName
		project   *sentryv1alpha1.Project
	)

	ctx := context.Background()

	BeforeEach(func() {
		lookupKey = types.NamespacedName{Name: projectName, Namespace: projectNamespace}
		project = new(sentryv1alpha1.Project)
	})

	Context("when creating a Project", func() {
		request := &sentryv1alpha1.Project{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "sentry.kubernetes.jaceys.me/v1alpha1",
				Kind:       "Project",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      projectName,
				Namespace: projectNamespace,
			},
			Spec: sentryv1alpha1.ProjectSpec{
				Team: "test-team",
				Name: "test-project",
				Slug: "test-project",
			},
		}

		BeforeEach(func() {
			created := testSentryProject("12345", request.Spec.Team, request.Spec.Name)
			fakeSentryTeams.CreateProjectReturns(created, newSentryResponse(http.StatusOK), nil)
		})

		It("the Project gets created successfully", func() {
			Expect(k8sClient.Create(ctx, request)).To(Succeed())

			By("with the expected status")
			Eventually(func() (*sentryv1alpha1.ProjectStatus, error) {
				err := k8sClient.Get(ctx, lookupKey, project)
				if err != nil {
					return nil, err
				}
				return &project.Status, nil
			}, timeout, interval).Should(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Condition": Equal(sentryv1alpha1.ProjectConditionCreated),
					"Message":   BeEmpty(),
					"ID":        Equal("12345"),
				})),
			)

			By("with the desired spec")
			Expect(project.Spec).To(Equal(request.Spec))

			By("with the expected finalizer")
			Expect(project.Finalizers).To(ContainElement(controllers.ProjectFinalizerName))

			By("invoked the Sentry client's .Teams.CreateProject method")
			organizationSlug, teamSlug, params := fakeSentryTeams.CreateProjectArgsForCall(fakeSentryTeams.CreateProjectCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(teamSlug).To(Equal(request.Spec.Team))
			Expect(params).To(Equal(&sentry.CreateProjectParams{
				Name: request.Spec.Name,
				Slug: request.Spec.Slug,
			}))
		})
	})

	Context("when updating a Project", func() {
		var (
			existing *sentry.Project
		)

		BeforeEach(func() {
			Expect(k8sClient.Get(ctx, lookupKey, project)).To(Succeed())

			existing = testSentryProject("12345", project.Spec.Team, project.Spec.Name)
			fakeSentryOrganizations.ListProjectsReturns([]sentry.Project{*existing}, newSentryResponse(http.StatusOK), nil)

			project.Spec.Name = "test-project-update"
			project.Spec.Slug = "test-project-update"
			fakeSentryProjects.UpdateReturns(testSentryProject("12345", project.Spec.Team, project.Spec.Name), newSentryResponse(http.StatusOK), nil)
		})

		Context("the Sentry client returns an error", func() {
			BeforeEach(func() {
				project.Spec.Name = "test-project-error"
				project.Spec.Slug = "test-project-error"
				fakeSentryProjects.UpdateReturns(nil, newSentryResponse(http.StatusBadRequest), errors.New("an error occurred"))
			})

			It("the Project gets updated unsuccessfully", func() {
				Expect(k8sClient.Update(ctx, project)).To(Succeed())

				By("with the expected status")
				Eventually(func() (*sentryv1alpha1.ProjectStatus, error) {
					err := k8sClient.Get(ctx, lookupKey, project)
					if err != nil {
						return nil, err
					}
					return &project.Status, nil
				}, timeout, interval).Should(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Condition": Equal(sentryv1alpha1.ProjectConditionError),
						"Message":   Equal("an error occurred"),
						"ID":        Equal("12345"),
					})),
				)

				By("with the desired spec")
				Expect(project.Spec).To(Equal(project.Spec))

				By("invoked the Sentry client's .Organizations.ListProjects method")
				organizationSlug := fakeSentryOrganizations.ListProjectsArgsForCall(fakeSentryOrganizations.ListProjectsCallCount() - 1)
				Expect(organizationSlug).To(Equal("organization"))

				By("invoked the Sentry client's .Projects.Update method")
				organizationSlug, projectSlug, params := fakeSentryProjects.UpdateArgsForCall(fakeSentryProjects.UpdateCallCount() - 1)
				Expect(organizationSlug).To(Equal("organization"))
				Expect(projectSlug).To(Equal(existing.Slug))
				Expect(params).To(Equal(&sentry.UpdateProjectParams{
					Name: project.Spec.Name,
					Slug: project.Spec.Slug,
				}))
			})
		})

		Context("the associated Sentry team has changed", func() {
			BeforeEach(func() {
				existing = testSentryProject("12345", "new-team", project.Spec.Name)
				fakeSentryOrganizations.ListProjectsReturns([]sentry.Project{*existing}, newSentryResponse(http.StatusOK), nil)
			})

			It("the Project gets updated unsuccessfully", func() {
				Expect(k8sClient.Update(ctx, project)).To(Succeed())

				By("with the expected status")
				Eventually(func() (*sentryv1alpha1.ProjectStatus, error) {
					err := k8sClient.Get(ctx, lookupKey, project)
					if err != nil {
						return nil, err
					}
					return &project.Status, nil
				}, timeout, interval).Should(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Condition": Equal(sentryv1alpha1.ProjectConditionError),
						"Message":   ContainSubstring("out of sync"),
						"ID":        Equal("12345"),
					})),
				)

				By("with the desired spec")
				Expect(project.Spec).To(Equal(project.Spec))

				By("invoked the Sentry client's .Organizations.ListProjects method")
				organizationSlug := fakeSentryOrganizations.ListProjectsArgsForCall(fakeSentryOrganizations.ListProjectsCallCount() - 1)
				Expect(organizationSlug).To(Equal("organization"))
			})
		})

		It("the Project gets updated successfully", func() {
			Expect(k8sClient.Update(ctx, project)).To(Succeed())

			By("with the expected status")
			Eventually(func() (*sentryv1alpha1.ProjectStatus, error) {
				err := k8sClient.Get(ctx, lookupKey, project)
				if err != nil {
					return nil, err
				}
				return &project.Status, nil
			}, timeout, interval).Should(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Condition": Equal(sentryv1alpha1.ProjectConditionCreated),
					"Message":   BeEmpty(),
					"ID":        Equal("12345"),
				})),
			)

			By("with the desired spec")
			Expect(project.Spec).To(Equal(project.Spec))

			By("invoked the Sentry client's .Organizations.ListProjects method")
			organizationSlug := fakeSentryOrganizations.ListProjectsArgsForCall(fakeSentryOrganizations.ListProjectsCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))

			By("invoked the Sentry client's .Projects.Update method")
			organizationSlug, projectSlug, params := fakeSentryProjects.UpdateArgsForCall(fakeSentryProjects.UpdateCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(projectSlug).To(Equal(existing.Slug))
			Expect(params).To(Equal(&sentry.UpdateProjectParams{
				Name: project.Spec.Name,
				Slug: project.Spec.Slug,
			}))
		})
	})

	Context("when deleting a Project", func() {
		var (
			existing *sentry.Project
		)

		BeforeEach(func() {
			Expect(k8sClient.Get(ctx, lookupKey, project)).To(Succeed())

			existing = testSentryProject("12345", project.Spec.Team, project.Spec.Name)
			fakeSentryOrganizations.ListProjectsReturns([]sentry.Project{*existing}, newSentryResponse(http.StatusOK), nil)
			fakeSentryProjects.DeleteReturns(newSentryResponse(http.StatusNoContent), nil)
		})

		It("the Project gets deleted succesfully", func() {
			Expect(k8sClient.Delete(ctx, project)).To(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, lookupKey, project)
			}, timeout, interval).ShouldNot(Succeed())

			By("invoked the Sentry client's .Organizations.ListProjects method")
			organizationSlug := fakeSentryOrganizations.ListProjectsArgsForCall(fakeSentryOrganizations.ListProjectsCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))

			By("invoked the Sentry client's .Projects.Delete method")
			organizationSlug, projectSlug := fakeSentryProjects.DeleteArgsForCall(fakeSentryProjects.DeleteCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(projectSlug).To(Equal(existing.Slug))
		})
	})
})
