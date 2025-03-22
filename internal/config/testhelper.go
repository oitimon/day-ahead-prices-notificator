package config

import "github.com/shopspring/decimal"

func GenerateTestConfig() *App {
	return &App{
		Analytics: Analytics{
			HighPrice: decimal.NewFromFloat(0.2),
			LowPrice:  decimal.NewFromFloat(0.1),
		},
		DataRepository: Repository{
			Driver: RepositoryDriverGroupCache,
			GroupCache: GroupCache{
				Me:     "http://localhost:9090",
				Listen: "localhost:9090",
			},
			Influx: Influx{
				Url:     "http://localhost:8086",
				Orgname: "test",
				Bucket:  "test",
				Token:   "test",
			},
		},
		Loader: Loader{
			InclBtw: true,
			Driver:  "energyzero",
			API: Api{
				Endpoint: "http://localhost:8080",
			},
		},
		Server: Server{
			Port: "8080",
		},
		Messenger: Messenger{
			Driver: "telegram",
			Telegram: Telegram{
				Token:  "test",
				ChatID: 123,
			},
		},
	}
}
