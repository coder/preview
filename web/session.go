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
	ID     int               `json:"id"`
	Inputs map[string]string `json:"inputs"`
}

type Response struct {
	ID          int               `json:"id"`
	Diagnostics types.Diagnostics `json:"diagnostics"`
	Parameters  []types.Parameter `json:"parameters"`
	// TODO: Workspace tags
}

// @typescript-ignore Session
type Session struct {
	logger       slog.Logger
	dir          fs.FS
	staticInputs SessionInputs

	requests  chan *Request
	responses chan *Response
}

type SessionInputs struct {
	PlanPath string
	User     types.WorkspaceOwner
}

func NewSession(logger slog.Logger, dir fs.FS, staticInputs SessionInputs) *Session {
	return &Session{
		logger:       logger,
		dir:          dir,
		staticInputs: staticInputs,
		requests:     make(chan *Request, 2),
		responses:    make(chan *Response, 2),
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
		PlanJSONPath:    s.staticInputs.PlanPath,
		ParameterValues: req.Inputs,
		Owner:           s.staticInputs.User,
	}, s.dir)

	r := Response{
		ID:          req.ID,
		Diagnostics: types.Diagnostics(diags),
	}
	if output == nil {
		return r
	}

	r.Parameters = output.Parameters

	return r
}
