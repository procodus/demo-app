// Package frontend provides the web frontend for the IoT dashboard.
package frontend

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"procodus.dev/demo-app/pkg/iot"
)

// handleIndex serves the main index page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("handling index request")

	// Render index template
	if err := renderIndex(r.Context(), w); err != nil {
		s.logger.Error("failed to render index", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleDevices serves the devices page.
func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("handling devices request")

	// Fetch devices from backend
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.grpcClient.GetAllDevice(ctx, &iot.GetAllDevicesRequest{})
	if err != nil {
		s.logger.Error("failed to fetch devices", "error", err)
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	// Render devices page
	if err := renderDevices(r.Context(), w, resp.GetDevices()); err != nil {
		s.logger.Error("failed to render devices", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleDevice serves a single device detail page.
func (s *Server) handleDevice(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("id")
	s.logger.Debug("handling device request", "device_id", deviceID)

	// Fetch device from backend
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	deviceResp, err := s.grpcClient.GetDevice(ctx, &iot.GetDeviceByIDRequest{
		DeviceId: deviceID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Device not found", http.StatusNotFound)
			return
		}
		s.logger.Error("failed to fetch device", "error", err, "device_id", deviceID)
		http.Error(w, "Failed to fetch device", http.StatusInternalServerError)
		return
	}

	// Fetch sensor readings for the device
	readingsResp, err := s.grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
		DeviceId: deviceID,
	})
	if err != nil {
		s.logger.Error("failed to fetch sensor readings", "error", err, "device_id", deviceID)
		http.Error(w, "Failed to fetch sensor readings", http.StatusInternalServerError)
		return
	}

	// Render device detail page
	if err := renderDevice(r.Context(), w, deviceResp.GetDevice(), readingsResp.GetReading()); err != nil {
		s.logger.Error("failed to render device", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleAPIDevices serves the devices list as HTML fragment for htmx.
func (s *Server) handleAPIDevices(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("handling API devices request")

	// Fetch devices from backend
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.grpcClient.GetAllDevice(ctx, &iot.GetAllDevicesRequest{})
	if err != nil {
		s.logger.Error("failed to fetch devices", "error", err)
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	// Render devices list fragment
	if err := renderDevicesList(r.Context(), w, resp.GetDevices()); err != nil {
		s.logger.Error("failed to render devices list", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleAPIDeviceReadings serves the device readings as HTML fragment for htmx.
func (s *Server) handleAPIDeviceReadings(w http.ResponseWriter, r *http.Request) {
	deviceID := r.PathValue("id")
	s.logger.Debug("handling API device readings request", "device_id", deviceID)

	// Get page token from query params
	pageToken := r.URL.Query().Get("page_token")

	// Fetch sensor readings from backend
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := s.grpcClient.GetSensorReadingByDeviceID(ctx, &iot.GetSensorReadingByDeviceIDRequest{
		DeviceId:  deviceID,
		PageToken: pageToken,
	})
	if err != nil {
		s.logger.Error("failed to fetch sensor readings", "error", err, "device_id", deviceID)
		http.Error(w, "Failed to fetch sensor readings", http.StatusInternalServerError)
		return
	}

	// Render readings list fragment
	if err := renderReadingsList(r.Context(), w, resp.GetReading(), resp.GetNextPageToken()); err != nil {
		s.logger.Error("failed to render readings list", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleStatic serves static files.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("handling static file request", "path", r.URL.Path)
	http.Error(w, "Not Found", http.StatusNotFound)
}

// handleHealth serves health check endpoint.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
		s.logger.Error("failed to write health response", "error", err)
	}
}
