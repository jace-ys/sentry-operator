package controllers_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	sentryv1alpha1 "github.com/jace-ys/sentry-operator/api/v1alpha1"
	"github.com/jace-ys/sentry-operator/controllers"
	"github.com/jace-ys/sentry-operator/pkg/sentry"
)

var _ = Describe("ProjectKeyReconciler", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250

		projectkeyName      = "test-projectkey"
		projectkeyNamespace = "test-projectkey-namespace"
	)

	var (
		lookupKey       types.NamespacedName
		secretLookupKey types.NamespacedName

		projectkey *sentryv1alpha1.ProjectKey
		secret     *corev1.Secret
	)

	ctx := context.Background()

	request := &sentryv1alpha1.ProjectKey{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sentry.kubernetes.jaceys.me/v1alpha1",
			Kind:       "ProjectKey",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      projectkeyName,
			Namespace: projectkeyNamespace,
			Labels: map[string]string{
				"label": "test-label",
			},
			Annotations: map[string]string{
				"annotation": "test-annotation",
			},
		},
		Spec: sentryv1alpha1.ProjectKeySpec{
			Project: "test-project",
			Name:    "test-projectkey",
		},
	}

	BeforeEach(func() {
		lookupKey = types.NamespacedName{Name: projectkeyName, Namespace: projectkeyNamespace}
		secretLookupKey = types.NamespacedName{Name: fmt.Sprintf("sentry-projectkey-%s", projectkeyName), Namespace: projectkeyNamespace}

		projectkey = new(sentryv1alpha1.ProjectKey)
		secret = new(corev1.Secret)
	})

	Context("when creating a ProjectKey", func() {
		var (
			created *sentry.ProjectKey
		)

		BeforeEach(func() {
			created = testSentryProjectKey("12345", 0, request.Spec.Name, "test-dsn")
			fakeSentryProjects.CreateKeyReturns(created, newSentryResponse(http.StatusOK), nil)

			project := testSentryProject("0", "test-team", request.Spec.Project)
			fakeSentryOrganizations.ListProjectsReturns([]sentry.Project{*project}, newSentryResponse(http.StatusOK), nil)

			fakeSentryProjects.ListKeysReturns([]sentry.ProjectKey{*created}, newSentryResponse(http.StatusOK), nil)
			fakeSentryProjects.UpdateKeyReturns(created, newSentryResponse(http.StatusOK), nil)
		})

		It("the ProjectKey gets created successfully", func() {
			Expect(k8sClient.Create(ctx, request)).To(Succeed())

			By("with the expected status")
			Eventually(func() (*sentryv1alpha1.ProjectKeyStatus, error) {
				err := k8sClient.Get(ctx, lookupKey, projectkey)
				if err != nil {
					return nil, err
				}
				return &projectkey.Status, nil
			}, timeout, interval).Should(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Condition": Equal(sentryv1alpha1.ProjectKeyConditionCreated),
					"Message":   BeEmpty(),
					"ID":        Equal("12345"),
					"ProjectID": Equal("0"),
				})),
			)

			By("with the desired spec")
			Expect(projectkey.Spec).To(Equal(sentryv1alpha1.ProjectKeySpec{
				Project: "test-project",
				Name:    "test-projectkey",
			}))

			By("with the expected finalizer")
			Expect(projectkey.Finalizers).To(ContainElement(controllers.ProjectKeyFinalizerName))

			By("invoked the Sentry client's .Projects.CreateKey method")
			organization, project, params := fakeSentryProjects.CreateKeyArgsForCall(fakeSentryProjects.CreateKeyCallCount() - 1)
			Expect(organization).To(Equal("organization"))
			Expect(project).To(Equal(request.Spec.Project))
			Expect(params).To(Equal(&sentry.CreateProjectKeyParams{
				Name: request.Spec.Name,
			}))
		})

		It("the Secret gets created successfully", func() {
			Eventually(func() (map[string][]byte, error) {
				err := k8sClient.Get(ctx, secretLookupKey, secret)
				if err != nil {
					return nil, err
				}
				return secret.Data, nil
			}, timeout, interval).Should(HaveKeyWithValue("SENTRY_DSN", []byte(created.DSN.Public)))

			By("with the expected metadata")
			Expect(secret.Name).To(Equal(fmt.Sprintf("sentry-projectkey-%s", projectkeyName)))
			Expect(secret.Namespace).To(Equal(projectkeyNamespace))

			By("with the desired labels and annotations")
			Expect(secret.Labels).To(HaveKeyWithValue("label", "test-label"))
			Expect(secret.Annotations).To(HaveKeyWithValue("annotation", "test-annotation"))

			By("with the expected owner reference")
			Expect(secret.ObjectMeta.OwnerReferences).To(
				ContainElement(
					MatchFields(IgnoreExtras, Fields{
						"APIVersion": Equal("sentry.kubernetes.jaceys.me/v1alpha1"),
						"Kind":       Equal("ProjectKey"),
						"Name":       Equal(request.GetName()),
						"UID":        Equal(request.GetUID()),
					}),
				),
			)
		})
	})

	Context("when updating a ProjectKey", func() {
		var (
			project  *sentry.Project
			existing *sentry.ProjectKey
			updated  *sentry.ProjectKey
		)

		BeforeEach(func() {
			Expect(k8sClient.Get(ctx, lookupKey, projectkey)).To(Succeed())

			project = testSentryProject("0", "test-team", projectkey.Spec.Project)
			fakeSentryOrganizations.ListProjectsReturns([]sentry.Project{*project}, newSentryResponse(http.StatusOK), nil)

			existing = testSentryProjectKey("12345", 0, projectkey.Spec.Name, "test-dsn")
			fakeSentryProjects.ListKeysReturns([]sentry.ProjectKey{*existing}, newSentryResponse(http.StatusOK), nil)

			projectkey.Spec.Name = "test-projectkey-update"

			updated = testSentryProjectKey("12345", 0, projectkey.Spec.Name, "test-dsn-update")
			fakeSentryProjects.UpdateKeyReturns(updated, newSentryResponse(http.StatusOK), nil)
		})

		Context("the Sentry client returns an error", func() {
			BeforeEach(func() {
				projectkey.Spec.Name = "test-projectkey-error"
				fakeSentryProjects.UpdateKeyReturns(nil, newSentryResponse(http.StatusBadRequest), errors.New("an error occurred"))
			})

			It("the ProjectKey gets updated unsuccessfully", func() {
				Expect(k8sClient.Update(ctx, projectkey)).To(Succeed())

				By("with the expected status")
				Eventually(func() (*sentryv1alpha1.ProjectKeyStatus, error) {
					err := k8sClient.Get(ctx, lookupKey, projectkey)
					if err != nil {
						return nil, err
					}
					return &projectkey.Status, nil
				}, timeout, interval).Should(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Condition": Equal(sentryv1alpha1.ProjectKeyConditionError),
						"Message":   Equal("an error occurred"),
						"ID":        Equal("12345"),
						"ProjectID": Equal("0"),
					})),
				)

				By("with the desired spec")
				Expect(projectkey.Spec).To(Equal(sentryv1alpha1.ProjectKeySpec{
					Project: "test-project",
					Name:    "test-projectkey-error",
				}))

				By("with the expected finalizer")
				Expect(projectkey.Finalizers).To(ContainElement(controllers.ProjectKeyFinalizerName))

				By("invoked the Sentry client's .Organizations.ListProjects method")
				organizationSlug, opts := fakeSentryOrganizations.ListProjectsArgsForCall(fakeSentryOrganizations.ListProjectsCallCount() - 1)
				Expect(organizationSlug).To(Equal("organization"))
				Expect(opts.Cursor).To(BeEmpty())

				By("invoked the Sentry client's .Projects.ListKeys method")
				organizationSlug, projectSlug, opts := fakeSentryProjects.ListKeysArgsForCall(fakeSentryProjects.ListKeysCallCount() - 1)
				Expect(organizationSlug).To(Equal("organization"))
				Expect(projectSlug).To(Equal(project.Slug))
				Expect(opts.Cursor).To(BeEmpty())

				By("invoked the Sentry client's .Projects.UpdateKey method")
				organizationSlug, projectSlug, keyID, params := fakeSentryProjects.UpdateKeyArgsForCall(fakeSentryProjects.UpdateKeyCallCount() - 1)
				Expect(organizationSlug).To(Equal("organization"))
				Expect(projectSlug).To(Equal(project.Slug))
				Expect(keyID).To(Equal(existing.ID))
				Expect(params).To(Equal(&sentry.UpdateProjectKeyParams{
					Name: projectkey.Spec.Name,
				}))
			})
		})

		Context("the associated Sentry project has changed", func() {
			BeforeEach(func() {
				project = testSentryProject("0", "test-team", "new-project")
				fakeSentryOrganizations.ListProjectsReturns([]sentry.Project{*project}, newSentryResponse(http.StatusOK), nil)

				existing = testSentryProjectKey("12345", 0, projectkey.Spec.Name, "test-dsn")
				fakeSentryProjects.ListKeysReturns([]sentry.ProjectKey{*existing}, newSentryResponse(http.StatusOK), nil)
			})

			It("the ProjectKey gets updated unsuccessfully", func() {
				Expect(k8sClient.Update(ctx, projectkey)).To(Succeed())

				By("with the expected status")
				Eventually(func() (*sentryv1alpha1.ProjectKeyStatus, error) {
					err := k8sClient.Get(ctx, lookupKey, projectkey)
					if err != nil {
						return nil, err
					}
					return &projectkey.Status, nil
				}, timeout, interval).Should(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Condition": Equal(sentryv1alpha1.ProjectKeyConditionError),
						"Message":   ContainSubstring("out of sync"),
						"ID":        Equal("12345"),
						"ProjectID": Equal("0"),
					})),
				)

				By("with the desired spec")
				Expect(projectkey.Spec).To(Equal(sentryv1alpha1.ProjectKeySpec{
					Project: "test-project",
					Name:    "test-projectkey-update",
				}))

				By("with the expected finalizer")
				Expect(projectkey.Finalizers).To(ContainElement(controllers.ProjectKeyFinalizerName))

				By("invoked the Sentry client's .Organizations.ListProjects method")
				organizationSlug, opts := fakeSentryOrganizations.ListProjectsArgsForCall(fakeSentryOrganizations.ListProjectsCallCount() - 1)
				Expect(organizationSlug).To(Equal("organization"))
				Expect(opts.Cursor).To(BeEmpty())

				By("invoked the Sentry client's .Projects.ListKeys method")
				organizationSlug, projectSlug, opts := fakeSentryProjects.ListKeysArgsForCall(fakeSentryProjects.ListKeysCallCount() - 1)
				Expect(organizationSlug).To(Equal("organization"))
				Expect(projectSlug).To(Equal(project.Slug))
				Expect(opts.Cursor).To(BeEmpty())
			})
		})

		It("the ProjectKey gets updated successfully", func() {
			Expect(k8sClient.Update(ctx, projectkey)).To(Succeed())

			By("with the expected status")
			Eventually(func() (*sentryv1alpha1.ProjectKeyStatus, error) {
				err := k8sClient.Get(ctx, lookupKey, projectkey)
				if err != nil {
					return nil, err
				}
				return &projectkey.Status, nil
			}, timeout, interval).Should(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Condition": Equal(sentryv1alpha1.ProjectKeyConditionCreated),
					"Message":   BeEmpty(),
					"ID":        Equal("12345"),
					"ProjectID": Equal("0"),
				})),
			)

			By("with the desired spec")
			Expect(projectkey.Spec).To(Equal(sentryv1alpha1.ProjectKeySpec{
				Project: "test-project",
				Name:    "test-projectkey-update",
			}))

			By("with the expected finalizer")
			Expect(projectkey.Finalizers).To(ContainElement(controllers.ProjectKeyFinalizerName))

			By("invoked the Sentry client's .Organizations.ListProjects method")
			organizationSlug, opts := fakeSentryOrganizations.ListProjectsArgsForCall(fakeSentryOrganizations.ListProjectsCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(opts.Cursor).To(BeEmpty())

			By("invoked the Sentry client's .Projects.ListKeys method")
			organizationSlug, projectSlug, opts := fakeSentryProjects.ListKeysArgsForCall(fakeSentryProjects.ListKeysCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(projectSlug).To(Equal(project.Slug))
			Expect(opts.Cursor).To(BeEmpty())

			By("invoked the Sentry client's .Projects.UpdateKey method")
			organizationSlug, projectSlug, keyID, params := fakeSentryProjects.UpdateKeyArgsForCall(fakeSentryProjects.UpdateKeyCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(projectSlug).To(Equal(project.Slug))
			Expect(keyID).To(Equal(existing.ID))
			Expect(params).To(Equal(&sentry.UpdateProjectKeyParams{
				Name: projectkey.Spec.Name,
			}))
		})

		It("the Secret gets reconciled successfully", func() {
			By("with the expected secret data")
			Eventually(func() (map[string][]byte, error) {
				err := k8sClient.Get(ctx, secretLookupKey, secret)
				if err != nil {
					return nil, err
				}
				return secret.Data, nil
			}, timeout, interval).Should(HaveKeyWithValue("SENTRY_DSN", []byte(updated.DSN.Public)))

			By("with the expected metadata")
			Expect(secret.ObjectMeta.Name).To(Equal(fmt.Sprintf("sentry-projectkey-%s", projectkeyName)))
			Expect(secret.ObjectMeta.Namespace).To(Equal(projectkeyNamespace))

			By("with the desired labels and annotations")
			Expect(secret.Labels).To(HaveKeyWithValue("label", "test-label"))
			Expect(secret.Annotations).To(HaveKeyWithValue("annotation", "test-annotation"))

			By("with the expected owner reference")
			Expect(secret.ObjectMeta.OwnerReferences).To(
				ContainElement(
					MatchFields(IgnoreExtras, Fields{
						"APIVersion": Equal("sentry.kubernetes.jaceys.me/v1alpha1"),
						"Kind":       Equal("ProjectKey"),
						"Name":       Equal(projectkey.GetName()),
						"UID":        Equal(projectkey.GetUID()),
					}),
				),
			)
		})
	})

	Context("when deleting a ProjectKey", func() {
		var (
			project  *sentry.Project
			existing *sentry.ProjectKey
		)

		BeforeEach(func() {
			Expect(k8sClient.Get(ctx, lookupKey, projectkey)).To(Succeed())

			project = testSentryProject("0", "test-team", projectkey.Spec.Project)
			fakeSentryOrganizations.ListProjectsReturns([]sentry.Project{*project}, newSentryResponse(http.StatusOK), nil)

			existing = testSentryProjectKey("12345", 0, projectkey.Spec.Name, "test-dsn")
			fakeSentryProjects.ListKeysReturns([]sentry.ProjectKey{*existing}, newSentryResponse(http.StatusOK), nil)
			fakeSentryProjects.DeleteReturns(newSentryResponse(http.StatusNoContent), nil)
		})

		It("the ProjectKey gets deleted successfully", func() {
			Expect(k8sClient.Delete(ctx, projectkey)).To(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, lookupKey, projectkey)
			}, timeout, interval).ShouldNot(Succeed())

			By("invoked the Sentry client's .Organizations.ListProjects method")
			organizationSlug, opts := fakeSentryOrganizations.ListProjectsArgsForCall(fakeSentryOrganizations.ListProjectsCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(opts.Cursor).To(BeEmpty())

			By("invoked the Sentry client's .Projects.ListKeys method")
			organizationSlug, projectSlug, opts := fakeSentryProjects.ListKeysArgsForCall(fakeSentryProjects.ListKeysCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(projectSlug).To(Equal(project.Slug))
			Expect(opts.Cursor).To(BeEmpty())

			By("invoked the Sentry client's .Projects.DeleteKey method")
			organizationSlug, projectSlug, keyID := fakeSentryProjects.DeleteKeyArgsForCall(fakeSentryProjects.DeleteKeyCallCount() - 1)
			Expect(organizationSlug).To(Equal("organization"))
			Expect(projectSlug).To(Equal(project.Slug))
			Expect(keyID).To(Equal(existing.ID))
		})
	})
})
