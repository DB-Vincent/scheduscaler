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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autoscalingv1 "github.com/DB-Vincent/scheduscaler/api/v1"
)

// ScheduledScalerReconciler reconciles a ScheduledScaler object
type ScheduledScalerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=autoscaling.vincentdeborger.be,resources=scheduledscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=autoscaling.vincentdeborger.be,resources=scheduledscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=autoscaling.vincentdeborger.be,resources=scheduledscalers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Gets the ScheduledScaler resource
func (r *ScheduledScalerReconciler) GetScheduledScaler(ctx context.Context, req ctrl.Request, scheduledScaler *autoscalingv1.ScheduledScaler) error {
	err := r.Get(ctx, req.NamespacedName, scheduledScaler)
	if err != nil {
		return err
	}

	return nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *ScheduledScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.V(10).Info("Starting reconcilation!")

	scheduledScaler := &autoscalingv1.ScheduledScaler{}

	// Check if a ScheduledScaler resource exists
	err := r.GetScheduledScaler(ctx, req, scheduledScaler)
	if err != nil {
		// Ignore if it doesn't exist yet
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		// Throw error when something goes wrong
		log.Error(err, "Could not retrieve ScheduledScaler resource!")
		return ctrl.Result{}, err
	}

	// Deployment does not exist
	ok, err := r.CreateDeploymentIfNotExists(ctx, req, scheduledScaler)
	if err != nil {
		log.Error(err, "Could not create deployment for scheduledScaler")

		return ctrl.Result{}, err
	}

	if ok {
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Update replica count if mismatched
	err = r.UpdateDeploymentReplicaCount(ctx, req, scheduledScaler)
	if err != nil {
		log.Error(err, "Could not update deployment for scheduledScaler")

		return ctrl.Result{}, err
	}

	log.V(10).Info("Stopping reconcilation")

	// Reconcile every 5 minutes
	return ctrl.Result{RequeueAfter: time.Minute * time.Duration(5)}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScheduledScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalingv1.ScheduledScaler{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
