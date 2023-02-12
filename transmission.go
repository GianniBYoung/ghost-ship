package main

import (
	"context"
	"fmt"
	"os"

	trans "github.com/hekmon/transmissionrpc/v2"
)

var TransmissionClient *trans.Client

func transmissionClientInit() {
	TransmissionClient = setupCheck()
}

func setupCheck() *trans.Client {
	transmissionPassword := os.Getenv("TRANSMISSIONPASSWORD")
	transmissionUserName := os.Getenv("TRANSMISSIONUSERNAME")
	transmissionIP := os.Getenv("TRANSMISSIONIP")
	client, err := trans.New(transmissionIP, transmissionUserName, transmissionPassword, nil)

	if err != nil {
		fmt.Println("Unable to create transmission client.")
		panic(err)
	}

	if transmissionPassword == "" || transmissionIP == "" || transmissionUserName == "" {
		panic(`Credentials error. Are the environmental variables set?
'TRANSMISSIONPASSWORD',
'TRANSMISSIONUSERNAME',
'TRANSMISSIONIP'`)
	}

	ok, serverVersion, serverMinimumVersion, err := client.RPCVersion(context.TODO())
	if err != nil {
		panic(err)
	}
	if !ok {
		panic(fmt.Sprintf("Remote transmission RPC version (v%d) is incompatible with the transmission library (v%d): remote needs at least v%d",
			serverVersion, trans.RPCVersion, serverMinimumVersion))
	}
	return client

}
