package main

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/aws/aws-lambda-go/lambda"
)

type CfnEvent struct {
	RequestType        string `json:"RequestType"`
	ResponseURL        string `json:"ResponseURL"`
	StackID            string `json:"StackId"`
	RequestID          string `json:"RequestId"`
	LogicalResourceID  string `json:"LogicalResourceId"`
	PhysicalResourceID string `json:"PhysicalResourceId"`
}

type CfnResponse struct {
	Status             string                 `json:"Status"`
	Reason             string                 `json:"Reason"`
	PhysicalResourceID string                 `json:"PhysicalResourceId"`
	StackID            string                 `json:"StackId"`
	RequestID          string                 `json:"RequestId"`
	LogicalResourceID  string                 `json:"LogicalResourceId"`
	Data               map[string]interface{} `json:"Data,omitempty"`
}

// FIX: putanja mora da odgovara stvarnoj lokaciji fajla
//go:embed sql/001_schema.sql
var schemaSQL string

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event CfnEvent) error {
	var resultErr error

	switch event.RequestType {
	case "Delete":
		resultErr = sendCfnResponse(event, "SUCCESS", "delete ignored")
		if resultErr != nil {
			return fmt.Errorf("failed to send CloudFormation delete response: %w", resultErr)
		}
		return nil

	case "Create", "Update":
		resultErr = initializeSchema(ctx)
		if resultErr != nil {
			sendErr := sendCfnResponse(event, "FAILED", resultErr.Error())
			if sendErr != nil {
				return fmt.Errorf("init error: %v; additionally failed to send CloudFormation response: %w", resultErr, sendErr)
			}
			return resultErr
		}

		sendErr := sendCfnResponse(event, "SUCCESS", "schema initialized")
		if sendErr != nil {
			return fmt.Errorf("schema initialized but failed to send CloudFormation response: %w", sendErr)
		}
		return nil

	default:
		resultErr = fmt.Errorf("unsupported request type: %s", event.RequestType)
		sendErr := sendCfnResponse(event, "FAILED", resultErr.Error())
		if sendErr != nil {
			return fmt.Errorf("unsupported request type: %v; additionally failed to send CloudFormation response: %w", resultErr, sendErr)
		}
		return resultErr
	}
}

func initializeSchema(ctx context.Context) error {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")

	if host == "" || port == "" || user == "" || pass == "" || name == "" {
		return fmt.Errorf("missing required database environment variables")
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		user, pass, host, port, name,
	)

	var db *sql.DB
	var err error

	for i := 1; i <= 30; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.PingContext(ctx)
			if err == nil {
				break
			}
			_ = db.Close()
		}

		if i == 30 {
			return fmt.Errorf("database not ready after retries: %w", err)
		}

		time.Sleep(15 * time.Second)
	}

	defer db.Close()

	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

func sendCfnResponse(event CfnEvent, status, reason string) error {
	resp := CfnResponse{
		Status:             status,
		Reason:             reason,
		PhysicalResourceID: physicalID(event),
		StackID:            event.StackID,
		RequestID:          event.RequestID,
		LogicalResourceID:  event.LogicalResourceID,
		Data: map[string]interface{}{
			"Message": reason,
		},
	}

	body, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, event.ResponseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "")
	req.ContentLength = int64(len(body))

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, _ = io.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("cloudformation response returned status %d", res.StatusCode)
	}

	return nil
}

func physicalID(event CfnEvent) string {
	if event.PhysicalResourceID != "" {
		return event.PhysicalResourceID
	}
	return "treuepunkte-schema-init"
}