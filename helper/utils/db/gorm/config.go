package gorm

import "time"

type Database struct {
	Driver             string   `json:"driver,default=postgres"`
	Host               string   `json:"host"`
	Port               int      `json:"port"`
	DBName             string   `json:"name"`
	Username           string   `json:"username"`
	Password           string   `json:"password"`
	SchemaName         string   `json:"schema-name,optional"`
	TimeZone           string   `json:"time-zone,default=Asia/Ho_Chi_Minh"`
	MaxIdleConnections int      `json:"max-idle-connections,default=20"`
	MaxOpenConnections int      `json:"max-open-connections,default=100"`
	ConnectTimeout     int      `json:"connect-timeout,default=5"`
	ConnMaxLifetime    int      `json:"connection-max-lifetime,default=1200"`
	ConnMaxIdleTime    int      `json:"connection-max-idle-time,default=60"`
	Replicas           []string `json:"replicas,optional"`
	SSLMode            string   `json:"ssl-mode,optional"`
	LogLevel           string   `json:"log-level,optional"`
	LogSlowThreshold   int      `json:"log-slow-threshold,default=3"`
	LogIgnoreNotFound  bool     `json:"log-ignore-not-found,default=false"`
}

func (c *Database) GetDriver() string          { return c.Driver }
func (c *Database) GetHost() string            { return c.Host }
func (c *Database) GetPort() int               { return c.Port }
func (c *Database) GetDBName() string          { return c.DBName }
func (c *Database) GetUsername() string        { return c.Username }
func (c *Database) GetPassword() string        { return c.Password }
func (c *Database) GetSchemaName() string      { return c.SchemaName }
func (c *Database) GetTimeZone() string        { return c.TimeZone }
func (c *Database) GetMaxIdleConnections() int { return c.MaxIdleConnections }
func (c *Database) GetMaxOpenConnections() int { return c.MaxOpenConnections }
func (c *Database) GetConnectTimeout() time.Duration {
	return time.Duration(c.ConnectTimeout) * time.Second
}
func (c *Database) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}
func (c *Database) GetConnMaxIdleTime() time.Duration {
	return time.Duration(c.ConnMaxIdleTime) * time.Second
}
func (c *Database) GetReplicas() []string      { return c.Replicas }
func (c *Database) GetSSLMode() string         { return c.SSLMode }
func (c *Database) GetLogLevel() string        { return c.LogLevel }
func (c *Database) GetLogSlowThreshold() int   { return c.LogSlowThreshold }
func (c *Database) GetLogIgnoreNotFound() bool { return c.LogIgnoreNotFound }
