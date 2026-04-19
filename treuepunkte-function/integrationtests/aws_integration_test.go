package integrationtests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type balanceResponse struct {
	CustomerID    string `json:"customer_id"`
	ActivePoints  int    `json:"active_points"`
	PendingPoints int    `json:"pending_points"`
}

type transaction struct {
	ID         int     `json:"id"`
	CustomerID string  `json:"customer_id"`
	OrderID    *string `json:"order_id"`
	Reference  *string `json:"reference"`
	ReturnID   *string `json:"return_id"`
	Kind       string  `json:"kind"`
	Status     string  `json:"status"`
	Points     int     `json:"points"`
	OccurredAt string  `json:"occurred_at"`
}

func TestAWSIntegration_AccrueConfirmBalance(t *testing.T) {
	if os.Getenv("RUN_AWS_INTEGRATION") != "1" {
		t.Skip("set RUN_AWS_INTEGRATION=1 to run live AWS integration test")
	}

	baseURL := strings.TrimRight(os.Getenv("AWS_BASE_URL"), "/")
	if baseURL == "" {
		t.Fatal("AWS_BASE_URL is required")
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	uniq := time.Now().UnixNano()
	customerID := fmt.Sprintf("it-aws-cust-%d", uniq)
	orderID := fmt.Sprintf("it-aws-order-%d", uniq)

	t.Logf("customer_id=%s", customerID)
	t.Logf("order_id=%s", orderID)

	{
		resp, body := doRequest(t, client, http.MethodGet, baseURL+"/health", "", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /health expected 200, got %d, body=%s", resp.StatusCode, body)
		}
	}

	{
		payload := map[string]any{
			"customer_id": customerID,
			"order_id":    orderID,
			"points":      120,
		}

		resp, body := doRequest(t, client, http.MethodPost, baseURL+"/v1/points/accrue", "aws-it-accrue-"+fmt.Sprint(uniq), payload)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("POST /v1/points/accrue expected 201, got %d, body=%s", resp.StatusCode, body)
		}
	}

	{
		resp, body := doRequest(t, client, http.MethodGet, baseURL+"/v1/customers/"+customerID+"/balance", "", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /balance after accrue expected 200, got %d, body=%s", resp.StatusCode, body)
		}

		var bal balanceResponse
		if err := json.Unmarshal([]byte(body), &bal); err != nil {
			t.Fatalf("failed to decode balance after accrue: %v, body=%s", err, body)
		}

		if bal.CustomerID != customerID {
			t.Fatalf("unexpected customer_id after accrue: got %s want %s", bal.CustomerID, customerID)
		}
		if bal.ActivePoints != 0 || bal.PendingPoints != 120 {
			t.Fatalf("unexpected balance after accrue: got active=%d pending=%d", bal.ActivePoints, bal.PendingPoints)
		}
	}

	{
		payload := map[string]any{
			"customer_id": customerID,
			"order_id":    orderID,
		}

		resp, body := doRequest(t, client, http.MethodPost, baseURL+"/v1/points/confirm", "aws-it-confirm-"+fmt.Sprint(uniq), payload)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("POST /v1/points/confirm expected 201, got %d, body=%s", resp.StatusCode, body)
		}
	}

	{
		resp, body := doRequest(t, client, http.MethodGet, baseURL+"/v1/customers/"+customerID+"/balance", "", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /balance after confirm expected 200, got %d, body=%s", resp.StatusCode, body)
		}

		var bal balanceResponse
		if err := json.Unmarshal([]byte(body), &bal); err != nil {
			t.Fatalf("failed to decode balance after confirm: %v, body=%s", err, body)
		}

		if bal.CustomerID != customerID {
			t.Fatalf("unexpected customer_id after confirm: got %s want %s", bal.CustomerID, customerID)
		}
		if bal.ActivePoints != 120 || bal.PendingPoints != 0 {
			t.Fatalf("unexpected balance after confirm: got active=%d pending=%d", bal.ActivePoints, bal.PendingPoints)
		}
	}

	{
		resp, body := doRequest(t, client, http.MethodGet, baseURL+"/v1/customers/"+customerID+"/transactions", "", nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /transactions expected 200, got %d, body=%s", resp.StatusCode, body)
		}

		var txs []transaction
		if err := json.Unmarshal([]byte(body), &txs); err != nil {
			t.Fatalf("failed to decode transactions: %v, body=%s", err, body)
		}

		if len(txs) != 2 {
			t.Fatalf("expected 2 transactions, got %d, body=%s", len(txs), body)
		}

		if txs[0].Kind != "accrue" || txs[0].Status != "pending" || txs[0].Points != 120 {
			t.Fatalf("unexpected first transaction: %+v", txs[0])
		}
		if txs[1].Kind != "confirm" || txs[1].Status != "active" || txs[1].Points != 120 {
			t.Fatalf("unexpected second transaction: %+v", txs[1])
		}
	}
}

func doRequest(t *testing.T, client *http.Client, method, url, idemKey string, payload any) (*http.Response, string) {
	t.Helper()

	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idemKey != "" {
		req.Header.Set("Idempotency-Key", idemKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		t.Fatalf("failed to read response body: %v", err)
	}

	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	return resp, string(respBody)
}