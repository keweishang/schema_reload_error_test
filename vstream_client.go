/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This is a sample vstream client. */

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"
	"os"
	"strings"

	binlogdatapb "vitess.io/vitess/go/vt/proto/binlogdata"
	logutilpb "vitess.io/vitess/go/vt/proto/logutil"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
	_ "vitess.io/vitess/go/vt/vtctl/grpcvtctlclient"
	"vitess.io/vitess/go/vt/vtctl/vtctlclient"
	_ "vitess.io/vitess/go/vt/vtgate/grpcvtgateconn"
	"vitess.io/vitess/go/vt/vtgate/vtgateconn"
)

func main() {
	var tabletType topodatapb.TabletType
	if strings.ToUpper(os.Args[1]) == "MASTER" {
		tabletType = topodatapb.TabletType_MASTER
	} else if strings.ToUpper(os.Args[1]) == "REPLICA" {
		tabletType = topodatapb.TabletType_REPLICA
	} else {
		panic("please specify tablet type: MASTER or REPLICA")
	}
	fmt.Printf("Subscribe to tablet type: %v\n", tabletType)

	ctx := context.Background()
	vgtid := &binlogdatapb.VGtid{
		ShardGtids: []*binlogdatapb.ShardGtid{{
			Keyspace: "test_sharded_keyspace",
			Gtid:  "current",
		}}}
	filter := &binlogdatapb.Filter{
		Rules: []*binlogdatapb.Rule{{
			Match: "/.*/",
		}},
	}

	conn, err := vtgateconn.Dial(ctx, "localhost:15991")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	
	
	reader, err := conn.VStream(ctx, tabletType, vgtid, filter)
	for {
		e, err := reader.Recv()
		switch err {
		case nil:
			_ = e
			//fmt.Printf("%v\n", e)
			fmt.Printf(".")
		case io.EOF:
			fmt.Printf("stream ended\n")
			return
		default:
			fmt.Printf("%s:: remote error: %v\n", time.Now(), err)
			return
		}
	}
}

func execVtctl(ctx context.Context, args []string) ([]string, error) {
	client, err := vtctlclient.New("localhost:15999")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	stream, err := client.ExecuteVtctlCommand(ctx, args, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot execute remote command: %v", err)
	}

	var results []string
	for {
		e, err := stream.Recv()
		switch err {
		case nil:
			if e.Level == logutilpb.Level_CONSOLE {
				results = append(results, e.Value)
			}
		case io.EOF:
			return results, nil
		default:
			return nil, fmt.Errorf("remote error: %v", err)
		}
	}
	return results, nil
}