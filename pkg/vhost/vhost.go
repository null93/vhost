package vhost

import (
	"crypto/sha256"
	"encoding/hex"
)

var (
	PATH_NGINX_DIR           = "/etc/nginx"
	PATH_NGINX_AVAILABLE_DIR = "/etc/nginx/sites-available"
	PATH_NGINX_ENABLED_DIR   = "/etc/nginx/sites-enabled"
	PATH_TEMPLATES_DIR       = "/var/lib/vhost/templates"
	PATH_CHECKPOINTS_DIR     = "/var/lib/vhost/checkpoints"
)

func HashData(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))[0:7]
}
