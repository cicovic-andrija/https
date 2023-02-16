package https

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type Config struct {
	Network          NetworkConfig    `json:"network"`
	EnableFileServer bool             `json:"enable_file_server"`
	FileServer       FileServerConfig `json:"file_server"`
	LogRequests      bool             `json:"log_requests"`
	LogDirectory     string           `json:"logs_directory"`
}

type NetworkConfig struct {
	IPAcceptHost string `json:"ip_accept_host"`
	TCPPort      int    `json:"tcp_port"`
	TLSCertPath  string `json:"tls_cert_path"`
	TLSKeyPath   string `json:"tls_key_path"`
}

type FileServerConfig struct {
	URLPrefix string `json:"url_prefix"`
	Directory string `json:"directory"`
}

func NewServer(config *Config) (server *HTTPSServer, err error) {
	initError := func(format string, v ...interface{}) error {
		return fmt.Errorf("https init: "+format, v...)
	}

	if config == nil {
		err = initError("empty config")
		return
	}

	var host string
	switch config.Network.IPAcceptHost {
	case "localhost":
		host = "127.0.0.1"
	case "any":
		host = "0.0.0.0"
	default:
		err = initError("invalid IP host descriptor")
		return
	}

	if config.Network.TCPPort < 0 || config.Network.TCPPort > 65535 {
		err = initError("invalid TCP port: %d", config.Network.TCPPort)
		return
	}

	if config.Network.TLSCertPath == "" {
		err = initError("TLS certificate not provided")
		return
	}

	if exists, _ := Exists(config.Network.TLSCertPath); !exists {
		err = initError("file not found: %s", config.Network.TLSCertPath)
		return
	}

	if config.Network.TLSKeyPath == "" {
		err = initError("TLS key not provided")
		return
	}

	if exists, _ := Exists(config.Network.TLSKeyPath); !exists {
		err = initError("file not found: %s", config.Network.TLSKeyPath)
		return
	}

	serveMux := http.NewServeMux()

	if config.EnableFileServer {
		if config.FileServer.Directory == "" {
			err = initError("file server directory not provided")
			return
		}
		if exists, _ := DirectoryExists(config.FileServer.Directory); !exists {
			err = initError("directory not found: %s", config.FileServer.Directory)
			return
		}
		if config.FileServer.URLPrefix == "" {
			err = initError("file server URL prefix not provided")
			return
		}
		fileServer := http.FileServer(http.Dir(config.FileServer.Directory))
		serveMux.Handle(
			config.FileServer.URLPrefix,
			Adapt(
				fileServer,
				StripPrefix(config.FileServer.URLPrefix),
			),
		)
	}

	server = &HTTPSServer{
		httpsImpl: &http.Server{
			Addr:     net.JoinHostPort(host, strconv.Itoa(config.Network.TCPPort)),
			Handler:  serveMux,
			ErrorLog: log.New(io.Discard, "", 0),
		},
		serveMux:    serveMux,
		tlsCertPath: config.Network.TLSCertPath,
		tlsKeyPath:  config.Network.TLSKeyPath,
		started:     false,
		shutdownSem: &sync.WaitGroup{},
	}

	return
}
