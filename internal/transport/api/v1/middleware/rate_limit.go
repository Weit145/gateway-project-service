package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/netip"
	"strings"

	errs "github.com/Weit145/gateway-startup/internal/lib/e"
	"github.com/Weit145/gateway-startup/internal/transport/api/v1/authctx"
	router_err "github.com/Weit145/gateway-startup/internal/transport/api/v1/err"
)

func CheckIPRateLimit(rd RateLimiter, trustedProxies []netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ip, err := clientIP(r, trustedProxies)
			if err != nil {
				slog.DebugContext(ctx, "invalid ip", slog.Any("error", err))
				router_err.MapErr(w, r, errs.ErrInvalidInput)
				return
			}

			if err = rd.LimiterIP(ctx, ip); err != nil {
				slog.DebugContext(ctx, "rate limit", slog.Any("error", err))
				router_err.MapErr(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CheckUserRateLimit(rd RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID, ok := authctx.UserID(ctx)
			if !ok {
				router_err.MapErr(w, r, errs.ErrUnauthorized)
				return
			}

			if err := rd.LimiterUserID(ctx, userID); err != nil {
				slog.DebugContext(ctx, "rate limit", slog.Any("error", err))
				router_err.MapErr(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request, trustedProxies []netip.Prefix) (string, error) {
	remoteIP, err := parseIP(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("parse remote address: %w", err)
	}
	if !isTrustedProxy(remoteIP, trustedProxies) {
		return remoteIP.String(), nil
	}

	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			ip, err := parseIP(parts[i])
			if err != nil {
				return "", fmt.Errorf("parse X-Forwarded-For: %w", err)
			}
			if !isTrustedProxy(ip, trustedProxies) || i == 0 {
				return ip.String(), nil
			}
		}
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		ip, err := parseIP(realIP)
		if err != nil {
			return "", fmt.Errorf("parse X-Real-IP: %w", err)
		}
		return ip.String(), nil
	}

	return remoteIP.String(), nil
}

func parseIP(value string) (netip.Addr, error) {
	value = strings.TrimSpace(value)

	if addrPort, err := netip.ParseAddrPort(value); err == nil {
		return addrPort.Addr().Unmap(), nil
	}

	ip, err := netip.ParseAddr(value)
	if err != nil {
		return netip.Addr{}, err
	}
	return ip.Unmap(), nil
}

func isTrustedProxy(ip netip.Addr, trustedProxies []netip.Prefix) bool {
	for _, proxy := range trustedProxies {
		if proxy.Contains(ip) {
			return true
		}
	}
	return false
}
