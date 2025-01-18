package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/miyunari/model-validation-controller/api/v1alpha1"
)

func main() {
	interceptor := &podInterceptor{}
	scheme, err := v1alpha1.SchemeBuilder.Build()
	if err != nil {
		panic(err)
	}
	decoder := admission.NewDecoder(scheme)
	interceptor.decoder = decoder

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		log.Println("new request, path: /webhook")
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
		var request admission.Request
		if err := json.Unmarshal(body, &request); err != nil {
			http.Error(w, fmt.Sprintf("Error unmarshalling admission request: %v", err), http.StatusInternalServerError)
			return
		}

		// Handle the request using podInterceptor
		response := interceptor.Handle(r.Context(), request)
		w.WriteHeader(int(response.Result.Code))
		_, _ = w.Write([]byte(response.Result.Message))
	})

	// Start the server
	log.Println("Starting webhook server on :8080")
	folder := "/etc/admission-webhook/tls/"
	if err := http.ListenAndServeTLS(":8080", folder+"tls.crt", folder+"tls.key", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
