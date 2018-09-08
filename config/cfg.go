package config

// C is the package config
var C Config

// Config is the server configurtion
type Config struct {
	Networks map[string]struct {
		Description string
		Type        string
		NetworkId   string
		URL         string
	}

	Accounts map[string]struct {
		Network string
	}

	Alerts map[string]struct {
		Network string
		Logs    bool
		Rule    string
		Message string
	}
}
