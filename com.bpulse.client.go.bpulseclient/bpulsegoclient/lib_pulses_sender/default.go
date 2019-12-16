package lib_pulses_sender

import (
	"io"

	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_send_channel_adapter"

	PROTO "github.com/golang/protobuf/proto"
	MDH_PROTO "github.com/pineda89/bpulse-protobuf-go/com.bpulse.protobuf/go"
)

var PULSERQ_VERSION string

/*
Function used by the sender to build the implementation specific pulse rq.

To provide this function is mandatory for each client,
because, only each specific client knows the
exact type of proto.Message implementation it want to send to pulse.

Parameter of the function: Slice with all of messages which will be encapsulated into a bpulse request.
*/
var DEFAULT_RQMAKER PulseRQMaker = func(pulses []PROTO.Message) (PROTO.Message, error) {
	a_pulses := make([]*MDH_PROTO.Pulse, len(pulses))
	for z, m := range pulses {
		a_pulses[z] = m.(*MDH_PROTO.Pulse)
	}
	rq := &MDH_PROTO.PulsesRQ{Pulse: a_pulses, Version: &PULSERQ_VERSION}
	return rq, nil
}

/*
   Function used to process the bpulse response. it's optional.

   Arguments:
       pulseRS : Message returned by bpulse
       rs_writer : If provided, it's used to save the unmarshaled responses of bpulses
*/
var DEFAULT_RSPROCESSOR PulseRSProcessor = func(pulseRS PROTO.Message, rs_writer io.Writer) error {
	return nil
}

var DEFAULT_RSINSTANCEGETTER lib_send_channel_adapter.PulseRSInstanceGetter = func() PROTO.Message {
	return &MDH_PROTO.PulsesRS{}
}
