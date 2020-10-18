/*
Copyright (c) 2020 Tom Doherty <tom@tomdoherty.io>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIEDi
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	api "github.com/tomdoherty/adssl-issuer/api/v1alpha2"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type adsslStatusReconciler struct {
	*AdsslIssuerReconciler
	issuer *api.AdsslIssuer
	logger logr.Logger
}

func newAdsslStatusReconciler(r *AdsslIssuerReconciler, iss *api.AdsslIssuer, log logr.Logger) *adsslStatusReconciler {
	return &adsslStatusReconciler{
		AdsslIssuerReconciler: r,
		issuer:                iss,
		logger:                log,
	}
}

func (r *adsslStatusReconciler) Update(ctx context.Context, status api.ConditionStatus, reason, message string, args ...interface{}) error {
	completeMessage := fmt.Sprintf(message, args...)
	r.setCondition(status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == api.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(r.issuer, eventType, reason, completeMessage)

	return r.Client.Status().Update(ctx, r.issuer)
}

func (r *adsslStatusReconciler) UpdateNoError(ctx context.Context, status api.ConditionStatus, reason, message string, args ...interface{}) {
	if err := r.Update(ctx, status, reason, message, args...); err != nil {
		r.logger.Error(err, "failed to update", "status", status, "reason", reason)
	}
}

// setCondition will set a 'condition' on the given api.AdsslIssuer resource.
//
// - If no condition of the same type already exists, the condition will be
//   inserted with the LastTransitionTime set to the current time.
// - If a condition of the same type and state already exists, the condition
//   will be updated but the LastTransitionTime will not be modified.
// - If a condition of the same type and different state already exists, the
//   condition will be updated and the LastTransitionTime set to the current
//   time.
func (r *adsslStatusReconciler) setCondition(status api.ConditionStatus, reason, message string) {
	now := meta.NewTime(r.Clock.Now())
	c := api.AdsslIssuerCondition{
		Type:               api.ConditionReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: &now,
	}

	// Search through existing conditions
	for idx, cond := range r.issuer.Status.Conditions {
		// Skip unrelated conditions
		if cond.Type != api.ConditionReady {
			continue
		}

		// If this update doesn't contain a state transition, we don't update
		// the conditions LastTransitionTime to Now()
		if cond.Status == status {
			c.LastTransitionTime = cond.LastTransitionTime
		} else {
			r.logger.Info("found status change for AdsslIssuer condition; setting lastTransitionTime", "condition", cond.Type, "old_status", cond.Status, "new_status", status, "time", now.Time)
		}

		// Overwrite the existing condition
		r.issuer.Status.Conditions[idx] = c
		return
	}

	// If we've not found an existing condition of this type, we simply insert
	// the new condition into the slice.
	r.issuer.Status.Conditions = append(r.issuer.Status.Conditions, c)
	r.logger.Info("setting lastTransitionTime for AdsslIssuer condition", "condition", api.ConditionReady, "time", now.Time)
}
