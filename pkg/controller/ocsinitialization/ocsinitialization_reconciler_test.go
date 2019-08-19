package ocsinitialization

import (
	v1 "github.com/openshift/ocs-operator/pkg/apis/ocs/v1alpha1"
	"github.com/openshift/ocs-operator/pkg/controller/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestOCSInitResourceNameOutsideWatchNotFound(t *testing.T) {

	cr := &v1.OCSInitialization{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.OCSInitializationSpec{},
	}

	scheme, err := v1.SchemeBuilder.Build()
	assert.Nil(t, err, "Failed to get scheme")

	mockService := test.MockService()
	mockService.GetSchemeFunc = func() *runtime.Scheme {
		return scheme
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "foo",
			Namespace: watchNamespace,
		},
	}

	reconciler := &OCSInitializationReconciler{mockService}
	_, err = reconciler.Reconcile(request)
	assert.Error(t, err)
}

func GetEnvironment(cr *v1.OCSInitialization, service v1.PlatformService) (v1.Environment, error) {
	envTemplate, err := getEnvTemplate(cr)
}

func getEnvTemplate(cr *v1.OCSInitialization)
