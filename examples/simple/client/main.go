package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc"

	"github.com/go-nacelle/scarf/examples/simple/proto"
)

func main() {
	client := newClient()
	reader := bufio.NewReader(os.Stdin)

	for {
		read(client, reader)
	}
}

func newClient() proto.SecretServiceClient {
	conn, err := grpc.Dial(getAddr(), grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}

	return proto.NewSecretServiceClient(conn)
}

func getAddr() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}

	return "localhost:6000"
}

func read(client proto.SecretServiceClient, reader *bufio.Reader) {
	fmt.Print("> ")
	text, err := reader.ReadString('\n')
	if err != nil {
		panic(err.Error())
	}

	resp, err := client.PostSecret(context.Background(), &proto.Secret{Secret: strings.TrimSpace(text)})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("< %s\n\n", resp.GetName())
}
