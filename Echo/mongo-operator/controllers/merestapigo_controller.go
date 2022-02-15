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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cachev1alpha1 "mongo-operator/api/v1alpha1"
)

// MeRestAPIGoReconciler reconciles a MeRestAPIGo object
type MeRestAPIGoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=cache.my.domain,resources=merestapigoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.my.domain,resources=merestapigoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.my.domain,resources=merestapigoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MeRestAPIGo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *MeRestAPIGoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	r.Log = ctrl.Log.WithValues("RestapiGo", req.NamespacedName)

	restapigo := &cachev1alpha1.MeRestAPIGo{}

	err := r.Get(ctx, req.NamespacedName, restapigo)

	if err != nil {
		if errors.IsNotFound(err) {
			r.Log.Info("RestAPiGo resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		r.Log.Error(err, "Failed to get RestapiGO")
		return ctrl.Result{}, err
	}
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: restapigo.Name, Namespace: restapigo.Namespace}, found)

	if err != nil && errors.IsNotFound(err) {
		dep := r.deploymentForrestapigo(restapigo)

		r.Log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)

		err = r.Create(ctx, dep)
		if err != nil {
			r.Log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		r.Log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	size := restapigo.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			r.Log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the Memcached status with the pod names
	// List the pods for this memcached's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(restapigo.Namespace),
		client.MatchingLabels(labelsForrestapigo(restapigo.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		r.Log.Error(err, "Failed to list pods", "restapigo.Namespace", restapigo.Namespace, "restapigo.Name", restapigo.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames1(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, restapigo.Status.Nodes) {
		restapigo.Status.Nodes = podNames
		err := r.Status().Update(ctx, restapigo)
		if err != nil {
			r.Log.Error(err, "Failed to update restapigo status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// deploymentForTamim returns a memcached Deployment object
func (r *MeRestAPIGoReconciler) deploymentForrestapigo(m *cachev1alpha1.MeRestAPIGo) *appsv1.Deployment {
	ls := labelsForrestapigo(m.Name)
	replicas := m.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
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
						Image: "tamim447/restapigo",
						Name:  "customer",
						//Command: []string{"customer", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "customer",
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

func labelsForrestapigo(name string) map[string]string {
	return map[string]string{"app": "restapi", "restapi_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames1(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *MeRestAPIGoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.MeRestAPIGo{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
