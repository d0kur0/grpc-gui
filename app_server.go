package main

import (
	"context"
	"encoding/json"
	"time"

	"grpc-gui/internal/grpcreflect"
	"grpc-gui/internal/grpcrequest"
	"grpc-gui/internal/models"
	"grpc-gui/internal/utils"
)

type ValidationStatus int

const (
	ValidationStatusSuccess ValidationStatus = iota
	ValidationStatusConnectionFailed
	ValidationStatusReflectionNotAvailable
	ValidationStatusNoServices
)

type ValidationResult struct {
	Status  ValidationStatus `json:"status"`
	Message string           `json:"message,omitempty"`
}

func (a *App) CreateServer(name, address string, useTLS, insecure bool) (uint, error) {
	server := &models.Server{
		Name:        name,
		Address:     address,
		OptUseTLS:   useTLS,
		OptInsecure: insecure,
	}
	err := a.storage.CreateServer(server)
	if err != nil {
		return 0, err
	}
	return server.ID, nil
}

func (a *App) GetServers() ([]models.Server, error) {
	return a.storage.GetServers()
}

type ServerWithReflection struct {
	Server     *models.Server            `json:"server"`
	Reflection *grpcreflect.ServicesInfo `json:"reflection"`
	Error      string                    `json:"error,omitempty"`
}

func (a *App) getServerReflection(ctx context.Context, server models.Server) ServerWithReflection {
	result := ServerWithReflection{
		Server:     &server,
		Reflection: &grpcreflect.ServicesInfo{Services: []grpcreflect.ServiceInfo{}},
	}

	reflector, err := grpcreflect.NewReflector(ctx, server.Address, &utils.GRPCConnectOptions{UseTLS: server.OptUseTLS, Insecure: server.OptInsecure})
	if err != nil {
		result.Error = utils.FormatConnectionError(err, server.Address, server.OptUseTLS, server.OptInsecure)
		return result
	}
	defer reflector.Close()

	services, err := reflector.GetAllServicesInfo()
	if err != nil {
		if utils.IsConnectionError(err) {
			result.Error = utils.FormatConnectionError(err, server.Address, server.OptUseTLS, server.OptInsecure)
		} else {
			result.Error = utils.FormatReflectionError(err)
		}
		return result
	}

	filteredServices := &grpcreflect.ServicesInfo{
		Services: []grpcreflect.ServiceInfo{},
	}

	for _, service := range services.Services {
		if !grpcreflect.IsReflectionService(service.Name) {
			filteredServices.Services = append(filteredServices.Services, service)
		}
	}

	result.Reflection = filteredServices
	return result
}

func (a *App) GetServersWithReflection() ([]ServerWithReflection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	servers, err := a.storage.GetServers()
	if err != nil {
		return nil, err
	}

	var serversWithReflection []ServerWithReflection
	for _, server := range servers {
		serversWithReflection = append(serversWithReflection, a.getServerReflection(ctx, server))
	}

	return serversWithReflection, nil
}

func (a *App) GetServerWithReflection(id uint) (*ServerWithReflection, error) {
	server, err := a.storage.GetServer(id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := a.getServerReflection(ctx, *server)
	return &result, nil
}

func (a *App) DeleteServer(id uint) error {
	return a.storage.DeleteServer(id)
}

func (a *App) UpdateServer(id uint, name, address string, useTLS, insecure bool) error {
	server := &models.Server{
		ID:          id,
		Name:        name,
		Address:     address,
		OptUseTLS:   useTLS,
		OptInsecure: insecure,
	}
	return a.storage.UpdateServer(server)
}

func (a *App) GetServerReflection(id uint) (*models.Server, error) {
	server, err := a.storage.GetServer(id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reflection, err := grpcreflect.NewReflector(ctx, server.Address, &utils.GRPCConnectOptions{UseTLS: server.OptUseTLS, Insecure: server.OptInsecure})
	if err != nil {
		return nil, err
	}
	defer reflection.Close()

	return server, nil
}

func (a *App) GetJsonExample(msg *grpcreflect.MessageInfo) (string, error) {
	json, err := grpcreflect.GenerateJSONExample(msg)
	if err != nil {
		return "{}", err
	}

	return string(json), nil
}

func (a *App) DoGRPCRequest(serverId uint, address, service, method, payload string, requestHeaders, contextValues map[string]string) (string, int32, error) {
	resp, code, respHeaders, err := grpcrequest.DoGRPCRequest(address, service, method, payload, requestHeaders, contextValues)

	var historyRecord models.History
	historyRecord.ServerID = serverId
	historyRecord.Service = service
	historyRecord.Method = method
	historyRecord.Request = payload
	historyRecord.Response = resp
	historyRecord.StatusCode = int32(code)

	if len(requestHeaders) > 0 {
		reqHeadersJSON, _ := json.Marshal(requestHeaders)
		historyRecord.RequestHeaders = string(reqHeadersJSON)
	}

	if len(respHeaders) > 0 {
		respHeadersMap := make(map[string]string)
		for k, v := range respHeaders {
			if len(v) > 0 {
				respHeadersMap[k] = v[0]
			}
		}
		respHeadersJSON, _ := json.Marshal(respHeadersMap)
		historyRecord.ResponseHeaders = string(respHeadersJSON)
	}

	if len(contextValues) > 0 {
		contextJSON, _ := json.Marshal(contextValues)
		historyRecord.ContextValues = string(contextJSON)
	}

	if err := a.storage.CreateHistory(&historyRecord); err != nil {
		return "", int32(code), err
	}

	return resp, int32(code), err
}

func (a *App) ValidateServerAddress(address string, useTLS, insecure bool) ValidationResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := &utils.GRPCConnectOptions{
		UseTLS:   useTLS,
		Insecure: insecure,
	}

	reflection, err := grpcreflect.NewReflector(ctx, address, opts)
	if err != nil {
		return ValidationResult{
			Status:  ValidationStatusConnectionFailed,
			Message: utils.FormatConnectionError(err, address, useTLS, insecure),
		}
	}
	defer reflection.Close()

	services, err := reflection.GetAllServicesInfo()
	if err != nil {
		if utils.IsConnectionError(err) {
			return ValidationResult{
				Status:  ValidationStatusConnectionFailed,
				Message: utils.FormatConnectionError(err, address, useTLS, insecure),
			}
		}
		return ValidationResult{
			Status:  ValidationStatusReflectionNotAvailable,
			Message: utils.FormatReflectionError(err),
		}
	}

	if services == nil || len(services.Services) == 0 {
		return ValidationResult{
			Status:  ValidationStatusNoServices,
			Message: "Сервер доступен, но не предоставляет сервисы через рефлексию",
		}
	}

	return ValidationResult{
		Status: ValidationStatusSuccess,
	}
}
