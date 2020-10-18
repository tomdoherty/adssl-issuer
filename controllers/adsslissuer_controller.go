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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/tomdoherty/adssl-issuer/api/v1alpha2"
)

// AdsslIssuerReconciler reconciles a AdsslIssuer object
type AdsslIssuerReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Clock    clock.Clock
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=certmanager.tomdoherty.io,resources=adsslissuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certmanager.tomdoherty.io,resources=adsslissuers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile will read and validate the AdsslIssuer resources, it will set the
// status condition ready to true if everything is right.
func (r *AdsslIssuerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("adsslissuer", req.NamespacedName)

	iss := new(api.AdsslIssuer)
	if err := r.Client.Get(ctx, req.NamespacedName, iss); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	statusReconciler := newAdsslStatusReconciler(r, iss, log)
	log.Info("Registered issuer")

	return ctrl.Result{}, statusReconciler.Update(ctx, api.ConditionTrue, "Verified", "AdsslIssuer verified and ready to sign certificates")
}

// SetupWithManager initializes the AdsslIssuer controller into the
// controller runtime.
func (r *AdsslIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.AdsslIssuer{}).
		Complete(r)
}
