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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autoscalingv1 "github.com/DB-Vincent/scheduscaler/api/v1"
)

var daysOfWeek = map[string]time.Weekday{
	"Sunday":    time.Sunday,
	"Monday":    time.Monday,
	"Tuesday":   time.Tuesday,
	"Wednesday": time.Wednesday,
	"Thursday":  time.Thursday,
	"Friday":    time.Friday,
	"Saturday":  time.Saturday,
}

func (r *ScheduledScalerReconciler) GetExpectedReplicaCount(ctx context.Context, req ctrl.Request, scheduledScaler *autoscalingv1.ScheduledScaler) (int32, error) {
	log := log.FromContext(ctx)

	if scheduledScaler.Spec.SchedulingConfig != nil {
		now := time.Now()
		day := now.Weekday()

		log.V(1).Info("current server", "day", day, "time", now)

		if day >= daysOfWeek[scheduledScaler.Spec.SchedulingConfig.StartTime] && day <= daysOfWeek[scheduledScaler.Spec.SchedulingConfig.EndTime] {
			return int32(scheduledScaler.Spec.SchedulingConfig.Replica), nil
		}
	}

	return scheduledScaler.Spec.DefaultReplica, nil
}
