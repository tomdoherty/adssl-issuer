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
	apiutil "github.com/jetstack/cert-manager/pkg/api/util"
	cmapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"github.com/tomdoherty/adssl"
	api "github.com/tomdoherty/adssl-issuer/api/v1alpha2"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CertificateRequestReconciler reconciles a AdsslIssuer object.
type CertificateRequestReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificaterequests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=patch

// Reconcile will read and validate the CertificateRequest resources, it will set the
// status condition ready to true if everything is right.
func (r *CertificateRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("certificaterequest", req.NamespacedName)

	cr := new(cmapi.CertificateRequest)
	if err := r.Client.Get(ctx, req.NamespacedName, cr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if cr.Spec.IssuerRef.Group != "" && cr.Spec.IssuerRef.Group != api.GroupVersion.Group {
		log.V(4).Info("resource does not specify an issuerRef group name that we are responsible for", "group", cr.Spec.IssuerRef.Group)
		return ctrl.Result{}, nil
	}

	if len(cr.Status.Certificate) > 0 {
		log.V(4).Info("existing certificate data found in status, skipping already completed CertificateRequest")
		return ctrl.Result{}, nil
	}
	iss := api.AdsslIssuer{}
	issNamespaceName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      cr.Spec.IssuerRef.Name,
	}
	if err := r.Client.Get(ctx, issNamespaceName, &iss); err != nil {
		log.Error(err, "failed to retrieve AdsslIssuer resource", "namespace", req.Namespace, "name", cr.Spec.IssuerRef.Name)
		_ = r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending, "Failed to retrieve AdsslIssuer resource %s: %v", issNamespaceName, err)
		return ctrl.Result{}, err
	}
	// Check if the AdsslIssuer resource has been marked Ready
	if !adsslIssuerHasCondition(iss, api.AdsslIssuerCondition{Type: api.ConditionReady, Status: api.ConditionTrue}) {
		err := fmt.Errorf("resource %s is not ready", issNamespaceName)
		log.Error(err, "failed to retrieve AdsslIssuer resource", "namespace", req.Namespace, "name", cr.Spec.IssuerRef.Name)
		_ = r.setStatus(ctx, cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending, "AdsslIssuer resource %s is not Ready", issNamespaceName)
		return ctrl.Result{}, err
	}

	var secret core.Secret
	secretNamespaceName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      iss.Spec.PasswordRef.Name,
	}
	if err := r.Client.Get(ctx, secretNamespaceName, &secret); err != nil {
		log.Error(err, "failed to retrieve AdsslIssuer secret", "namespace", secretNamespaceName.Namespace, "name", secretNamespaceName.Name)
		return ctrl.Result{}, err
	}

	password, ok := secret.Data[iss.Spec.PasswordRef.Key]
	if !ok {
		err := fmt.Errorf("secret %s does not contain key %s", secret.Name, iss.Spec.PasswordRef.Key)
		log.Error(err, "failed to retrieve AdsslIssuer provisioner secret", "namespace", secretNamespaceName.Namespace, "name", secretNamespaceName.Name)
		return ctrl.Result{}, err
	}

	server := adssl.Server{
		Endpoint: iss.Spec.URL,
		Username: iss.Spec.User,
		Password: string(password),
	}

	cert := adssl.Certificate{
		CertificateRequest: string(cr.Spec.CSRPEM),
	}
	certResult, err := adssl.GenerateCertFromCsr(server, cert)
	if err != nil {
		log.Error(err, "failed to request new cert", "namespace", req.Namespace)
		return ctrl.Result{}, err
	}

	cr.Status.Certificate = []byte(certResult.Result)

	return ctrl.Result{}, r.setStatus(ctx, cr, cmmeta.ConditionTrue, cmapi.CertificateRequestReasonIssued, "Certificate issued")

}

// SetupWithManager initializes the CertificateRequest controller into the
// controller runtime.
func (r *CertificateRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cmapi.CertificateRequest{}).
		Complete(r)
}

// adsslIssuerHasCondition will return true if the given AdsslIssuer resource has
// a condition matching the provided AdsslIssuerCondition. Only the Type and
// Status field will be used in the comparison, meaning that this function will
// return 'true' even if the Reason, Message and LastTransitionTime fields do
// not match.
func adsslIssuerHasCondition(iss api.AdsslIssuer, c api.AdsslIssuerCondition) bool {
	existingConditions := iss.Status.Conditions
	for _, cond := range existingConditions {
		if c.Type == cond.Type && c.Status == cond.Status {
			return true
		}
	}
	return false
}

func (r *CertificateRequestReconciler) setStatus(ctx context.Context, cr *cmapi.CertificateRequest, status cmmeta.ConditionStatus, reason, message string, args ...interface{}) error {
	completeMessage := fmt.Sprintf(message, args...)
	apiutil.SetCertificateRequestCondition(cr, cmapi.CertificateRequestConditionReady, status, reason, completeMessage)

	// Fire an Event to additionally inform users of the change
	eventType := core.EventTypeNormal
	if status == cmmeta.ConditionFalse {
		eventType = core.EventTypeWarning
	}
	r.Recorder.Event(cr, eventType, reason, completeMessage)

	return r.Client.Status().Update(ctx, cr)
}
