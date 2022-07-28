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

package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var cleanerlog = logf.Log.WithName("cleaner-resource")

const ConfigResourceName = "cleaner"

func (r *Cleaner) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-actions-cnative-dev-v1alpha1-cleaner,mutating=true,failurePolicy=fail,sideEffects=None,groups=actions.cnative.dev,resources=cleaners,verbs=create;update,versions=v1alpha1,name=mcleaner.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Cleaner{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Cleaner) Default() {
	cleanerlog.Info("default", "name", r.Name)

	if r.Spec.TTL == 0 {
		r.Spec.TTL = 3600
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-actions-cnative-dev-v1alpha1-cleaner,mutating=false,failurePolicy=fail,sideEffects=None,groups=actions.cnative.dev,resources=cleaners,verbs=create;update,versions=v1alpha1,name=vcleaner.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Cleaner{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Cleaner) ValidateCreate() error {
	cleanerlog.Info("validate create", "name", r.Name)
	errs := field.ErrorList{}
	// TODO(user): fill in your validation logic upon object creation.
	if r.Name != ConfigResourceName {
		errMsg := fmt.Sprintf("metadata.name,  Only one instance of Cleaner is allowed by name, %s", ConfigResourceName)
		errs = append(errs, &field.Error{
			Type:     field.ErrorTypeInvalid,
			Field:    "name",
			BadValue: r.Name,
			Detail:   errMsg})
		return apierrors.NewInvalid(schema.GroupKind{Group: "actions.cnative.dev", Kind: r.Kind}, r.Name, errs)
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Cleaner) ValidateUpdate(old runtime.Object) error {
	cleanerlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cleaner) ValidateDelete() error {
	cleanerlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
