/*
Copyright 2015 - Serhstourism, S.A

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

package lib_gc_metrics

import (
    "expvar"
    "fmt"
    "github.com/youtube/vitess/go/stats"
    "time"
)

func init() {
    //ML = &MemoryLoaderImpl{}
    ME = &MetricsExposerImpl{}
    BUFFER_MAX_NUM_TRACES = 1000
    working = true
    fmt.Println("Metrics package initiated.")
}

// --> http://{mdh server}:{mdh port}/debug/vars

var metricsChannel chan *traceMetric

var mtm = stats.NewHistogram("TimeResponses", []int64{10000000, 50000000, 100000000, 500000000, 1000000000, 1500000000, 2000000000, 2500000000, 3000000000, 3500000000, 4000000000, 4500000000, 5000000000, 5500000000, 6000000000, 10000000000})

var working bool

var ME MetricsExposer
var BUFFER_MAX_NUM_TRACES int

// Two metrics, these are exposed by "magic" :)
// Number of calls to our server.
var numCalls = expvar.NewInt("num_calls")
var numError = expvar.NewInt("num_errs")
var numPanic = expvar.NewInt("num_panics")
var numTimeout = expvar.NewInt("num_timeouts")

// ME Status
type MEStatus int

const (
    NOT_INITIALIZED MEStatus = 0
    INITIALIZED     MEStatus = 1
    STARTED         MEStatus = 2
    STOPPED         MEStatus = 3
    RESUMED         MEStatus = 4
    SHUTDOWN        MEStatus = 5
)

var Status MEStatus = NOT_INITIALIZED

// Estructura para encapsular las metricas
type traceMetric struct {
    traceType int8 //'0': Call ; '1': Both (Call+Error)
    elapsed   time.Duration
}

type MetricsExposer interface {
    Initialize() error
    Start() error
    Stop() error
    Resume() error
    Shutdown() *[]error
    GetAsJSON() (string, error)
    Expose(traceType int8, elapsed time.Duration) error
}

//type MemoryLoaderImpl struct{}
type MetricsExposerImpl struct {}

func (mei *MetricsExposerImpl) GetWorking() bool {
    return working
}

func (mei *MetricsExposerImpl) Initialize() error {

    fmt.Println("Metrics: Starting initialization...")

    //Inicializamos metricas
    metricsChannel = make(chan *traceMetric, BUFFER_MAX_NUM_TRACES)

    // Set end status
    defer func() {
        Status = INITIALIZED
    }()
    fmt.Println("Inititialized finished")
    return nil
}

func (mei *MetricsExposerImpl) Start() error {

    // Set end status
    defer func() {
        Status = STARTED
    }()

    go exposeMetrics()

    return nil
}

func (mei *MetricsExposerImpl) Stop() error {

    // Set end status
    defer func() {
        Status = STOPPED
    }()

    return nil
}

func (mei *MetricsExposerImpl) Shutdown() *[]error {

    // Set end status
    defer func() {
        Status = SHUTDOWN
    }()

    return nil
}

func (mei *MetricsExposerImpl) Resume() error {

    // Set end status
    defer func() {
        Status = RESUMED
    }()

    return nil
}

func (mei *MetricsExposerImpl) GetAsJSON() (string, error) {
    return mtm.String(), nil
}

func (mei *MetricsExposerImpl) GetHistory() (*[]string, error) {
    return nil, nil
}

func (mei *MetricsExposerImpl) Expose(traceType int8, elapsed time.Duration) error {
    tm := new(traceMetric)
    tm.traceType = traceType
    tm.elapsed = elapsed
    metricsChannel <- tm
    return nil
}

//
// ----------------------------------------------------------------
//

func exposeMetrics() {
    for {
        trace := <-metricsChannel
        switch trace.traceType {
            case 0: //Total Requests with response with no error :-)
            mtm.Add(trace.elapsed.Nanoseconds())
            numCalls.Add(1)
            case 1: //Total Requests with error with some error :-(
            numCalls.Add(1)
            numError.Add(1)
            case 2: //Total panics
            numPanic.Add(1)
            case 3: //Total Requests with timeouts
            numTimeout.Add(1)
        }
    }
}
