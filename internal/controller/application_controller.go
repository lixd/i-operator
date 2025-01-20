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

package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"k8s.io/client-go/tools/record"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	v1 "github.com/lixd/i-operator/api/v1"
	pkgerror "github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	AppFinalizer = "lixueduan.com/application"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// See that we added the following code to allow us to pass the record.EventRecorder
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=core.crd.lixueduan.com,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.crd.lixueduan.com,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.crd.lixueduan.com,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx)
	log := logger.WithValues("application", req.NamespacedName)

	log.Info("start reconcile")
	// query app
	var app v1.Application
	err := r.Get(ctx, req.NamespacedName, &app)
	if err != nil {
		log.Error(err, "unable to fetch application")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if app.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// to registering our finalizer.
		if !controllerutil.ContainsFinalizer(&app, AppFinalizer) {
			controllerutil.AddFinalizer(&app, AppFinalizer)
			if err = r.Update(ctx, &app); err != nil {
				log.Error(err, "unable to add finalizer to application")
				return ctrl.Result{}, err
			}
			r.Recorder.Eventf(&app, corev1.EventTypeNormal, "AddFinalizer", fmt.Sprintf("add finalizer %s", AppFinalizer))
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(&app, AppFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			if err = r.deleteExternalResources(&app); err != nil {
				log.Error(err, "unable to cleanup application")
				// if fail to delete the external dependency here, return with error
				// so that it can be retried.
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&app, AppFinalizer)
			if err = r.Update(ctx, &app); err != nil {
				return ctrl.Result{}, err
			}
			r.Recorder.Eventf(&app, corev1.EventTypeNormal, "RemoveFinalizer", fmt.Sprintf("remove finalizer %s", AppFinalizer))
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Your reconcile logic
	log.Info("run reconcile logic")
	if err = r.syncApp(ctx, app); err != nil {
		log.Error(err, "unable to sync application")
		return ctrl.Result{}, err
	}

	// sync status
	var deploy appsv1.Deployment
	objKey := client.ObjectKey{Namespace: app.Namespace, Name: deploymentName(app.Name)}
	err = r.Get(ctx, objKey, &deploy)
	if err != nil {
		log.Error(err, "unable to fetch deployment", "deployment", objKey.String())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	copyApp := app.DeepCopy()
	// now,if ready replicas is gt 1,set status to true
	copyApp.Status.Ready = deploy.Status.ReadyReplicas >= 1
	if !reflect.DeepEqual(app, copyApp) { // update when changed
		log.Info("app changed,update app status")
		if err = r.Client.Status().Update(ctx, copyApp); err != nil {
			log.Error(err, "unable to update application status")
			return ctrl.Result{}, err
		}
		r.Recorder.Eventf(&app, corev1.EventTypeNormal, "UpdateStatus", fmt.Sprintf("update status from %v to %v", app.Status, copyApp.Status))
	}

	// Requeue every 5 minutes,to keep application always ready
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Application{}).
		Watches(&appsv1.Deployment{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []ctrl.Request {
				app, ok := obj.GetLabels()["app"]
				if !ok { // if no app label,means not owned by app,do nothing
					return nil
				}
				return []ctrl.Request{{NamespacedName: types.NamespacedName{
					Namespace: obj.GetNamespace(),
					Name:      app, // return name is app name,not deployment name
				},
				}}
			})).
		Named("application").
		Complete(r)
}

func (r *ApplicationReconciler) syncApp(ctx context.Context, app v1.Application) error {
	if app.Spec.Enabled {
		// if enabled, create deployment
		// if deployment already exists, update it,if need
		return r.syncAppEnabled(ctx, app)
	}
	// if not enabled, delete deployment
	return r.syncAppDisable(ctx, app)
}

func (r *ApplicationReconciler) syncAppDisable(ctx context.Context, app v1.Application) error {
	var deploy appsv1.Deployment
	objKey := client.ObjectKey{Namespace: app.Namespace, Name: deploymentName(app.Name)}
	err := r.Get(ctx, objKey, &deploy)
	if err != nil {
		if errors.IsNotFound(err) { // if not found, maybe already deleted,do nothing
			return nil
		}
		return pkgerror.WithMessagef(err, "unable to fetch deployment [%s]", objKey.String())
	}

	log.Log.Info("reconcile application delete deployment", "app", app.Namespace, "deployment", objKey.Name)

	if err = r.Delete(ctx, &deploy); err != nil {
		return pkgerror.WithMessage(err, "unable to delete deployment")
	}
	return nil
}

func (r *ApplicationReconciler) syncAppEnabled(ctx context.Context, app v1.Application) error {
	var deploy appsv1.Deployment
	objKey := client.ObjectKey{Namespace: app.Namespace, Name: deploymentName(app.Name)}
	err := r.Get(ctx, objKey, &deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Log.Info("reconcile application create deployment", "app", app.Namespace, "deployment", objKey.Name)
			deploy = r.generateDeployment(app)
			if err = r.Create(ctx, &deploy); err != nil {
				return pkgerror.WithMessage(err, "unable to create deployment")
			}
		}
		return pkgerror.WithMessagef(err, "unable to fetch deployment [%s]", objKey.String())
	}
	// update deployment if needed
	if !equal(app, deploy) {
		log.Log.Info("reconcile application update deployment", "app", app.Namespace, "deployment", objKey.Name)
		deploy.Spec.Template.Spec.Containers[0].Image = app.Spec.Image
		if err = r.Update(ctx, &deploy); err != nil {
			return pkgerror.WithMessage(err, "unable to update deployment")
		}
	}
	return nil
}

func (r *ApplicationReconciler) generateDeployment(app v1.Application) appsv1.Deployment {
	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName(app.Name),
			Namespace: app.Namespace,
			Labels: map[string]string{
				"app": app.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)), // 副本数
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": app.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": app.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  app.Name,
							Image: app.Spec.Image,
							// Ports: []corev1.ContainerPort{
							//	{
							//		ContainerPort: 80,
							//	},
							// },
						},
					},
				},
			},
		},
	}
	// Set the ownerRef for the Deployment, ensuring that the Deployment
	// will be deleted when the Application CR is deleted.
	_ = controllerutil.SetControllerReference(&app, &deploy, r.Scheme)
	return deploy
}

func deploymentName(app string) string {
	return fmt.Sprintf("app-%s", app)
}

func equal(app v1.Application, deploy appsv1.Deployment) bool {
	// only check image for now
	return deploy.Spec.Template.Spec.Containers[0].Image == app.Spec.Image
}

func (r *ApplicationReconciler) deleteExternalResources(app *v1.Application) error {
	//
	// delete any external resources associated with the cronJob
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple times for same object.
	return nil
}
