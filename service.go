package register

type (
	Service struct {
		Name        string      `mapstructure:"name"`
		Protocol    string      `mapstructure:"protocol"`
		Url         string      `mapstructure:"url"`
		HealthCheck HealthCheck `mapstructure:"healthcheck"`
		Routes      []Route     `mapstructure:"routes"`
	}

	HealthCheck struct {
		Method   string `mapstructure:"method"`
		Path     string `mapstructure:"path"`
		Port     int    `mapstructure:"port"`
		Interval string `mapstructure:"interval"`
		Timeout  string `mapstructure:"timeout"`
		Status   int    `mapstructure:"status"`
	}

	Route struct {
		Name        string   `mapstructure:"name"`
		Rule        string   `mapstructure:"rule"`
		EntryPoints []string `mapstructure:"entry-points"`
		Priority    int64    `mapstructure:"priority"`
		Middlewares []string `mapstructure:"middlewares"`
	}
)
