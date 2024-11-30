package reservation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Service struct {
	baseURL    string
	httpClient *http.Client
}

type reservationRequest struct {
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

type reservationResponse struct {
	ReservationID string `json:"reservation_id"`
	Available     bool   `json:"available"`
}

func NewService(baseURL string) *Service {
	return &Service{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Service) CheckAvailability(ctx context.Context, itemName string, quantity int) (bool, error) {
	reqBody, err := json.Marshal(reservationRequest{
		Item:     itemName,
		Quantity: quantity,
	})
	if err != nil {
		return false, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/availability", s.baseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error checking availability: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response reservationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("error decoding response: %w", err)
	}

	return response.Available, nil
}

func (s *Service) ReserveItem(ctx context.Context, itemName string, quantity int) (string, error) {
	reqBody, err := json.Marshal(reservationRequest{
		Item:     itemName,
		Quantity: quantity,
	})
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/reserve", s.baseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error reserving item: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response reservationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	return response.ReservationID, nil
}
