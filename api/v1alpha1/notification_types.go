package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatusMapping defines how Pod phases map to notification messages.
type StatusMapping struct {
	Running   string `json:"Running,omitempty"`
	Pending   string `json:"Pending,omitempty"`
	Succeeded string `json:"Succeeded,omitempty"`
	Failed    string `json:"Failed,omitempty"`
	Unknown   string `json:"Unknown,omitempty"`
}

// DiscordConfig contains configuration for Discord notifications.
type DiscordConfig struct {
	Enabled   bool          `json:"enabled"`
	BotToken  string        `json:"discordBotToken"`
	ChannelID string        `json:"discordChannelID"`
	Message   string        `json:"message"`
	Status    StatusMapping `json:"status"`
}

// TelegramConfig contains configuration for Telegram notifications.
type TelegramConfig struct {
	Enabled  bool          `json:"enabled"`
	BotToken string        `json:"telegramBotToken"`
	ChatID   string        `json:"telegramChatID"`
	Message  string        `json:"message"`
	Status   StatusMapping `json:"status"`
}

// NotificationSpec defines the desired state of Notification.
type NotificationSpec struct {
	Namespaces []string       `json:"namespaces"`
	Discord    DiscordConfig  `json:"discord"`
	Telegram   TelegramConfig `json:"telegram"`
}

// NotificationStatus defines the observed state of Notification.
type NotificationStatus struct {
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Notification is the Schema for the notifications API.
type Notification struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NotificationSpec   `json:"spec,omitempty"`
	Status NotificationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NotificationList contains a list of Notification.
type NotificationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Notification `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Notification{}, &NotificationList{})
}
