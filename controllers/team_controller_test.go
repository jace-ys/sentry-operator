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

var _ = Describe("TestReconciler", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250

		teamName      = "test-team"
		teamNamespace = "test-team-namespace"
	)

	var (
		lookupKey types.NamespacedName
		team      *sentryv1alpha1.Team
	)

	ctx := context.Background()

	request := &sentryv1alpha1.Team{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sentry.kubernetes.jaceys.me/v1alpha1",
			Kind:       "Team",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      teamName,
			Namespace: teamNamespace,
		},
		Spec: sentryv1alpha1.TeamSpec{
			Name: "test-team",
			Slug: "test-team",
		},
	}

	BeforeEach(func() {
		lookupKey = types.NamespacedName{Name: teamName, Namespace: teamNamespace}
		team = new(sentryv1alpha1.Team)
	})

	Context("when creating a Team", func() {
		BeforeEach(func() {
			created := testSentryTeam("12345", request.Spec.Name)
			fakeSentryTeams.CreateReturns(created, newSentryResponse(http.StatusOK), nil)
		})

		It("the Team gets created successfully", func() {
			Expect(k8sClient.Create(ctx, request)).To(Succeed())

			By("with the expected status")
			Eventually(func() (*sentryv1alpha1.TeamStatus, error) {
				err := k8sClient.Get(ctx, lookupKey, team)
				if err != nil {
					return nil, err
				}
				return &team.Status, nil
			}, timeout, interval).Should(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Condition": Equal(sentryv1alpha1.TeamConditionCreated),
					"Message":   BeEmpty(),
					"ID":        Equal("12345"),
				})),
			)

			By("with the desired spec")
			Expect(team.Spec).To(Equal(sentryv1alpha1.TeamSpec{
				Name: "test-team",
				Slug: "test-team",
			}))

			By("with the expected finalizer")
			Expect(team.Finalizers).To(ContainElement(controllers.TeamFinalizerName))

			By("invoked the Sentry client's .Teams.Create method")
			organizationSlug, params := fakeSentryTeams.CreateArgsForCall(fakeSentryTeams.CreateCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(params).To(Equal(&sentry.CreateTeamParams{
				Name: request.Spec.Name,
				Slug: request.Spec.Slug,
			}))
		})
	})

	Context("when updating a Team", func() {
		var (
			existing *sentry.Team
		)

		BeforeEach(func() {
			Expect(k8sClient.Get(ctx, lookupKey, team)).To(Succeed())

			existing = testSentryTeam("12345", team.Spec.Name)
			fakeSentryTeams.ListReturns([]sentry.Team{*existing}, newSentryResponse(http.StatusOK), nil)

			team.Spec.Name = "test-team-update"
			team.Spec.Slug = "test-team-update"

			updated := testSentryTeam("12345", team.Spec.Name)
			fakeSentryTeams.UpdateReturns(updated, newSentryResponse(http.StatusOK), nil)
		})

		Context("the Sentry client returns an error", func() {
			BeforeEach(func() {
				team.Spec.Name = "test-team-error"
				team.Spec.Slug = "test-team-error"
				fakeSentryTeams.UpdateReturns(nil, newSentryResponse(http.StatusBadRequest), errors.New("an error occurred"))
			})

			It("the Team gets updated unsuccessfully", func() {
				Expect(k8sClient.Update(ctx, team)).To(Succeed())

				By("with the expected status")
				Eventually(func() (*sentryv1alpha1.TeamStatus, error) {
					err := k8sClient.Get(ctx, lookupKey, team)
					if err != nil {
						return nil, err
					}
					return &team.Status, nil
				}, timeout, interval).Should(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Condition": Equal(sentryv1alpha1.TeamConditionError),
						"Message":   Equal("an error occurred"),
						"ID":        Equal("12345"),
					})),
				)

				By("with the desired spec")
				Expect(team.Spec).To(Equal(sentryv1alpha1.TeamSpec{
					Name: "test-team-error",
					Slug: "test-team-error",
				}))

				By("with the expected finalizer")
				Expect(team.Finalizers).To(ContainElement(controllers.TeamFinalizerName))

				By("invoked the Sentry client's .Teams.List method")
				organizationSlug, opts := fakeSentryTeams.ListArgsForCall(fakeSentryTeams.ListCallCount() - 1)
				Expect(organizationSlug).To(Equal("organization"))
				Expect(opts.Cursor).To(BeEmpty())

				By("invoked the Sentry client's .Teams.Update method")
				organizationSlug, teamSlug, params := fakeSentryTeams.UpdateArgsForCall(fakeSentryTeams.UpdateCallCount() - 1)
				Expect(organizationSlug).To(Equal("organization"))
				Expect(teamSlug).To(Equal(existing.Slug))
				Expect(params).To(Equal(&sentry.UpdateTeamParams{
					Name: team.Spec.Name,
					Slug: team.Spec.Slug,
				}))
			})
		})

		It("the Team gets updated successfully", func() {
			Expect(k8sClient.Update(ctx, team)).To(Succeed())

			By("with the expected status")
			Eventually(func() (*sentryv1alpha1.TeamStatus, error) {
				err := k8sClient.Get(ctx, lookupKey, team)
				if err != nil {
					return nil, err
				}
				return &team.Status, nil
			}, timeout, interval).Should(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Condition": Equal(sentryv1alpha1.TeamConditionCreated),
					"Message":   BeEmpty(),
					"ID":        Equal("12345"),
				})),
			)

			By("with the desired spec")
			Expect(team.Spec).To(Equal(sentryv1alpha1.TeamSpec{
				Name: "test-team-update",
				Slug: "test-team-update",
			}))

			By("with the expected finalizer")
			Expect(team.Finalizers).To(ContainElement(controllers.TeamFinalizerName))

			By("invoked the Sentry client's .Teams.List method")
			organizationSlug, opts := fakeSentryTeams.ListArgsForCall(fakeSentryTeams.ListCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(opts.Cursor).To(BeEmpty())

			By("invoked the Sentry client's .Teams.Update method")
			organizationSlug, teamSlug, params := fakeSentryTeams.UpdateArgsForCall(fakeSentryTeams.UpdateCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(teamSlug).To(Equal(existing.Slug))
			Expect(params).To(Equal(&sentry.UpdateTeamParams{
				Name: team.Spec.Name,
				Slug: team.Spec.Slug,
			}))
		})
	})

	Context("when deleting a Team", func() {
		var (
			existing *sentry.Team
		)

		BeforeEach(func() {
			Expect(k8sClient.Get(ctx, lookupKey, team)).To(Succeed())

			existing = testSentryTeam("12345", team.Spec.Name)
			fakeSentryTeams.ListReturns([]sentry.Team{*existing}, newSentryResponse(http.StatusOK), nil)
			fakeSentryTeams.DeleteReturns(newSentryResponse(http.StatusNoContent), nil)
		})

		It("the Team gets deleted successfully", func() {
			Expect(k8sClient.Delete(ctx, team)).To(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, lookupKey, team)
			}, timeout, interval).ShouldNot(Succeed())

			By("invoked the Sentry client's .Teams.List method")
			organizationSlug, opts := fakeSentryTeams.ListArgsForCall(fakeSentryTeams.ListCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(opts.Cursor).To(BeEmpty())

			By("invoked the Sentry client's .Teams.Delete method")
			organizationSlug, teamSlug := fakeSentryTeams.DeleteArgsForCall(fakeSentryTeams.DeleteCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(teamSlug).To(Equal(existing.Slug))
		})
	})
})
