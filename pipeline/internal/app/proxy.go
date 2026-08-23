package app

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type AuditProxy struct {
	HTTPAddress   string
	SOCKSAddress  string
	httpListener  net.Listener
	socksListener net.Listener
	server        *http.Server
	transport     *ScopeTransport
	logPath       string
	logMu         sync.Mutex
}

func NewPolicyScopeTransport(policy EngagementPolicy, maxRequests int, rps float64) *ScopeTransport {
	hosts := []string{}
	ports := []int{}
	pins := pinnedResolver{}
	for _, target := range policy.Environment.Targets {
		host := strings.ToLower(target.Host)
		hosts = append(hosts, host)
		ports = append(ports, target.Port)
		pins[host] = append(pins[host], net.IPAddr{IP: net.ParseIP(target.IP)})
	}
	for _, external := range policy.Environment.ExternalAPIs {
		host := strings.ToLower(external.Host)
		hosts = append(hosts, host)
		ports = append(ports, external.Port)
		pins[host] = append(pins[host], net.IPAddr{IP: net.ParseIP(external.IP)})
	}
	client := newScopedClient([]string{"http", "https"}, hosts, ports, true, maxRequests, rps, 30*time.Second, pins)
	return client.Transport.(*ScopeTransport)
}

func StartAuditProxy(ctx context.Context, httpAddress, socksAddress, logPath string, transport *ScopeTransport) (*AuditProxy, error) {
	if transport == nil {
		return nil, fmt.Errorf("ScopeTransport é obrigatório")
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		return nil, err
	}
	httpListener, err := net.Listen("tcp", httpAddress)
	if err != nil {
		return nil, err
	}
	socksListener, err := net.Listen("tcp", socksAddress)
	if err != nil {
		_ = httpListener.Close()
		return nil, err
	}
	proxy := &AuditProxy{HTTPAddress: httpListener.Addr().String(), SOCKSAddress: socksListener.Addr().String(), httpListener: httpListener, socksListener: socksListener, transport: transport, logPath: logPath}
	proxy.server = &http.Server{Handler: http.HandlerFunc(proxy.handleHTTP), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	go func() { _ = proxy.server.Serve(httpListener) }()
	go proxy.serveSOCKS(ctx)
	go func() { <-ctx.Done(); _ = proxy.Close(context.Background()) }()
	return proxy, nil
}

func StartEngagementAuditProxy(ctx context.Context, state *NetworkState, policy EngagementPolicy, maxRequests int, rps float64) (*AuditProxy, error) {
	command := exec.CommandContext(ctx, "docker", "network", "inspect", state.NetworkName, "--format", "{{(index .IPAM.Config 0).Gateway}}")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("obter gateway da rede: %w", err)
	}
	gateway := strings.TrimSpace(string(output))
	if net.ParseIP(gateway) == nil {
		return nil, fmt.Errorf("gateway Docker inválido: %s", gateway)
	}
	proxy, err := StartAuditProxy(ctx, net.JoinHostPort(gateway, "0"), net.JoinHostPort(gateway, "0"), state.ProxyLog, NewPolicyScopeTransport(policy, maxRequests, rps))
	if err != nil {
		return nil, err
	}
	if err := allowAuditProxyInput(ctx, *state, gateway, proxy.HTTPAddress, proxy.SOCKSAddress); err != nil {
		_ = proxy.Close(context.Background())
		return nil, fmt.Errorf("liberar somente as portas do proxy no INPUT: %w", err)
	}
	state.HTTPProxyURL = "http://" + proxy.HTTPAddress
	state.SOCKSProxyURL = "socks5://" + proxy.SOCKSAddress
	return proxy, nil
}

func allowAuditProxyInput(ctx context.Context, state NetworkState, gateway string, addresses ...string) error {
	bridgeData, err := os.ReadFile(filepath.Join(filepath.Dir(state.SetupScript), "bridge.name"))
	if err != nil {
		return err
	}
	bridge := strings.TrimSpace(string(bridgeData))
	if bridge == "" || state.InputChain == "" {
		return fmt.Errorf("bridge ou input_chain ausente")
	}
	for _, address := range addresses {
		_, port, splitErr := net.SplitHostPort(address)
		if splitErr != nil {
			return splitErr
		}
		args := []string{"-w", "-I", state.InputChain, "1", "-i", bridge, "-d", gateway, "-p", "tcp", "--dport", port, "-j", "ACCEPT"}
		if output, runErr := exec.CommandContext(ctx, "iptables", args...).CombinedOutput(); runErr != nil {
			return fmt.Errorf("%w: %s", runErr, strings.TrimSpace(string(output)))
		}
	}
	return nil
}

func (p *AuditProxy) Close(ctx context.Context) error {
	if p.socksListener != nil {
		_ = p.socksListener.Close()
	}
	if p.server != nil {
		return p.server.Shutdown(ctx)
	}
	return nil
}

func (p *AuditProxy) handleHTTP(w http.ResponseWriter, req *http.Request) {
	started := time.Now()
	toolRunID := req.Header.Get("X-Empresa-Tool-Run")
	req.Header.Del("X-Empresa-Tool-Run")
	if toolRunID == "" {
		toolRunID = proxyToolRunID(req.Header.Get("Proxy-Authorization"))
	}
	req.Header.Del("Proxy-Authorization")
	if req.Method == http.MethodConnect {
		p.handleConnect(w, req, started, toolRunID)
		return
	}
	if req.URL == nil || req.URL.Hostname() == "" {
		http.Error(w, "proxy exige URL absoluta", http.StatusBadRequest)
		p.writeRequestLog(toolRunID, req.Method, requestURL(req), http.StatusBadRequest, started, "rejected")
		return
	}
	outbound := req.Clone(req.Context())
	outbound.RequestURI = ""
	outbound.Header.Del("Proxy-Connection")
	response, err := p.transport.RoundTrip(outbound)
	if err != nil {
		http.Error(w, "destino recusado", http.StatusBadGateway)
		p.writeRequestLog(toolRunID, req.Method, req.URL.String(), http.StatusBadGateway, started, "rejected")
		return
	}
	defer response.Body.Close()
	copyHeaders(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
	p.writeRequestLog(toolRunID, req.Method, req.URL.String(), response.StatusCode, started, "allowed_pinned")
}

func proxyToolRunID(header string) string {
	if !strings.HasPrefix(header, "Basic ") {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(header, "Basic "))
	if err != nil {
		return ""
	}
	user := strings.SplitN(string(decoded), ":", 2)[0]
	if strings.HasPrefix(user, "run_") {
		return user
	}
	return ""
}

func (p *AuditProxy) handleConnect(w http.ResponseWriter, req *http.Request, started time.Time, toolRunID string) {
	host, portText, err := net.SplitHostPort(req.Host)
	if err != nil {
		http.Error(w, "CONNECT inválido", http.StatusBadRequest)
		p.writeRequestLog(toolRunID, req.Method, "https://"+req.Host, http.StatusBadRequest, started, "rejected")
		return
	}
	port, _ := strconv.Atoi(portText)
	scheme := "https"
	if port == 80 {
		scheme = "http"
	}
	target := &url.URL{Scheme: scheme, Host: net.JoinHostPort(host, portText)}
	upstream, err := p.transport.DialValidated(req.Context(), target)
	if err != nil {
		http.Error(w, "destino recusado", http.StatusBadGateway)
		p.writeRequestLog(toolRunID, req.Method, target.String(), http.StatusBadGateway, started, "rejected")
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		_ = upstream.Close()
		http.Error(w, "hijack indisponível", http.StatusInternalServerError)
		return
	}
	client, buffered, err := hijacker.Hijack()
	if err != nil {
		_ = upstream.Close()
		return
	}
	_, _ = buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
	_ = buffered.Flush()
	p.writeRequestLog(toolRunID, req.Method, target.String(), http.StatusOK, started, "allowed_pinned")
	go tunnel(client, upstream)
}

func tunnel(a, b net.Conn) {
	defer a.Close()
	defer b.Close()
	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(a, b); done <- struct{}{} }()
	go func() { _, _ = io.Copy(b, a); done <- struct{}{} }()
	<-done
}

func (p *AuditProxy) serveSOCKS(ctx context.Context) {
	for {
		conn, err := p.socksListener.Accept()
		if err != nil {
			return
		}
		go p.handleSOCKS(ctx, conn)
	}
}

func (p *AuditProxy) handleSOCKS(ctx context.Context, client net.Conn) {
	defer client.Close()
	_ = client.SetDeadline(time.Now().Add(30 * time.Second))
	reader := bufio.NewReader(client)
	header := make([]byte, 2)
	if _, err := io.ReadFull(reader, header); err != nil || header[0] != 5 {
		return
	}
	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(reader, methods); err != nil {
		return
	}
	_, _ = client.Write([]byte{5, 0})
	request := make([]byte, 4)
	if _, err := io.ReadFull(reader, request); err != nil || request[0] != 5 || request[1] != 1 {
		return
	}
	var host string
	switch request[3] {
	case 3:
		length, err := reader.ReadByte()
		if err != nil {
			return
		}
		name := make([]byte, int(length))
		if _, err := io.ReadFull(reader, name); err != nil {
			return
		}
		host = string(name)
	default:
		_, _ = client.Write([]byte{5, 8, 0, 1, 0, 0, 0, 0, 0, 0})
		return
	}
	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(reader, portBytes); err != nil {
		return
	}
	port := int(binary.BigEndian.Uint16(portBytes))
	scheme := "https"
	if port == 80 {
		scheme = "http"
	}
	target := &url.URL{Scheme: scheme, Host: net.JoinHostPort(host, strconv.Itoa(port))}
	started := time.Now()
	upstream, err := p.transport.DialValidated(ctx, target)
	if err != nil {
		_, _ = client.Write([]byte{5, 2, 0, 1, 0, 0, 0, 0, 0, 0})
		p.writeRequestLog("", "CONNECT", target.String(), 502, started, "rejected")
		return
	}
	_, _ = client.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0})
	p.writeRequestLog("", "CONNECT", target.String(), 200, started, "allowed_pinned")
	_ = client.SetDeadline(time.Time{})
	tunnel(client, upstream)
}

func (p *AuditProxy) writeRequestLog(toolRunID, method, rawURL string, status int, started time.Time, decision string) {
	redacted := redactURL(rawURL)
	stamp := time.Now().UTC().Format(time.RFC3339Nano)
	requestHash := HashBytes([]byte(strings.ToUpper(method) + "|" + redacted))
	log := RequestLog{ID: stableID("request", stamp+"|"+requestHash), ToolRunID: toolRunID, ObservedAt: stamp, Method: strings.ToUpper(method), URL: redacted, Status: status, DurationMS: time.Since(started).Milliseconds(), RequestSHA256: requestHash, ScopeDecision: decision}
	data, err := json.Marshal(log)
	if err != nil {
		return
	}
	p.logMu.Lock()
	defer p.logMu.Unlock()
	file, err := os.OpenFile(p.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	_, _ = file.Write(append(data, '\n'))
	_ = file.Close()
}

func copyHeaders(destination, source http.Header) {
	for key, values := range source {
		if strings.EqualFold(key, "Connection") || strings.EqualFold(key, "Proxy-Connection") {
			continue
		}
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}

func requestURL(req *http.Request) string {
	if req.URL == nil {
		return ""
	}
	return req.URL.String()
}
