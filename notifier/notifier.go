package notifier

import (
	"dhtc/config"
)

type Notifier interface {
	Notify(message string) error
}

type Manager struct {
	notifiers []Notifier
}

func (m *Manager) Notify(message string) {
	for _, n := range m.notifiers {
		_ = n.Notify(message)
	}
}

func SetupNotifiers(cfg *config.Configuration) *Manager {
	m := &Manager{}

	if cfg.TelegramToken != "" {
		bot := SetupTelegramBot(cfg)
		if bot != nil {
			m.notifiers = append(m.notifiers, &TelegramNotifier{
				config: cfg,
				bot:    bot,
			})
		}
	}

	if cfg.DiscordWebhook != "" {
		m.notifiers = append(m.notifiers, &DiscordNotifier{
			WebhookURL: cfg.DiscordWebhook,
		})
	}

	if cfg.SlackWebhook != "" {
		m.notifiers = append(m.notifiers, &SlackNotifier{
			WebhookURL: cfg.SlackWebhook,
		})
	}

	if cfg.GotifyURL != "" && cfg.GotifyToken != "" {
		m.notifiers = append(m.notifiers, &GotifyNotifier{
			URL:   cfg.GotifyURL,
			Token: cfg.GotifyToken,
		})
	}

	return m
}
