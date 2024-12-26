/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	corev1 "github.com/lixd/i-operator/api/v1"
)

// nolint:unused
// log is for logging in this package.
var applicationlog = logf.Log.WithName("application-resource")

// SetupApplicationWebhookWithManager registers the webhook for Application in the manager.
func SetupApplicationWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.Application{}).
		WithValidator(&ApplicationCustomValidator{}).
		WithDefaulter(&ApplicationCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-core-crd-lixueduan-com-v1-application,mutating=true,failurePolicy=fail,sideEffects=None,groups=core.crd.lixueduan.com,resources=applications,verbs=create;update,versions=v1,name=mapplication-v1.lixueduan.com,admissionReviewVersions=v1

// ApplicationCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Application when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ApplicationCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &ApplicationCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Application.
func (d *ApplicationCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	application, ok := obj.(*corev1.Application)

	if !ok {
		return fmt.Errorf("expected an Application object but got %T", obj)
	}
	applicationlog.Info("Defaulting for Application", "name", application.GetName())

	// TODO(user): fill in your defaulting logic.

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-core-crd-lixueduan-com-v1-application,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.crd.lixueduan.com,resources=applications,verbs=create;update,versions=v1,name=vapplication-v1.lixueduan.com,admissionReviewVersions=v1

// ApplicationCustomValidator struct is responsible for validating the Application resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ApplicationCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &ApplicationCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Application.
func (v *ApplicationCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	application, ok := obj.(*corev1.Application)
	if !ok {
		return nil, fmt.Errorf("expected a Application object but got %T", obj)
	}
	applicationlog.Info("Validation for Application upon creation", "name", application.GetName())

	if !isValidImageName(application.Spec.Image) {
		return nil, fmt.Errorf("invalid image name: %s", application.Spec.Image)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Application.
func (v *ApplicationCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	application, ok := newObj.(*corev1.Application)
	if !ok {
		return nil, fmt.Errorf("expected a Application object for the newObj but got %T", newObj)
	}
	applicationlog.Info("Validation for Application upon update", "name", application.GetName())

	if !isValidImageName(application.Spec.Image) {
		return nil, fmt.Errorf("invalid image name: %s", application.Spec.Image)
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Application.
func (v *ApplicationCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	application, ok := obj.(*corev1.Application)
	if !ok {
		return nil, fmt.Errorf("expected a Application object but got %T", obj)
	}
	applicationlog.Info("Validation for Application upon deletion", "name", application.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}

// IsValidImageName checks if the given image name is a valid image name.
func isValidImageName(imageName string) bool {
	_, err := name.ParseReference(imageName)
	return err == nil
}
