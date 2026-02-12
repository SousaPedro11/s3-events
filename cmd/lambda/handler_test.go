package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"
)

func TestHandler_LogsMessage(t *testing.T) {
	// capture logs
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)
	// construct a minimal S3 event wrapped in SNS
	s3json := `{"Records":[{"eventVersion":"2.1","eventSource":"aws:s3","awsRegion":"us-east-1","eventTime":"2026-02-11T00:00:00.000Z","eventName":"ObjectCreated:Put","s3":{"bucket":{"name":"my-bucket"},"object":{"key":"uploads/foo.jpg","size":123}}}]}`

	// build SNS wrapper programmatically so the inner JSON is properly escaped
	wrapper := map[string]any{
		"Records": []map[string]any{
			{"Sns": map[string]any{"Message": s3json}},
		},
	}
	b, err := json.Marshal(wrapper)
	if err != nil {
		t.Fatalf("failed to marshal sns wrapper: %v", err)
	}

	var event SNSEvent
	if err := json.Unmarshal(b, &event); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}

	if err := handler(context.Background(), event); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"status":"uploaded"`) {
		t.Fatalf("expected log to contain uploaded status, got: %s", out)
	}
	if !strings.Contains(out, "my-bucket/uploads/foo.jpg") {
		t.Fatalf("expected log to contain path, got: %s", out)
	}
}

func TestHandler_MultipleEventTypes(t *testing.T) {
	cases := []struct {
		name      string
		eventName string
		expect    []string
	}{
		{"created_put", "ObjectCreated:Put", []string{"\"status\":\"uploaded\"", "my-bucket/uploads/foo.jpg"}},
		{"removed_delete", "ObjectRemoved:Delete", []string{"\"status\":\"deleted\"", "my-bucket/uploads/foo.jpg"}},
		{"restore_post", "ObjectRestore:Post", []string{"\"status\":\"restoring\"", "my-bucket/uploads/foo.jpg"}},
		{"restore_completed", "ObjectRestore:Completed", []string{"\"status\":\"uploaded\"", "my-bucket/uploads/foo.jpg"}},
		{"replication_completed", "Replication:OperationCompletedReplication", []string{"\"status\":\"replicated\"", "my-bucket/uploads/foo.jpg"}},
		{"acl_put", "ObjectAcl:Put", []string{"\"status\":\"acl_changed\"", "my-bucket/uploads/foo.jpg"}},
		{"lifecycle_expire", "LifecycleExpiration:Delete", []string{"\"status\":\"deleted\"", "my-bucket/uploads/foo.jpg"}},
		{"rr_lost", "ObjectReducedRedundancyLostObject", []string{"\"status\":\"lost\"", "my-bucket/uploads/foo.jpg"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// capture logs
			var buf bytes.Buffer
			old := log.Writer()
			log.SetOutput(&buf)
			defer log.SetOutput(old)

			s3map := map[string]any{
				"Records": []map[string]any{{
					"eventVersion": "2.1",
					"eventSource":  "aws:s3",
					"awsRegion":    "us-east-1",
					"eventTime":    "2026-02-11T00:00:00.000Z",
					"eventName":    tc.eventName,
					"s3": map[string]any{
						"bucket": map[string]any{"name": "my-bucket"},
						"object": map[string]any{"key": "uploads/foo.jpg", "size": 123},
					},
				}},
			}
			s3b, _ := json.Marshal(s3map)
			wrapper := map[string]any{"Records": []map[string]any{{"Sns": map[string]any{"Message": string(s3b)}}}}
			wb, _ := json.Marshal(wrapper)

			var event SNSEvent
			if err := json.Unmarshal(wb, &event); err != nil {
				t.Fatalf("unmarshal event: %v", err)
			}

			if err := handler(context.Background(), event); err != nil {
				t.Fatalf("handler returned error: %v", err)
			}

			out := buf.String()
			for _, want := range tc.expect {
				if !strings.Contains(out, want) {
					t.Fatalf("expected log to contain %q, got: %s", want, out)
				}
			}
		})
	}
}

func TestHandler_FallbackUnknownMessage(t *testing.T) {
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	// SNS message that is not an S3 event
	wrapper := map[string]any{"Records": []map[string]any{{"Sns": map[string]any{"Message": "not-a-json"}}}}
	wb, _ := json.Marshal(wrapper)
	var event SNSEvent
	if err := json.Unmarshal(wb, &event); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}

	if err := handler(context.Background(), event); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "\"event_type\":\"Unknown\"") {
		t.Fatalf("expected Unknown event_type in log, got: %s", out)
	}
}
