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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const ActionName = "cleaner"
const DefaultTTL = 3600

// CleanerReconciler reconciles a Cleaner object
type CleanerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;delete

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

	namespace := &corev1.Namespace{}

	if err := r.Get(ctx, req.NamespacedName, namespace); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Error Get Namespace: "+namespace.GetName())
		return ctrl.Result{}, err
	}

	if namespace.Status.Phase == corev1.NamespaceTerminating {
		return ctrl.Result{}, nil
	}

	if enable, exists := namespace.GetLabels()["cnative/operator.actions."+ActionName]; !exists || enable == "false" {
		logger.Info(namespace.GetName() + " cleaner not activated")
		return ctrl.Result{}, nil
	}
	ttl := DefaultTTL
	if customTTL, exists := namespace.GetLabels()["cnative/operator.actions."+ActionName+".ttl"]; exists {
		if intCustomTTL, err := strconv.Atoi(customTTL); err == nil {
			ttl = intCustomTTL
		}
	}

	ddl := namespace.CreationTimestamp.Add(time.Duration(ttl) * time.Duration(time.Second))
	if ddl.Before(time.Now()) {
		if err := r.Delete(ctx, namespace); err != nil {
			if errors.IsNotFound(err) {
				logger.Error(err, "Not Found when Delete, skip")
				return ctrl.Result{}, nil
			}
			logger.Error(err, "Ops")
			return ctrl.Result{}, err
		}
		logger.Info(namespace.GetName() + " deleting...")
		return ctrl.Result{}, nil
	} else {
		logger.Info(namespace.GetName() + " not expired, requeue later...")
		return ctrl.Result{RequeueAfter: time.Until(ddl) + time.Second*1}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CleanerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
