package wordle

import (
	"github.com/vivym/chinanews-spider/internal/api/protobuf/wordle"
	"google.golang.org/grpc"
)

type WordleToolkit struct {
	conn   *grpc.ClientConn
	client wordle.WordleClient
}

func New(config Config) (*WordleToolkit, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	conn, err := grpc.Dial(config.Address, opts...)
	if err != nil {
		return nil, err
	}

	client := wordle.NewWordleClient(conn)

	return &WordleToolkit{
		conn:   conn,
		client: client,
	}, nil
}

func (w *WordleToolkit) Release() {
	if w.conn != nil {
		_ = w.conn.Close()
	}
}
