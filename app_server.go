package main

import (
	"encoding/json"

	"grpc-gui/internal/grpcreflect"
	"grpc-gui/internal/grpcrequest"
	"grpc-gui/internal/models"
	"grpc-gui/internal/utils"
)

func (a *App) CreateServer(server *models.Server) error {
	return a.storage.CreateServer(server)
}

func (a *App) GetServers() ([]models.Server, error) {
	return a.storage.GetServers()
}

func (a *App) DeleteServer(id uint) error {
	return a.storage.DeleteServer(id)
}

func (a *App) UpdateServer(server *models.Server) error {
	return a.storage.UpdateServer(server)
}

func (a *App) GetServerReflection(id uint) (*models.Server, error) {
	server, err := a.storage.GetServer(id)
	if err != nil {
		return nil, err
	}

	reflection, err := grpcreflect.NewReflector(a.ctx, server.Address, &utils.GRPCConnectOptions{UseTLS: server.OptUseTLS, Insecure: server.OptInsecure})
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
