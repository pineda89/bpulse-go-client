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

package lib_pulse_storage

import (
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_config"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_log"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_event"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_timeout_wrapper"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container"
    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulse_storage"
    "time"
    "errors"
    "container/ring"

    "github.com/pineda89/bpulse-go-client/com.bpulse.client.go.bpulseclient/common/lib_gc_container/lib_gc_contents_loader"
    "fmt"
    "sync"
    "sync/atomic"
)

func init(){
    // Locks the container until this content has been put in it.
    wg := lib_pulse_storage.Container.(*lib_pulse_storage.PulseStorageContainer).Container.GetWg()
    wg.Add(1)
    lib_gc_contents_loader.AddInitiator("DEFAULT_STORAGE", Initialize)
    once = &sync.Once{}
    wg.Done()
}

// Maximum time that the storage will wait for shutdown
var shutdown_timeout int64
var checker_interval int

const DEFAULT_STORAGE_NAME = "DEFAULT"
var storage_semaphore = &sync.Mutex{}
var once *sync.Once

type DefaultPulseStorage struct{
    lib_gc_container.Contents
    ring_size int
    records_ring *ring.Ring
    next_pending_record *ring.Ring
    count_stored int32
    count_taken int32

    checker *time.Ticker
}

func Initialize() error{
    if lib_pulse_storage.Container.IsStorageActivated(DEFAULT_STORAGE_NAME) {

        pending_keys_size := lib_gc_config.ConfigurationProvider.GetPropertyINT("PULSE_STORAGE_DEFAULT", "PULSE_STORAGE_DEFAULT_PENDING_KEYS_SIZE")
        checker_interval = lib_gc_config.ConfigurationProvider.GetPropertyINT("PULSE_STORAGE_DEFAULT", "PULSE_STORAGE_WRITE_EXCEEDS_READER_CHECK_INTERVAL_MS")
        shutdown_timeout = lib_gc_config.ConfigurationProvider.GetPropertyINT64("PULSE_STORAGE_DEFAULT", "PULSE_STORAGE_DEFAULT_TIMEOUT_TIME_MS_FOR_SHUTDOWN")
        storage := &DefaultPulseStorage{records_ring:ring.New(pending_keys_size),ring_size:pending_keys_size}
        storage.Contents = lib_gc_container.Contents{lib_pulse_storage.Container.(*lib_pulse_storage.PulseStorageContainer).Container,lib_gc_container.CONTAINER_STATUS_STOPPED, storage}
        lib_pulse_storage.Container.AddPulseStorage(DEFAULT_STORAGE_NAME, storage)

        //storage_semaphoreW = lib_gc_semaphore.SemaphoreFactory.GetSemaphore(1)
        //storage_semaphoreR = lib_gc_semaphore.SemaphoreFactory.GetSemaphore(1)

        lib_gc_log.Info.Printf("%s storage initalized. \n",DEFAULT_STORAGE_NAME )
    }
    return nil
}

func (dps *DefaultPulseStorage) Shutdown() error{

    lib_gc_log.Info.Printf("STORAGE content [%p, %s], container [%p, %s] SHUTDOWN catched. \n",&dps.Contents,DEFAULT_STORAGE_NAME,dps.Container,dps.Container.GetName())

    // Wait for all pulses has been sent
    if _,_, timeout_err := lib_gc_timeout_wrapper.TimeoutWrapper.Wrap(time.Duration(shutdown_timeout)*time.Millisecond , dps);timeout_err!=nil{
        _msg,_ := lib_gc_event.NotifyEvent("007-001","",&[]string{DEFAULT_STORAGE_NAME})
        return errors.New(_msg)
    }else{
        return nil
    }
}

func (dps *DefaultPulseStorage) Start() error {
    lib_gc_log.Info.Printf("STORAGE content [%p, %s], container [%p, %s] START catched. \n", &dps.Contents, DEFAULT_STORAGE_NAME, dps.Container, dps.Container.GetName())

    /*
        The checker that monitors if the pulses's writer reaches the pulses's reader in the ring,
        must be executed once.
    */
    f := func() {
        // Start the write vs reader relation checker.
        dps.checker = time.NewTicker(time.Duration(checker_interval) *time.Millisecond)
        go func(dps *DefaultPulseStorage) {
            for {
                select {
                case <-dps.checker.C:
                    dps.checkForWriterReaderRelation()
                }
            }
        }(dps)
    }
    once.Do(f)

    return nil
}

func (dps *DefaultPulseStorage) Stop() error{
    lib_gc_log.Info.Printf("STORAGE content [%p, %s], container [%p, %s] STOP catched. \n",&dps.Contents,DEFAULT_STORAGE_NAME,dps.Container,dps.Container.GetName())
    return nil
}

func (dps *DefaultPulseStorage) AddPulseRecord(record *lib_pulse_storage.PulseRecord) error {
    storage_semaphore.Lock()
    defer storage_semaphore.Unlock()
    
    if dps.next_pending_record == nil{
        dps.next_pending_record = dps.records_ring
    }
    lib_gc_log.Trace.Printf("STORAGE content [%p, %s], container [%p, %s] ADD pulse record %p. Ring: %p. Pending record ring %p \n",&dps.Contents,DEFAULT_STORAGE_NAME,dps.Container,dps.Container.GetName(), record, dps.records_ring, dps.next_pending_record)
    dps.records_ring.Value = record
    dps.records_ring = dps.records_ring.Next()
    dps.count_stored = dps.count_stored + 1

    lib_gc_log.Trace.Printf("STORAGE content [%p, %s], container [%p, %s] ADDED pulse record %p. New Ring: %p \n",&dps.Contents,DEFAULT_STORAGE_NAME,dps.Container,dps.Container.GetName(), record, dps.records_ring)
    return nil
}

func (dps *DefaultPulseStorage) GetPulseRecord() (*lib_pulse_storage.PulseRecord, error) {
    storage_semaphore.Lock()
    defer storage_semaphore.Unlock()
    
    if dps.next_pending_record != nil{
        ppulse := dps.next_pending_record.Value
        if ppulse!=nil {
            var pulse lib_pulse_storage.PulseRecord = *(ppulse.(*lib_pulse_storage.PulseRecord))
            dps.next_pending_record.Value = nil
            dps.next_pending_record = dps.next_pending_record.Next()

            lib_gc_log.Trace.Printf("STORAGE content [%p, %s], container [%p, %s] TAKE pulse record %p from ring: %p, new ring %p \n",&dps.Contents,DEFAULT_STORAGE_NAME,dps.Container,dps.Container.GetName(), pulse, dps.next_pending_record, dps.next_pending_record.Next())

            dps.incTaken(1)
            return &pulse, nil
        }else{
            return nil, nil
        }
    }else{
        return nil,nil
    }
}

// ----------------------------------------------------------------------------
func (dps *DefaultPulseStorage) Execute() (*[]interface{},error){
    for {
        if dps.next_pending_record.Value == nil{
            break
        }
    }
    return nil,nil
}

/**
    This method checks if the writer (store pulses) has exceeded the reader (send pulses) on the ring
*/
func (dps *DefaultPulseStorage) checkForWriterReaderRelation() {
    exceeded := dps.count_stored-dps.count_taken >= int32(dps.ring_size)
    lib_gc_log.Trace.Printf("STORAGE content [%p, %s], container [%p, %s] CHECK for writer-reader relation [%d - %d] [%t] \n",&dps.Contents,DEFAULT_STORAGE_NAME,dps.Container,dps.Container.GetName(), dps.count_stored, dps.count_taken, exceeded)
    if exceeded{
        // The writer has reached the reader.
        msg,_ := lib_gc_event.NotifyEvent("007-002","",&[]string{fmt.Sprint(time.Now())})
        lib_gc_log.Error.Println(msg)

        // Normalize  the counters. Skip the lost changes.
        dps.incTaken((dps.count_stored % int32(dps.ring_size)) * int32(dps.ring_size))
    }
}

func (dps *DefaultPulseStorage) incTaken(delta int32){
  atomic.AddInt32(&dps.count_taken,delta)
}
