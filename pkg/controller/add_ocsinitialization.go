package controller

import (
	"github.com/openshift/ocs-operator/pkg/controller/ocsinitialization"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	addManager := func(mgr manager.Manager) error {
		reconciler := ocsinitialization.ReconcileOCSInitialization{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
		return ocsinitialization.Add(mgr, &reconciler)
	}
	AddToManagerFuncs = []func(manager manager.Manager) error{addManager}
}
