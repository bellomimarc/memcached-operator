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
	"regexp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cachev1alpha1 "github.com/bellomimarc/memcached-operator.git/api/v1alpha1"
)

var contollerLog = ctrl.Log.WithName("controller")

// MemcachedReconciler reconciles a Memcached object
type MemcachedReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
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

	// check if the name is in the form memcached-%s-%d
	matched, err := regexp.Match("memcached-\\w+-\\d+", []byte(req.Name))
	if err != nil {
		panic(err)
	}
	if matched {
		contollerLog.Info(">>>>>>>>>Reconciliation starts for config map change", "Name", req.Name)

		// TODO: extract CR an call CreateResources method

		return ctrl.Result{}, nil
	}

	memcached := &cachev1alpha1.Memcached{}
	err = r.Client.Get(ctx, req.NamespacedName, memcached)
	if err != nil {
		contollerLog.Error(err, "unable to fetch Memcached")
		return ctrl.Result{}, nil
	}

	err = r.CreateResources(ctx, memcached)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MemcachedReconciler) CreateResources(ctx context.Context, memcached *cachev1alpha1.Memcached) error {
	contollerLog.Info(">>>>>>>>>>Memcached fetched", "memcached", memcached)

	for i := range make([]int, memcached.Spec.Size) {
		contollerLog.Info(">>>>>>>>>>ConfigMap", "configMap", i)

		// create configmap with name memcached-<memcached.Name>-<i>
		annotation := make(map[string]string)
		annotation["corvina/ownedby"] = "my-super-operator"

		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        fmt.Sprintf("memcached-%s-%d", memcached.Name, i),
				Namespace:   memcached.Namespace,
				Annotations: annotation,
			},
		}

		err := r.Client.Create(ctx, cm)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				contollerLog.Info(">>>>>>>>>>ConfigMap already exists", "configMap", cm)
				continue
			}
			contollerLog.Error(err, "unable to create ConfigMap")
			return err
		}

		contollerLog.Info(">>>>>>>>>>ConfigMap created", "configMap", cm)
	}

	return nil
}

func (r *MemcachedReconciler) filterConfigMap(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}
	// Check if the desired annotation is present
	if _, ok := annotations["corvina/ownedby"]; ok {
		return true
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	contollerLog.Info(">>>>>>>>>>Setting up with manager")

	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Memcached{}).
		Owns(&corev1.ConfigMap{}).
		Watches(
			&corev1.ConfigMap{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return r.filterConfigMap(e.Object.GetAnnotations())
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return r.filterConfigMap(e.Object.GetAnnotations())
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					return r.filterConfigMap(e.ObjectNew.GetAnnotations())
				},
			})).
		Complete(r)
}
