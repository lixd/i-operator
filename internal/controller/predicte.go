package controller

import (
	v1 "github.com/lixd/i-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Predicate to trigger reconciliation only on size changes in the Busybox spec
var updatePred = predicate.Funcs{
	// Only allow updates when the spec.image or spec.enabled of the Application resource changes
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldObj := e.ObjectOld.(*v1.Application)
		newObj := e.ObjectNew.(*v1.Application)

		// Trigger reconciliation only if the spec.size field has changed
		return oldObj.Spec.Image != newObj.Spec.Image || oldObj.Spec.Enabled != newObj.Spec.Enabled
	},

	// Allow create events
	CreateFunc: func(e event.CreateEvent) bool {
		return true
	},

	// Allow delete events
	DeleteFunc: func(e event.DeleteEvent) bool {
		return true
	},

	// Allow generic events (e.g., external triggers)
	GenericFunc: func(e event.GenericEvent) bool {
		return true
	},
}
