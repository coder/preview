package web

import (
	"context"
	"io/fs"

	"cdr.dev/slog"
	"github.com/coder/preview"
	"github.com/coder/preview/types"
)

type Request struct {
	// ID identifies the request. The response contains the same
	// ID so that the client can match it to the request.
	ID     int
	Inputs map[string]string
}

type Response struct {
	Diagnostics Diagnostics       `json:"diagnostics"`
	Paramaters  []types.Parameter `json:"paramaters"`
	// TODO: Workspace tags
}

type Session struct {
	logger   slog.Logger
	dir      fs.FS
	planPath string

	requests  chan *Request
	responses chan *Response
}

func NewSession(logger slog.Logger, dir fs.FS, planPath string) *Session {
	return &Session{
		logger:    logger,
		dir:       dir,
		planPath:  planPath,
		requests:  make(chan *Request, 2),
		responses: make(chan *Response, 2),
	}
}

func (s *Session) handleRequests(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-s.requests:
			resp := s.preview(ctx, req)
			// TODO: If this blocks, that is unfortunate. We should drop the
			// oldest requests.
			s.responses <- &resp
		}
	}
}

func (s *Session) sendRequest(ctx context.Context, req Request) {
	select {
	case <-ctx.Done():
		return
	case s.requests <- &req:
	}
}

func (s *Session) preview(ctx context.Context, req *Request) Response {
	output, diags := preview.Preview(ctx, preview.Input{
		PlanJSONPath:    s.planPath,
		ParameterValues: req.Inputs,
	}, s.dir)

	r := Response{
		Diagnostics: FromHCLDiagnostics(diags),
	}
	if output == nil {
		return r
	}

	r.Paramaters = output.Parameters

	return r
}
