/*
Copyright 2022.

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

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cachev1alpha1 "mongo-operator/api/v1alpha1"
)

// MongoserviceReconciler reconciles a Mongoservice object
type MongoserviceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=cache.my.domain,resources=mongoservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.my.domain,resources=mongoservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.my.domain,resources=mongoservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Mongoservice object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *MongoserviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	r.Log = ctrl.Log.WithValues("tamim", req.NamespacedName)

	mongoservice := &cachev1alpha1.Mongoservice{}

	err := r.Get(ctx, req.NamespacedName, mongoservice)

	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("Mongoservice resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		r.Log.Error(err, "Failed to get Mongoservice")
		return ctrl.Result{}, err
	}
	found := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: mongoservice.Name, Namespace: mongoservice.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		dep := r.ServiceForMongo(mongoservice)

		r.Log.Error(err, "Failed to create new Service", "Service.Namespace", dep.Namespace, "Service.Name", dep.Name)

		err = r.Create(ctx, dep)
		if err != nil {
			r.Log.Error(err, "Failed to create new Service", "Service.Namespace", dep.Namespace, "Service.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// ServiceForMongo returns a service object
func (r *MongoserviceReconciler) ServiceForMongo(m *cachev1alpha1.Mongoservice) *corev1.Service {
	ls := labelsForMongoService(m.Name)
	//replicas := m.Spec.Size

	dep := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{Port: 27017,
					TargetPort: intstr.IntOrString{IntVal: 27017},
					Protocol:   "TCP"},
			},
		},
	}

	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func labelsForMongoService(name string) map[string]string {
	return map[string]string{"app": "mongo"}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoserviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Mongoservice{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
