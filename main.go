package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-logr/logr"
	admissionv1 "k8s.io/api/admission/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/miyunari/model-validation-controller/api/v1alpha1"
)

func main() {
	if err := v1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
	decoder := admission.NewDecoder(scheme.Scheme)

	c, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme.Scheme})
	if err !=  nil {
		panic(err)
	}

	interceptor := NewPodInterceptorWebhook(c, decoder)

	logger := logr.FromSlogHandler(slog.NewTextHandler(os.Stdout, nil))

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("new request, path: /webhook")
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusInternalServerError)
			return
		}

		// Parse the body into an AdmissionRequest
		var review admissionv1.AdmissionReview
		if err = json.Unmarshal(body, &review); err != nil {
			http.Error(w, fmt.Sprintf("Error unmarshalling admission request: %v", err), http.StatusInternalServerError)
			return
		}

		if review.Request == nil {
			http.Error(w, fmt.Sprintf("Missing request: %v", err), http.StatusBadRequest)
			return
		}

		ctx := logr.NewContext(r.Context(), logger)
		// Handle the request using podInterceptor
		response := interceptor.Handle(ctx, admission.Request{AdmissionRequest: *review.Request})

		review.Response = &response.AdmissionResponse
		patch, err := json.Marshal(response.Patches)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid patch: %v", err), http.StatusInternalServerError)
			return
		}
		review.Response.Patch  = patch
		review.Response.UID = review.Request.UID
		b, err := json.Marshal(review)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid response: %v", err), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(b)
	})

	// Start the server
	logger.Info("Starting webhook server on :8080")
	folder := "/etc/admission-webhook/tls/"
	if err := http.ListenAndServeTLS(":8080", folder+"tls.crt", folder+"tls.key", nil); err != nil {
		logger.Error(err, "failed to start server")
	}
}
