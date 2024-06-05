package session

import (
	bytes "bytes"
	context "context"
	"errors"
	fmt "fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	strings "strings"

	"github.com/dagger/dagger/engine"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/moby/buildkit/util/grpcerrors"
	"golang.org/x/sync/errgroup"
	codes "google.golang.org/grpc/codes"
)

func (s TerminalAttachable) remoteSession(srv Terminal_SessionServer) error {
	dialer := &websocket.Dialer{}

	sessionID := uuid.New().String()
	slog.Info(fmt.Sprintf("remote terminal session started\n\nssh -p 2222 %s@ssh.dagger.cloud", sessionID))

	reqHeader := http.Header{}
	// if ts.Client.SecretToken != "" {
	// 	reqHeader["Authorization"] = []string{"Basic " + base64.StdEncoding.EncodeToString([]byte(ts.Client.SecretToken+":"))}
	// }

	wsconn, errResp, err := dialer.DialContext(srv.Context(), "ws://localhost:9009/"+sessionID, reqHeader)
	if err != nil {
		if errors.Is(err, websocket.ErrBadHandshake) {
			return fmt.Errorf("dial error %d: %w", errResp.StatusCode, err)
		}
		return fmt.Errorf("dial: %w", err)
	}
	if errResp != nil {
		defer errResp.Body.Close()
	}

	// Handle incoming messages
	eg := errgroup.Group{}

	// gRPC -> WebSocket
	eg.Go(func() error {
		defer func() {
			wsconn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			wsconn.Close()
		}()

		for {
			req, err := srv.Recv()
			if err != nil {
				if errors.Is(err, context.Canceled) || grpcerrors.Code(err) == codes.Canceled {
					// canceled
					return nil
				}

				if errors.Is(err, io.EOF) {
					// stopped
					return nil
				}

				if grpcerrors.Code(err) == codes.Unavailable {
					// client disconnected (i.e. quitting Dagger out)
					return nil
				}

				return err
			}
			switch msg := req.GetMsg().(type) {
			case *SessionRequest_Stdout:
				message := []byte(engine.StdoutPrefix)
				message = append(message, msg.Stdout...)
				err = wsconn.WriteMessage(websocket.BinaryMessage, message)
				if err != nil {
					fmt.Fprintf(os.Stderr, "write: %v\n", err)
					continue
				}
			case *SessionRequest_Stderr:
				message := []byte(engine.StderrPrefix)
				message = append(message, msg.Stderr...)
				err = wsconn.WriteMessage(websocket.BinaryMessage, message)
				if err != nil {
					fmt.Fprintf(os.Stderr, "write: %v\n", err)
					continue
				}
			case *SessionRequest_Exit:
				message := []byte(engine.ExitPrefix)
				message = append(message, []byte(fmt.Sprintf("%d", msg.Exit))...)
				err = wsconn.WriteMessage(websocket.BinaryMessage, message)
				if err != nil {
					fmt.Fprintf(os.Stderr, "write: %v\n", err)
					continue
				}

				return nil
			}
		}
	})

	// WebSocket -> gRPC
	eg.Go(func() error {
		for {
			_, buff, err := wsconn.ReadMessage()
			if err != nil {
				wsCloseErr := &websocket.CloseError{}
				if errors.As(err, &wsCloseErr) && wsCloseErr.Code == websocket.CloseNormalClosure {
					return nil
				}
				return err
			}
			switch {
			case bytes.HasPrefix(buff, []byte(engine.StdinPrefix)):
				err = srv.Send(&SessionResponse{
					Msg: &SessionResponse_Stdin{
						Stdin: bytes.TrimPrefix(buff, []byte(engine.StdinPrefix)),
					},
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "write: %v\n", err)
					continue
				}
			case bytes.HasPrefix(buff, []byte(engine.ResizePrefix)):
				sizeMessage := string(bytes.TrimPrefix(buff, []byte(engine.ResizePrefix)))
				size := strings.SplitN(sizeMessage, ";", 2)
				cols, err := strconv.Atoi(size[0])
				if err != nil {
					return err
				}
				rows, err := strconv.Atoi(size[1])
				if err != nil {
					return err
				}
				err = srv.Send(&SessionResponse{
					Msg: &SessionResponse_Resize{
						Resize: &Resize{
							Width:  int32(cols),
							Height: int32(rows),
						},
					},
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "write: %v\n", err)
					continue
				}
			}
		}
	})

	return eg.Wait()
}
