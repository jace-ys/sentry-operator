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
	"time"

	"github.com/go-logr/logr"
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
	ProjectFinalizerName = "finalizers.sentry.kubernetes.jaceys.me/project"
)

// ProjectReconciler reconciles a Project object
type ProjectReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Sentry *Sentry
}

func (r *ProjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sentryv1alpha1.Project{}).
		WithEventFilter(&predicate.GenerationChangedPredicate{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=sentry.kubernetes.jaceys.me,resources=projects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sentry.kubernetes.jaceys.me,resources=projects/status,verbs=get;update;patch

func (r *ProjectReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("project", req.NamespacedName)

	var project sentryv1alpha1.Project
	if err := r.Get(ctx, req.NamespacedName, &project); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "failed to fetch Project")
		return ctrl.Result{}, err
	}

	hasFinalizer := containsFinalizer(project.GetFinalizers(), ProjectFinalizerName)

	if project.Status.LastSynced.IsZero() {
		if _, err := r.handleCreate(ctx, &project, hasFinalizer); err != nil {
			log.Error(err, "failed to create Project")
			return ctrl.Result{}, r.handleError(ctx, &project, err)
		}

		log.Info("successfully created Project")
		return ctrl.Result{}, nil
	}

	existing, err := r.getExistingState(project)
	if err != nil {
		log.Error(err, "failed to fetch Sentry project state")

		if !errors.Is(err, ErrOutOfSync) {
			return ctrl.Result{}, r.handleError(ctx, &project, err)
		}

		if existing, err = r.handleCreate(ctx, &project, hasFinalizer); err != nil {
			log.Error(err, "failed to recreate Project")
			return ctrl.Result{}, r.handleError(ctx, &project, err)
		}

		log.Info("successfully recreated Project")
	}

	if !project.ObjectMeta.DeletionTimestamp.IsZero() {
		if hasFinalizer {
			if err := r.handleDelete(ctx, &project, existing); err != nil {
				log.Error(err, "failed to delete Project")
				return ctrl.Result{}, r.handleError(ctx, &project, err)
			}
		}

		log.Info("successfully deleted Project")
		return ctrl.Result{}, nil
	}

	if _, err := r.handleUpdate(ctx, &project, existing); err != nil {
		log.Error(err, "failed to update Project")
		return ctrl.Result{}, r.handleError(ctx, &project, err)
	}

	log.Info("successfully updated Project")
	return ctrl.Result{}, nil
}

// getExistingState retrieves the true state of the project that exists in Sentry, using its constant project
// ID. It returns an ErrOutOfSync error when it can't find a project with the expected ID, allowing us to
// correct drift in cases where the project was deleted externally of the controller.
func (r *ProjectReconciler) getExistingState(project sentryv1alpha1.Project) (*sentry.Project, error) {
	sProjects, resp, err := r.Sentry.Client.Teams.ListProjects(r.Sentry.Organization, project.Spec.Team)
	if err != nil {
		return nil, retryableHTTPError(resp.Response, err)
	}

	for idx, sProject := range sProjects {
		if sProject.ID == project.Status.ID {
			return &sProjects[idx], nil
		}
	}

	return nil, fmt.Errorf("%w: could not find Sentry project with ID %s", ErrOutOfSync, project.Status.ID)
}

func (r *ProjectReconciler) handleCreate(ctx context.Context, project *sentryv1alpha1.Project, hasFinalizer bool) (*sentry.Project, error) {
	sProject, resp, err := r.Sentry.Client.Teams.CreateProject(r.Sentry.Organization, project.Spec.Team, &sentry.CreateProjectParams{
		Name: project.Spec.Name,
		Slug: project.Spec.Slug,
	})
	if err != nil {
		return nil, retryableHTTPError(resp.Response, err)
	}

	project.Status.Condition = sentryv1alpha1.ProjectConditionCreated
	project.Status.ID = sProject.ID
	project.Status.LastSynced = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, project); err != nil {
		return nil, retryableError{err}
	}

	if !hasFinalizer {
		project.SetFinalizers(append(project.GetFinalizers(), ProjectFinalizerName))
		if err := r.Update(ctx, project); err != nil {
			return nil, retryableError{err}
		}
	}

	return sProject, nil
}

func (r *ProjectReconciler) handleDelete(ctx context.Context, project *sentryv1alpha1.Project, existing *sentry.Project) error {
	resp, err := r.Sentry.Client.Projects.Delete(r.Sentry.Organization, existing.Slug)
	if err != nil {
		if resp.StatusCode != http.StatusNotFound {
			return retryableHTTPError(resp.Response, err)
		}
	}

	project.SetFinalizers(removeFinalizer(project.GetFinalizers(), ProjectFinalizerName))
	if err := r.Update(ctx, project); err != nil {
		return retryableError{err}
	}

	return nil
}

func (r *ProjectReconciler) handleUpdate(ctx context.Context, project *sentryv1alpha1.Project, existing *sentry.Project) (*sentry.Project, error) {
	sProject, resp, err := r.Sentry.Client.Projects.Update(r.Sentry.Organization, existing.Slug, &sentry.UpdateProjectParams{
		Name: project.Spec.Name,
		Slug: project.Spec.Slug,
	})
	if err != nil {
		return nil, retryableHTTPError(resp.Response, err)
	}

	project.Status.Condition = sentryv1alpha1.ProjectConditionCreated
	project.Status.Message = ""
	project.Status.ID = sProject.ID
	project.Status.LastSynced = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, project); err != nil {
		return nil, retryableError{err}
	}

	return sProject, nil
}

// handleError is a helper function for annotating our Project status with ProjectConditionError and the error message.
// It also checks if the error is retryable, ignoring non-retryable ones so we don't requeue our reconcile key.
func (r *ProjectReconciler) handleError(ctx context.Context, project *sentryv1alpha1.Project, err error) error {
	project.Status.Condition = sentryv1alpha1.ProjectConditionError
	project.Status.Message = err.Error()
	if err := r.Status().Update(ctx, project); err != nil {
		return err
	}

	var re retryableError
	if errors.As(err, &re) {
		return re.err
	}

	return nil
}
