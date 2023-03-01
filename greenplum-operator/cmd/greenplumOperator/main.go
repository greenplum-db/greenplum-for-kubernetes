package main

import (
	"flag"
	"os"
	goruntime "runtime"

	// Enable auth plugin for GCP
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/go-logr/logr"
	"github.com/jessevdk/go-flags"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/sshkeygen"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/multidaemon"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	// +kubebuilder:scaffold:imports
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	err := Run()
	if err != nil {
		setupLog.Error(err, "error")
		os.Exit(1)
	}
}

func Run() error {
	options := GreenplumOperatorOptions{}
	parseCommandLine(&options)
	ctrl.SetLogger(gplog.ForProd(options.LogLevel == "debug"))

	logGoInfo(setupLog)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: ":8080",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	podExec := executor.NewPodExec(mgr.GetScheme(), mgr.GetConfig())

	instanceImage, err := GetInstanceImageFromEnv(os.Getenv)
	if err != nil {
		return errors.Wrap(err, "getting greenplum image name")
	}

	operatorImage, err := GetOperatorImageFromEnv(os.Getenv)
	if err != nil {
		return errors.Wrap(err, "getting operator image name")
	}

	if err = (&controllers.GreenplumPXFServiceReconciler{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName("GreenplumPXFService"),
		InstanceImage: instanceImage,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GreenplumPXFService")
		return err
	}

	if err = (&greenplumcluster.GreenplumClusterReconciler{
		Client:        mgr.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName("GreenplumCluster"),
		SSHCreator:    sshkeygen.New(),
		InstanceImage: instanceImage,
		OperatorImage: operatorImage,
		PodExec:       podExec,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GreenplumCluster")
		return err
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if errs := multidaemon.InitializeDaemons(ctrl.SetupSignalHandler(), mgr.Start); len(errs) > 0 {
		return k8serrors.NewAggregate(errs)
	}
	return nil
}

func logGoInfo(log logr.Logger) {
	log.Info("Go Info", "Version", goruntime.Version(), "GOOS", goruntime.GOOS, "GOARCH", goruntime.GOARCH)
}

type GreenplumOperatorOptions struct {
	LogLevel string `short:"v" long:"logLevel" default:"info" description:"Log verbosity" choice:"info" choice:"debug"`
}

// Parse with both jessevdk/go-flags and the golang flag package
func parseCommandLine(options *GreenplumOperatorOptions) {
	args, err := flags.ParseArgs(options, os.Args[1:])
	if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		flag.CommandLine.Usage() // Also print usage/help for golang flag package
		os.Exit(1)
	} else if err != nil {
		os.Exit(1) // ParseArgs() will have printed the error already.
	}
	// must do this for libraries that use glog, even if we don't use glog in our code
	_ = flag.CommandLine.Parse(args) // See flag.Parse() for why we can ignore the return.
}

func GetInstanceImageFromEnv(getenv func(key string) string) (string, error) {
	return getImageNameFromEnv("GREENPLUM_IMAGE", getenv)
}

func GetOperatorImageFromEnv(getenv func(key string) string) (string, error) {
	return getImageNameFromEnv("OPERATOR_IMAGE", getenv)
}

func getImageNameFromEnv(envPrefix string, getenv func(key string) string) (string, error) {
	imageRepo := getenv(envPrefix + "_REPO")
	if imageRepo == "" {
		return "", errors.New(envPrefix + "_REPO cannot be empty")
	}

	imageTag := getenv(envPrefix + "_TAG")
	if imageTag == "" {
		return "", errors.New(envPrefix + "_TAG cannot be empty")
	}

	imageName := imageRepo + ":" + imageTag

	return imageName, nil
}
