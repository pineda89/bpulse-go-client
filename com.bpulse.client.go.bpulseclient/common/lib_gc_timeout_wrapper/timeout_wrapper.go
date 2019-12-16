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

package lib_gc_timeout_wrapper

import(

    "time"
    "errors"
)

func init(){
    TimeoutWrapper = &timeoutWrapper{}
}

type TIMEOUT_ERROR error

var TimeoutWrapper TimeoutWrapper_I = nil

type TimeoutWrapper_I interface{
    Wrap(timeout time.Duration, wrappable Wrappable_I) (*[]interface{}, error, TIMEOUT_ERROR)
}

type f_result struct{
    result *[]interface{}
    err error
}

type Wrappable_I interface{
    Execute() (*[]interface{},error)
}

type timeoutWrapper struct{}

func (tw *timeoutWrapper) Wrap(timeout time.Duration, wrappable Wrappable_I) (*[]interface{}, error, TIMEOUT_ERROR) {

    var c_func chan *f_result = make(chan *f_result)
    var c_timeout chan int = make(chan int)
    var err_timeout string
    var ret *f_result


    go func() {
        result,err := wrappable.Execute()
        c_func <- &f_result{result,err}
    }()

    go func(){
        timer := time.NewTimer(timeout)
        <-timer.C
        close(c_timeout)
    }()

    cancel := false
    for !cancel {
        select {
        case <-c_timeout:
            err_timeout = "*Timeout has been reached*"
            cancel = true
            //println("Entro por TO")
        case ret = <-c_func:
        // func done
            cancel = true
            //println("Entro por FN")
        }
    }

    if err_timeout != "" {
        return nil, nil, errors.New(err_timeout).(TIMEOUT_ERROR)
    }else {
        return ret.result, ret.err, nil
    }
}

