/*
* Copyright 2019 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package outputs

import (
	"encoding/json"
	"fmt"

	"github.com/newrelic/nri-flex/internal/load"
	"github.com/sirupsen/logrus"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
)

// SendToInsights - Send processed events to insights
// loop through integration entities as there could be multiple that have been set
// when posted they are batched by entity
func SendToInsights() {
	for _, entity := range load.Integration.Entities {
		// split the payload into smaller batches so that the payload size does not exceed the Insight endpoint size limit
		batchSize := load.Args.InsightBatchSize
		sizeMetrics := len(entity.Metrics)
		batches := sizeMetrics/batchSize + 1

		modifyEventType(entity)

		for i := 0; i < batches; i++ {
			start := i * batchSize
			end := start + batchSize
			if end > sizeMetrics {
				end = sizeMetrics
			}
			thisBatch := (entity.Metrics)[start:end]
			jsonData, err := json.Marshal(thisBatch)
			if err != nil {
				load.Logrus.WithFields(logrus.Fields{
					"err": err,
				}).Error("insights: failed to marshal json")
				break
			} else {
				load.Logrus.Debug(fmt.Sprintf("posting %d events to insights", len(thisBatch)))
				postRequest(load.Args.InsightsURL, load.Args.InsightsAPIKey, jsonData)
				if load.Args.InsightsOutput {
					fmt.Println(string(jsonData))
				}
			}

		}
		// empty the infrastructure entity metrics by default
		entity.Metrics = []*metric.Set{}
	}
}

// modifyEventType insights uses eventType key in camel case whereas infrastructure uses event_type
func modifyEventType(entity *integration.Entity) {
	for _, event := range entity.Metrics {
		if event.Metrics["event_type"] != nil {
			event.Metrics["eventType"] = event.Metrics["event_type"].(string)
		}
		delete(event.Metrics, "event_type")
	}
}

// // postRequest wraps request and attaches needed headers and zlib compression
// func postRequest(entity *integration.Entity) {
// 	jsonData, err := json.Marshal(entity.Metrics)
// 	if err != nil {
// 		logger.Flex("error", err, "failed to marshal", false)
// 	} else {
// 		var zlibCompressedPayload bytes.Buffer
// 		w := zlib.NewWriter(&zlibCompressedPayload)
// 		_, err := w.Write(jsonData)
// 		logger.Flex("error", err, "unable to write zlib compressed form", false)
// 		logger.Flex("error", w.Close(), "unable to close zlib writer", false)
// 		if err != nil {
// 			logger.Flex("error", fmt.Errorf("failed to compress payload"), "", false)
// 		} else {
// 			tr := &http.Transport{IdleConnTimeout: 15 * time.Second}
// 			client := &http.Client{Transport: tr}
// 			req, err := http.NewRequest("POST", load.Args.InsightsURL, bytes.NewBuffer(zlibCompressedPayload.Bytes()))
// 			logger.Flex("debug", nil, fmt.Sprintf("insights: bytes %d events %d", len(zlibCompressedPayload.Bytes()), len(load.Entity.Metrics)), false)

// 			if err != nil {
// 				logger.Flex("error", err, "unable to create http.Request", false)
// 			} else {
// 				req.Header.Set("Content-Encoding", "deflate")
// 				req.Header.Set("Content-Type", "application/json")
// 				req.Header.Set("X-Insert-Key", load.Args.InsightsAPIKey)
// 				_, err := client.Do(req)
// 				logger.Flex("error", err, "send failure", false)
// 			}
// 		}
// 	}
// }
