# Kubernetes Notification Operator

The **Kubernetes Notification Operator** is a custom operator that sends notifications to a specified **Messenger** whenever a pod is in chosen status in the monitored namespaces. The operator retrieves the **Messenger Conf, pod statuses and namespaces to watch** from a CRD.

## ✨ Features

- 🚀 **Automatic Notifications** – Sends messages to a enabled and configured messenger when pods are detected.
- 🔧 **Configurable via Custom CRD** – Messengers, pod statues and namespaces to monitor.
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
kubectl apply -f .\config\default
kubectl apply -f .\config\notification-config.yaml #before applying this file, update auth tokens for enabled messengers
```

## 🛠 Configuration

Modify the `config/notification-config.yaml` file to set up your Discord bot:

```yaml
apiVersion: notification.res-x.com/v1alpha1
kind: Notification
metadata:
  name: notification-config
  namespace: operator-system
spec:
  namespaces: ["default", "custom-namespace"]
  discord:
    enabled: true
    discordBotToken: "<YOUR_DISCORD_BOT_TOKEN>"
    discordChannelID: "<YOUR_DISCORD_CHANNEL_ID>"
    message: "Pod {{.Name}} in namespace {{.Namespace}} is {{.Status}}"
    status:
      Running: "running"
      Pending: "pending"
      Succeeded: "succeeded"
      Failed: "failed"
      Unknown: "unknown"
  telegram:
    enabled: true
    telegramBotToken: "<YOUR_TELEGRAM_BOT_TOKEN>"
    telegramChatID: "<YOUR_TELEGRAM_CHAT_ID>"
    message: "Pod {{.Name}} in namespace {{.Namespace}} is {{.Status}}"
    status:
      Running: "running"
      Succeeded: "succeeded"

```

- **`discord`** – Configuration for Discord notifications.
- **`discordBotToken`** – The Discord bot token to use for sending messages.
- **`discordChannelID`** – The Discord channel ID where notifications will be sent.
- **`telegram`** – Configuration for Telegram notifications.
- **`telegramBotToken`** – The Telegram bot token to use for sending messages.
- **`telegramChatID`** – The Telegram chat ID where notifications will be sent.
- **`message`** – The message template to use for notifications.
- **`status`** – Mapping of pod statuses to human-readable strings.
- **`enabled`** – Whether to enable the Discord or Telegram notification feature.
- **`watchNamespaces`** – Comma-separated list of namespaces to monitor for pod creation.

## 🎯 How It Works

1. The operator watches for pods with described in config statuses in the specified namespaces.
2. When a new pod is detected, it sends a message to the configured **Messengers**.
3. The message contains details about the pod, including its name, namespace, and status.

## 🔗 Useful Links

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Discord Developer Portal](https://discord.com/developers/applications)

## 📜 License

This project is licensed under the MIT License.

---

🚀 **Deploy the Kubernetes Notification Operator and stay updated on pod creation events effortlessly!**
