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
	"k8s.io/apimachinery/pkg/types"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cachev1alpha1 "mongo-operator/api/v1alpha1"
)

// MongoReconciler reconciles a Mongo object
type MongoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=cache.my.domain,resources=mongoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.my.domain,resources=mongoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.my.domain,resources=mongoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Mongo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *MongoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	r.Log = ctrl.Log.WithValues("mongo", req.NamespacedName)

	mongo := &cachev1alpha1.Mongo{}

	err := r.Get(ctx, req.NamespacedName, mongo)

	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("Mongo resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		r.Log.Error(err, "Failed to get Tamim")
		return ctrl.Result{}, err
	}

	found := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{Name: mongo.Name, Namespace: mongo.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		dep := r.StateFulSetForMongo(mongo)

		r.Log.Error(err, "Failed to create new StateFulSet", "StateFulSet.Namespace", dep.Namespace, "StateFulSet.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			r.Log.Error(err, "Failed to create new StateFulSet", "StateFulSet.Namespace", dep.Namespace, "StateFulSet.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// StateFulSet created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get StateFulSet")
		return ctrl.Result{}, err
	}

	// Ensure the stateFulSet size is the same as the spec
	size := mongo.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			r.Log.Error(err, "Failed to update StateFulSet", "StateFulSet.Namespace", found.Namespace, "StateFulSet.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(mongo.Namespace),
		client.MatchingLabels(labelsForMongo(mongo.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		r.Log.Error(err, "Failed to list pods", "Tamim.Namespace", mongo.Namespace, "Tamim.Name", mongo.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)
	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, mongo.Status.Nodes) {
		mongo.Status.Nodes = podNames
		err := r.Status().Update(ctx, mongo)
		if err != nil {
			r.Log.Error(err, "Failed to update Memcached status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// StateFulSet returns a Mongo StateFulSet  object

func (r *MongoReconciler) StateFulSetForMongo(m *cachev1alpha1.Mongo) *appsv1.StatefulSet {
	ls := labelsForMongo(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "mongo",
						Name:  "mongo",
						//Command: []string{"customer", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 27017,
							Name:          "mongo",
						}},
						Env: []corev1.EnvVar{{
							Name:  "MONGO_INITDB_ROOT_USERNAME",
							Value: "mongoadmin",
						},
							{
								Name:  "MONGO_INITDB_ROOT_PASSWORD",
								Value: "abc123",
							}},
					}},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func labelsForMongo(name string) map[string]string {
	return map[string]string{"app": "mongo"}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Mongo{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
