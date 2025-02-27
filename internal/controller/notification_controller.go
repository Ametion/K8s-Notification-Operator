package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	notification "github.com/Ametion/k8s-notification-operator/api/v1alpha1"
)

type NotificationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	config notification.NotificationSpec
)

// Reconcile watches for Pod events and sends notifications if conditions are met.
func (r *NotificationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.Log.WithValues("pod", req.NamespacedName)

	// Load Notification CR configuration
	err := reloadConfig(r.Client)
	if err != nil {
		log.Error(err, "Failed to load Notification configuration")
		return ctrl.Result{}, err
	}

	// Handle Pod Events
	pod := &corev1.Pod{}
	err = r.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Pod deleted, ignoring")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Check if the namespace is watched
	namespaceWatched := false
	for _, ns := range config.Namespaces {
		if ns == pod.Namespace {
			namespaceWatched = true
			break
		}
	}
	if !namespaceWatched {
		log.Info("Ignoring pod from untracked namespace", "namespace", pod.Namespace)
		return ctrl.Result{}, nil
	}

	// Get the status message for the Pod phase
	var status string
	switch pod.Status.Phase {
	case corev1.PodRunning:
		status = config.Discord.Status.Running
	case corev1.PodPending:
		status = config.Discord.Status.Pending
	case corev1.PodSucceeded:
		status = config.Discord.Status.Succeeded
	case corev1.PodFailed:
		status = config.Discord.Status.Failed
	default:
		status = config.Discord.Status.Unknown
	}

	// Create message template data
	messageData := struct {
		Name      string
		Namespace string
		Status    string
	}{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    status,
	}

	// Send notifications based on enabled platforms
	if config.Discord.Enabled {
		discordMessage := renderMessage(config.Discord.Message, messageData)
		err := sendDiscordMessage(discordMessage)
		if err != nil {
			log.Error(err, "Failed to send Discord message")
		}
	}

	if config.Telegram.Enabled {
		telegramMessage := renderMessage(config.Telegram.Message, messageData)
		err := sendTelegramMessage(telegramMessage)
		if err != nil {
			log.Error(err, "Failed to send Telegram message")
		}
	}

	return ctrl.Result{}, nil
}

// sendDiscordMessage sends a message to Discord
func sendDiscordMessage(content string) error {
	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", config.Discord.ChannelID)

	payload := map[string]string{"content": content}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+config.Discord.BotToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to send message to Discord, status code: %d", resp.StatusCode)
	}

	return nil
}

// sendTelegramMessage sends a message to Telegram
func sendTelegramMessage(content string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.Telegram.BotToken)

	payload := map[string]string{
		"chat_id": config.Telegram.ChatID,
		"text":    content,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message to Telegram, status code: %d", resp.StatusCode)
	}

	return nil
}

// SetupWithManager initializes the controller manager
func (r *NotificationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&notification.Notification{}).
		Watches(
			&corev1.Pod{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(),
		).
		Complete(r)
}

// reloadConfig loads the Notification CRD configuration
func reloadConfig(k8sClient client.Client) error {
	ctx := context.Background()
	notificationResource := &notification.Notification{}

	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      "notification-config",
		Namespace: "k8s-notification-operator-system",
	}, notificationResource)

	if err != nil {
		return fmt.Errorf("failed to load Notification CR: %w", err)
	}

	config = notificationResource.Spec
	return nil
}

// renderMessage parses a message template
func renderMessage(tmpl string, data interface{}) string {
	t, err := template.New("msg").Parse(tmpl)
	if err != nil {
		return "Failed to generate message"
	}
	var msg bytes.Buffer
	err = t.Execute(&msg, data)
	if err != nil {
		return "Failed to generate message"
	}
	return msg.String()
}
