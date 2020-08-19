/*

MIT License

Copyright (c) 2020 Jace Tan

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	sentryv1alpha1 "github.com/jace-ys/sentry-operator/api/v1alpha1"
	"github.com/jace-ys/sentry-operator/pkg/sentry"
)

const (
	ProjectKeyFinalizerName = "finalizers.sentry.kubernetes.jaceys.me/projectkey"
)

// ProjectKeyReconciler reconciles a ProjectKey object
type ProjectKeyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Sentry *Sentry
}

func (r *ProjectKeyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sentryv1alpha1.ProjectKey{}).
		Owns(&corev1.Secret{}).
		WithEventFilter(&predicate.GenerationChangedPredicate{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=sentry.kubernetes.jaceys.me,resources=projectkeys,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sentry.kubernetes.jaceys.me,resources=projectkeys/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secret,verbs=get;list;watch;create;update;patch;delete

func (r *ProjectKeyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("projectkey", req.NamespacedName)

	var projectkey sentryv1alpha1.ProjectKey
	if err := r.Get(ctx, req.NamespacedName, &projectkey); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "failed to fetch ProjectKey")
		return ctrl.Result{}, err
	}

	hasFinalizer := containsFinalizer(projectkey.GetFinalizers(), ProjectKeyFinalizerName)

	// Create our Sentry resource and secret if we have not been synced before
	if projectkey.Status.LastSynced.IsZero() {
		sProjectKey, err := r.handleCreate(ctx, &projectkey, hasFinalizer)
		if err != nil {
			log.Error(err, "failed to create ProjectKey")
			return ctrl.Result{}, r.handleError(ctx, &projectkey, err)
		}

		if err := r.reconcileSecret(ctx, &projectkey, sProjectKey); err != nil {
			log.Error(err, "failed to create Secret for ProjectKey")
			return ctrl.Result{}, r.handleError(ctx, &projectkey, err)
		}

		log.Info("successfully created ProjectKey")
		return ctrl.Result{}, nil
	}

	// Get the existing state of our Sentry resource as it might have drifted. Ignore ErrOutOfSync errors for now as we
	// will handle this accordingly below.
	existing, projectSlug, err := r.getExistingState(projectkey)
	if err != nil && !errors.Is(err, ErrOutOfSync) {
		log.Error(err, "failed to fetch Sentry project key state")
		return ctrl.Result{}, r.handleError(ctx, &projectkey, err)
	}

	// Attempt to delete our Sentry resource and remove our finalizer if we receive a delete request
	if !projectkey.ObjectMeta.DeletionTimestamp.IsZero() {
		if hasFinalizer {
			if err := r.handleDelete(ctx, &projectkey, existing, projectSlug); err != nil {
				log.Error(err, "failed to delete ProjectKey")
				return ctrl.Result{}, r.handleError(ctx, &projectkey, err)
			}
		}

		log.Info("successfully deleted ProjectKey")
		return ctrl.Result{}, nil
	}

	// Our Sentry resource might have been deleted externally of the controller, so attempt to recreate it
	if errors.Is(err, ErrOutOfSync) {
		sProjectKey, err := r.handleCreate(ctx, &projectkey, hasFinalizer)
		if err != nil {
			log.Error(err, "failed to recreate ProjectKey")
			return ctrl.Result{}, r.handleError(ctx, &projectkey, err)
		}

		log.Info("successfully recreated ProjectKey")

		// Reconcile our secret to ensure that its data matches that found in our Sentry project key
		if err := r.reconcileSecret(ctx, &projectkey, sProjectKey); err != nil {
			log.Error(err, "failed to reconcile Secret for ProjectKey")
			return ctrl.Result{}, r.handleError(ctx, &projectkey, err)
		}

		log.Info("successfully reconciled Secret for ProjectKey")

		return ctrl.Result{}, nil
	}

	// Reconcile any differences between our spec and the existing state of our Sentry resource
	sProjectKey, err := r.handleUpdate(ctx, &projectkey, existing, projectSlug)
	if err != nil {
		log.Error(err, "failed to update ProjectKey")
		return ctrl.Result{}, r.handleError(ctx, &projectkey, err)
	}

	log.Info("successfully updated ProjectKey")

	// Reconcile our secret to ensure that its data matches that found in our Sentry project key
	if err := r.reconcileSecret(ctx, &projectkey, sProjectKey); err != nil {
		log.Error(err, "failed to reconcile Secret for ProjectKey")
		return ctrl.Result{}, r.handleError(ctx, &projectkey, err)
	}

	log.Info("successfully reconciled Secret for ProjectKey")

	return ctrl.Result{}, nil
}

// getExistingState retrieves the true state of the resource that exists in Sentry using its constant resource ID, and
// returns an ErrOutOfSync error if the resource cannot be found. It also returns our associated project's slug, as it's
// not part of the payload returned when listing a Sentry project's keys.
func (r *ProjectKeyReconciler) getExistingState(projectkey sentryv1alpha1.ProjectKey) (*sentry.ProjectKey, string, error) {
	sProjects, resp, err := r.Sentry.Client.Organizations.ListProjects(r.Sentry.Organization)
	if err != nil {
		switch {
		case resp.StatusCode >= 500:
			return nil, "", retryableError{err}
		default:
			// Don't retry on 4XX errors as these indicate that there might be an issue with our organization
			return nil, "", err
		}
	}

	var projectSlug string
	for _, sProject := range sProjects {
		if sProject.ID == projectkey.Status.ProjectID {
			projectSlug = sProject.Slug
			break
		}

		return nil, "", ErrOutOfSync
	}

	sProjectKeys, resp, err := r.Sentry.Client.Projects.ListKeys(r.Sentry.Organization, projectSlug)
	if err != nil {
		switch {
		case resp.StatusCode >= 500:
			return nil, "", retryableError{err}
		case resp.StatusCode == http.StatusNotFound:
			return nil, "", ErrOutOfSync
		default:
			// Don't retry on other 4XX errors as these indicate that there might be an issue with our spec
			return nil, "", err
		}
	}

	for idx, sProjectKey := range sProjectKeys {
		if sProjectKey.ID == projectkey.Status.ID {
			return &sProjectKeys[idx], projectSlug, nil
		}
	}

	return nil, "", ErrOutOfSync
}

func (r *ProjectKeyReconciler) handleCreate(ctx context.Context, projectkey *sentryv1alpha1.ProjectKey, hasFinalizer bool) (*sentry.ProjectKey, error) {
	sProjectKey, resp, err := r.Sentry.Client.Projects.CreateKey(r.Sentry.Organization, projectkey.Spec.Project, &sentry.CreateProjectKeyParams{
		Name: projectkey.Spec.Name,
	})
	if err != nil {
		switch {
		case resp.StatusCode >= 500:
			return nil, retryableError{err}
		case resp.StatusCode == http.StatusNotFound:
			// Retry on 404 errors as the error might get resolved once dependencies are satisfied
			return nil, retryableError{err}
		default:
			// Don't retry on other 4XX errors as these indicate that we might have an issue with our spec
			return nil, err
		}
	}

	projectkey.Status.Condition = sentryv1alpha1.ProjectKeyConditionCreated
	projectkey.Status.Message = ""
	projectkey.Status.ID = sProjectKey.ID
	projectkey.Status.LastSynced = &metav1.Time{Time: time.Now()}
	projectkey.Status.ProjectID = strconv.Itoa(sProjectKey.ProjectID)
	if err := r.Status().Update(ctx, projectkey); err != nil {
		return nil, retryableError{err}
	}

	if !hasFinalizer {
		projectkey.SetFinalizers(append(projectkey.GetFinalizers(), ProjectKeyFinalizerName))
		if err := r.Update(ctx, projectkey); err != nil {
			return nil, retryableError{err}
		}
	}

	return sProjectKey, nil
}

func (r *ProjectKeyReconciler) reconcileSecret(ctx context.Context, projectkey *sentryv1alpha1.ProjectKey, sProjectKey *sentry.ProjectKey) error {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("sentry-projectkey-%s", projectkey.Name),
			Namespace: projectkey.Namespace,
		},
		Data: make(map[string][]byte),
	}

	if err := ctrl.SetControllerReference(projectkey, secret, r.Scheme); err != nil {
		return err
	}

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, secret, func() error {
		secret.Data["SENTRY_DSN"] = []byte(sProjectKey.DSN.Public)
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *ProjectKeyReconciler) handleDelete(ctx context.Context, projectkey *sentryv1alpha1.ProjectKey, existing *sentry.ProjectKey, projectSlug string) error {
	// Our resource might no longer exist so check that it's not nil to avoid panicking below
	if existing != nil {
		resp, err := r.Sentry.Client.Projects.DeleteKey(r.Sentry.Organization, projectSlug, existing.ID)
		if err != nil {
			switch {
			case resp.StatusCode >= 500:
				return retryableError{err}
			case resp.StatusCode == http.StatusNotFound:
				// Ignore 404 errors as our resource might have already been deleted
			default:
				// Don't retry on other 4XX errors as these indicate that we might have an issue with our spec
				return err
			}
		}
	}

	projectkey.SetFinalizers(removeFinalizer(projectkey.GetFinalizers(), ProjectKeyFinalizerName))
	if err := r.Update(ctx, projectkey); err != nil {
		return retryableError{err}
	}

	return nil
}

func (r *ProjectKeyReconciler) handleUpdate(ctx context.Context, projectkey *sentryv1alpha1.ProjectKey, existing *sentry.ProjectKey, projectSlug string) (*sentry.ProjectKey, error) {
	// Error if our spec's project doesn't match reality as updating a project key's project is not a valid operation.
	// This helps highlight configuration drift where a user forgets to update our spec's project after modifying the
	// associated project's slug.
	if projectkey.Spec.Project != projectSlug {
		return nil, retryableError{fmt.Errorf("%w: ProjectKey's project could not be updated", ErrOutOfSync)}
	}

	sProjectKey, resp, err := r.Sentry.Client.Projects.UpdateKey(r.Sentry.Organization, projectSlug, existing.ID, &sentry.UpdateProjectKeyParams{
		Name: projectkey.Spec.Name,
	})
	if err != nil {
		switch {
		case resp.StatusCode >= 500:
			return nil, retryableError{err}
		case resp.StatusCode == http.StatusNotFound:
			// Retry on 404 errors as the error might get resolved once dependencies are satisfied
			return nil, retryableError{err}
		default:
			// Don't retry on 4XX errors as these indicate that we might have an issue with our spec
			return nil, err
		}
	}

	projectkey.Status.Condition = sentryv1alpha1.ProjectKeyConditionCreated
	projectkey.Status.Message = ""
	projectkey.Status.ID = sProjectKey.ID
	projectkey.Status.LastSynced = &metav1.Time{Time: time.Now()}
	projectkey.Status.ProjectID = strconv.Itoa(sProjectKey.ProjectID)
	if err := r.Status().Update(ctx, projectkey); err != nil {
		return nil, retryableError{err}
	}

	return sProjectKey, nil
}

// handleError is a helper function for annotating our Custom Resource status with the error condition and message. It
// also checks if the error is retryable, ignoring non-retryable ones so we don't requeue our reconcile key.
func (r *ProjectKeyReconciler) handleError(ctx context.Context, projectkey *sentryv1alpha1.ProjectKey, err error) error {
	projectkey.Status.Condition = sentryv1alpha1.ProjectKeyConditionError
	projectkey.Status.Message = err.Error()
	if err := r.Status().Update(ctx, projectkey); err != nil {
		return err
	}

	var re retryableError
	if errors.As(err, &re) {
		return re.err
	}

	return nil
}
