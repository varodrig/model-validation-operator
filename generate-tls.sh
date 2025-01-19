#!/bin/bash
cfssl gencert -initca ./tls/ca-csr.json | cfssljson -bare /tmp/ca
cfssl gencert \
  -ca=/tmp/ca.pem \
  -ca-key=/tmp/ca-key.pem \
  -config=./tls/ca-config.json \
  -hostname="model-validation-controller.model-validation-controller.svc.cluster.local,model-validation-controller.model-validation-controller.svc,localhost,127.0.0.1" \
  -profile=default \
  ./tls/ca-csr.json | cfssljson -bare /tmp/example-webhook
cat <<EOF > ./manifests/example-webhook-tls.yaml 
apiVersion: v1
kind: Secret
metadata:
  name: example-webhook-tls
type: Opaque
data:
  tls.crt: $(cat /tmp/example-webhook.pem | base64 | tr -d '\n')
  tls.key: $(cat /tmp/example-webhook-key.pem | base64 | tr -d '\n')
EOF

#generate CA Bundle + inject into template
ca_pem_b64="$(openssl base64 -A <"/tmp/ca.pem")"
sed -e 's@${CA_PEM_B64}@'"$ca_pem_b64"'@g' <"tls/mutation_webhook-template.yaml" \
    > manifests/mutation_webhook.yaml
