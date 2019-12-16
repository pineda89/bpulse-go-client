# What is this?
This is a Go client for [BPulse monitoring tool](http://www.bpulse.io).

# Use it!
To use it, the next environment variables must be set:
- **GO_COMMON_LIB_CONFIGURATION_FILE** This variable is set to the path of the configurtion file *BPulseGoClient.ini*
- **GO_COMMON_LIB_EVENTS_FILE** This variable is set to the path of the configuration file *EventsConfiguration.json*

In addition, the steps to integrate this BPulse client in your project is:

- Define your pulse in BPulse:
  * Its typeId
  * Its attributes list
- Integrating Go BPulse client in your project:
  * Add the BPulse client libraries to your Godeps (or another libraries manager you use)
      ```sh
      # BPulse
      github.com/pineda89/bpulse-go-client # Go client for BPulse 
      github.com/pineda89/bpulse-protobuf-go # MDH protobuf messages
      ```
  * Implement the necessary interfaces:
    * rq maker: com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulses_sender.PulseRQMaker
    * rs instance builder: com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_send_channel_adapter.PulseRSInstanceGetter
    * rs processor (optional): com.bpulse.client.go.bpulseclient/bpulsegoclient/lib_pulses_sender.PulseRSProcessor
  
  * Start the bpulse client:
  
     ```go
        if err := bpulsegoclient.StartBPulseClient(rq_maker, rs_processor,rs_instance_getter); err!=nil {
          panic(err)
        }
      ```
  
  * Send pulses:
      ```go
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

                // Take the pulses aggregator from the aggregator's container.
                if aggregator, err := AGG_CONTAINER.Container.GetPulseAggregator(AGG_CONTENT.DEFAULT_AGGREGATOR_NAME); err!=nil {
                    panic(err)
                }else {

                    // Take the aggregator channel
                    if ch, err := aggregator.GetPulsesAggregationChannel(); err!=nil {
                        panic(err)
                    }else {

                        // Send the pulse
                        ch <- pulse

                        // Restore the aggregator to the aggregator's container
                        AGG_CONTAINER.Container.ReleasePulseAggregator(AGG_CONTENT.DEFAULT_AGGREGATOR_NAME, aggregator)
                    }
                }
      ```
      
  * Shutdown the client:
    ```go
    errors_list := bpulsegoclient.ShutdownBPulseClient()
    for e:=errors_list.Front();e!=nil;e=e.Next(){
        fmt.Println(e.Value.(error).Error())
    }
    ```
    
#Example
There is an example of use at *bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/example* 
To execute this example:

- Clone the project and go to its root directory:
    ```sh
    git clone https://github.com/pineda89/bpulse-go-client.git
    cd bpulse-go-client
    ```
- Set the correct values in the configuration file *bpulse-go-client/conf/BPulseGoClient.ini* :
*RESTFUL_SEND_CHANNEL_ADAPTER_HOST* BPulse host name.
*RESTFUL_SEND_CHANNEL_ADAPTER_URL_PORT* BPulse host port.
*RESTFUL_SEND_CHANNEL_ADAPTER_USER* BPulse user login.
*RESTFUL_SEND_CHANNEL_ADAPTER_URL_PWD* BPulse user password.

- Execute the example:
    ```sh
    cd bpulse-go-client/com.bpulse.client.go.bpulseclient/bpulsegoclient/example
    go build ("go build -tags debug" if we want debug traces) 
    source ./set_env.sh
    ./example
   ```

# Start as a Docker container
The example also can be started as a docker container. To do it, you must go to *com.bpulse.client.go.pbulseclient/bpulseclient/example/assets* and do this:

- Install **Docker Engine** and **Docker Compose** in your system.

- Edit the *docker-compose.yml* file to set the properties:
    * BPULSEID
    * BPULSE_HOST
    * BPULSE_PORT
    * BPULSE_USER
    * BPULSE_PWD

- Build the Docker image and run the container:
    ```sh
    docker build -t theskyinflames/bpc
    docker-compose up -d bpc
    ```
- Examine BPulse client logs.

