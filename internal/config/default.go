package config

var defaultConfig = map[string]interface{}{
	"server.port":             9090,
	"server.timeoutSecs":      20,
	"server.readTimeoutSecs":  20,
	"server.writeTimeoutSecs": 40,

	"jwt.secret":      "secret-key",
	"jwt.sessionTime": 864000,

	"db.dataSourceName":   "postgres://common:@localhost:5432/kek?sslmode=disable",
	"db.migrate.enable":   false,
	"db.migrate.dir":      "/migrations",
	"db.pool.maxOpen":     50,
	"db.pool.maxIdle":     5,
	"db.pool.maxLifetime": 5,

	"metrics.namespace": "kek_server",
	"metrics.subsystem": "",
}
