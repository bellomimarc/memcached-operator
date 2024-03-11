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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cachev1alpha1 "github.com/bellomimarc/memcached-operator.git/api/v1alpha1"
)

var contollerLog = ctrl.Log.WithName("controller")

// MemcachedReconciler reconciles a Memcached object
type MemcachedReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.g691465.artifact-registry.corvina.cloud,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.g691465.artifact-registry.corvina.cloud,resources=memcacheds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.g691465.artifact-registry.corvina.cloud,resources=memcacheds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Memcached object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *MemcachedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	contollerLog.Info(">>>>>>>>>>Reconciling Memcached")

	memcached := &cachev1alpha1.Memcached{}
	err := r.Client.Get(ctx, req.NamespacedName, memcached)
	if err != nil {
		contollerLog.Error(err, "unable to fetch Memcached")
		return ctrl.Result{}, err
	}

	contollerLog.Info(">>>>>>>>>>Memcached fetched", "memcached", memcached)

	for i := range make([]int, memcached.Spec.Size) {
		contollerLog.Info(">>>>>>>>>>ConfigMap", "configMap", i)

		// create configmap with name memcached-<memcached.Name>-<i>
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("memcached-%s-%d", memcached.Name, i),
				Namespace: memcached.Namespace,
			},
		}
		err = r.Client.Create(ctx, cm)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				contollerLog.Info(">>>>>>>>>>ConfigMap already exists", "configMap", cm)
				continue
			}
			contollerLog.Error(err, "unable to create ConfigMap")
			return ctrl.Result{}, err
		}

		contollerLog.Info(">>>>>>>>>>ConfigMap created", "configMap", cm)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	contollerLog.Info(">>>>>>>>>>Setting up with manager")

	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Memcached{}).
		Complete(r)
}
