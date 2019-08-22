package ocsinitialization

import (
	v1 "github.com/openshift/ocs-operator/pkg/apis/ocs/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestReconcilerImplementsInterface(t *testing.T) {
	reconciler := ReconcileOCSInitialization{}
	var i interface{} = reconciler
	_, ok := i.(reconcile.Reconciler)
	assert.True(t, ok)

}

func TestNonWatchedResourceNameNotFound(t *testing.T) {
	ocs := v1.OCSInitialization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
	}
	reconciler := getReconcilerWithSchemeObject(t, &ocs)

	_, err := reconciler.Reconcile(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "foo",
			Namespace: "test-ns",
		},
	})
	assert.NoError(t, err)
}

func TestNonWatchedResourceNamespaceNotFound(t *testing.T) {
	ocs := v1.OCSInitialization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
	}
	reconciler := getReconcilerWithSchemeObject(t, &ocs)

	result, err := reconciler.Reconcile(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "foo",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)
}

func TestNonWatchedResourceStatusUpdated(t *testing.T) {
	ocs := v1.OCSInitialization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test-ns",
		},
	}
	reconciler := getReconcilerWithSchemeObject(t, &ocs)

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "test-ns",
		},
	}
	result, err := reconciler.Reconcile(request)
	assert.NoError(t, err)
	assert.Equal(t, reconcile.Result{}, result)
}

func getReconcilerWithSchemeObject(t *testing.T, obj runtime.Object) ReconcileOCSInitialization {
	scheme := getScheme(t)
	client := fake.NewFakeClientWithScheme(scheme, obj)

	return ReconcileOCSInitialization{
		Scheme: scheme,
		Client: client,
	}
}

func getScheme(t *testing.T) *runtime.Scheme {
	registerObjs := []runtime.Object{&v1.OCSInitialization{}}
	registerObjs = append(registerObjs)
	v1.SchemeBuilder.Register(registerObjs...)
	scheme, err := v1.SchemeBuilder.Build()
	if err != nil {
		assert.Fail(t, "unable to build scheme")
	}
	return scheme
}
