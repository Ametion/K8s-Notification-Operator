package controller

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	discordBotToken  string
	discordChannelID string
	watchNamespaces  map[string]bool
	lastConfigHash   string
)

func (r *NotificationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.Log.WithValues("pod", req.NamespacedName)

	if req.NamespacedName.Name == "discord-config" && req.NamespacedName.Namespace == "operator-system" {
		log.Info("ConfigMap update detected, reloading configuration...")
		changed, err := reloadConfig(r.Client)
		if err != nil {
			log.Error(err, "Failed to reload ConfigMap")
			return ctrl.Result{}, err
		}

		if !changed {
			log.Info("ConfigMap updated, but no relevant changes detected.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	pod := &corev1.Pod{}
	err := r.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Pod deleted, ignoring")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if pod.Status.Phase != corev1.PodRunning {
		log.Info("Pod is not in Running state yet, ignoring", "pod", pod.Name)
		return ctrl.Result{}, nil
	}

	if !watchNamespaces[pod.Namespace] {
		log.Info("Ignoring pod from untracked namespace", "namespace", pod.Namespace)
		return ctrl.Result{}, nil
	}

	message := fmt.Sprintf("🚀 **New Pod Created**\n**Name:** %s\n**Namespace:** %s", pod.Name, pod.Namespace)
	err = sendDiscordMessage(message)
	if err != nil {
		log.Error(err, "Failed to send Discord message")
	}

	return ctrl.Result{}, nil
}

func sendDiscordMessage(content string) error {
	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", discordChannelID)

	payload := map[string]string{"content": content}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bot "+discordBotToken)
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

func (r *NotificationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Wait until the cache is fully started before reading ConfigMap
	go func() {
		time.Sleep(5 * time.Second) // Small delay to ensure cache is ready
		_, err := reloadConfig(mgr.GetClient())
		if err != nil {
			ctrl.Log.Error(err, "Failed to reload ConfigMap after cache was initialized")
		}
	}()

	return ctrl.NewControllerManagedBy(mgr).
		For(&notification.Notification{}).
		Watches(
			&corev1.Pod{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(),
		).
		Watches(
			&corev1.ConfigMap{},
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(),
		).
		Complete(r)
}

func reloadConfig(k8sClient client.Client) (bool, error) {
	ctx := context.Background()
	configMap := &corev1.ConfigMap{}

	err := k8sClient.Get(ctx, types.NamespacedName{
		Name:      "notification-config",
		Namespace: "k8s-notification-operator-system",
	}, configMap)

	if err != nil {
		return false, fmt.Errorf("failed to load ConfigMap: %w", err)
	}

	newHash := computeConfigHash(configMap.Data)

	if newHash == lastConfigHash {
		return false, nil
	}

	discordBotToken = configMap.Data["discordBotToken"]
	discordChannelID = configMap.Data["discordChannelID"]

	namespacesList := strings.Split(configMap.Data["watchNamespaces"], ",")
	watchNamespaces = make(map[string]bool)
	for _, ns := range namespacesList {
		watchNamespaces[strings.TrimSpace(ns)] = true
	}

	lastConfigHash = newHash
	return true, nil
}

func computeConfigHash(data map[string]string) string {
	hasher := sha256.New()
	for key, value := range data {
		hasher.Write([]byte(key + value))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
