package forza

import (
	"errors"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/utils"
)

func RunPlaybackServer(key string, ch chan TelemetryFrame, errCh chan error, dvr utils.DiskvRecorder) {
	// Read the value back out of the store.
	buffer, _ := dvr.Store.Read(key)
	fmt.Printf("%v\n", buffer)
	dataPoints := len(buffer) / PacketSize
	log.DefaultLogger.Info("Starting playback server for Forza Horizon 5", "key", key, "datapoints", dataPoints, "bufferSize", len(buffer))

	i := 0
	previousTimestampMS := uint32(0)
	diff := 17 // ~60 fps
	for i+PacketSize <= len(buffer) {
		packetBuffer := buffer[i : i+PacketSize]
		p, err := ReadPacket(packetBuffer)
		if err != nil {
			log.DefaultLogger.Warn("Some error happened..", "err", err)
			errCh <- err
			return
		}
		i += PacketSize

		if previousTimestampMS > 0 {
			diff = int(p.TimestampMS - previousTimestampMS)
			previousTimestampMS = p.TimestampMS
		}

		time.Sleep(time.Duration(diff) * time.Millisecond)
		ch <- *p
	}
	// raise an error so we stop this stream, and re-start a new one for new subscribers
	errCh <- errors.New("stream playback is finished")
}
