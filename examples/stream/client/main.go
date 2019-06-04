package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	"github.com/go-nacelle/scarf/examples/stream/proto"
)

func main() {
	client := newClient()
	reader := bufio.NewReader(os.Stdin)

	for {
		read(client, reader)
	}
}

func newClient() proto.ValueServiceClient {
	conn, err := grpc.Dial(getAddr(), grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}

	return proto.NewValueServiceClient(conn)
}

func getAddr() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}

	return "localhost:6000"
}

func read(client proto.ValueServiceClient, reader *bufio.Reader) {
	stream, err := client.NoisyUpdate(context.Background())
	if err != nil {
		panic(err.Error())
	}

	go func() {
		for {
			summary, err := stream.Recv()
			if err != nil {
				panic(err.Error())
			}

			fmt.Printf("> update #%d = %d\n", summary.GetUpdates(), summary.GetValue())
		}
	}()

	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err.Error())
		}

		value, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			panic(err.Error())
		}

		if err := stream.Send(&proto.Update{Delta: int32(value)}); err != nil {
			panic(err.Error())
		}
	}
}
