# Distributed Auth Operator

A custom Kubernetes Operator that manages a distributed authentication and authorization cluster, inspired by blockchain-style consistency and Raft consensus.

This project uses a Custom Resource Definition (CRD) called `AuthCluster` to automate the creation and lifecycle of a secure, peer-aware, fault-tolerant node cluster.

---

## 🚀 Features

- **CRD: `AuthCluster`**
  - Defines cluster size and Redis configuration
- **Operator-powered automation**
  - Dynamically creates Pods, PVCs, and ConfigMaps
  - Injects peer lists and config into containers
- **Distributed node management**
  - Supports Raft-style consensus and internal peer discovery
- **Works with Docker Desktop or remote K8s clusters**

---

## 🧱 Architecture

Each `AuthCluster` CR results in:
- `N` Pods named `auth-node-0`, `auth-node-1`, ...
- One shared `ConfigMap` with peer addresses
- One `PersistentVolumeClaim` per node for LevelDB storage


---

## ⚙️ Local Development

### Prerequisites
- Go 1.20+
- Docker + Docker Desktop with Kubernetes enabled
- [`kubebuilder`](https://book.kubebuilder.io/quick-start.html)

### Running Locally
```bash
make install      # install CRDs
make run          # run controller locally
kubectl apply -f kubectl-apply/redis-deployment.yaml
kubectl apply -f kubectl-apply/authcluster-demo.yaml
```

Check resources:
```bash
kubectl get pods
kubectl get pvc
kubectl get configmaps
```

---

## 🐳 Docker Compose Support

The project also supports local Docker Compose testing:
- Build and run 3 auth nodes and a client
- Each node shares the same image with unique `NODE_ID`

```bash
docker-compose up --build
```

---

## 🔐 Tech Stack
- Kubernetes Operator SDK (`kubebuilder`)
- Raft consensus for node coordination
- Redis for caching and pub/sub
- LevelDB for persistent node storage
- gRPC for peer-to-peer communication

---

## 📁 Project Structure

```
api/                   # CRD definition (AuthCluster)
controllers/           # Reconcile logic
config/                # Deployment YAMLs
Dockerfile             # Operator image
Makefile               # Dev workflows
main.go                # Operator entrypoint
```

---

## ✨ Credits
Built with ❤️ by [@GooseFuse](https://github.com/GooseFuse)

Inspired by:
- Operator SDK
- HashiCorp Raft
- Blockchain-inspired consensus designs

---

## 📜 License
MIT

