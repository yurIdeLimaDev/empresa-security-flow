package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ipResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

type pinnedResolver map[string][]net.IPAddr

func (r pinnedResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	addresses := r[strings.ToLower(strings.TrimSuffix(host, "."))]
	if len(addresses) == 0 {
		return nil, fmt.Errorf("host sem IP pinado na policy: %s", host)
	}
	result := make([]net.IPAddr, len(addresses))
	copy(result, addresses)
	return result, nil
}

type ScopeTransport struct {
	Base            *http.Transport
	Resolver        ipResolver
	AllowedSchemes  map[string]struct{}
	AllowedHosts    map[string]struct{}
	AllowedPorts    map[int]struct{}
	AllowPrivateIPs bool
	MaxRequests     int
	Interval        time.Duration

	mu        sync.Mutex
	requests  int
	lastStart time.Time
}

func NewScopedClient(schemes, hosts []string, ports []int, allowPrivate bool, maxRequests int, rps float64, timeout time.Duration) *http.Client {
	return newScopedClient(schemes, hosts, ports, allowPrivate, maxRequests, rps, timeout, nil)
}

func NewScopedClientWithPins(schemes, hosts []string, ports []int, allowPrivate bool, maxRequests int, rps float64, timeout time.Duration, pins map[string][]string) *http.Client {
	resolver := pinnedResolver{}
	for host, values := range pins {
		key := strings.ToLower(strings.TrimSuffix(host, "."))
		for _, value := range values {
			ip := net.ParseIP(value)
			if ip != nil {
				resolver[key] = append(resolver[key], net.IPAddr{IP: ip})
			}
		}
	}
	return newScopedClient(schemes, hosts, ports, allowPrivate, maxRequests, rps, timeout, resolver)
}

func newScopedClient(schemes, hosts []string, ports []int, allowPrivate bool, maxRequests int, rps float64, timeout time.Duration, resolver ipResolver) *http.Client {
	allowedSchemes := make(map[string]struct{})
	for _, value := range schemes {
		allowedSchemes[strings.ToLower(value)] = struct{}{}
	}
	allowedHosts := make(map[string]struct{})
	for _, value := range hosts {
		allowedHosts[strings.ToLower(strings.TrimSuffix(value, "."))] = struct{}{}
	}
	allowedPorts := make(map[int]struct{})
	for _, value := range ports {
		allowedPorts[value] = struct{}{}
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	if maxRequests <= 0 {
		maxRequests = 1
	}
	if rps <= 0 {
		rps = 1
	}
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = nil
	transport := &ScopeTransport{
		Base:            base,
		Resolver:        resolver,
		AllowedSchemes:  allowedSchemes,
		AllowedHosts:    allowedHosts,
		AllowedPorts:    allowedPorts,
		AllowPrivateIPs: allowPrivate,
		MaxRequests:     maxRequests,
		Interval:        time.Duration(float64(time.Second) / rps),
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return errors.New("limite de redirecionamentos atingido")
			}
			_, _, err := transport.resolveURL(req.Context(), req.URL)
			return err
		},
	}
}

func (t *ScopeTransport) ResolveURL(ctx context.Context, u *url.URL) ([]net.IPAddr, int, error) {
	return t.resolveURL(ctx, u)
}

func (t *ScopeTransport) DialValidated(ctx context.Context, u *url.URL) (net.Conn, error) {
	addresses, port, err := t.resolveURL(ctx, u)
	if err != nil {
		return nil, err
	}
	if err := t.acquire(ctx); err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	var lastErr error
	for _, address := range addresses {
		conn, dialErr := dialer.DialContext(ctx, "tcp", net.JoinHostPort(address.IP.String(), strconv.Itoa(port)))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	return nil, fmt.Errorf("nenhum IP validado aceitou conexão: %w", lastErr)
}

func (t *ScopeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	addresses, port, err := t.resolveURL(req.Context(), req.URL)
	if err != nil {
		return nil, err
	}
	if err := t.acquire(req.Context()); err != nil {
		return nil, err
	}

	host := strings.ToLower(strings.TrimSuffix(req.URL.Hostname(), "."))
	transport := t.Base.Clone()
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		dialHost, dialPort, splitErr := net.SplitHostPort(address)
		if splitErr != nil {
			return nil, fmt.Errorf("destino de conexão inválido: %w", splitErr)
		}
		if !strings.EqualFold(strings.TrimSuffix(dialHost, "."), host) || dialPort != strconv.Itoa(port) {
			return nil, fmt.Errorf("tentativa de conexão fora do destino validado: %s", address)
		}
		var lastErr error
		dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
		for _, candidate := range addresses {
			conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(candidate.IP.String(), dialPort))
			if dialErr == nil {
				return conn, nil
			}
			lastErr = dialErr
		}
		return nil, fmt.Errorf("nenhum IP validado aceitou conexão: %w", lastErr)
	}
	return transport.RoundTrip(req)
}

func (t *ScopeTransport) acquire(ctx context.Context) error {
	t.mu.Lock()
	if t.requests >= t.MaxRequests {
		t.mu.Unlock()
		return errors.New("budget máximo de requisições atingido")
	}
	wait := t.Interval - time.Since(t.lastStart)
	if wait > 0 {
		t.mu.Unlock()
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return ctx.Err()
		}
		t.mu.Lock()
		if t.requests >= t.MaxRequests {
			t.mu.Unlock()
			return errors.New("budget máximo de requisições atingido")
		}
	}
	t.requests++
	t.lastStart = time.Now()
	t.mu.Unlock()
	return nil
}

func (t *ScopeTransport) resolveURL(ctx context.Context, u *url.URL) ([]net.IPAddr, int, error) {
	scheme := strings.ToLower(u.Scheme)
	if _, ok := t.AllowedSchemes[scheme]; !ok {
		return nil, 0, fmt.Errorf("scheme fora do escopo: %s", scheme)
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if _, ok := t.AllowedHosts[host]; !ok {
		return nil, 0, fmt.Errorf("host fora do escopo: %s", host)
	}
	port := 80
	if scheme == "https" {
		port = 443
	}
	if u.Port() != "" {
		value, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, 0, errors.New("porta inválida")
		}
		port = value
	}
	if _, ok := t.AllowedPorts[port]; !ok {
		return nil, 0, fmt.Errorf("porta fora do escopo: %d", port)
	}
	resolver := t.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	addresses, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, 0, fmt.Errorf("resolver host: %w", err)
	}
	if len(addresses) == 0 {
		return nil, 0, errors.New("host sem endereço resolvido")
	}
	for _, address := range addresses {
		if !t.AllowPrivateIPs && isPrivateOrLocal(address.IP) {
			return nil, 0, fmt.Errorf("host resolve para IP privado/local: %s", address.IP)
		}
	}
	return addresses, port, nil
}

func isPrivateOrLocal(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()
}
