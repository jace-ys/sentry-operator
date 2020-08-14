package controllers_test

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	sentryv1alpha1 "github.com/jace-ys/sentry-operator/api/v1alpha1"
	"github.com/jace-ys/sentry-operator/controllers"
	"github.com/jace-ys/sentry-operator/controllers/controllersfakes"
	"github.com/jace-ys/sentry-operator/pkg/sentry"
	// +kubebuilder:scaffold:imports
)

var (
	testEnv   *envtest.Environment
	k8sClient client.Client
)

var (
	fakeSentryProjects *controllersfakes.FakeSentryProjects
	fakeSentryTeams    *controllersfakes.FakeSentryTeams
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	log.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	var err error
	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = sentryv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).ToNot(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	fakeSentryProjects = new(controllersfakes.FakeSentryProjects)
	fakeSentryTeams = new(controllersfakes.FakeSentryTeams)

	err = (&controllers.ProjectReconciler{
		Client: k8sManager.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Project"),
		Sentry: &controllers.Sentry{
			Organization: "organization",
			Client: &controllers.SentryClient{
				Projects: fakeSentryProjects,
				Teams:    fakeSentryTeams,
			},
		},
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	time.Sleep(3 * time.Second)
})

var _ = AfterSuite(func() {
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func testSentryProject(id, name string) *sentry.Project {
	return &sentry.Project{
		DateCreated: time.Now(),
		ID:          id,
		Name:        name,
		Slug:        name,
	}
}

func newSentryResponse(statusCode int) *sentry.Response {
	return &sentry.Response{
		Response: &http.Response{
			StatusCode: statusCode,
		},
	}
}
