package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"github.com/a-berahman/shopping-cart/internal/core/domain"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCartService struct {
	mock.Mock
}

func (m *MockCartService) AddItemToCart(ctx context.Context, name string, quantity int) (*domain.Item, error) {
	args := m.Called(ctx, name, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Item), args.Error(1)
}

func (m *MockCartService) ListCartItems(ctx context.Context) ([]domain.Item, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Item), args.Error(1)
}

// create a custom validator here to support failed validation scneario
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func setupTest() (*echo.Echo, *MockCartService, *Handler) {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	mockService := new(MockCartService)
	handler := NewHandler(mockService)
	handler.Register(e)
	return e, mockService, handler
}

func TestAddItem(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockCartService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful item addition",
			requestBody: AddItemRequest{
				Name:     "Test Item",
				Quantity: 1,
			},
			setupMock: func(ms *MockCartService) {
				ms.On("AddItemToCart", mock.Anything, "Test Item", 1).
					Return(&domain.Item{
						ID:       123,
						Name:     "Test Item",
						Quantity: 1,
						Status:   domain.StatusReservationPending,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":123,"name":"Test Item","quantity":1,"status":"PENDING"}`,
		},
		{
			name: "invalid request - missing name",
			requestBody: AddItemRequest{
				Quantity: 1,
			},
			setupMock: func(_ *MockCartService) {
				// no mock setup needed as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"Key: 'AddItemRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"}`,
		},
		{
			name: "service error",
			requestBody: AddItemRequest{
				Name:     "Test Item",
				Quantity: 1,
			},
			setupMock: func(ms *MockCartService) {
				ms.On("AddItemToCart", mock.Anything, "Test Item", 1).
					Return(nil, errors.New("internal server error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockService, h := setupTest()
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBuffer(jsonBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err = h.AddItem(c)
			if tt.expectedStatus >= http.StatusBadRequest {
				assert.Error(t, err)
				he, ok := err.(*echo.HTTPError)
				if ok {
					assert.Equal(t, tt.expectedStatus, he.Code)
					assert.JSONEq(t, tt.expectedBody, fmt.Sprintf(`{"message":%q}`, he.Message))
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var expected, actual map[string]interface{}
				err = json.Unmarshal([]byte(tt.expectedBody), &expected)
				require.NoError(t, err)
				err = json.Unmarshal(rec.Body.Bytes(), &actual)
				require.NoError(t, err)

				delete(actual, "created_at")
				delete(actual, "updated_at")
				assert.Equal(t, expected, actual)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestListItems(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockCartService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful items listing",
			setupMock: func(ms *MockCartService) {
				ms.On("ListCartItems", mock.Anything).
					Return([]domain.Item{
						{
							ID:       123,
							Name:     "Test Item 1",
							Quantity: 1,
							Status:   domain.StatusReservationPending,
						},
						{
							ID:       456,
							Name:     "Test Item 2",
							Quantity: 2,
							Status:   domain.StatusReservationReserved,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `[
				{"id":123,"name":"Test Item 1","quantity":1,"status":"PENDING"},
				{"id":456,"name":"Test Item 2","quantity":2,"status":"RESERVED"}
			]`,
		},
		{
			name: "empty list",
			setupMock: func(ms *MockCartService) {
				ms.On("ListCartItems", mock.Anything).
					Return([]domain.Item{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name: "service error",
			setupMock: func(ms *MockCartService) {
				ms.On("ListCartItems", mock.Anything).
					Return([]domain.Item{}, errors.New("internal server error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, mockService, h := setupTest()
			tt.setupMock(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.ListItems(c)
			if tt.expectedStatus >= http.StatusBadRequest {
				assert.Error(t, err)
				he, ok := err.(*echo.HTTPError)
				if ok {
					assert.Equal(t, tt.expectedStatus, he.Code)
					assert.JSONEq(t, tt.expectedBody, fmt.Sprintf(`{"message":%q}`, he.Message))
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				var expected, actual interface{}
				err = json.Unmarshal([]byte(tt.expectedBody), &expected)
				require.NoError(t, err)
				err = json.Unmarshal(rec.Body.Bytes(), &actual)
				require.NoError(t, err)

				// remove time fields for array of items
				// this is not a good practice but for the sake of this project and the test I will do this
				// in a real application we should use a different approach to handle time fields
				if items, ok := actual.([]interface{}); ok {
					for _, item := range items {
						if m, ok := item.(map[string]interface{}); ok {
							delete(m, "created_at")
							delete(m, "updated_at")
						}
					}
				}
				assert.Equal(t, expected, actual)
			}

			mockService.AssertExpectations(t)
		})
	}
}
