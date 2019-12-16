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

package contents

import (
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_storage"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_pool"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_event"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_aggregator"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container/lib_gc_contents_loader"
    "github.com/golang/protobuf/proto"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_panic_catching"

    "fmt"
    "strconv"
    "time"
    "sync"
)

func init() {

    // Locks the container until this content has been put in it.
    wg := lib_pulse_aggregator.Container.(*lib_pulse_aggregator.PulseAggregatorContainer).Container.GetWg()
    wg.Add(1)
    lib_gc_contents_loader.AddInitiator("DEFAULT_AGGREGATOR", Initialize)
    wg.Done()
}

const DEFAULT_AGGREGATOR_NAME = "DEFAULT"

func Initialize() error {
    if lib_pulse_aggregator.Container.IsAggregatorActivated(DEFAULT_AGGREGATOR_NAME) {
        pool_size := lib_gc_config.ConfigurationProvider.GetPropertyINT("PULSE_AGGREGATOR_DEFAULT", "PULSE_AGGREGATOR_DEFAULT_POOL_SIZE")
        record_size := lib_gc_config.ConfigurationProvider.GetPropertyINT64("PULSE_AGGREGATOR_DEFAULT", "PULSE_AGGREGATOR_DEFAULT_RECORD_SIZE")
        aggregation_channel_size := lib_gc_config.ConfigurationProvider.GetPropertyINT64("PULSE_AGGREGATOR_DEFAULT", "PULSE_AGGREGATOR_DEFAULT_AGGREGATION_CHANNEL_SIZE")
        max_time_without_store := lib_gc_config.ConfigurationProvider.GetPropertyINT64("PULSE_AGGREGATOR_DEFAULT", "PULSE_AGGREGATOR_DEFAULT_MAX_TIME_MS_WITHOUT_STORE")
        pool := lib_gc_pool.POOLMaker.GetPool(int32(pool_size))
        for z := 0; z<pool_size; z++ {
            aggregator := &DefaultPulseAggregator{aggregation_channel: make(chan proto.Message, int(aggregation_channel_size)), record_size: record_size, max_time_without_store:max_time_without_store,isrunning:false,once:&sync.Once{},once_run:&sync.Once{}}
            aggregator.Contents = lib_gc_container.Contents{lib_pulse_aggregator.Container.(*lib_pulse_aggregator.PulseAggregatorContainer).Container, lib_gc_container.CONTAINER_STATUS_STOPPED, aggregator}
            pool.GetPool() <- &lib_gc_pool.Poolable{Item:aggregator}
        }
        lib_pulse_aggregator.Container.AddPulseAggregatorPool(DEFAULT_AGGREGATOR_NAME, &lib_gc_container.PooledContents{Pooler:pool})

        lib_gc_log.Info.Printf("%s aggregator initiated, with a pool of %d instances and record size %d . \n", DEFAULT_AGGREGATOR_NAME, pool_size, record_size)
    }
    return nil
}

type DefaultPulseAggregator struct {
    lib_gc_container.Contents
    storage lib_pulse_storage.IPulsesStorageAdapter
    aggregation_channel chan proto.Message
    record_size int64
    max_time_without_store int64
    isrunning bool
    once_run *sync.Once
    once *sync.Once
}

func (dpa *DefaultPulseAggregator) SetPulsesStorage(storage lib_pulse_storage.IPulsesStorageAdapter) error {
    dpa.storage = storage
    return nil
}

func (dpa *DefaultPulseAggregator) GetPulsesAggregationChannel() (chan <- proto.Message, error) {
    return dpa.aggregation_channel, nil
}


func (dpa *DefaultPulseAggregator) Shutdown() error {
    lib_gc_log.Info.Printf("AGGREGATOR content [%p, %s], container [%p, %s] SHUTDOWN catched. \n", &dpa.Contents, DEFAULT_AGGREGATOR_NAME, dpa.Container, dpa.Container.GetName())
    println("Default aggregator Shutdown, ", dpa.Contents.Status)
    close(dpa.aggregation_channel)
    if err := dpa.storage.Shutdown(); err!=nil {
        return err
    }else {
        return nil
    }
}

func (dpa *DefaultPulseAggregator) Start() error {

    // Before start the content, its dependencies must be satisfied.
    // This dependencies aggregation will be executed once.
    dependencies_solver := func(){
        lib_gc_log.Info.Printf("AGGREGATOR content [%p, %s], container [%p, %s] START catched. \n", &dpa.Contents, DEFAULT_AGGREGATOR_NAME, dpa.Container, dpa.Container.GetName())

        storage_name := lib_gc_config.ConfigurationProvider.GetPropertySTRING("PULSE_AGGREGATOR_DEFAULT", "PULSE_AGGREGATOR_DEFAULT_STORAGE")
        if storage, err := lib_pulse_storage.Container.GetPulseStorageAdapter(storage_name); err!=nil {
            panic(err)
        }else {
            dpa.storage = storage
        }
    }
    dpa.once.Do(dependencies_solver)

    aggregator_main := func(){
        // With all of its dependencies satisfied, the content can start. These will be executed once.
        go dpa.checkForRunning()
    }
    dpa.once_run.Do(aggregator_main)

    return nil
}

func (dpa *DefaultPulseAggregator) Stop() error {
    lib_gc_log.Info.Printf("AGGREGATOR content [%p, %s], container [%p, %s] STOP catched. \n", &dpa.Contents, DEFAULT_AGGREGATOR_NAME, dpa.Container, dpa.Container.GetName())
    return nil
}

func (dpa *DefaultPulseAggregator) checkForRunning(){
   ticker := time.NewTicker(time.Millisecond * 500)
    for  _ = range ticker.C{
      if !dpa.isrunning{
          dpa.isrunning = true
          go dpa.run()
      }
   }
}

func (dpa *DefaultPulseAggregator) run() {
    // panic catching
    defer func(){
      lib_gc_panic_catching.PanicCatching("DefaultSender.run")
      dpa.isrunning = false
    }()

    lib_gc_log.Info.Printf("AGGREGATOR content [%p, %s], container [%p, %s] RUN STARTED. \n", &dpa.Contents, DEFAULT_AGGREGATOR_NAME, dpa.Container, dpa.Container.GetName())
    var record *lib_pulse_storage.PulseRecord
    var last_storage_at time.Time = time.Now()
    ticker := time.NewTicker(time.Nanosecond * 100)
    for  _ = range ticker.C{
        if dpa.Contents.Status == lib_gc_container.CONTAINER_STATUS_STARTED && dpa.storage!=nil {
            select {
            case pulse, ok := <-dpa.aggregation_channel:
                _ = pulse
                _ = ok
                lib_gc_log.Trace.Printf("AGGREGATOR recived pulse %p on content %p, container %p [%s] \n", pulse, &dpa.Contents, dpa.Container, dpa.Container.GetName())
                lib_gc_log.Trace.Printf("AGGREGATOR recived pulse %p : %v+ \n", pulse.String())

            // Check for channel keeps opened
                if !ok {
                    _msg, _ := lib_gc_event.NotifyEvent("003-001", "", &[]string{})
                    lib_gc_log.Warning.Println(_msg)
                    dpa.Status = lib_gc_container.CONTAINER_STATUS_STOPPED
                    break
                }

            // Check for an active record
                if record == nil {
                    key := strconv.FormatInt(time.Now().Unix(), 10)
                    pulses := make([]proto.Message, dpa.record_size)
                    record = &lib_pulse_storage.PulseRecord{Key:key, Pulses:pulses, Pos:0}
                }

            // Add the pulse to the record.
                record.Pulses[record.Pos] = pulse
                lib_gc_log.Trace.Printf("AGGREGATOR added pulse %p on content %p, container %p [%s], pos %d / %d \n", pulse, &dpa.Contents, dpa.Container, dpa.Container.GetName(), record.Pos, dpa.record_size)
                record.Pos = record.Pos + 1

            // Check if the record pulses have reached the block_size size
                if record.Pos == dpa.record_size {
                    // Store
                    lib_gc_log.Trace.Printf("AGGREGATOR storing BY BLOCK SIZE [%d - %d] pulse %p on content %p, container %p [%s] \n", record.Pos, dpa.record_size, pulse, &dpa.Contents, dpa.Container, dpa.Container.GetName())
                    record.Pulses = record.Pulses[:record.Pos]
                    dpa.storage.AddPulseRecord(record)
                    last_storage_at = time.Now()
                    record = nil
                }

            default:
                delay := int64(time.Now().Sub(last_storage_at).Nanoseconds()/1000000)
                if record !=nil && delay >= dpa.max_time_without_store {
                    lib_gc_log.Info.Printf("AGGREGATOR storing BY TIMING pulses on content %p, container %p [%s] with delay %s \n", &dpa.Contents, dpa.Container, dpa.Container.GetName(), fmt.Sprint(time.Now().Sub(last_storage_at)))
                    // Store
                    record.Pulses = record.Pulses[:record.Pos]
                    dpa.storage.AddPulseRecord(record)
                    last_storage_at = time.Now()
                    record = nil
                }
            }
        }
    }
}
