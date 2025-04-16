# Sigstore Model-Validation-Operator Roadmap

## Objective

This roadmap outlines the community direction for the Sigstore Model-Validation-Operator from April 2025 through 2026 and beyond. It focuses on short-term deliverables with a long-term vision for secure and flexible AI model validation on Kubernetes.

## Background

### Problem

There is currently no standard method for validating AI/ML models on Kubernetes, whether sourced from volumes, OCI images, or remote APIs. As model deployment becomes increasingly central, ensuring model integrity is critical.

### Current Status

With this repository, a proof of concept exists that injects an `initContainer` into workloads to validate signed models from mounted volumes using the [Sigstore Model Transparency CLI](https://github.com/sigstore/model-transparency/tree/main/src/model_signing).

## Vision

As the project is still in its early stages, we should further evaluate how models are typically consumed ([model-transparency#435](https://github.com/sigstore/model-transparency/issues/435)) — e.g., via volumes, OCI images, remote APIs, or a combination. Based on these findings, we may consider supporting partial verification to accommodate mixed sources ([model-transparency#434](https://github.com/sigstore/model-transparency/issues/434)). In the long term, this effort could be integrated into the Sigstore policy-controller ([model-transparency#436](https://github.com/sigstore/model-transparency/issues/436)) or remain as a standalone operator.
