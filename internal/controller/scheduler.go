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

// Mapping so week days start in a logical way. Weeks start on monday, not on sunday.
var daysOfWeek = map[string]time.Weekday{
	"Monday":    time.Sunday,
	"Tuesday":   time.Monday,
	"Wednesday": time.Tuesday,
	"Thursday":  time.Wednesday,
	"Friday":    time.Thursday,
	"Saturday":  time.Friday,
	"Sunday":    time.Saturday,
}

func (r *ScheduledScalerReconciler) GetExpectedReplicaCount(ctx context.Context, req ctrl.Request, scheduledScaler *autoscalingv1.ScheduledScaler) (int32, error) {
	log := log.FromContext(ctx)

	if scheduledScaler.Spec.SchedulingConfig != nil {
		if scheduledScaler.Spec.SchedulingConfig.StartDate != "" && scheduledScaler.Spec.SchedulingConfig.EndDate != "" {
			now := time.Now().AddDate(0, 0, -1) // Weeks start on monday
			day := now.Weekday()

			log.V(10).Info("current server", "day", day, "startDate", daysOfWeek[scheduledScaler.Spec.SchedulingConfig.StartDate], "endDate", daysOfWeek[scheduledScaler.Spec.SchedulingConfig.EndDate])
			if day >= daysOfWeek[scheduledScaler.Spec.SchedulingConfig.StartDate] && day <= daysOfWeek[scheduledScaler.Spec.SchedulingConfig.EndDate] {
				return int32(scheduledScaler.Spec.SchedulingConfig.Replica), nil
			}
		} else if scheduledScaler.Spec.SchedulingConfig.StartTime != 0 && scheduledScaler.Spec.SchedulingConfig.EndTime != 0 {
			now := time.Now()
			hour := now.Hour()

			log.V(10).Info("current server", "hour", hour, "startTime", scheduledScaler.Spec.SchedulingConfig.StartTime, "endTime", scheduledScaler.Spec.SchedulingConfig.EndTime)
			if hour >= scheduledScaler.Spec.SchedulingConfig.StartTime && hour <= scheduledScaler.Spec.SchedulingConfig.EndTime {
				return int32(scheduledScaler.Spec.SchedulingConfig.Replica), nil
			}
		}
	}

	return scheduledScaler.Spec.DefaultReplica, nil
}
