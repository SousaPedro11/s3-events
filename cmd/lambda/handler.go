package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

type SNSEvent struct {
	Records []struct {
		Sns struct {
			Message string `json:"Message"`
		} `json:"Sns"`
	} `json:"Records"`
}

type ESCommand struct {
	Action string `json:"action"`
	Status string `json:"status"`
}

type ESEvent struct {
	Source      string    `json:"source"`
	Aggregate   string    `json:"aggregate"`
	AggregateID string    `json:"aggregate_id"`
	EventType   string    `json:"event_type"`
	Command     ESCommand `json:"command"`
	Path        string    `json:"path"`
	Timestamp   string    `json:"timestamp"`
	Raw         any       `json:"raw,omitempty"`
}

func handler(ctx context.Context, event SNSEvent) error {
	for _, record := range event.Records {
		msg := record.Sns.Message

		var s3events events.S3Event
		if err := json.Unmarshal([]byte(msg), &s3events); err != nil {
			ev := ESEvent{
				Source:    "s3-events-handler",
				Aggregate: "file",
				EventType: "Unknown",
				Command: ESCommand{
					Action: "noop",
					Status: "unknown",
				},
				Path:      "",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Raw:       msg,
			}
			b, _ := json.Marshal(ev)
			log.Println(string(b))
			continue
		}

		for _, rec := range s3events.Records {
			bucket := rec.S3.Bucket.Name
			key := rec.S3.Object.Key
			eventName := rec.EventName

			status := "unknown"
			cmdAction := "set_status"
			if strings.HasPrefix(eventName, "ObjectCreated") {
				status = "uploaded"
			} else if strings.HasPrefix(eventName, "ObjectRemoved") {
				status = "deleted"
			} else if strings.HasPrefix(eventName, "ObjectRestore") {
				if strings.HasSuffix(eventName, "Completed") {
					status = "uploaded"
				} else {
					status = "restoring"
				}
			} else if strings.HasPrefix(eventName, "Replication") {
				status = "replicated"
			} else if strings.HasPrefix(eventName, "ObjectAcl") {
				status = "acl_changed"
			} else if strings.HasPrefix(eventName, "LifecycleExpiration") {
				status = "deleted"
			} else if strings.HasPrefix(eventName, "ObjectReducedRedundancyLostObject") {
				status = "lost"
			}

			path := bucket + "/" + key

			ev := ESEvent{
				Source:      "s3-events-handler",
				Aggregate:   "file",
				AggregateID: bucket + ":" + key,
				EventType:   eventName,
				Command: ESCommand{
					Action: cmdAction,
					Status: status,
				},
				Path:      path,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Raw:       rec,
			}

			b, err := json.Marshal(ev)
			if err != nil {
				log.Printf("failed to marshal event: %v", err)
				continue
			}
			log.Println(string(b))
		}
	}
	return nil
}
