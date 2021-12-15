package forza

import (
	"fmt"
	"net"

	"github.com/gofrs/uuid"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/utils"
	"github.com/smallnest/ringbuffer"
)

const (
	serverPort = "20777"
)

func RunTelemetryServer(ch chan TelemetryFrame, errCh chan error, dvr utils.DiskvRecorder) {
	port := fmt.Sprintf(":%s", serverPort)
	s, err := net.ResolveUDPAddr("udp4", port)
	if err != nil {
		errCh <- err
		return
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		errCh <- err
		return
	}
	log.DefaultLogger.Info("Starting telemetry server for Forza Horizon 5")

	defer connection.Close()
	buffer := make([]byte, 1024)
	// number of datapoints to store (60 fps * 10 minutes)
	rb := ringbuffer.New(PacketSize * int(dvr.RecordingBufferDataPoints))

	// Initialize a new diskv store, rooted at "my-data-dir", with a 1MB cache.
	log.DefaultLogger.Info("Storing data with Diskv", "d", dvr.Store.BasePath, "cacheSize", dvr.RecordingBufferDataPoints)

	for {
		n, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			errCh <- err
			return
		}
		// log.DefaultLogger.Info("Read bytes", "n", n)

		packetBuffer := buffer[0:n]
		rb.Write(packetBuffer)
		p, err := ReadPacket(packetBuffer)
		// log.DefaultLogger.Info("Ringbuffer state", "total", rb.Length(), "free", rb.Free())

		if err != nil {
			errCh <- err
			return
		}
		// log.DefaultLogger.Info("Packet read")
		select {
		case doRecord := <-dvr.RecordingChan:
			if doRecord {
				log.DefaultLogger.Info("Recording called!", "doRecord", doRecord)
				u, err := uuid.NewV4()
				key := "forza-" + u.String()
				if err != nil {
					errCh <- err
					return
				}
				log.DefaultLogger.Info("Dumping ringbuffer to file", "key", key, "total", rb.Length())
				dvr.Store.Write(key, rb.Bytes())
				// rb.Reset()
				dvr.RecordingChan <- false
			}
		default:
		}

		ch <- *p
	}
}
