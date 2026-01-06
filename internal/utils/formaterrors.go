package utils

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FormatConnectionError(err error, address string, useTLS, insecure bool) string {
	if err == nil {
		return "Неизвестная ошибка подключения"
	}

	errStr := err.Error()
	baseMsg := "Не удалось подключиться к серверу"

	if useTLS {
		if insecure {
			baseMsg += " (TLS с пропуском проверки сертификата)"
		} else {
			baseMsg += " (TLS с проверкой сертификата)"
		}
	} else {
		baseMsg += " (без TLS)"
	}

	errLower := strings.ToLower(errStr)
	if strings.Contains(errLower, "first record does not look like a tls handshake") {
		return baseMsg + ": сервер не использует TLS, но вы пытаетесь подключиться с TLS. Отключите опцию TLS или используйте сервер с поддержкой TLS"
	}

	if st, ok := status.FromError(err); ok {
		code := st.Code()
		switch code {
		case codes.Unavailable:
			msg := extractMainError(st.Message())
			if strings.Contains(strings.ToLower(msg), "tls") || strings.Contains(strings.ToLower(msg), "handshake") {
				if useTLS {
					return baseMsg + ": ошибка TLS handshake. Возможно, сервер не использует TLS. " + msg
				}
			}
			return baseMsg + ": сервер недоступен. " + msg
		case codes.DeadlineExceeded:
			return baseMsg + ": превышено время ожидания подключения"
		case codes.Canceled:
			return baseMsg + ": подключение отменено"
		case codes.PermissionDenied:
			return baseMsg + ": доступ запрещен. " + extractMainError(st.Message())
		}
	}

	if strings.Contains(errLower, "tls") || strings.Contains(errLower, "certificate") {
		return baseMsg + ": ошибка TLS/сертификата - " + extractMainError(errStr)
	}
	if strings.Contains(errLower, "connection refused") {
		return baseMsg + ": соединение отклонено. Проверьте, что сервер запущен на " + address
	}
	if strings.Contains(errLower, "dial tcp") || strings.Contains(errLower, "connectex") {
		return baseMsg + ": не удалось установить TCP соединение с " + address + " - " + extractMainError(errStr)
	}
	if strings.Contains(errLower, "no such host") {
		return baseMsg + ": хост не найден - " + address
	}

	return baseMsg + ": " + extractMainError(errStr)
}

func extractMainError(errStr string) string {
	errStr = strings.TrimSpace(errStr)

	if strings.Contains(errStr, "desc =") {
		parts := strings.Split(errStr, "desc =")
		for i := len(parts) - 1; i >= 0; i-- {
			if i < len(parts) {
				desc := strings.TrimSpace(parts[i])
				desc = strings.Trim(desc, "\"")
				if desc != "" && !strings.Contains(desc, "rpc error:") {
					return desc
				}
			}
		}
	}

	if strings.Contains(errStr, "rpc error:") {
		parts := strings.Split(errStr, "rpc error:")
		if len(parts) > 1 {
			mainErr := strings.TrimSpace(parts[len(parts)-1])
			mainErr = strings.Trim(mainErr, "\"")
			return mainErr
		}
	}

	return errStr
}

func FormatReflectionError(err error) string {
	if err == nil {
		return "Неизвестная ошибка рефлексии"
	}

	errStr := err.Error()
	if st, ok := status.FromError(err); ok {
		code := st.Code()
		switch code {
		case codes.Unimplemented:
			return "Сервер не поддерживает gRPC рефлексию (метод не реализован)"
		case codes.PermissionDenied:
			return "Доступ к рефлексии запрещен: " + st.Message()
		case codes.NotFound:
			return "Сервис рефлексии не найден: " + st.Message()
		}
		return "Ошибка рефлексии: " + st.Message()
	}

	return "Ошибка получения рефлексии: " + errStr
}

func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	if st, ok := status.FromError(err); ok {
		code := st.Code()
		return code == codes.Unavailable || code == codes.DeadlineExceeded || code == codes.Canceled
	}

	errStr := strings.ToLower(err.Error())
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"no connection",
		"unavailable",
		"deadline exceeded",
		"context canceled",
		"dial tcp",
		"connectex",
		"connection",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(errStr, connErr) {
			return true
		}
	}

	return false
}
