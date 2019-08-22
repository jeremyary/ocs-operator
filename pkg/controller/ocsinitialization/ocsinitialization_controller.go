package ocsinitialization

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	ocsv1alpha1 "github.com/openshift/ocs-operator/pkg/apis/ocs/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_ocsinitialization")

// watchNamespace is the namespace the operator is watching.
var watchNamespace string

const wrongNamespacedName = "Ignoring this resource. Only one should exist, and this one has the wrong name and/or namespace."

// InitNamespacedName returns a NamespacedName for the singleton instance that
// should exist.
func InitNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      "ocsinit",
		Namespace: watchNamespace,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func Add(mgr manager.Manager, r reconcile.Reconciler) error {
	// set the watchNamespace so we know where to create the OCSInitialization resource
	ns, err := k8sutil.GetWatchNamespace()
	if err != nil {
		return err
	}
	watchNamespace = ns

	// Create a new controller
	c, err := controller.New("ocsinitialization-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource OCSInitialization
	return c.Watch(&source.Kind{Type: &ocsv1alpha1.OCSInitialization{}}, &handler.EnqueueRequestForObject{})
}

// ReconcileOCSInitialization reconciles a OCSInitialization object
type ReconcileOCSInitialization struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a OCSInitialization object and makes changes based on the state read
// and what is in the OCSInitialization.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileOCSInitialization) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling OCSInitialization")

	initNamespacedName := InitNamespacedName()
	instance := &ocsv1alpha1.OCSInitialization{}
	if initNamespacedName.Name != request.Name || initNamespacedName.Namespace != request.Namespace {
		// Ignoring this resource because it has the wrong name or namespace
		reqLogger.Info(wrongNamespacedName)
		err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
		if err != nil {
			// the resource probably got deleted
			if errors.IsNotFound(err) {
				return reconcile.Result{}, nil
			}
			return reconcile.Result{}, err
		}
		instance.Status.ErrorMessage = wrongNamespacedName

		err = r.Client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "failed to update ignored resource")
		}
		return reconcile.Result{}, err
	}

	// Fetch the OCSInitialization instance
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Recreating since we depend on this to exist. A user may delete it to
			// induce a reset of all initial data.
			reqLogger.Info("recreating OCSInitialization resource")
			return reconcile.Result{}, r.Client.Create(context.TODO(), &ocsv1alpha1.OCSInitialization{
				ObjectMeta: metav1.ObjectMeta{
					Name:      initNamespacedName.Name,
					Namespace: initNamespacedName.Namespace,
				},
			})
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Status.StorageClassesCreated == true {
		// we only create the data once and then allow changes or even deletion, so we
		// return here without inspecting or modifying the initial data.
		return reconcile.Result{}, nil
	}

	err = r.ensureStorageClasses(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	instance.Status.StorageClassesCreated = true
	err = r.Client.Status().Update(context.TODO(), instance)

	return reconcile.Result{}, err
}

// ensureStorageClasses ensures that StorageClass resources exist in the desired
// state.
func (r *ReconcileOCSInitialization) ensureStorageClasses(initialdata *ocsv1alpha1.OCSInitialization, reqLogger logr.Logger) error {
	scs, err := r.newStorageClasses(initialdata)
	if err != nil {
		return err
	}
	for _, sc := range scs {
		existing := storagev1.StorageClass{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{Name: sc.Name, Namespace: sc.Namespace}, &existing)

		switch {
		case err == nil:
			reqLogger.Info(fmt.Sprintf("Restoring original StorageClass %s", sc.Name))
			sc.DeepCopyInto(&existing)
			err = r.Client.Update(context.TODO(), &existing)
			if err != nil {
				return err
			}
		case errors.IsNotFound(err):
			reqLogger.Info(fmt.Sprintf("Creating StorageClass %s", sc.Name))
			err = r.Client.Create(context.TODO(), &sc)
			if err != nil {
				return err
			}
		default:
			return err
		}
	}
	return nil
}

// newStorageClasses returns the StorageClass instances that should be created
// on first run.
func (r *ReconcileOCSInitialization) newStorageClasses(initdata *ocsv1alpha1.OCSInitialization) ([]storagev1.StorageClass, error) {
	// TODO add the real values OCS wants to ship with.
	ret := []storagev1.StorageClass{}
	return ret, nil
}
