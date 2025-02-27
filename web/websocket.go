package web

import (
	"context"
	"fmt"
	"time"

	"cdr.dev/slog"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func (s *Session) Listen(ctx context.Context, conn *websocket.Conn) {
	s.logger.Info(ctx, "new connection")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// req -> responses
	go s.handleRequests(ctx)

	// read the requests
	go s.readLoop(ctx, cancel, conn)

	// write responses back
	go s.writeLoop(ctx, cancel, conn)

	time.Sleep(time.Second)
	fmt.Println("CLOSE")

	// Always close the connection at the end of the Listen.
	defer conn.Close(websocket.StatusNormalClosure, "closing connection")
	<-ctx.Done()
	return
}

func (s *Session) readLoop(ctx context.Context, cancel func(), conn *websocket.Conn) {
	defer cancel()

	for {
		var req Request
		err := wsjson.Read(ctx, conn, &req)
		if err != nil {
			s.logger.Error(ctx, "failed to read request", slog.F("err", err))
			return
		}

		s.sendRequest(ctx, req)
	}
}

func (s *Session) writeLoop(ctx context.Context, cancel func(), conn *websocket.Conn) {
	defer cancel()

	for {
		select {
		case resp := <-s.responses:
			err := wsjson.Write(ctx, conn, resp)
			if err != nil {
				s.logger.Error(ctx, "failed to write response", slog.F("err", err))
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
