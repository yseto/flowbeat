package beater

import (
	"bytes"
	"net"
	"time"

	"github.com/fstelzer/sflow"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
)

type Flowbeat struct {
	FbConfig ConfigSettings
	events   publisher.Client

	listen string
	conn   *net.UDPConn

	done chan struct{}
}

func New() *Flowbeat {
	return &Flowbeat{}
}

func (fb *Flowbeat) Config(b *beat.Beat) error {

	err := cfgfile.Read(&fb.FbConfig, "")
	if err != nil {
		logp.Err("Error reading configuration file: %v", err)
		return err
	}

	if fb.FbConfig.Input.Listen != nil {
		fb.listen = *fb.FbConfig.Input.Listen
	} else {
		fb.listen = ":6343"
	}

	logp.Debug("flowbeat", "Init flowbeat")
	logp.Debug("flowbeat", "Listening on %s\n", fb.listen)

	return nil
}

func (fb *Flowbeat) Setup(b *beat.Beat) error {
	fb.events = b.Events
	fb.done = make(chan struct{})

	addr, err := net.ResolveUDPAddr("udp", fb.listen)
	if err != nil {
		return err
	}
	fb.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	return nil
}

func (fb *Flowbeat) Run(b *beat.Beat) error {
	var err error
	packetbuffer := make([]byte, 65535)

	for {
		select {
		case <-fb.done:
			return nil
		default:
		}

		// Listen for sflow datagrams
		size, _, err := fb.conn.ReadFromUDP(packetbuffer)
		logp.Debug("flowbeat", "Received UDP Packet with Size: %d", size)
		if err != nil {
			return err
		}

		reader := bytes.NewReader(packetbuffer)
		decoder := sflow.NewDecoder(reader)
		dgram, err := decoder.Decode()
		if err != nil {
			logp.Warn("Error decoding sflow packet: %s", err)
			continue
		}

		for _, sample := range dgram.Samples {
			var sampleType string
			switch sample.SampleType() {
			case sflow.TypeFlowSample:
				sampleType = "flow"
			case sflow.TypeCounterSample:
				sampleType = "counter"
			case sflow.TypeExpandedFlowSample:
				sampleType = "extended_flow"
			case sflow.TypeExpandedCounterSample:
				sampleType = "extended_counter"
			default:
				sampleType = "unknown"
			}

			//TODO: Sanitize / Beautify / Convert some of the sample data here for easier analytics
			event := common.MapStr{
				"@timestamp": common.Time(time.Now()),
				"type":       sampleType,
				"sflowdata":  sample,
			}

			fb.events.PublishEvent(event)
		}
	}

	return err
}

func (fb *Flowbeat) Cleanup(b *beat.Beat) error {
	if fb.conn != nil {
		fb.conn.Close()
	}
	return nil
}

func (fb *Flowbeat) Stop() {
	close(fb.done)
}