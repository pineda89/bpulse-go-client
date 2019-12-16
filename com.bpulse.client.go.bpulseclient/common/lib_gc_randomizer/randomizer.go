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

package lib_gc_randomizer

import (
	M_RAND "math/rand"
	"sync"
	"time"

    "crypto/rand"
    "encoding/binary"
    "bytes"
)

func init() {
	// Set random seed
	M_RAND.Seed(time.Now().Unix())

	mutex = &sync.Mutex{}
}

var mutex *sync.Mutex

func GetRandom() int64 {

	return M_RAND.Int63()
}

func GetRingRandom(ring_size int32) (int32, error){

    b := make([]byte,32)
    if _,err := rand.Read(b);err!=nil{
        return 0, err
    }

    var p int32
    buf := bytes.NewBuffer(b)
    binary.Read(buf, binary.BigEndian, &p)

    p = p % ring_size
    if p<0{
        p = p * -1
    }
    return p, nil
}
