# Kubernetes Notification Operator

The **Kubernetes Notification Operator** is a custom operator that sends notifications to a specified **Discord channel** whenever a pod is created in the monitored namespaces. The operator retrieves the **Discord bot token, channel ID, and namespaces to watch** from a ConfigMap. **You need your own **Discord Bot** added on your own server to use this operator.**

## ✨ Features

- 🚀 **Automatic Notifications** – Sends messages to a Discord channel when new pods are created.
- 🔧 **Configurable via ConfigMap** – Customize the Discord bot token, channel ID, and namespaces to monitor.
- 🔄 **Easy Deployment** – Deploy the operator with simple Kubernetes commands. 


## 📦 Installation

Follow these steps to install and configure the operator inside your Kubernetes cluster.

### 1️⃣ Build and Load the Docker Image

```sh
make docker-build IMG=k8s-notification-operator:latest
kind load docker-image k8s-notification-operator:latest
```

### 2️⃣ Apply the Required Kubernetes Resources

Run the following commands to set up the necessary roles, bindings, namespace, and ConfigMap.

```sh
kubectl apply -f .\config\rbac\cluster_role.yaml
kubectl apply -f .\config\rbac\cluster_role_binding.yaml
kubectl create namespace k8s-notification-operator-system
kubectl apply -f .\config\notification-config.yaml #before applying this file, make sure to update the discordBotToken, discordChannelID, and watchNamespaces
kubectl apply -f .\config\default
```

## 🛠 Configuration

Modify the `config/notification-config.yaml` file to set up your Discord bot:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: notification-config
  namespace: k8s-notification-operator-system
data:
  discordBotToken: "<YOUR_DISCORD_BOT_TOKEN>"
  discordChannelID: "<YOUR_DISCORD_CHANNEL_ID>"
  watchNamespaces: "default,custom-namespace"
```

- **`discordBotToken`** – Your Discord bot token.
- **`discordChannelID`** – The Discord channel ID where notifications will be sent.
- **`watchNamespaces`** – Comma-separated list of namespaces to monitor for pod creation.

## 🎯 How It Works

1. The operator watches for newly created pods in the specified namespaces.
2. When a new pod is detected, it sends a message to the configured **Discord channel**.
3. The message contains details about the new pod, including its name, namespace, and creation time.

## 🔗 Useful Links

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Discord Developer Portal](https://discord.com/developers/applications)

## 📜 License

This project is licensed under the MIT License.

---

🚀 **Deploy the Kubernetes Notification Operator and stay updated on pod creation events effortlessly!**
