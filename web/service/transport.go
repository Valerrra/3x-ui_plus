package service

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/pelletier/go-toml/v2"
)

type ManagedTransportStatus struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
	State       string `json:"state"`
	Active      bool   `json:"active"`
	ConfigPath  string `json:"configPath"`
	BinaryPath  string `json:"binaryPath"`
}

type ManagedTransportConfig struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Path        string `json:"path"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

type ManagedTransportService struct{}

type TrustTunnelServiceConfig struct {
	ListenHost      string `json:"listenHost"`
	ListenPort      int    `json:"listenPort"`
	CredentialsFile string `json:"credentialsFile"`
	Hostname        string `json:"hostname"`
	CertChainPath   string `json:"certChainPath"`
	PrivateKeyPath  string `json:"privateKeyPath"`
	PublicAddress   string `json:"publicAddress"`
}

type MTProtoServiceConfig struct {
	BindHost       string `json:"bindHost"`
	Port           int    `json:"port"`
	Secret         string `json:"secret"`
	ConfigPath     string `json:"configPath"`
	FrontingDomain string `json:"frontingDomain"`
}

type TrustTunnelClient struct {
	Username string `json:"username" toml:"username"`
	Password string `json:"password" toml:"password"`
}

type TrustTunnelCertificatePaths struct {
	CertChainPath  string `json:"certChainPath"`
	PrivateKeyPath string `json:"privateKeyPath"`
	Source         string `json:"source"`
}

type TrustTunnelExportConfig struct {
	Hostname           string   `toml:"hostname"`
	Addresses          []string `toml:"addresses"`
	CustomSNI          string   `toml:"custom_sni"`
	HasIPv6            bool     `toml:"has_ipv6"`
	Username           string   `toml:"username"`
	Password           string   `toml:"password"`
	ClientRandomPrefix string   `toml:"client_random_prefix"`
	SkipVerification   bool     `toml:"skip_verification"`
	Certificate        string   `toml:"certificate"`
	UpstreamProtocol   string   `toml:"upstream_protocol"`
	AntiDPI            bool     `toml:"anti_dpi"`
}

func (s *ManagedTransportService) ListStatuses() []ManagedTransportStatus {
	items := []ManagedTransportStatus{
		{
			Key:         "trusttunnel",
			Title:       "TrustTunnel Core",
			Description: "Endpoint service and core tunnel runtime.",
			Unit:        "trusttunnel",
			ConfigPath:  "/opt/trusttunnel/vpn.toml",
			BinaryPath:  "/usr/local/bin/trusttunnel_endpoint",
		},
		{
			Key:         "trusttunnel-webui",
			Title:       "TrustTunnel WebUI",
			Description: "Panel service for TrustTunnel server management.",
			Unit:        "trusttunnel-webui",
			ConfigPath:  "/opt/trusttunnel-suite/apps/webui/deploy/systemd/trusttunnel-webui.env",
			BinaryPath:  "/opt/trusttunnel-suite/apps/webui/build/trusttunnel-webui",
		},
		{
			Key:         "trusttunnel-mtproto",
			Title:       "MTProto Access",
			Description: "Telegram MTProto access bridge managed alongside TrustTunnel.",
			Unit:        "trusttunnel-mtproto",
			ConfigPath:  "/opt/trusttunnel/access/mtproto.toml",
			BinaryPath:  "/usr/local/bin/mtg",
		},
	}

	for i := range items {
		items[i].State = s.serviceState(items[i].Unit)
		items[i].Active = items[i].State == "active"
	}

	return items
}

func (s *ManagedTransportService) RunAction(key, action string) error {
	unit, ok := map[string]string{
		"trusttunnel":         "trusttunnel",
		"trusttunnel-webui":   "trusttunnel-webui",
		"trusttunnel-mtproto": "trusttunnel-mtproto",
	}[key]
	if !ok {
		return fmt.Errorf("unknown service key: %s", key)
	}

	switch action {
	case "start", "stop", "restart":
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}

	cmd := exec.Command("systemctl", action, unit)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl %s %s failed: %s", action, unit, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func (s *ManagedTransportService) ListConfigs() []ManagedTransportConfig {
	return []ManagedTransportConfig{
		{
			Key:         "trusttunnel-vpn",
			Title:       "TrustTunnel vpn.toml",
			Path:        "/opt/trusttunnel/vpn.toml",
			Description: "Основной runtime-конфиг TrustTunnel endpoint.",
		},
		{
			Key:         "trusttunnel-hosts",
			Title:       "TrustTunnel hosts.toml",
			Path:        "/opt/trusttunnel/hosts.toml",
			Description: "TLS hostnames, certificate paths и endpoint host mapping.",
		},
		{
			Key:         "trusttunnel-webui-env",
			Title:       "TrustTunnel WebUI env",
			Path:        "/opt/trusttunnel-suite/apps/webui/deploy/systemd/trusttunnel-webui.env",
			Description: "Переменные окружения WebUI deployment.",
		},
		{
			Key:         "mtproto-service",
			Title:       "MTProto systemd unit",
			Path:        "/etc/systemd/system/trusttunnel-mtproto.service",
			Description: "Systemd unit для MTProto bridge.",
		},
	}
}

func (s *ManagedTransportService) GetConfig(key string) (*ManagedTransportConfig, error) {
	for _, item := range s.ListConfigs() {
		if item.Key != key {
			continue
		}
		raw, err := os.ReadFile(item.Path)
		if err != nil {
			return nil, err
		}
		item.Content = string(raw)
		return &item, nil
	}
	return nil, fmt.Errorf("unknown config key: %s", key)
}

func (s *ManagedTransportService) SaveConfig(key, content string) error {
	item, err := s.GetConfigMeta(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(item.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(item.Path, []byte(content), 0o644)
}

func (s *ManagedTransportService) GetConfigMeta(key string) (*ManagedTransportConfig, error) {
	for _, item := range s.ListConfigs() {
		if item.Key == key {
			copy := item
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("unknown config key: %s", key)
}

func (s *ManagedTransportService) GetTrustTunnelConfig() (*TrustTunnelServiceConfig, error) {
	type vpnConfig struct {
		ListenAddress   string `toml:"listen_address"`
		CredentialsFile string `toml:"credentials_file"`
	}
	type hostEntry struct {
		Hostname       string `toml:"hostname"`
		CertChainPath  string `toml:"cert_chain_path"`
		PrivateKeyPath string `toml:"private_key_path"`
	}
	type hostsConfig struct {
		MainHosts []hostEntry `toml:"main_hosts"`
	}

	cfg := &TrustTunnelServiceConfig{
		ListenHost: "0.0.0.0",
		ListenPort: 443,
	}

	rawVPN, err := os.ReadFile("/opt/trusttunnel/vpn.toml")
	if err != nil {
		return nil, err
	}
	var vpn vpnConfig
	if err := toml.Unmarshal(rawVPN, &vpn); err != nil {
		return nil, fmt.Errorf("parse vpn.toml: %w", err)
	}
	cfg.CredentialsFile = strings.TrimSpace(vpn.CredentialsFile)
	if host, port := splitHostPort(vpn.ListenAddress, "0.0.0.0", 443); port > 0 {
		cfg.ListenHost = host
		cfg.ListenPort = port
	}

	rawHosts, err := os.ReadFile("/opt/trusttunnel/hosts.toml")
	if err != nil {
		return nil, err
	}
	var hosts hostsConfig
	if err := toml.Unmarshal(rawHosts, &hosts); err != nil {
		return nil, fmt.Errorf("parse hosts.toml: %w", err)
	}
	if len(hosts.MainHosts) > 0 {
		cfg.Hostname = strings.TrimSpace(hosts.MainHosts[0].Hostname)
		cfg.CertChainPath = strings.TrimSpace(hosts.MainHosts[0].CertChainPath)
		cfg.PrivateKeyPath = strings.TrimSpace(hosts.MainHosts[0].PrivateKeyPath)
	}

	envMap, err := readEnvFile("/opt/trusttunnel-suite/apps/webui/deploy/systemd/trusttunnel-webui.env")
	if err == nil {
		cfg.PublicAddress = strings.TrimSpace(envMap["TT_WEBUI_ENDPOINT_PUBLIC_ADDRESS"])
	}

	return cfg, nil
}

func (s *ManagedTransportService) SaveTrustTunnelConfig(cfg *TrustTunnelServiceConfig) error {
	if err := s.writeTrustTunnelRuntimeFiles(cfg); err != nil {
		return err
	}
	return s.updateTrustTunnelWebUIEnv(cfg)
}

func (s *ManagedTransportService) GetMTProtoConfig() (*MTProtoServiceConfig, error) {
	type mtprotoConfig struct {
		Secret string `toml:"secret"`
		BindTo string `toml:"bind-to"`
	}

	cfg := &MTProtoServiceConfig{
		BindHost:   "0.0.0.0",
		Port:       2443,
		ConfigPath: "/opt/trusttunnel/access/mtproto.toml",
	}

	raw, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		return nil, err
	}
	var parsed mtprotoConfig
	if err := toml.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse mtproto config: %w", err)
	}
	cfg.Secret = strings.TrimSpace(parsed.Secret)
	cfg.BindHost, cfg.Port = splitHostPort(parsed.BindTo, "0.0.0.0", 2443)
	return cfg, nil
}

func (s *ManagedTransportService) SaveMTProtoConfig(cfg *MTProtoServiceConfig) error {
	type networkConfig struct {
		DNS string `toml:"dns"`
	}
	type mtprotoConfig struct {
		Secret  string        `toml:"secret"`
		BindTo  string        `toml:"bind-to"`
		Network networkConfig `toml:"network"`
	}

	bindHost := strings.TrimSpace(cfg.BindHost)
	if bindHost == "" {
		bindHost = "0.0.0.0"
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("invalid MTProto port")
	}
	secret := strings.TrimSpace(cfg.Secret)
	if secret == "" {
		generatedSecret, err := s.GenerateMTProtoSecret(cfg.FrontingDomain)
		if err != nil {
			return err
		}
		secret = generatedSecret
		cfg.Secret = secret
	}
	configPath := strings.TrimSpace(cfg.ConfigPath)
	if configPath == "" {
		configPath = "/opt/trusttunnel/access/mtproto.toml"
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	body, err := toml.Marshal(mtprotoConfig{
		Secret: secret,
		BindTo: fmt.Sprintf("%s:%d", bindHost, cfg.Port),
		Network: networkConfig{
			DNS: "https://1.1.1.1",
		},
	})
	if err != nil {
		return fmt.Errorf("encode mtproto config: %w", err)
	}
	return os.WriteFile(configPath, body, 0o600)
}

func (s *ManagedTransportService) GenerateMTProtoSecret(frontingDomain string) (string, error) {
	domain := strings.TrimSpace(frontingDomain)
	if domain == "" {
		domain = "www.cloudflare.com"
	}

	mtgPath := "/usr/local/bin/mtg"
	if _, err := os.Stat(mtgPath); err == nil {
		cmd := exec.Command(mtgPath, "generate-secret", domain)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil {
			secret := strings.TrimSpace(stdout.String())
			if secret != "" {
				return secret, nil
			}
		}
	}

	randomPart := make([]byte, 16)
	if _, err := rand.Read(randomPart); err != nil {
		return "", fmt.Errorf("generate MTProto secret: %w", err)
	}

	return "ee" + hex.EncodeToString(randomPart) + hex.EncodeToString([]byte(domain)), nil
}

func (s *ManagedTransportService) ApplyTrustTunnelInbound(inbound *model.Inbound) error {
	if inbound == nil {
		return fmt.Errorf("inbound is nil")
	}

	settings := struct {
		Hostname        string `json:"hostname"`
		PublicAddress   string `json:"publicAddress"`
		CredentialsFile string `json:"credentialsFile"`
		CertChainPath   string `json:"certChainPath"`
		PrivateKeyPath  string `json:"privateKeyPath"`
	}{}
	if strings.TrimSpace(inbound.Settings) != "" {
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			return fmt.Errorf("parse TrustTunnel inbound settings: %w", err)
		}
	}

	listenHost := strings.TrimSpace(inbound.Listen)
	if listenHost == "" {
		listenHost = "0.0.0.0"
	}

	cfg := &TrustTunnelServiceConfig{
		ListenHost:      listenHost,
		ListenPort:      inbound.Port,
		CredentialsFile: defaultString(strings.TrimSpace(settings.CredentialsFile), "/opt/trusttunnel/credentials.toml"),
		Hostname:        strings.TrimSpace(settings.Hostname),
		CertChainPath:   strings.TrimSpace(settings.CertChainPath),
		PrivateKeyPath:  strings.TrimSpace(settings.PrivateKeyPath),
		PublicAddress:   strings.TrimSpace(settings.PublicAddress),
	}
	if err := s.writeTrustTunnelRuntimeFiles(cfg); err != nil {
		return err
	}
	if err := s.updateTrustTunnelWebUIEnv(cfg); err != nil {
		return err
	}

	clients, err := trustTunnelClientsFromInbound(inbound)
	if err != nil {
		return err
	}
	if err := s.saveTrustTunnelClientsFromConfig(cfg.CredentialsFile, clients, false); err != nil {
		return err
	}

	if s.serviceState("trusttunnel") != "not-installed" {
		if err := s.RunAction("trusttunnel", "restart"); err != nil {
			return err
		}
	}

	return nil
}

func (s *ManagedTransportService) writeTrustTunnelRuntimeFiles(cfg *TrustTunnelServiceConfig) error {
	listenHost := strings.TrimSpace(cfg.ListenHost)
	if listenHost == "" {
		listenHost = "0.0.0.0"
	}
	if cfg.ListenPort < 1 || cfg.ListenPort > 65535 {
		return fmt.Errorf("invalid TrustTunnel listen port")
	}
	hostname := strings.TrimSpace(cfg.Hostname)
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	certChainPath := strings.TrimSpace(cfg.CertChainPath)
	if certChainPath == "" {
		return fmt.Errorf("certificate chain path is required")
	}
	privateKeyPath := strings.TrimSpace(cfg.PrivateKeyPath)
	if privateKeyPath == "" {
		return fmt.Errorf("private key path is required")
	}
	credentialsFile := defaultString(strings.TrimSpace(cfg.CredentialsFile), "/opt/trusttunnel/credentials.toml")
	if err := os.MkdirAll("/opt/trusttunnel", 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(credentialsFile), 0o755); err != nil {
		return err
	}
	vpnBody := fmt.Sprintf(`# The address to listen on
listen_address = %q

# The path to a TOML file with endpoint users.
credentials_file = %q

# The path to a TOML file for connection filtering rules.
rules_file = "/opt/trusttunnel/rules.toml"

# Whether IPv6 connections can be routed or rejected with unreachable status
ipv6_available = true

# Whether connections to private network of the endpoint are allowed
allow_private_network_connections = false

# Timeout of an incoming TLS handshake. In seconds.
tls_handshake_timeout_secs = 10

# Timeout of a client listener. In seconds.
client_listener_timeout_secs = 600

# Timeout of outgoing connection establishment.
connection_establishment_timeout_secs = 30

# Idle timeout of tunneled TCP connections. In seconds.
tcp_connections_timeout_secs = 604800

# Timeout of tunneled UDP "connections". In seconds.
udp_connections_timeout_secs = 300

# Whether speedtest is available on the main hosts.
speedtest_enable = false

# Optional path prefix for speedtest requests on main hosts.
speedtest_path = "/speedtest"

# Whether ping is available on the main hosts.
ping_enable = false

# Optional path prefix for ping requests on main hosts.
ping_path = "/ping"

# HTTP status code returned on authentication failure.
auth_failure_status_code = 407

[forward_protocol]
[forward_protocol.direct]

[listen_protocols]

[listen_protocols.http1]
upload_buffer_size = 32768

[listen_protocols.http2]
initial_connection_window_size = 8388608
initial_stream_window_size = 131072
max_concurrent_streams = 1000
max_frame_size = 16384
header_table_size = 65536

[listen_protocols.quic]
recv_udp_payload_size = 1350
send_udp_payload_size = 1350
initial_max_data = 104857600
initial_max_stream_data_bidi_local = 1048576
initial_max_stream_data_bidi_remote = 1048576
initial_max_stream_data_uni = 1048576
initial_max_streams_bidi = 4096
initial_max_streams_uni = 4096
max_connection_window = 25165824
max_stream_window = 16777216
disable_active_migration = true
enable_early_data = true
message_queue_capacity = 4096
`, fmt.Sprintf("%s:%d", listenHost, cfg.ListenPort), credentialsFile)
	if err := os.WriteFile("/opt/trusttunnel/vpn.toml", []byte(vpnBody), 0o644); err != nil {
		return err
	}

	hostsBody := fmt.Sprintf(`[[main_hosts]]
hostname = %q
cert_chain_path = %q
private_key_path = %q
`, hostname, certChainPath, privateKeyPath)
	if err := os.WriteFile("/opt/trusttunnel/hosts.toml", []byte(hostsBody), 0o644); err != nil {
		return err
	}

	return nil
}

func (s *ManagedTransportService) DisableMTProto() error {
	if s.serviceState("trusttunnel-mtproto") == "not-installed" {
		return nil
	}
	return s.RunAction("trusttunnel-mtproto", "stop")
}

func (s *ManagedTransportService) ListTrustTunnelClients() ([]TrustTunnelClient, error) {
	type credentialsConfig struct {
		Clients []TrustTunnelClient `toml:"client"`
	}

	ttCfg, err := s.GetTrustTunnelConfig()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(ttCfg.CredentialsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []TrustTunnelClient{}, nil
		}
		return nil, err
	}
	var cfg credentialsConfig
	if err := toml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse credentials file: %w", err)
	}
	return cfg.Clients, nil
}

func (s *ManagedTransportService) AddTrustTunnelClient(username, password string) (*TrustTunnelClient, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		username = "tt-" + randomAlphaNum(8)
	}
	password = strings.TrimSpace(password)
	if password == "" {
		password = randomAlphaNum(18)
	}

	clients, err := s.ListTrustTunnelClients()
	if err != nil {
		return nil, err
	}
	for _, item := range clients {
		if strings.EqualFold(item.Username, username) {
			return nil, fmt.Errorf("client already exists: %s", username)
		}
	}
	clients = append(clients, TrustTunnelClient{
		Username: username,
		Password: password,
	})
	if err := s.saveTrustTunnelClients(clients); err != nil {
		return nil, err
	}
	client := &TrustTunnelClient{Username: username, Password: password}
	return client, nil
}

func (s *ManagedTransportService) DeleteTrustTunnelClient(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username is required")
	}
	clients, err := s.ListTrustTunnelClients()
	if err != nil {
		return err
	}
	filtered := make([]TrustTunnelClient, 0, len(clients))
	found := false
	for _, item := range clients {
		if strings.EqualFold(item.Username, username) {
			found = true
			continue
		}
		filtered = append(filtered, item)
	}
	if !found {
		return fmt.Errorf("client not found: %s", username)
	}
	return s.saveTrustTunnelClients(filtered)
}

func (s *ManagedTransportService) ExportTrustTunnelClient(username string) (string, error) {
	return s.ExportTrustTunnelClientWithOptions(username, "http2", false)
}

func (s *ManagedTransportService) ExportTrustTunnelClientWithOptions(username, upstreamProtocol string, antiDPI bool) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", fmt.Errorf("username is required")
	}
	cfg, err := s.GetTrustTunnelConfig()
	if err != nil {
		return "", err
	}
	publicAddress := strings.TrimSpace(cfg.PublicAddress)
	if publicAddress == "" {
		publicAddress = cfg.Hostname
	}
	if publicAddress == "" {
		return "", fmt.Errorf("TrustTunnel public address is required")
	}
	cmd := exec.Command(
		s.trustTunnelEndpointBinary(),
		"/opt/trusttunnel/vpn.toml",
		"/opt/trusttunnel/hosts.toml",
		"-c", username,
		"-a", publicAddress,
		"--format", "toml",
	)
	cmd.Dir = "/opt/trusttunnel"
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("trusttunnel export failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	var parsed TrustTunnelExportConfig
	if err := toml.Unmarshal(out, &parsed); err != nil {
		return "", fmt.Errorf("parse trusttunnel export: %w", err)
	}
	if strings.TrimSpace(upstreamProtocol) != "" {
		parsed.UpstreamProtocol = strings.TrimSpace(upstreamProtocol)
	}
	parsed.AntiDPI = antiDPI
	return encodeTrustTunnelDeepLink(parsed)
}

func (s *ManagedTransportService) trustTunnelEndpointBinary() string {
	for _, candidate := range []string{
		"/usr/local/bin/trusttunnel_endpoint",
		"/opt/trusttunnel/trusttunnel_endpoint",
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "/usr/local/bin/trusttunnel_endpoint"
}

func (s *ManagedTransportService) DetectTrustTunnelCertificatePaths(hostname string) (*TrustTunnelCertificatePaths, error) {
	trimmed := strings.TrimSpace(hostname)
	candidates := []TrustTunnelCertificatePaths{}
	if trimmed != "" {
		candidates = append(candidates, TrustTunnelCertificatePaths{
			CertChainPath:  filepath.Join("/etc/letsencrypt/live", trimmed, "fullchain.pem"),
			PrivateKeyPath: filepath.Join("/etc/letsencrypt/live", trimmed, "privkey.pem"),
			Source:         "letsencrypt",
		})
	}
	candidates = append(candidates, TrustTunnelCertificatePaths{
		CertChainPath:  "/opt/trusttunnel/certs/fullchain.pem",
		PrivateKeyPath: "/opt/trusttunnel/certs/privkey.pem",
		Source:         "trusttunnel-default",
	})

	for _, item := range candidates {
		if fileExists(item.CertChainPath) && fileExists(item.PrivateKeyPath) {
			return &item, nil
		}
	}

	if len(candidates) > 0 {
		return &candidates[0], nil
	}
	return &TrustTunnelCertificatePaths{}, nil
}

func (s *ManagedTransportService) updateTrustTunnelWebUIEnv(cfg *TrustTunnelServiceConfig) error {
	envPath := "/opt/trusttunnel-suite/apps/webui/deploy/systemd/trusttunnel-webui.env"
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		return err
	}
	envMap, err := readEnvFile(envPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if envMap == nil {
		envMap = map[string]string{}
	}
	publicAddress := strings.TrimSpace(cfg.PublicAddress)
	if publicAddress == "" {
		publicAddress = strings.TrimSpace(cfg.Hostname)
	}
	envMap["TT_WEBUI_ENDPOINT_PUBLIC_ADDRESS"] = publicAddress
	return writeEnvFile(envPath, envMap)
}

func (s *ManagedTransportService) DisableTrustTunnel() error {
	if s.serviceState("trusttunnel") == "not-installed" {
		return nil
	}
	return s.RunAction("trusttunnel", "stop")
}

func (s *ManagedTransportService) ApplyMTProtoInbound(inbound *model.Inbound) error {
	if inbound == nil {
		return fmt.Errorf("inbound is nil")
	}

	settings := struct {
		Secret         string `json:"secret"`
		ConfigPath     string `json:"configPath"`
		FrontingDomain string `json:"frontingDomain"`
	}{}
	if strings.TrimSpace(inbound.Settings) != "" {
		if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
			return fmt.Errorf("parse MTProto inbound settings: %w", err)
		}
	}

	cfgPath := strings.TrimSpace(settings.ConfigPath)
	if cfgPath == "" {
		cfgPath = "/opt/trusttunnel/access/mtproto.toml"
	}
	cfg := &MTProtoServiceConfig{
		BindHost:       defaultString(strings.TrimSpace(inbound.Listen), "0.0.0.0"),
		Port:           inbound.Port,
		Secret:         settings.Secret,
		ConfigPath:     cfgPath,
		FrontingDomain: defaultString(strings.TrimSpace(settings.FrontingDomain), "www.cloudflare.com"),
	}
	if err := s.SaveMTProtoConfig(cfg); err != nil {
		return err
	}

	settings.Secret = cfg.Secret
	updatedSettings, err := json.Marshal(settings)
	if err == nil {
		inbound.Settings = string(updatedSettings)
	}

	if _, err := os.Stat("/usr/local/bin/mtg"); err == nil {
		if err := s.ensureMTProtoUnit(cfgPath); err != nil {
			return err
		}
		if s.serviceState("trusttunnel-mtproto") != "not-installed" {
			if err := s.RunAction("trusttunnel-mtproto", "restart"); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ManagedTransportService) serviceState(unit string) string {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return "systemctl-unavailable"
	}
	cmd := exec.Command("systemctl", "is-active", unit)
	out, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		trimmed := strings.TrimSpace(string(exitErr.Stderr))
		if trimmed != "" {
			return trimmed
		}
	}
	if _, statErr := os.Stat("/etc/systemd/system/" + unit + ".service"); statErr != nil {
		return "not-installed"
	}
	return "inactive"
}

func splitHostPort(value, defaultHost string, defaultPort int) (string, int) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return defaultHost, defaultPort
	}
	idx := strings.LastIndex(trimmed, ":")
	if idx <= 0 || idx == len(trimmed)-1 {
		return defaultHost, defaultPort
	}
	host := strings.TrimSpace(trimmed[:idx])
	port, err := strconv.Atoi(strings.TrimSpace(trimmed[idx+1:]))
	if err != nil || port < 1 || port > 65535 {
		return defaultHost, defaultPort
	}
	return host, port
}

func readEnvFile(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result, nil
}

func writeEnvFile(path string, env map[string]string) error {
	lines := make([]string, 0, len(env))
	for _, key := range []string{
		"TT_WEBUI_ADDR",
		"TT_WEBUI_DB_PATH",
		"TT_WEBUI_SESSION_SECRET",
		"TT_WEBUI_BOOTSTRAP_ADMIN_USER",
		"TT_WEBUI_BOOTSTRAP_ADMIN_PASS",
		"TT_WEBUI_ENDPOINT_LIVE_MODE",
		"TT_WEBUI_ENDPOINT_BIN",
		"TT_WEBUI_ENDPOINT_VPN_CONFIG",
		"TT_WEBUI_ENDPOINT_HOSTS_CONFIG",
		"TT_WEBUI_ENDPOINT_CREDENTIALS_FILE",
		"TT_WEBUI_ENDPOINT_PUBLIC_ADDRESS",
		"TT_WEBUI_ENDPOINT_CLIENT_BIN",
		"TT_WEBUI_ENDPOINT_PORT",
		"TT_WEBUI_CLIENT_STATS_FILE",
		"TT_WEBUI_METRICS_DISK_PATH",
		"TT_WEBUI_MTPROTO_BIN",
	} {
		if value, ok := env[key]; ok && strings.TrimSpace(value) != "" {
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
			delete(env, key)
		}
	}
	for key, value := range env {
		if strings.TrimSpace(key) == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}
	body := strings.Join(lines, "\n")
	if body != "" {
		body += "\n"
	}
	return os.WriteFile(path, []byte(body), 0o600)
}

func (s *ManagedTransportService) ensureMTProtoUnit(configPath string) error {
	servicePath := "/etc/systemd/system/trusttunnel-mtproto.service"
	serviceBody := fmt.Sprintf(`[Unit]
Description=TrustTunnel MTProto Access
Documentation=https://github.com/9seconds/mtg
After=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/mtg run %s
Restart=always
RestartSec=3
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
`, configPath)
	if err := os.WriteFile(servicePath, []byte(serviceBody), 0o644); err != nil {
		return err
	}
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "daemon-reload")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("systemctl daemon-reload failed: %s", strings.TrimSpace(stderr.String()))
		}
	}
	return nil
}

func (s *ManagedTransportService) saveTrustTunnelClients(clients []TrustTunnelClient) error {
	type credentialsConfig struct {
		Clients []TrustTunnelClient `toml:"client"`
	}
	cfg, err := s.GetTrustTunnelConfig()
	if err != nil {
		return err
	}
	return s.saveTrustTunnelClientsFromConfig(cfg.CredentialsFile, clients, true)
}

func (s *ManagedTransportService) saveTrustTunnelClientsFromConfig(path string, clients []TrustTunnelClient, restart bool) error {
	if strings.TrimSpace(path) == "" {
		path = "/opt/trusttunnel/credentials.toml"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body := renderTrustTunnelCredentials(clients)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		return err
	}
	if restart {
		if _, err := exec.LookPath("systemctl"); err == nil {
			cmd := exec.Command("systemctl", "restart", "trusttunnel")
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to restart trusttunnel: %s", strings.TrimSpace(stderr.String()))
			}
		}
	}
	return nil
}

func renderTrustTunnelCredentials(clients []TrustTunnelClient) []byte {
	var body strings.Builder
	for idx, client := range clients {
		if idx > 0 {
			body.WriteString("\n")
		}
		body.WriteString("[[client]]\n")
		body.WriteString("username = ")
		body.WriteString(strconv.Quote(strings.TrimSpace(client.Username)))
		body.WriteString("\n")
		body.WriteString("password = ")
		body.WriteString(strconv.Quote(strings.TrimSpace(client.Password)))
		body.WriteString("\n")
	}
	return []byte(body.String())
}

const (
	ttTagHostname           = 0x01
	ttTagAddress            = 0x02
	ttTagCustomSNI          = 0x03
	ttTagHasIPv6            = 0x04
	ttTagUsername           = 0x05
	ttTagPassword           = 0x06
	ttTagSkipVerification   = 0x07
	ttTagCertificate        = 0x08
	ttTagUpstreamProtocol   = 0x09
	ttTagAntiDPI            = 0x0A
	ttTagClientRandomPrefix = 0x0B
)

func encodeTrustTunnelDeepLink(cfg TrustTunnelExportConfig) (string, error) {
	if strings.TrimSpace(cfg.Hostname) == "" {
		return "", fmt.Errorf("hostname is required")
	}
	if strings.TrimSpace(cfg.Username) == "" {
		return "", fmt.Errorf("username is required")
	}
	if cfg.Password == "" {
		return "", fmt.Errorf("password is required")
	}

	var payload []byte
	payload = append(payload, ttTLV(ttTagHostname, []byte(strings.TrimSpace(cfg.Hostname)))...)
	payload = append(payload, ttTLV(ttTagUsername, []byte(strings.TrimSpace(cfg.Username)))...)
	payload = append(payload, ttTLV(ttTagPassword, []byte(cfg.Password))...)

	for _, address := range cfg.Addresses {
		address = strings.TrimSpace(address)
		if address == "" {
			continue
		}
		payload = append(payload, ttTLV(ttTagAddress, []byte(address))...)
	}
	if cfg.CustomSNI != "" {
		payload = append(payload, ttTLV(ttTagCustomSNI, []byte(cfg.CustomSNI))...)
	}
	if cfg.ClientRandomPrefix != "" {
		payload = append(payload, ttTLV(ttTagClientRandomPrefix, []byte(cfg.ClientRandomPrefix))...)
	}
	if !cfg.HasIPv6 {
		payload = append(payload, ttTLV(ttTagHasIPv6, []byte{0x00})...)
	}
	if cfg.SkipVerification {
		payload = append(payload, ttTLV(ttTagSkipVerification, []byte{0x01})...)
	}
	if cfg.AntiDPI {
		payload = append(payload, ttTLV(ttTagAntiDPI, []byte{0x01})...)
	}

	protocol := strings.TrimSpace(cfg.UpstreamProtocol)
	if protocol == "" {
		protocol = "http2"
	}
	switch protocol {
	case "http2":
	case "http3":
		if protocol == "http3" {
			payload = append(payload, ttTLV(ttTagUpstreamProtocol, []byte{0x02})...)
		}
	default:
		return "", fmt.Errorf("unknown upstream protocol: %s", protocol)
	}

	if strings.TrimSpace(cfg.Certificate) != "" {
		der, err := pemToDER(cfg.Certificate)
		if err != nil {
			return "", err
		}
		payload = append(payload, ttTLV(ttTagCertificate, der)...)
	}

	return "tt://?" + base64.RawURLEncoding.EncodeToString(payload), nil
}

func ttTLV(tag int, value []byte) []byte {
	out := append(ttVarint(uint64(tag)), ttVarint(uint64(len(value)))...)
	out = append(out, value...)
	return out
}

func ttVarint(value uint64) []byte {
	switch {
	case value <= 0x3F:
		return []byte{byte(value)}
	case value <= 0x3FFF:
		value |= 0x4000
		return []byte{byte(value >> 8), byte(value)}
	case value <= 0x3FFFFFFF:
		value |= 0x80000000
		return []byte{byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)}
	default:
		value |= 0xC000000000000000
		return []byte{
			byte(value >> 56), byte(value >> 48), byte(value >> 40), byte(value >> 32),
			byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value),
		}
	}
}

func pemToDER(pemText string) ([]byte, error) {
	data := []byte(strings.TrimSpace(pemText))
	if len(data) == 0 {
		return nil, nil
	}

	var out []byte
	found := false
	for len(data) > 0 {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		if strings.Contains(block.Type, "CERTIFICATE") {
			out = append(out, block.Bytes...)
			found = true
		}
		data = rest
	}
	if !found {
		return nil, fmt.Errorf("certificate field does not contain valid PEM certificate blocks")
	}
	return out, nil
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func trustTunnelClientsFromInbound(inbound *model.Inbound) ([]TrustTunnelClient, error) {
	if inbound == nil || strings.TrimSpace(inbound.Settings) == "" {
		return []TrustTunnelClient{}, nil
	}
	settings := struct {
		Clients []model.Client `json:"clients"`
	}{}
	if err := json.Unmarshal([]byte(inbound.Settings), &settings); err != nil {
		return nil, fmt.Errorf("parse TrustTunnel clients from inbound settings: %w", err)
	}
	result := make([]TrustTunnelClient, 0, len(settings.Clients))
	for _, client := range settings.Clients {
		username := strings.TrimSpace(client.Email)
		password := strings.TrimSpace(client.Password)
		if username == "" || password == "" {
			continue
		}
		result = append(result, TrustTunnelClient{
			Username: username,
			Password: password,
		})
	}
	return result, nil
}

func randomAlphaNum(length int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if length <= 0 {
		return ""
	}
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "trusttunnel"
	}
	for i := range bytes {
		bytes[i] = alphabet[int(bytes[i])%len(alphabet)]
	}
	return string(bytes)
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
