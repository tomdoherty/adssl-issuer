module github.com/tomdoherty/adssl-issuer

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/jetstack/cert-manager v0.16.1
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/smallstep/step-issuer v0.2.0
	github.com/tomdoherty/adssl v0.0.11
	gonum.org/v1/netlib v0.0.0-20190331212654-76723241ea4e // indirect
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.5
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
	sigs.k8s.io/controller-runtime v0.5.1-0.20200416234307-5377effd4043
	sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06 // indirect
)
