package check

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"syscall"
	"time"
	"url-monitor/internal/monitor"
)

type CheckRunner struct{}

func (c *CheckRunner) Check(ctx context.Context, m monitor.Monitor) monitor.MonitorCheck {
	client := &http.Client{
		Timeout: time.Duration(m.IntervalSeconds/2) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	check := monitor.MonitorCheck{
		MonitorID: m.ID,
		StartedAt: time.Now(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, nil)
	if err != nil {
		check.Status = monitor.MonitorCheckStatusError
		check.ErrorMessage = err.Error()
		check.FinishedAt = time.Now()
		check.ResponseTimeMS = check.FinishedAt.Sub(check.StartedAt).Milliseconds()
		return check
	}

	resp, err := client.Do(req)
	if err != nil {
		check.Status = monitor.MonitorCheckStatusError
		check.ErrorKind = c.classifyCheckErrorKind(err)
		check.ErrorMessage = err.Error()
		check.FinishedAt = time.Now()
		check.ResponseTimeMS = check.FinishedAt.Sub(check.StartedAt).Milliseconds()
		return check
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	check.FinishedAt = time.Now()
	check.ResponseTimeMS = check.FinishedAt.Sub(check.StartedAt).Milliseconds()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		check.Status = monitor.MonitorCheckStatusUp
	} else {
		check.Status = monitor.MonitorCheckStatusDown
	}

	check.HTTPStatusCode = int16(resp.StatusCode)

	return check
}

func (c *CheckRunner) classifyCheckErrorKind(err error) monitor.CheckErrorKind {
	var dnsErr *net.DNSError

	var certificateInvalidErr *x509.CertificateInvalidError
	var hostnameErr *x509.HostnameError
	var unknownAuthorityErr *x509.UnknownAuthorityError
	var certificateVerificationErr *tls.CertificateVerificationError
	var recordHeaderErr tls.RecordHeaderError

	var opErr *net.OpError
	var sysErr *os.SyscallError
	var errnoErr syscall.Errno

	var urlErr *url.Error

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return monitor.CheckErrorTimeout
	case errors.Is(err, context.Canceled):
		return monitor.CheckErrorCanceled
	case errors.As(err, &dnsErr):
		return monitor.CheckErrorDNS
	case errors.As(err, &certificateInvalidErr):
	case errors.As(err, &hostnameErr):
	case errors.As(err, &unknownAuthorityErr):
	case errors.As(err, &certificateVerificationErr):
	case errors.As(err, &recordHeaderErr):
		return monitor.CheckErrorTLS
	case errors.As(err, &opErr):
	case errors.As(err, &sysErr):
	case errors.As(err, &errnoErr):
		return monitor.CheckErrorConnection
	case errors.As(err, &urlErr):
		if urlErr.Timeout() || errors.Is(urlErr.Err, context.DeadlineExceeded) {
			return monitor.CheckErrorTimeout
		}

		return c.classifyCheckErrorKind(urlErr.Err)
	default:
		return monitor.CheckErrorKindUnknown
	}

	return monitor.CheckErrorKindUnknown
}
