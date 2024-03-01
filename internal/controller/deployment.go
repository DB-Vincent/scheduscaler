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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autoscalingv1 "github.com/DB-Vincent/scheduscaler/api/v1"
)

func (r *ScheduledScalerReconciler) Deployment(ctx context.Context, req ctrl.Request, scheduledScaler *autoscalingv1.ScheduledScaler) (*appsv1.Deployment, error) {
	log := log.FromContext(ctx)

	replicas, err := r.GetExpectedReplicaCount(ctx, req, scheduledScaler)
	if err != nil {
		log.Error(err, "Could not retrieve replica count")

		return nil, err
	}

	labels := map[string]string{
		"app.kubernetes.io/name":       "scheduledScaler",
		"app.kubernetes.io/instance":   scheduledScaler.Name,
		"app.kubernetes.io/version":    "v1",
		"app.kubernetes.io/part-of":    "scheduledScaler-operator",
		"app.kubernetes.io/created-by": "controller-manager",
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      scheduledScaler.Name,
			Namespace: scheduledScaler.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           scheduledScaler.Spec.Container.Image,
						Name:            scheduledScaler.Name,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{{
							ContainerPort: int32(scheduledScaler.Spec.Container.Port),
						}},
					}},
				},
			},
		},
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(scheduledScaler, dep, r.Scheme); err != nil {
		log.Error(err, "Could not set controller owner reference")
		return nil, err
	}

	return dep, nil
}

func (r *ScheduledScalerReconciler) CreateDeploymentIfNotExists(ctx context.Context, req ctrl.Request, scheduledScaler *autoscalingv1.ScheduledScaler) (bool, error) {
	log := log.FromContext(ctx)

	dep := &appsv1.Deployment{}

	err := r.Get(ctx, types.NamespacedName{Name: scheduledScaler.Name, Namespace: scheduledScaler.Namespace}, dep)
	if err != nil && apierrors.IsNotFound(err) {
		dep, err := r.Deployment(ctx, req, scheduledScaler)
		if err != nil {
			log.Error(err, "Could not create new deployment definition for scheduledScaler")
		}

		log.Info("Creating new deployment")

		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Could not create new deployment")
			r.Recorder.Event(scheduledScaler, "Warning", "Failed", fmt.Sprintf("Could not create new deployment %s", scheduledScaler.Name))

			return false, err
		}

		r.Recorder.Event(scheduledScaler, "Normal", "Created", fmt.Sprintf("Created new deployment %s", scheduledScaler.Name))

		return true, nil
	}

	if err != nil {
		log.Error(err, "Failed to get Deployment")
		r.Recorder.Event(scheduledScaler, "Warning", "Failed", fmt.Sprintf("Failed to retrieve deployment %s", scheduledScaler.Name))

		return false, err
	}

	return false, nil
}

func (r *ScheduledScalerReconciler) UpdateDeploymentReplicaCount(ctx context.Context, req ctrl.Request, scheduledScaler *autoscalingv1.ScheduledScaler) error {
	log := log.FromContext(ctx)

	dep := &appsv1.Deployment{}

	err := r.Get(ctx, types.NamespacedName{Name: scheduledScaler.Name, Namespace: scheduledScaler.Namespace}, dep)
	if err != nil {
		log.Error(err, "Failed to get Deployment")
		r.Recorder.Event(scheduledScaler, "Warning", "Failed", fmt.Sprintf("Failed to retrieve deployment %s", scheduledScaler.Name))

		return err
	}

	replicas, err := r.GetExpectedReplicaCount(ctx, req, scheduledScaler)
	if err != nil {
		log.Error(err, "Could not retrieve replica count")

		return err
	}

	// Nothing to do here
	if replicas == *dep.Spec.Replicas {
		return nil
	}
	dep.Spec.Replicas = &replicas

	err = r.Update(ctx, dep)
	if err != nil {
		log.Error(err, "Failed to update deployment replicas for deployment")
		r.Recorder.Event(scheduledScaler, "Warning", "Failed", fmt.Sprintf("Failed to update deployment %s", scheduledScaler.Name))

		return err
	}

	err = r.GetScheduledScaler(ctx, req, scheduledScaler)
	if err != nil {
		log.Error(err, "Could not re-retrieve scheduledScaler")
		r.Recorder.Event(scheduledScaler, "Warning", "Failed", fmt.Sprintf("Failed to retrieve deployment %s", scheduledScaler.Name))

		return err
	}

	r.Recorder.Event(scheduledScaler, "Normal", "Scaled", fmt.Sprintf("Scaled deployment %s to %d replicas", scheduledScaler.Name, *dep.Spec.Replicas))

	return nil
}
