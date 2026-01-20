package notifier

import (
	"dhtc/config"
	"sync"
)

type Notifier interface {
	Notify(message string) error
}

type Manager struct {
	notifiers []Notifier
	mu        sync.RWMutex
}

func (m *Manager) Notify(message string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, n := range m.notifiers {
		_ = n.Notify(message)
	}
}

func (m *Manager) Setup(cfg *config.Configuration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifiers = nil

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
}

func SetupNotifiers(cfg *config.Configuration) *Manager {
	m := &Manager{}
	m.Setup(cfg)
	return m
}
