package main

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	// acc "github.com/grafana/grafana-starter-datasource-backend/pkg/acc/sharedmemory"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/dirtrally"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/forza"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/utils"

	// iracing "github.com/grafana/grafana-starter-datasource-backend/pkg/iracing/sharedmemory"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
)

var SharedMemoryUpdateInterval = time.Second / 60

// Make sure SimracingTelemetryDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler      = (*SimracingTelemetryDatasource)(nil)
	_ backend.CallResourceHandler   = (*SimracingTelemetryDatasource)(nil)
	_ backend.CheckHealthHandler    = (*SimracingTelemetryDatasource)(nil)
	_ backend.StreamHandler         = (*SimracingTelemetryDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*SimracingTelemetryDatasource)(nil)
)

// NewSimracingTelemetryDatasource creates a new datasource instance.
func NewSimracingTelemetryDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	recordingChan := make(chan bool, 1)
	return &SimracingTelemetryDatasource{recordingChan: recordingChan}, nil
}

// SimracingTelemetryDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type SimracingTelemetryDatasource struct {
	recordingChan chan bool
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSimracingTelemetryDatasource factory function.
func (d *SimracingTelemetryDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SimracingTelemetryDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	WithStreaming bool   `json:"withStreaming"`
	Telemetry     string `json:"telemetry"`
}

func (d *SimracingTelemetryDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	log.DefaultLogger.Info("Query sent", "pCtx", pCtx, "query", query)

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// create data frame response.
	frame := data.NewFrame("response")

	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
		data.NewField("values", nil, []float32{0, 0}),
	)

	// If query called with streaming on then return a channel
	// to subscribe on a client-side and consume updates from a plugin.
	// Feel free to remove this if you don't need streaming for your datasource.
	//streamPath := qm.Telemetry
	//if streamPath == "" {
	//	streamPath = "stream"
	//}
	streamPath := "dirt"
	if qm.WithStreaming {
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: pCtx.DataSourceInstanceSettings.UID,
			Path:      streamPath,
		}
		frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
	}

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

// Resource handler

type simpleJsonResponse struct {
	Message string `json:"message"`
}
type recordingsJsonData struct {
	Recordings []string `json:"recordings"`
}

func (d *SimracingTelemetryDatasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Info("CallResource called", "request", req)
	dsjd := dataSourceJSONData{}
	err := json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &dsjd)
	if err != nil {
		log.DefaultLogger.Error("Error parsing DataSourceInstanceSettings", "error", err)
		return err
	}

	dvr := utils.CreateDiskvRecorder(dsjd.RecordingBasePath, dsjd.RecordingBufferDataPoints, d.recordingChan)
	log.DefaultLogger.Info("Creating DVR", "DataSourceInstanceSettings", dsjd, "BasePath", dvr.Store.BasePath, "RecordingBufferDataPoints", dvr.RecordingBufferDataPoints)

	if req.Path == "recordings" {
		diskvIndexChan := dvr.Store.KeysPrefix("forza", nil)
		var diskvIndex []string
		diskvIndex = append(diskvIndex, "live")
		for v := range diskvIndexChan {
			diskvIndex = append(diskvIndex, v)
		}
		jd := recordingsJsonData{Recordings: diskvIndex}
		resource.SendJSON(sender, jd)
	} else if req.Path == "record" {
		d.recordingChan <- true
		return resource.SendJSON(sender, simpleJsonResponse{Message: "Recording started"})
	}
	return nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *SimracingTelemetryDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (d *SimracingTelemetryDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	status := backend.SubscribeStreamStatusOK
	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

type dataSourceJSONData struct {
	RecordingBasePath         string `json:"recordingBasePath"`
	RecordingBufferDataPoints int    `json:"recordingBufferDataPoints"`
}

// RunStream is called once for any open channel. Results are shared with everyone
// subscribed to the same channel.
func (d *SimracingTelemetryDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)
	dsjd := dataSourceJSONData{}
	err := json.Unmarshal(req.PluginContext.DataSourceInstanceSettings.JSONData, &dsjd)
	if err != nil {
		log.DefaultLogger.Error("Error parsing DataSourceInstanceSettings", "error", err)
		return err
	}
	dvr := utils.CreateDiskvRecorder(dsjd.RecordingBasePath, dsjd.RecordingBufferDataPoints, d.recordingChan)
	split := strings.Split(req.Path, "/")
	gamePath := split[0]

	recordingPath := ""
	if len(split) > 1 {
		recordingPath = split[1]
	}

	telemetryChan := make(chan dirtrally.TelemetryFrame)
	telemetryErrorChan := make(chan error)

	// accTelemetryChan := make(chan acc.ACCTelemetry)
	// accCtrlChan := make(chan string)

	// iracingTelemetryChan := make(chan iracing.IRacingTelemetryMap)
	// iracingCtrlChan := make(chan string)

	forzaTelemetryChan := make(chan forza.TelemetryFrame)
	forzaTelemetryErrorChan := make(chan error)

	forzaPlaybackTelemetryChan := make(chan forza.TelemetryFrame)
	forzaPlaybackTelemetryErrorChan := make(chan error)

	if gamePath == "dirtRally2" {
		go dirtrally.RunTelemetryServer(telemetryChan, telemetryErrorChan)
		// } else if gamePath == "acc" {
		// 	//go udpclient.RunClient(telemetryErrorChan)
		// 	go acc.RunSharedMemoryClient(accTelemetryChan, accCtrlChan, SharedMemoryUpdateInterval)
		// } else if gamePath == "iRacing" {
		// 	go iracing.RunSharedMemoryClient(iracingTelemetryChan, iracingCtrlChan, SharedMemoryUpdateInterval)
	} else if gamePath == "forzaHorizon5" {
		log.DefaultLogger.Info("Forza Horizon 5 MODE", "recordingPath", recordingPath)
		if recordingPath != "" {
			go forza.RunPlaybackServer(recordingPath, forzaPlaybackTelemetryChan, forzaPlaybackTelemetryErrorChan, dvr)
		} else {
			go forza.RunTelemetryServer(forzaTelemetryChan, forzaTelemetryErrorChan, dvr)
		}
	}
	lastTimeSent := time.Now()

	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case err := <-forzaPlaybackTelemetryErrorChan:
			log.DefaultLogger.Info("Playback error", "err", err)
			return nil

		case <-ctx.Done():
			log.DefaultLogger.Info("Context done, finish streaming", "path", req.Path)
			// if req.Path == "acc" {
			// 	accCtrlChan <- "stop"
			// } else if req.Path == "iRacing" {
			// 	iracingCtrlChan <- "stop"
			// }
			return nil

		case telemetryFrame := <-telemetryChan:
			if time.Now().Before(lastTimeSent.Add(time.Second / 60)) {
				// Drop frame
				continue
			}

			frame := dirtrally.TelemetryToDataFrame(telemetryFrame)
			lastTimeSent = time.Now()
			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}

		case telemetryFrame := <-forzaTelemetryChan:
			if time.Now().Before(lastTimeSent.Add(time.Second / 60)) {
				// Drop frame
				continue
			}

			frame := forza.TelemetryToDataFrame(telemetryFrame)
			lastTimeSent = time.Now()
			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}

		case telemetryFrame := <-forzaPlaybackTelemetryChan:
			if time.Now().Before(lastTimeSent.Add(time.Second / 60)) {
				// Drop frame
				continue
			}

			frame := forza.TelemetryToDataFrame(telemetryFrame)
			lastTimeSent = time.Now()
			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}

			// case mmapFrame := <-accTelemetryChan:
			// 	// Add a throttling for smooth animations
			// 	if time.Now().Before(lastTimeSent.Add(time.Second / 60)) {
			// 		// Drop frame
			// 		continue
			// 	}

			// 	frame, err := acc.ACCTelemetryToDataFrame(mmapFrame)
			// 	if err != nil {
			// 		log.DefaultLogger.Debug("Error converting telemetry frame", "error", err)
			// 		continue
			// 	}

			// 	lastTimeSent = time.Now()
			// 	err = sender.SendFrame(frame, data.IncludeAll)
			// 	if err != nil {
			// 		log.DefaultLogger.Error("Error sending frame", "error", err)
			// 		continue
			// 	}

			// case telemetryFrame := <-iracingTelemetryChan:
			// 	// Add a throttling for smooth animations
			// 	if time.Now().Before(lastTimeSent.Add(time.Second / 60)) {
			// 		// Drop frame
			// 		continue
			// 	}

			// 	frame, err := iracing.TelemetryToDataFrame(telemetryFrame)
			// 	if err != nil {
			// 		log.DefaultLogger.Debug("Error converting telemetry frame", "error", err)
			// 		continue
			// 	}

			// 	lastTimeSent = time.Now()
			// 	err = sender.SendFrame(frame, data.IncludeAll)
			// 	if err != nil {
			// 		log.DefaultLogger.Error("Error sending frame", "error", err)
			// 		continue
			// 	}
		}
	}
}

// PublishStream is called when a client sends a message to the stream.
func (d *SimracingTelemetryDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}
