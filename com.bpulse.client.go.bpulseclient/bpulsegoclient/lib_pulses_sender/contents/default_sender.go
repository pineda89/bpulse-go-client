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

package lib_pulses_sender

import (
	"io"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_storage"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulses_sender"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_send_channel_adapter"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container/lib_gc_contents_loader"
	EVENT "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_event"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_panic_catching"
	"github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_semaphore"
)

func init() {
	wg := lib_pulses_sender.Container.(*lib_pulses_sender.PulseSenderContainer).Container.GetWg()
	wg.Add(1)
	lib_gc_contents_loader.AddInitiator("DEFAULT_SENDER", Initialize)
	once = &sync.Once{}
	once_main = &sync.Once{}
	once_rq_rs_functions = &sync.Once{}
	defer wg.Done()
}

const DEFAULT_SENDER = "DEFAULT"

var adapter_name string
var max_concurrent_sents_semaphore lib_gc_semaphore.Semaphore_I
var max_concurrent_records_semaphore lib_gc_semaphore.Semaphore_I
var once, once_main, once_rq_rs_functions *sync.Once
var msg_sending_reintries int

func Initialize() error {
	if lib_pulses_sender.Container.IsPulseSenderActivated(DEFAULT_SENDER) {

		// Set the maximum number of concurrent RQ being sent to BPulse
		max_concurrent_sents := lib_gc_config.ConfigurationProvider.GetPropertyINT("DEFAULT_SENDER", "DEFAULT_SENDER_MAX_CONCURRENT_SENTS")
		max_concurrent_sents_semaphore = lib_gc_semaphore.SemaphoreFactory.GetSemaphore(max_concurrent_sents)

		// Set the maximum number of concurrent Records (blocks of pulses) processed concurrently
		max_concurrent_records := lib_gc_config.ConfigurationProvider.GetPropertyINT("DEFAULT_SENDER", "DEFAULT_SENDER_MAX_CONCURRENT_RECORDS")
		max_concurrent_records_semaphore = lib_gc_semaphore.SemaphoreFactory.GetSemaphore(max_concurrent_records)

		msg_sending_reintries = lib_gc_config.ConfigurationProvider.GetPropertyINT("DEFAULT_SENDER", "DEFAULT_SENDER_NUMBER_OF_PULSE_SENDING_REINTRIES")

		// Instantiate the sender
		sender := &DefaultSender{isrunning: false}
		sender.Contents = lib_gc_container.Contents{Container: lib_pulses_sender.Container.(*lib_pulses_sender.PulseSenderContainer).Container, Status: lib_gc_container.CONTAINER_STATUS_STOPPED, StatusListener: sender}

		// Add the new sender to the container
		lib_pulses_sender.Container.AddPulseSender(DEFAULT_SENDER, sender)

		// Set the name of the channel adapter used to send the pulses to BPulse
		adapter_name = lib_gc_config.ConfigurationProvider.GetPropertySTRING("DEFAULT_SENDER", "DEFAULT_SENDER_CHANNEL_ADAPTER")

		lib_gc_log.Info.Printf("%s sender initialized.\n", DEFAULT_SENDER)
	}
	return nil
}

type DefaultSender struct {
	lib_gc_container.Contents
	storage            lib_pulse_storage.IPulsesStorageAdapter
	rs_writer          io.Writer
	pulse_rq_maker     lib_pulses_sender.PulseRQMaker
	pulse_rs_processor lib_pulses_sender.PulseRSProcessor
	isrunning          bool
}

func (ds *DefaultSender) Shutdown() error {
	lib_gc_log.Info.Printf("SENDER content [%p, %s], container [%p, %s] SHUTDOWN catched. \n", &ds.Contents, DEFAULT_SENDER, ds.Container, ds.Container.GetName())
	return nil
}

func (ds *DefaultSender) Start() error {
	lib_gc_log.Info.Printf("SENDER content [%p, %s], container [%p, %s] START catched. \n", &ds.Contents, DEFAULT_SENDER, ds.Container, ds.Container.GetName())

	// The component's dependencies must be satisfied only the first time
	// this component is executed.
	dependencies_solver := func() {
		// Set the storage from which the sender will take the records (blocks of pulses) to send to BPulse
		storage_name := lib_gc_config.ConfigurationProvider.GetPropertySTRING("DEFAULT_SENDER", "DEFAULT_SENDER_STORAGE")

		if storage, err := lib_pulse_storage.Container.GetPulseStorageAdapter(storage_name); err != nil {
			panic(err)
		} else {
			ds.storage = storage
		}
	}
	once.Do(dependencies_solver)

	rq_rs_functions := func() {
		// Set the rq maker
		if rq_maker, err := ds.Contents.Container.GetParameter(lib_pulses_sender.PULSE_RQ_MAKER); err != nil {
			lib_gc_log.Error.Println(err.Error())
		} else {
			ds.pulse_rq_maker = rq_maker.(lib_pulses_sender.PulseRQMaker)
		}

		// Set the rs processor
		if rs_processor, err := ds.Contents.Container.GetParameter(lib_pulses_sender.PULSE_RS_PROCESSOR); err != nil {
			lib_gc_log.Warning.Println(err.Error())
		} else {
			ds.pulse_rs_processor = rs_processor.(lib_pulses_sender.PulseRSProcessor)
		}
	}
	once_rq_rs_functions.Do(rq_rs_functions)

	// The main algorithm of the sender, must be started once.
	sender_main := func() {
		go ds.checkForRunning()
	}
	once_main.Do(sender_main)

	return nil
}

func (ds *DefaultSender) Stop() error {
	lib_gc_log.Info.Printf("SENDER content [%p, %s], container [%p, %s] STOP catched.\n", &ds.Contents, DEFAULT_SENDER, ds.Container, ds.Container.GetName())
	return nil
}

func (ds *DefaultSender) SetPulseRQMaker(rq_maker lib_pulses_sender.PulseRQMaker) error {
	ds.pulse_rq_maker = rq_maker
	return nil
}

func (ds *DefaultSender) SetPulseRSProcessor(rs_processor lib_pulses_sender.PulseRSProcessor) error {
	ds.pulse_rs_processor = rs_processor
	return nil
}

func (ds *DefaultSender) writeToRSWriter(t []byte) {
	if ds.rs_writer != nil {
		ds.rs_writer.Write(t)
	}
}

func (ds *DefaultSender) checkForRunning() {
	ticker := time.NewTicker(time.Millisecond * 500)
	for t := range ticker.C {
		_ = t
		if !ds.isrunning {
			go ds.run()
		}
	}
}

func (ds *DefaultSender) run() {
	// panic catching
	defer func() {
		lib_gc_panic_catching.PanicCatching("DefaultSender.run")
		ds.isrunning = false
	}()

	lib_gc_log.Info.Printf("SENDER content [%p, %s], container [%p, %s] RUN STARTED.\n", &ds.Contents, DEFAULT_SENDER, ds.Container, ds.Container.GetName())
	for {
		ds.isrunning = true
		if ds.Contents.Status == lib_gc_container.CONTAINER_STATUS_STARTED && ds.storage != nil {
			max_concurrent_records_semaphore.Lock(1)
			if record, err := ds.storage.GetPulseRecord(); err != nil {
				msg, _ := EVENT.NotifyEvent("006-001", "", &[]string{err.Error()})
				lib_gc_log.Error.Printf("%s\n", msg)
				ds.writeToRSWriter([]byte(msg))
			} else {
				if record != nil {
					lib_gc_log.Trace.Printf("SENDER content [%p, %s], container [%p, %s] Taken pulse record, with %d pulses.\n", &ds.Contents, DEFAULT_SENDER, ds.Container, ds.Container.GetName(), len(record.Pulses))
					if ds.pulse_rq_maker == nil {
						msg, _ := EVENT.NotifyEvent("006-004", "", &[]string{DEFAULT_SENDER})
						lib_gc_log.Error.Println(msg)
					} else {
						if rq, err := ds.pulse_rq_maker(record.Pulses); err != nil {
							msg, _ := EVENT.NotifyEvent("006-005", "", &[]string{DEFAULT_SENDER})
							lib_gc_log.Error.Println(msg)
						} else {

							go func(rq proto.Message, rs_processor lib_pulses_sender.PulseRSProcessor, semaphore lib_gc_semaphore.Semaphore_I) {
								// panic catching
								defer lib_gc_panic_catching.PanicCatching("DefaultSender.run - go")
								if adapter, err := lib_send_channel_adapter.Container.GetSendChannelAdapter(adapter_name); err != nil {
									msg, _ := EVENT.NotifyEvent("006-002", "", &[]string{err.Error()})
									lib_gc_log.Error.Printf("%s\n", msg)
									ds.writeToRSWriter([]byte(msg))
								} else {
									defer lib_send_channel_adapter.Container.ReleaseSendChannelAdapter(adapter_name, adapter)
									semaphore.Lock(1)
									sent := false
									retries := 0
									for (!sent) && (retries < msg_sending_reintries) {
										retries += 1
										if rs, err := adapter.SendPulseRQ(rq); err != nil {
											if msg, err2 := EVENT.NotifyEvent("006-003", "", &[]string{err.Error()}); err2 == nil {
												lib_gc_log.Error.Printf("%s \n !!", msg)
												ds.writeToRSWriter([]byte(msg))
											} else {
												lib_gc_log.Error.Printf("%s \n !!", err2.Error())
												ds.writeToRSWriter([]byte(msg))
											}
										} else {
											sent = true
											if rs_processor != nil {
												if err := rs_processor(rs, ds.rs_writer); err != nil {
													ds.writeToRSWriter([]byte(err.Error()))
												}
											}
										}
									}
									semaphore.ForceToUnLock(1)
								}
							}(rq, ds.pulse_rs_processor, max_concurrent_sents_semaphore)

						}
					}
				}
			}
			max_concurrent_records_semaphore.ForceToUnLock(1)
		}
	}
}
