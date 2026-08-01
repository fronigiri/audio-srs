package main

type Config struct {
	LibraryPath string
}

func NewConfig() *Config {
	return &Config{
		LibraryPath: "./library",
	}
}

func SaveConfig(cfg *Config) {

}
