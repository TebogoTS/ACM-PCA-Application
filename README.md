### ACM PCA Application**

#### **Overview**
This application is a Go-based HTTP server that interacts with AWS ACM Private Certificate Authority (ACM PCA). It provides two API endpoints:

1. `/generate-csr`: Generates a Certificate Signing Request (CSR).
2. `/issue-certificate`: Uses the generated CSR to request a certificate from ACM PCA.

The application is containerized for deployment in Kubernetes and integrates with AWS services using IAM roles and Kubernetes service accounts.

---

### **Features**
- Generate CSRs dynamically.
- Issue certificates using AWS ACM PCA.
- Deployable on Amazon EKS with IAM-based authentication.

---

### **Deployment Instructions**

#### **1. Pre-requisites**
- An AWS account with ACM PCA set up.
- An EKS cluster with OIDC provider configured.
- AWS CLI and `kubectl` installed locally.
- Docker for containerizing the application.

---

#### **2. Setup Values to Substitute**

Before deploying, replace the placeholders in the Kubernetes YAML files and IAM role with your values:

| Placeholder                     | Description                                           |
|---------------------------------|-------------------------------------------------------|
| `your-region`                   | AWS region where ACM PCA is located.                 |
| `your-account-id`               | Your AWS account ID.                                 |
| `arn:aws:acm-pca:region:...`    | ARN of the ACM PCA to be used for issuing certificates. |
| `your-repo/go-acmpca-app:latest`| Docker image repository and tag for the application. |
| `acmpca-app-role`               | Name of the IAM role for the service account.        |

---

#### **3. Building and Deploying the Application**

##### **Build and Push the Docker Image**
```bash
# Build the Docker image
docker build -t your-repo/go-acmpca-app:latest .

# Push the image to a container registry (Docker Hub or Amazon ECR)
docker push your-repo/go-acmpca-app:latest
```

##### **Apply Kubernetes Resources**
1. **Service Account**:
   Replace `arn:aws:iam::your-account-id:role/acmpca-app-role` in `k8s-sa.yaml` with your IAM role ARN.

   Apply the Service Account:
   ```bash
   kubectl apply -f k8s-sa.yaml
   ```

2. **Deployment**:
   Replace `your-repo/go-acmpca-app:latest` in `8s-deploy.yaml` with your Docker image.

   Apply the Deployment:
   ```bash
   kubectl apply -f k8s-deploy.yaml
   ```

3. **Service**:
   Apply the Service to expose the application:
   ```bash
   kubectl apply -f k8s-svc.yaml
   ```

---

#### **4. Testing the Application**

1. Retrieve the external IP or hostname of the LoadBalancer:
   ```bash
   kubectl port-forward svc/go-acmpca-app 8080:8080
   ```

2. Test the endpoints:
   - Generate a CSR:
     ```bash
     curl http://localhost:8080/generate-csr
     ```
   - Issue a certificate:
     ```bash
     curl http://localhost:8080/issue-certificate
     ```

---

### **IAM Permissions**

Ensure the IAM role associated with the Kubernetes Service Account has the following policy:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "acm-pca:IssueCertificate",
        "acm-pca:GetCertificate",
        "acm-pca:DescribeCertificateAuthority"
      ],
      "Resource": "arn:aws:acm-pca:region:account-id:certificate-authority/CA-ID"
    }
  ]
}
```

---

### **Cleanup**

To remove all resources:
```bash
kubectl delete -f k8s-svc.yaml
kubectl delete -f k8s-deploy.yaml
kubectl delete -f k8s-sa.yaml
```
