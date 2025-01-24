# Model Validation Controller

This project is a proof of concept based on the [sigstore/model-transperency-cli](https://github.com/sigstore/model-transparency).
It offers a Kubernetes/OpenShift controller designed to validate AI models before they are picked up by actual
workload. This project provides a webhook that adds an initcontainer to perform model validation. 
The controller uses a custom resource to define how the models should be validated, such as utilizing [Sigstore](https://www.sigstore.dev/) or public keys.

### Features

- Model Validation: Ensures AI models are validated before they are used by workloads.
- Webhook Integration: A webhook automatically injects an initcontainer into pods to perform the validation step.
- Custom Resource: Configurable `ModelValidation` custom resource to specify how models should be validated. 
    - Supports methods like [Sigstore](https://www.sigstore.dev/), pki or public key validation.

### Prerequisites

- Kubernetes 1.29+ or OpenShift 4.16+
- Proper configuration for model validation (e.g., Sigstore, public keys)
- A singed model (e.g. check the `testdata` or `examples` folder)

### Installation

The controller can be installed in the `model-validation-controller` namespace via [kustomize](https://kustomize.io/).
```bash
kubectl apply -k https://raw.githubusercontent.com/miyunari/model-validation-controller/main/manifests
# or local
kubectl apply -k manifests
```

Run delete to uninstall the controller.
```bash
kubectl delete -k https://raw.githubusercontent.com/miyunari/model-validation-controller/main/manifests
# or local
kubectl delete -k manifests
```
### Known limitations

The project is at an early stage and therefore has some limitations.

- There is no validation or defaulting for the custom resource.
- The validation is namespace scoped and cannot be used across multiple namespaces.
- No more than one validation resource can be used per namespace.
- There are no status fields for the custom resource.
- The model and signature path must be specified, there is no auto discovery.
- TLS certificates used by the webhook are self generated.

### Usage

First, a ModelValidation CR must be created as follows:
```yaml
apiVersion: rhtas.redhat.com/v1alpha1
kind: ModelValidation
metadata:
  name: demo
spec:
  config:
    sigstoreConfig:
      certificateIdentity: "nolear@redhat.com"
      certificateOidcIssuer: "https://token.actions.githubusercontent.com"
  model:
    path: /data/tensorflow_saved_model
    signaturePath: /data/tensorflow_saved_model/model.sig
```

All pods in the namespace where the custom resource exists that have this label `validation.rhtas.redhat.com/ml: "true"` will be validated.
It should be noted that this does not apply to subsequently labeled pods.

```diff
apiVersion: v1
kind: Pod
metadata:
  name: whatever-workload
+  labels:
+    validation.rhtas.redhat.com/ml: "true"
spec:
  restartPolicy: Never
  containers:
  - name: whatever-workload
    image: nginx
    ports:
    - containerPort: 80
    volumeMounts:
    - name: model-storage
      mountPath: /data
  volumes:
  - name: model-storage
    persistentVolumeClaim:
      claimName: models
```

### Examples

The example folder contains two files `prepare.yaml` and `singed.yaml`.

- prepare: contains a persistent volume claim and a job that downloads a signed test model.
```bash
kubectl apply -f https://raw.githubusercontent.com/miyunari/model-validation-controller/main/examples/prepare.yaml
# or local
kubectl apply -f examples/prepare.yaml
```
- signed: contains a model validation manifest for the validation of this model and a demo pod, which is provided with the appropriate label for validation.
```bash
kubectl apply -f https://raw.githubusercontent.com/miyunari/model-validation-controller/main/examples/verify.yaml
# or local
kubectl apply -f examples/verify.yaml
```

After the example installation, the logs of the generated job should show a successful download:
```bash
kubectl logs -n testing job/download-extract-model 
```

The controller logs should show that a pod has been modified:
```bash
kubectl logs -n model-validation-controller deploy/model-validation-controller
```

Finally, the test pod should be running and the injected initcontainer should have been successfully validated.
```bash
kubectl logs -n testing whatever-workload model-validation
```