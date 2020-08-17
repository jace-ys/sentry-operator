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

package main

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	sentryv1alpha1 "github.com/jace-ys/sentry-operator/api/v1alpha1"
	"github.com/jace-ys/sentry-operator/controllers"
	"github.com/jace-ys/sentry-operator/pkg/sentry"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	cmd = kingpin.New("sentry-operator", "A Kubernetes operator for Sentry.").Version("v0.0.0")

	metricsAddr    = cmd.Flag("metrics-address", "Address to bind the metrics endpoint to.").Default("127.0.0.1:8080").String()
	leaderElection = cmd.Flag("leader-election", "Enable leader election for controller manager.").Bool()

	sentryURL          = cmd.Flag("sentry-url", "The URL to use to connect to Sentry.").Envar("SENTRY_URL").Default("https://sentry.io/").URL()
	sentryToken        = cmd.Flag("sentry-token", "Authentication token belonging to a user under the Sentry organization.").Envar("SENTRY_TOKEN").Required().String()
	sentryOrganization = cmd.Flag("sentry-organization", "Name of the Sentry organization to be managed.").Envar("SENTRY_ORGANIZATION").Required().String()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = sentryv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	kingpin.MustParse(cmd.Parse(os.Args[1:]))

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: *metricsAddr,
		Port:               9443,
		LeaderElection:     *leaderElection,
		LeaderElectionID:   "sentry.kubernetes.jaceys.me",
	})
	if err != nil {
		exit(err, "unable to start manager")
	}

	sentryClient := sentry.NewClient(*sentryToken, sentry.WithSentryURL(*sentryURL))
	ctrlSentry := &controllers.Sentry{
		Organization: *sentryOrganization,
		Client: &controllers.SentryClient{
			Organizations: sentryClient.Organizations,
			Projects:      sentryClient.Projects,
			Teams:         sentryClient.Teams,
		},
	}

	if err = (&controllers.ProjectReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Project"),
		Scheme: mgr.GetScheme(),
		Sentry: ctrlSentry,
	}).SetupWithManager(mgr); err != nil {
		exit(err, "unable to create controller", "controller", "Project")
	}

	if err = (&controllers.ProjectKeyReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ProjectKey"),
		Scheme: mgr.GetScheme(),
		Sentry: ctrlSentry,
	}).SetupWithManager(mgr); err != nil {
		exit(err, "unable to create controller", "controller", "ProjectKey")
	}

	if err = (&controllers.TeamReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Team"),
		Scheme: mgr.GetScheme(),
		Sentry: ctrlSentry,
	}).SetupWithManager(mgr); err != nil {
		exit(err, "unable to create controller", "controller", "Team")
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		exit(err, "problem running manager")
	}
}

func exit(err error, msg string, keysAndValues ...interface{}) {
	setupLog.Error(err, msg, keysAndValues...)
	os.Exit(1)
}
