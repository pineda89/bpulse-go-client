/*
Copyright 2015 - Serhstourism, S.A

Author Jaume Arús - jaumearus@gmail.com

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/example/pulse/go"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulses_sender"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"

	AGG_CONTAINER "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_aggregator"
	AGG_CONTENT "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_aggregator/contents"

	"github.com/golang/protobuf/proto"

	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_send_channel_adapter"
)

func init() {

	// Set PulseRQ version
	pulserq_version = lib_gc_config.ConfigurationProvider.GetPropertySTRING("DEFAULT_SENDER", "DEFAULT_SENDER_PULSERQ_VERSION")

	// Take pulse type id from bpulseid commmand line flag
	flag.StringVar(&bpulseid, "bpulseid", "", "The bpulse typeId")
	flag.Parse()
	if bpulseid == "" {
		panic("The pbulse Type ID must be specified: ./example -bpulseid=bpulse_serhs_MDHPLTEST")
	} else {
		GLOBAL_PULSE_TYPEID = bpulseid
		fmt.Println("BPulse typeId: ", GLOBAL_PULSE_TYPEID)
	}

	var err error
	if hostname, err = os.Hostname(); err != nil {
		panic(err)
	}

	once = &sync.Once{}

	// Set GOMAXPROCS
	maxCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(maxCpu)
}

var GLOBAL_PULSE_TYPEID string // Pulse type ID
const (
	GLOBAL_AGGREGATOR_NAME  = "DEFAULT"    // Aggregator implementation to use with
	GLOBAL_ENTITY_LABEL     = "entity"     // Pulse entity field
	GLOBAL_PK_LABEL         = "pk"         // Pulse pk field
	GLOBAL_ACTION_LABEL     = "action"     // Pulse action field
	GLOBAL_CHECKPOINT_LABEL = "checkpoint" // Pulse checkpoint field
	GLOBAL_DELAY_LABEL      = "delay"      // Pulse delay field
)

// Pulse sending loop delay between cicles in milliseconds
const LOOP_DELAY_IN_MS int = 10

// bpulse type id filled from 'bpulseid' command line flag
var bpulseid string

// Pulse value's labels
var hostname string
var pulserq_version string

// Pulse sender RS processor
var once *sync.Once
var rs_file *os.File

// Send pulses flag
var send_pulses = true

/*
   Helper method to build a pulse rq message based on our example pulse rq message.
*/
func HCreatePulse(checkpoint, entity, pk, action string, delay int64) (*collector.Pulse, error) {

	// Create the Values of the pulse
	ventity := collector.Value{Name: proto.String(GLOBAL_ENTITY_LABEL), Values: []string{entity}}
	vpk := collector.Value{Name: proto.String(GLOBAL_PK_LABEL), Values: []string{pk}}
	vaction := collector.Value{Name: proto.String(GLOBAL_ACTION_LABEL), Values: []string{action}}
	vchkpoint := collector.Value{Name: proto.String(GLOBAL_CHECKPOINT_LABEL), Values: []string{checkpoint}}
	vdelay := collector.Value{Name: proto.String(GLOBAL_DELAY_LABEL), Values: []string{strconv.FormatInt(delay, 10)}}

	// Create the pulse
	t := time.Now().UnixNano() / 1000000
	pulse := collector.Pulse{Time: proto.Int64(t), InstanceId: proto.String(hostname), TypeId: proto.String(GLOBAL_PULSE_TYPEID), Values: []*collector.Value{&ventity, &vpk, &vaction, &vchkpoint, &vdelay}}
	return &pulse, nil
}

/*
   Function used by the sender to build the implementation specific pulse rq.

   To provide this function is mandatory for each client,
   because, only each specific client knows the
   exact type of proto.Message implementation it want to send to pulse.

   Parameter of the function: Slice with all of messages which will be encapsulated into a bpulse request.
*/
var rq_maker lib_pulses_sender.PulseRQMaker = func(pulses []proto.Message) (proto.Message, error) {
	a_pulses := make([]*collector.Pulse, len(pulses))
	for z, m := range pulses {
		a_pulses[z] = m.(*collector.Pulse)
	}
	rq := &collector.PulsesRQ{Pulse: a_pulses, Version: &pulserq_version}
	return rq, nil
}

/*
   Function used to process the bpulse response. it's optional.

   Arguments:
       pulseRS : Message returned by bpulse
       rs_writer : If provided, it's used to save the unmarshaled responses of bpulses
*/
var rs_processor lib_pulses_sender.PulseRSProcessor = func(pulseRS proto.Message, rs_writer io.Writer) error {

	open_file := func() {

		// With response writer. Using a writer to store the responses asynchronously (It's optional)
		var err error
		if rs_file, err = os.OpenFile("/tmp/BPulse_responses.txt", os.O_CREATE, 0777); err != nil {
			panic(err)
		} else {

			defer rs_file.Close()

		}
	}
	once.Do(open_file)

	lib_gc_log.Trace.Printf("Process response %+v\n", pulseRS.String())
	rs_file.Write([]byte(fmt.Sprintf("%+v\n", pulseRS)))
	rs_file.Sync()
	return nil
}

var rs_instance_getter lib_send_channel_adapter.PulseRSInstanceGetter = func() proto.Message {
	return &collector.PulsesRS{}
}

func handleSingnals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		shutdown()
	}()
}

func shutdown() {
	fmt.Println("Shutdown ....")

	send_pulses = false

	errors_list := bpulsegoclient.ShutdownBPulseClient()
	for e := errors_list.Front(); e != nil; e = e.Next() {
		lib_gc_log.Error.Println(e.Value.(error).Error())
	}
}

func main() {

	// Cathc shutdown signals
	handleSingnals()

	// Start the Go BPulseClient
	lib_gc_log.Info.Println("Starting Go BPulse client ...")
	if err := bpulsegoclient.StartBPulseClient(rq_maker, rs_processor, rs_instance_getter); err != nil {
		panic(err)
	}
	lib_gc_log.Info.Println("Go BPulse client started.")

	// Simulate app logic ....
	lib_gc_log.Info.Println("Sending pulses each ", fmt.Sprint(LOOP_DELAY_IN_MS), " milliseconds...")
	checkpoints := []string{"DBAL", "ALKF", "KFML", "MLRD"}
	entities := []string{"CNPRECIO", "CNEDAD", "NEWPAROS", "CNCLIENTE", "ALIASOPACO", "CNDISTRIBUCION", "CNCONCEPTO"}
	action := []string{"SET", "UNSET"}
	delays := []int64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	z := 0
	for send_pulses {
		go func() {
			t := time.Now()
			if pulse, err := HCreatePulse(checkpoints[z%len(checkpoints)], entities[z%len(entities)], "MyPk", action[z%len(action)], delays[z%len(delays)]); err != nil {
				panic(err)
			} else {

				/**
				 * There are several way to use the aggregators:
				 *
				 * 1.- The first option is taking an aggregator once, and to use it for all of the application.
				 *     In this case, we'll use always the same aggregator. By this way, we'll have only a unique block of pulse to send.
				 *     It is necessary to keep in mind that the input channel of the aggregator is buffered to the Pulse RQ block size.
				 *     Therefore, if we think there will be more concurrent pulse than the block size, this option is not a good idea
				 *     unless we expand the size of the pulse RQ size.
				 *
				 * 2.- Another option is to take an aggregator from the pool every time we need it. In this case, we must to understand
				 *     that each aggregator has its own slice of pulses, and that, until it not have as many pulses as RQ size,
				 *     the aggregator does not put all of them into a PulseRQ to send it.
				 */

				// Take the pulses aggregator.
				if aggregator, err := AGG_CONTAINER.Container.GetPulseAggregator(AGG_CONTENT.DEFAULT_AGGREGATOR_NAME); err != nil {
					panic(err)
				} else {

					// Take the aggregator channel
					if ch, err := aggregator.GetPulsesAggregationChannel(); err != nil {
						panic(err)
					} else {

						// Send the pulse
						//_ = ch
						//_ = pulse
						ch <- pulse

						// Each 1000 pulses, write a log
						//if z%1000 == 0 {
						//event := &lib_bpulse_event.Event_message{Type: lib_bpulse_event.NEWS, Title: "Message sent", Description: "Sent message " + strconv.FormatInt(int64(z), 10), StartDate: time.Now().Unix(), EndDate: time.Now().Unix()}
						//lib_bpulse_event.Messages_chan <- event
						//}

						// Restore the aggregator to the aggregators pool
						AGG_CONTAINER.Container.ReleasePulseAggregator(AGG_CONTENT.DEFAULT_AGGREGATOR_NAME, aggregator)
					}
				}

			}
			_ = t
			//lib_gc_log.Trace.Println("Time to aggregate the pulse: ", fmt.Sprint(time.Now().Sub(t)))
		}()
		time.Sleep(time.Duration(LOOP_DELAY_IN_MS) * time.Millisecond)
		z = z + 1

	}

}
