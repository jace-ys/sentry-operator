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
	TeamFinalizerName = "finalizers.sentry.kubernetes.jaceys.me/team"
)

// TeamReconciler reconciles a Team object
type TeamReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Sentry *Sentry
}

func (r *TeamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sentryv1alpha1.Team{}).
		WithEventFilter(&predicate.GenerationChangedPredicate{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=sentry.kubernetes.jaceys.me,resources=teams,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sentry.kubernetes.jaceys.me,resources=teams/status,verbs=get;update;patch

func (r *TeamReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("team", req.NamespacedName)

	var team sentryv1alpha1.Team
	if err := r.Get(ctx, req.NamespacedName, &team); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "failed to fetch Team")
		return ctrl.Result{}, err
	}

	hasFinalizer := containsFinalizer(team.GetFinalizers(), TeamFinalizerName)

	// Create our Sentry resource if we have not been synced before
	if team.Status.LastSynced.IsZero() {
		if err := r.handleCreate(ctx, &team, hasFinalizer); err != nil {
			log.Error(err, "failed to create Team")
			return ctrl.Result{}, r.handleError(ctx, &team, err)
		}

		log.Info("successfully created Team")
		return ctrl.Result{}, nil
	}

	// Get the existing state of our Sentry resource as it might have drifted. Ignore ErrOutOfSync errors for now as we
	// will handle this accordingly below.
	existing, err := r.getExistingState(team)
	if err != nil && !errors.Is(err, ErrOutOfSync) {
		log.Error(err, "failed to fetch Sentry team state")
		return ctrl.Result{}, r.handleError(ctx, &team, err)
	}

	// Attempt to delete our Sentry resource and remove our finalizer if we receive a delete request
	if !team.ObjectMeta.DeletionTimestamp.IsZero() {
		if hasFinalizer {
			if err := r.handleDelete(ctx, &team, existing); err != nil {
				log.Error(err, "failed to delete Team")
				return ctrl.Result{}, r.handleError(ctx, &team, err)
			}
		}

		log.Info("successfully deleted Team")
		return ctrl.Result{}, nil
	}

	// Our Sentry resource might have been deleted externally of the controller, so attempt to recreate it
	if errors.Is(err, ErrOutOfSync) {
		if err := r.handleCreate(ctx, &team, hasFinalizer); err != nil {
			log.Error(err, "failed to recreate Team")
			return ctrl.Result{}, r.handleError(ctx, &team, err)
		}

		log.Info("successfully recreated Team")
		return ctrl.Result{}, nil
	}

	// Reconcile any differences between our spec and the existing state of our Sentry resource
	if err := r.handleUpdate(ctx, &team, existing); err != nil {
		log.Error(err, "failed to update Team")
		return ctrl.Result{}, r.handleError(ctx, &team, err)
	}

	log.Info("successfully updated Team")

	return ctrl.Result{}, nil
}

// getExistingState retrieves the true state of the resource that exists in Sentry using its constant resource ID, and
// returns an ErrOutOfSync error if the resource cannot be found.
func (r *TeamReconciler) getExistingState(team sentryv1alpha1.Team) (*sentry.Team, error) {
	opts := &sentry.ListOptions{}
	var sTeams []sentry.Team
	for {
		teams, resp, err := r.Sentry.Client.Teams.List(r.Sentry.Organization, opts)
		if err != nil {
			switch {
			case resp.StatusCode >= 500:
				return nil, retryableError{err}
			default:
				// Don't retry on 4XX errors as these indicate that there might be an issue with our organization
				return nil, err
			}
		}

		sTeams = append(sTeams, teams...)
		if !resp.NextPage.Results {
			break
		}
		opts.Cursor = resp.NextPage.Cursor
	}

	for idx, sTeam := range sTeams {
		if sTeam.ID == team.Status.ID {
			return &sTeams[idx], nil
		}
	}

	return nil, ErrOutOfSync
}

func (r *TeamReconciler) handleCreate(ctx context.Context, team *sentryv1alpha1.Team, hasFinalizer bool) error {
	sTeam, resp, err := r.Sentry.Client.Teams.Create(r.Sentry.Organization, &sentry.CreateTeamParams{
		Name: team.Spec.Name,
		Slug: team.Spec.Slug,
	})
	if err != nil {
		switch {
		case resp.StatusCode >= 500:
			return retryableError{err}
		case resp.StatusCode == http.StatusNotFound:
			// Retry on 404 errors as the error might get resolved once dependencies are satisfied
			return retryableError{err}
		default:
			// Don't retry on other 4XX errors as these indicate that we might have an issue with our spec
			return err
		}
	}

	team.Status.Condition = sentryv1alpha1.TeamConditionCreated
	team.Status.Message = ""
	team.Status.ID = sTeam.ID
	team.Status.LastSynced = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, team); err != nil {
		return retryableError{err}
	}

	if !hasFinalizer {
		team.SetFinalizers(append(team.GetFinalizers(), TeamFinalizerName))
		if err := r.Update(ctx, team); err != nil {
			return retryableError{err}
		}
	}

	return nil
}

func (r *TeamReconciler) handleDelete(ctx context.Context, team *sentryv1alpha1.Team, existing *sentry.Team) error {
	// Our resource might no longer exist so check that it's not nil to avoid panicking below
	if existing != nil {
		resp, err := r.Sentry.Client.Teams.Delete(r.Sentry.Organization, existing.Slug)
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

	team.SetFinalizers(removeFinalizer(team.GetFinalizers(), TeamFinalizerName))
	if err := r.Update(ctx, team); err != nil {
		return retryableError{err}
	}

	return nil
}

func (r *TeamReconciler) handleUpdate(ctx context.Context, team *sentryv1alpha1.Team, existing *sentry.Team) error {
	sTeam, resp, err := r.Sentry.Client.Teams.Update(r.Sentry.Organization, existing.Slug, &sentry.UpdateTeamParams{
		Name: team.Spec.Name,
		Slug: team.Spec.Slug,
	})
	if err != nil {
		switch {
		case resp.StatusCode >= 500:
			return retryableError{err}
		case resp.StatusCode == http.StatusNotFound:
			// Retry on 404 errors as the error might get resolved once dependencies are satisfied
			return retryableError{err}
		default:
			// Don't retry on 4XX errors as these indicate that we might have an issue with our spec
			return err
		}
	}

	team.Status.Condition = sentryv1alpha1.TeamConditionCreated
	team.Status.Message = ""
	team.Status.ID = sTeam.ID
	team.Status.LastSynced = &metav1.Time{Time: time.Now()}
	if err := r.Status().Update(ctx, team); err != nil {
		return retryableError{err}
	}

	return nil
}

// handleError is a helper function for annotating our Custom Resource status with the error condition and message. It
// also checks if the error is retryable, ignoring non-retryable ones so we don't requeue our reconcile key.
func (r *TeamReconciler) handleError(ctx context.Context, team *sentryv1alpha1.Team, err error) error {
	team.Status.Condition = sentryv1alpha1.TeamConditionError
	team.Status.Message = err.Error()
	if err := r.Status().Update(ctx, team); err != nil {
		return err
	}

	var re retryableError
	if errors.As(err, &re) {
		return re.err
	}

	return nil
}
