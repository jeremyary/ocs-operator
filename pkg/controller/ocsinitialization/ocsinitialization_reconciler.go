package ocsinitialization

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/openshift/ocs-operator/pkg/apis/ocs/v1alpha1"
	v12 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

const wrongNamespacedName = "Ignoring resource. Only one should exist, this one has wrong name and/or namespace."

type OCSInitializationReconciler struct {
	Service v1alpha1.PlatformService
}

// read state of cluster for an OCSInitialization object and makes changes based on comparison to OCSInitialization.spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (reconciler *OCSInitializationReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling OCSInitialization")

	namespacedName := InitNamespacedName()
	instance := &v1alpha1.OCSInitialization{}

	// Ignore requests for resources outside of operator's watched namespace
	if namespacedName.Name != request.Name || namespacedName.Namespace != request.Namespace {
		reqLogger.Info(wrongNamespacedName)

		// attempt fetch of resource from requested (non-watched) namespace for error status update
		err := reconciler.getObj(request.NamespacedName, instance)
		if err != nil {
			// resource probably got deleted
			if errors.IsNotFound(err) {
				return reconcile.Result{}, nil
			}
			return reconcile.Result{}, err
		}
		instance.Status.ErrorMessage = wrongNamespacedName

		// non-watched resource located, attempt to update resource with new error status
		return reconciler.updateObj(instance)
	}

	// resource namespaced correctly, attempt fetch of resource from watched namespace
	err := reconciler.getObj(request.NamespacedName, instance)
	// Request object not found, could have been deleted after reconcile request.
	// Recreating since we depend on this to exist. A user may delete it to
	// induce a reset of all initial data.
	reqLogger.Info("recreating OCSInitialization resource")
	_, err = reconciler.createObj(&v1alpha1.OCSInitialization{
		ObjectMeta: v1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
	}, err)
	if err != nil {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.Status.StorageClassesCreated == true {
		// we only create the data once and then allow changes or even deletion, so we
		// return here without inspecting or modifying the initial data.
		return reconcile.Result{}, nil
	}

	err = reconciler.ensureStorageClasses(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	instance.Status.StorageClassesCreated = true
	return reconciler.updateObj(instance)
}

// ensure StorageClass resources exist in the desired state
func (reconciler *OCSInitializationReconciler) ensureStorageClasses(initialdata *v1alpha1.OCSInitialization, reqLogger logr.Logger) error {
	storageClasses, err := reconciler.newStorageClasses(initialdata)
	if err != nil {
		return err
	}
	for _, sc := range storageClasses {
		existing := v12.StorageClass{}
		err := reconciler.getObj(types.NamespacedName{Name: sc.Name, Namespace: sc.Namespace}, &existing)

		switch {
		case err == nil:
			reqLogger.Info(fmt.Sprintf("Restoring original StorageClass %s", sc.Name))
			sc.DeepCopyInto(&existing)
			_, err = reconciler.updateObj(&existing)
			if err != nil {
				return err
			}
		case errors.IsNotFound(err):
			reqLogger.Info(fmt.Sprintf("Creating StorageClass %s", sc.Name))
			_, err = reconciler.createObj(&sc, err)
			if err != nil {
				return err
			}
		default:
			return err
		}
	}
	return nil
}

// return StorageClass instances that should be created on first run
func (reconciler *OCSInitializationReconciler) newStorageClasses(initdata *v1alpha1.OCSInitialization) ([]v12.StorageClass, error) {
	// TODO add the real values OCS wants to ship with.
	ret := []v12.StorageClass{}
	return ret, nil
}

// createObj creates an object based on the error passed in from a `client.Get`
func (reconciler *OCSInitializationReconciler) createObj(obj v1alpha1.OpenShiftObject, err error) (reconcile.Result, error) {
	logger := log.WithValues("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName(), "namespace", obj.GetNamespace())

	if err != nil && errors.IsNotFound(err) {
		// Define a new Object
		logger.Info("Creating")
		err = reconciler.Service.Create(context.TODO(), obj)
		if err != nil {
			logger.Info("Failed to create object. ", err)
			return reconcile.Result{}, err
		}
		// Object created successfully - return and requeue
		return reconcile.Result{RequeueAfter: time.Duration(200) * time.Millisecond}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get object. ", err)
		return reconcile.Result{}, err
	}
	logger.Info("Skip reconcile - object already exists")
	return reconcile.Result{}, nil
}

// updateObj reconciles the given object
func (reconciler *OCSInitializationReconciler) updateObj(obj v1alpha1.OpenShiftObject) (reconcile.Result, error) {
	logger := log.WithValues("kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName(), "namespace", obj.GetNamespace())
	logger.Info("Updating")

	err := reconciler.Service.Update(context.TODO(), obj)
	if err != nil {
		logger.Info("Failed to update object. ", err)
		return reconcile.Result{}, err
	}
	// Object updated - return and requeue
	return reconcile.Result{Requeue: true}, nil
}

// getObj returns an object from `r.client.Get`
func (reconciler *OCSInitializationReconciler) getObj(key client.ObjectKey, obj runtime.Object) error {
	return reconciler.Service.Get(context.TODO(), key, obj)
}