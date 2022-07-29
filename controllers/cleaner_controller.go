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
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	actionsv1alpha1 "github.com/cnative-dev/cnative-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const ActionName = "cleaner"
const resync = time.Duration(10) * time.Minute

// CleanerReconciler reconciles a Cleaner object
type CleanerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=actions.cnative.dev,resources=cleaners,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=actions.cnative.dev,resources=cleaners/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=actions.cnative.dev,resources=cleaners/finalizers,verbs=update
//+kubebuilder:rbac:groups=v1,resources=namespaces,verbs=get;list;watch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cleaner object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *CleanerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconcile...")
	ttl := actionsv1alpha1.DefaultTTL
	cleaner := &actionsv1alpha1.Cleaner{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: "",
		Name:      actionsv1alpha1.ConfigResourceName,
	}, cleaner); err == nil {
		ttl = cleaner.Spec.TTL
	} else {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error Get Cleaner")
			return ctrl.Result{}, err
		}
	}

	namespace := &corev1.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, namespace); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error Get Namespace")
		return ctrl.Result{}, err
	}
	logger.Info("Get Namespace:" + namespace.GetName())

	if enable, exists := namespace.GetLabels()["cnative/operator.actions."+ActionName]; !exists || enable == "false" {
		logger.Info("Cleaner not activated")
		return ctrl.Result{}, nil
	}

	if customTTL, exists := namespace.GetLabels()["cnative/operator.actions."+ActionName+".ttl"]; exists {
		if intCustomTTL, err := strconv.Atoi(customTTL); err == nil {
			ttl = intCustomTTL
		}
	}

	if namespace.CreationTimestamp.Add(time.Duration(ttl) * time.Duration(time.Second)).Before(time.Now()) {
		if err := r.Delete(ctx, namespace); err != nil {
			if errors.IsNotFound(err) {
				logger.Error(err, "Not Found when Delete, skip")
				return ctrl.Result{}, nil
			}
			logger.Error(err, "Ops")
			return ctrl.Result{}, err
		}
		logger.Info("Deleted")
		return ctrl.Result{}, nil
	} else {
		logger.Info("No expired, requeue later...")
		return ctrl.Result{RequeueAfter: resync}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CleanerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
