package monitor

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
)

type CheckRunner struct{}

func (c *CheckRunner) Check(ctx context.Context, m Monitor) MonitorCheck {
	client := &http.Client{
		Timeout: time.Duration(m.IntervalSeconds/2) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	check := MonitorCheck{
		MonitorID: m.ID,
		StartedAt: time.Now(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, nil)
	if err != nil {
		check.Status = MonitorCheckStatusError
		check.ErrorMessage = err.Error()
		check.FinishedAt = time.Now()
		check.ResponseTimeMS = check.FinishedAt.Sub(check.StartedAt).Milliseconds()
		return check
	}

	resp, err := client.Do(req)
	if err != nil {
		check.Status = MonitorCheckStatusError
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
		check.Status = MonitorCheckStatusUp
	} else {
		check.Status = MonitorCheckStatusDown
	}

	check.HTTPStatusCode = int16(resp.StatusCode)

	return check
}

func (c *CheckRunner) classifyCheckErrorKind(err error) CheckErrorKind {
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
		return CheckErrorTimeout
	case errors.Is(err, context.Canceled):
		return CheckErrorCanceled
	case errors.As(err, &dnsErr):
		return CheckErrorDNS
	case errors.As(err, &certificateInvalidErr):
	case errors.As(err, &hostnameErr):
	case errors.As(err, &unknownAuthorityErr):
	case errors.As(err, &certificateVerificationErr):
	case errors.As(err, &recordHeaderErr):
		return CheckErrorTLS
	case errors.As(err, &opErr):
	case errors.As(err, &sysErr):
	case errors.As(err, &errnoErr):
		return CheckErrorConnection
	case errors.As(err, &urlErr):
		if urlErr.Timeout() || errors.Is(urlErr.Err, context.DeadlineExceeded) {
			return CheckErrorTimeout
		}

		return c.classifyCheckErrorKind(urlErr)
	default:
		return CheckErrorKindUnknown
	}

	return CheckErrorKindUnknown
}
