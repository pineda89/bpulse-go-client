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

package lib_gc_event

import (
	"errors"
	"fmt"
	"testing"
)

func init() {

}

//
// Be careful !! To launch the test, you must to set the environment variable GO_COMMON_LIB_EVENTS_FILE
//


// Test for events loading from file
func Test_EventsLoaded(t *testing.T) {
	if ptrEH == nil {
		t.Error(errors.New("Events has not been loaded"))
	} else {
		fmt.Println(ptrEH.Configuration)
		for _, v := range ptrEH.Configuration.Events {
			fmt.Println("Loaded event: " + v.Code)
		}
		for k, v := range ptrEH.Configuration.EventMap {
			fmt.Println("Loaded event by code: " + k + " -> " + v.Code)
		}
	}
}

// Test events notification
func Test_NotifyEvent(t *testing.T) {

	// Test for an existing event
	if msg, err := NotifyEvent("000-000", "", &[]string{"ItDoesExist"}); err != nil {
		t.Error(err)
	} else {
		fmt.Println(msg)
	}

	// Test for an existing event with customized message
	if msg, err := NotifyEvent("000-000", "Customized message for the event #  .", &[]string{"ItDoesExist"}); err != nil {
		t.Error(err)
	} else {
        fmt.Println(msg)
	}

	// Test for a not existing event
	if _, err := NotifyEvent("XXX-001", "", &[]string{"ItDoesNotExist"}); err != nil {
        fmt.Println("The event does not exist, the error it's ok: " + err.Error())
	} else {
		t.Error("The event dos not exist. It must fail !!!")
	}
}
