package router

import (
	_ "github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
)

// WebSocketHandler godoc
// @Summary      Connect to WebSocket
// @Description  Upgrades the HTTP connection to a WebSocket connection for real-time notifications and presence. Requires a valid ticket obtained from /auth/ticket.
// @Tags         realtime
// @Param        ticket query string true "WebSocket authentication ticket"
// @Success      101 {string} string "Switching Protocols"
// @Failure      401 {object} response.SwaggerErrorResponseWrapper "Unauthorized"
// @Router       /ws [get]
//
//nolint:unused
func websocketDocs() {}

// SSEHandler godoc
// @Summary      Subscribe to SSE events
// @Description  Establishes a Server-Sent Events stream for real-time updates.
// @Tags         realtime
// @Security     BearerAuth
// @Produce      text/event-stream
// @Success      200 {string} string "Event stream"
// @Failure      401 {object} response.SwaggerErrorResponseWrapper "Unauthorized"
// @Router       /events [get]
//
//nolint:unused
func sseDocs() {}
