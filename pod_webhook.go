package main

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	v1alpha1 "github.com/miyunari/model-validation-controller/api/v1alpha1"
)

// NewPodInterceptorWebhook creates a new pod mutating webhook to be registered
func NewPodInterceptorWebhook(c client.Client) webhook.AdmissionHandler {
	return &podInterceptor{
		client: c,
	}
}

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,sideEffects=None,verbs=create;update,versions=v1,name=pods.model-validation.rhtas.redhat.com,admissionReviewVersions=v1

// +kubebuilder:rbac:groups=rhtasv1alpha1,resources=ModelValidation,verbs=get;list;watch
// +kubebuilder:rbac:groups=rhtasv1alpha1,resources=ModelValidation/status,verbs=get;update;patch

// podInterceptor extends pods with Model Validation Init-Container if annotation is specified.
type podInterceptor struct {
	client  client.Client
	decoder admission.Decoder
}

// Handle extends pods with Model Validation Init-Container if annotation is specified.
func (p *podInterceptor) Handle(ctx context.Context, req admission.Request) admission.Response {
	fmt.Println("Execute webhook")
	logger := log.FromContext(ctx)
	pod := &corev1.Pod{}

	if err := p.decoder.Decode(req, pod); err != nil {
		logger.Error(err, "failed to decode pod")
		return admission.Errored(http.StatusBadRequest, err)
	}

	fmt.Println("For pod: ", pod.Name, "in namespace: ", pod.Namespace)

	rhmv := &v1alpha1.ModelValidation{} // NOTE: Search for definition in Namespace.
	if err := p.client.Get(ctx, types.NamespacedName{Name: req.Namespace}, rhmv); err != nil {
		msg := "failed to get the ModelValidation Spec for the pod, skipping injection"
		logger.Error(err, msg)
		return admission.Errored(http.StatusNotFound, err)
	}

	// NOTE: check if validation sidecar is already injected. Then no action needed.
	for _, c := range pod.Spec.InitContainers {
		if c.Name == modelValidationInitContainerName {
			return admission.Allowed("no action needed")
		}
	}

	args := []string{"verify", fmt.Sprintf("--model_path=%s", rhmv.Spec.Model.Path)}
	args = append(args, validationConfigToArgs(rhmv.Spec.Config)...)

	pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
		Name:    modelValidationInitContainerName,
		Image:   "ghcr.io/miyunari/model-transparency-cli:latest", // TODO: get image from operator config.
		Command: args,
	})

	return admission.Allowed("Init container for model validation successfully injected")
}

func validationConfigToArgs(cfg v1alpha1.ValidationConfig) []string {
	res := []string{}
	if cfg.SigstoreConfig != nil {
		res = append(res,
			"sigstore",
			"--identity", cfg.SigstoreConfig.CertificateIdentity,
			"--identity-provider", cfg.SigstoreConfig.CertificateOidcIssuer,
		)
		return res
	}

	if cfg.PrivateKeyConfig != nil {
		res = append(res,
			"pki",
			"--public_key", cfg.PrivateKeyConfig.KeyPath,
		)
		return res
	}

	if cfg.PkiConfig != nil {
		res = append(res,
			"pki",
			"--root_certs", cfg.PkiConfig.CertificateAuthority,
		)
		return res
	}
	return []string{}
}

const modelValidationInitContainerName = "modelValidation"
