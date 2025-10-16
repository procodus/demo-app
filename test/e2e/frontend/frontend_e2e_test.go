package frontend_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Frontend E2E Tests", func() {
	var (
		httpClient *http.Client
		ctx        context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	})

	Describe("Health Check", func() {
		It("should return OK status", func() {
			url := getFrontendURL("/health")
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(ContainSubstring(`"status":"ok"`))
		})
	})

	Describe("Index Page", func() {
		It("should render the index page", func() {
			url := getFrontendURL("/")
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).To(ContainSubstring("text/html"))

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(body)).To(BeNumerically(">", 0))
		})
	})

	Describe("Devices Page", func() {
		Context("with no devices", func() {
			It("should render empty devices list", func() {
				url := getFrontendURL("/devices")
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				Expect(err).NotTo(HaveOccurred())

				resp, err := httpClient.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp.Header.Get("Content-Type")).To(ContainSubstring("text/html"))

				body, err := io.ReadAll(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(body)).To(BeNumerically(">", 0))
			})
		})

		Context("with devices in database", func() {
			BeforeEach(func() {
				// Create test device
				createTestDevice(ctx, "test-device-001")
			})

			It("should render devices list with device data", func() {
				url := getFrontendURL("/devices")
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				Expect(err).NotTo(HaveOccurred())

				resp, err := httpClient.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				body, err := io.ReadAll(resp.Body)
				Expect(err).NotTo(HaveOccurred())

				bodyStr := string(body)
				Expect(bodyStr).To(ContainSubstring("test-device-001"))
				Expect(bodyStr).To(ContainSubstring("Test Location"))
			})
		})
	})

	Describe("Device Detail Page", func() {
		var deviceID string

		BeforeEach(func() {
			deviceID = fmt.Sprintf("test-device-%d-%d", time.Now().Unix(), time.Now().UnixNano()%1000000)
			createTestDevice(ctx, deviceID)
		})

		It("should render device detail page", func() {
			url := getFrontendURL(fmt.Sprintf("/device/%s", deviceID))
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			bodyStr := string(body)
			Expect(bodyStr).To(ContainSubstring(deviceID))
			Expect(bodyStr).To(ContainSubstring("Test Location"))
		})

		It("should return 404 for non-existent device", func() {
			url := getFrontendURL("/device/non-existent-device")
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		Context("with sensor readings", func() {
			BeforeEach(func() {
				// Create test sensor readings
				now := time.Now()
				createTestSensorReading(ctx, deviceID, now.Add(-1*time.Hour))
				createTestSensorReading(ctx, deviceID, now.Add(-30*time.Minute))
				createTestSensorReading(ctx, deviceID, now)
			})

			It("should display sensor readings", func() {
				url := getFrontendURL(fmt.Sprintf("/device/%s", deviceID))
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				Expect(err).NotTo(HaveOccurred())

				resp, err := httpClient.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				body, err := io.ReadAll(resp.Body)
				Expect(err).NotTo(HaveOccurred())

				bodyStr := string(body)
				// Check for temperature value
				Expect(bodyStr).To(ContainSubstring("25.5"))
				// Check for humidity value
				Expect(bodyStr).To(ContainSubstring("65"))
			})
		})
	})

	Describe("API Endpoints (HTMX)", func() {
		Describe("GET /api/devices", func() {
			It("should return devices as HTML fragment", func() {
				// Create test device
				deviceID := fmt.Sprintf("api-device-%d-%d", time.Now().Unix(), time.Now().UnixNano()%1000000)
				createTestDevice(ctx, deviceID)

				url := getFrontendURL("/api/devices")
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				Expect(err).NotTo(HaveOccurred())

				resp, err := httpClient.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				body, err := io.ReadAll(resp.Body)
				Expect(err).NotTo(HaveOccurred())

				bodyStr := string(body)
				Expect(bodyStr).To(ContainSubstring(deviceID))
			})
		})

		Describe("GET /api/device/{id}/readings", func() {
			var deviceID string

			BeforeEach(func() {
				deviceID = fmt.Sprintf("api-reading-device-%d-%d", time.Now().Unix(), time.Now().UnixNano()%1000000)
				createTestDevice(ctx, deviceID)

				// Create sensor readings
				now := time.Now()
				for i := 0; i < 5; i++ {
					createTestSensorReading(ctx, deviceID, now.Add(-time.Duration(i)*time.Minute))
				}
			})

			It("should return sensor readings as HTML fragment", func() {
				url := getFrontendURL(fmt.Sprintf("/api/device/%s/readings", deviceID))
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				Expect(err).NotTo(HaveOccurred())

				resp, err := httpClient.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				body, err := io.ReadAll(resp.Body)
				Expect(err).NotTo(HaveOccurred())

				bodyStr := string(body)
				// Should contain temperature and humidity values
				Expect(bodyStr).To(ContainSubstring("25.5"))
				Expect(bodyStr).To(ContainSubstring("65"))
			})
		})
	})

	Describe("Static Files", func() {
		It("should return 404 for non-existent static files", func() {
			url := getFrontendURL("/static/nonexistent.js")
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("Error Handling", func() {
		It("should handle invalid routes", func() {
			url := getFrontendURL("/nonexistent-route")
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			// Should either redirect to index or return 404
			Expect(resp.StatusCode).To(SatisfyAny(
				Equal(http.StatusNotFound),
				Equal(http.StatusOK), // May redirect to index
			))
		})

		It("should handle POST requests appropriately", func() {
			url := getFrontendURL("/devices")
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader("test"))
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			// Should return method not allowed or similar error
			Expect(resp.StatusCode).NotTo(Equal(http.StatusOK))
		})
	})

	Describe("Content-Type Headers", func() {
		It("should return correct content type for HTML pages", func() {
			url := getFrontendURL("/devices")
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			contentType := resp.Header.Get("Content-Type")
			Expect(contentType).To(ContainSubstring("text/html"))
		})

		It("should return correct content type for health endpoint", func() {
			url := getFrontendURL("/health")
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())

			resp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			contentType := resp.Header.Get("Content-Type")
			Expect(contentType).To(ContainSubstring("application/json"))
		})
	})
})
