package framework

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SongOf/edge-storage-core/core"
	"github.com/SongOf/edge-storage-core/core/eserrors"
	"github.com/SongOf/edge-storage-core/core/i18n"
	"github.com/SongOf/edge-storage-core/core/recovery"
	"github.com/SongOf/edge-storage-core/mq"
	"github.com/SongOf/edge-storage-core/pkg/eslog"
	"github.com/SongOf/edge-storage-core/storage"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
	"time"
)

const ConnKey = "http-conn"
const DefaultMaxBodySize = 10 * 1024 * 1024

type ServerResponse struct {
	Content    interface{} `json:"Response"`
	ctx        *core.Context
	writer     http.ResponseWriter
	translator *i18n.Translator
}

type ErrorCode struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type ErrorCodeWithRequestId struct {
	Error     ErrorCode `json:"Error"`
	RequestId string    `json:"RequestId"`
}

func (sr *ServerResponse) WithError(err error) *ServerResponse {

	var (
		code, message string
		data          interface{}
	)

	if mxErr, ok := err.(eserrors.EsError); ok {
		code, message = mxErr.Format()
		data = mxErr.GetData()
	} else {
		code, message = eserrors.InternalError().Format()
	}

	if sr.translator != nil {
		if lang, ok := sr.ctx.Params["Language"]; ok {
			language, ok := lang.(string)
			if !ok {
				eslog.C(sr.ctx).Warn("language is not string")
				goto End
			}
			transMessage, err := sr.translator.Translate(language, code, message, data)
			if err != nil {
				eslog.C(sr.ctx).Warn("translate error", eslog.Field("Error", err))
			} else {
				message = transMessage
			}
		}
	}

End:
	sr.Content = ErrorCodeWithRequestId{
		Error:     ErrorCode{Code: code, Message: message},
		RequestId: sr.ctx.TraceId,
	}
	return sr
}

func (sr *ServerResponse) WithResult(response ControllerResult) *ServerResponse {
	requestId := sr.ctx.TraceId
	response.WithRequestId(requestId)
	sr.Content = response
	return sr
}

func (sr *ServerResponse) Reply() {
	body, _ := json.Marshal(sr)
	eslog.L().Info("server reply", eslog.Field("Response", string(body)))
	_, err := fmt.Fprint(sr.writer, string(body))
	if err != nil {
		eslog.L().Error("fmt response error")
	}
}

type Server struct {
	// http server
	server *http.Server

	// framework.Server router
	// dispatch request to Controller by Action name
	router *Router

	// parser is response to parse request body to map
	parser    *Parser
	validator *Validator

	translator *i18n.Translator

	// collector used to collect server metrics
	collector *ServerCollector

	// framework server option
	Option ServerOption
}

type ServerOption struct {
	ListenAddr      string
	StoreNetConn    bool
	PidFile         string
	ShutdownTimeout int
	TLSOption       TLSOption
	MaxBodySize     int64
	EntryList       []*Entry
	Reporter        core.Reporter
	TranslateDir    string
	Middlewares     []core.Middleware
}

type Entry struct {
	Path    string
	Handler http.Handler
}

type TLSOption struct {
	CertFile string
	KeyFile  string
}

func SaveConnInContext(ctx context.Context, c net.Conn) context.Context {
	return context.WithValue(ctx, ConnKey, c)
}

func NewServer(router *Router, option ServerOption) *Server {

	httpServer := http.Server{Addr: option.ListenAddr}
	if option.StoreNetConn {
		httpServer.ConnContext = SaveConnInContext
	}

	parser, err := NewParser()
	if err != nil {
		eslog.L().Panic("new parser error", eslog.Err(err))
	}

	newValidator, err := NewValidator()
	if err != nil {
		eslog.L().Panic("new validator error", eslog.Err(err))
	}

	var translator *i18n.Translator
	if option.TranslateDir != "" {
		translator, err = i18n.NewTranslator(option.TranslateDir)
		if err != nil {
			eslog.L().Panic("new translator error", eslog.Err(err))
		}
	}

	if option.MaxBodySize == 0 {
		option.MaxBodySize = DefaultMaxBodySize
	}

	s := &Server{
		server:     &httpServer,
		router:     router,
		Option:     option,
		parser:     parser,
		validator:  newValidator,
		collector:  NewCollector(),
		translator: translator,
	}

	return s
}

func (s *Server) Start(tls, graceful bool) (listener net.Listener, err error) {
	for _, entry := range s.Option.EntryList {
		http.Handle(entry.Path, entry.Handler)
		eslog.L().Info("add handler", eslog.Field("Path", entry.Path))
	}
	defaultHandler := http.HandlerFunc(s.defaultEntrypoint)
	http.Handle("/", defaultHandler)
	eslog.L().Info("Serving on " + s.Option.ListenAddr)

	if graceful {
		log.Print("main: Listening to existing file descriptor 3.")
		f := os.NewFile(3, "")
		ln, err := net.FileListener(f)
		if err != nil {
			eslog.L().Warn("listen error", eslog.Field("Error", err))
			return nil, err
		}
		// Closing ln does not affect f, and closing f does not affect ln.
		_ = f.Close()
		listener = ln

		// kill ppid
		ppid := syscall.Getppid()
		log.Printf("main: Killing parent pid: %v", ppid)
		_ = syscall.Kill(ppid, syscall.SIGTERM)
	} else {
		log.Print("main: Listening on a new file descriptor.")
		ln, err := net.Listen("tcp", s.server.Addr)
		if err != nil {
			eslog.L().Warn("listen error", eslog.Err(err))
			return nil, err
		}
		listener = ln
	}

	// update pidfile
	if s.Option.PidFile != "" {
		pid := syscall.Getpid()
		content := []byte(strconv.Itoa(pid))
		err := ioutil.WriteFile(s.Option.PidFile, content, 0644)
		if err != nil {
			eslog.L().Warn("write pidfile failed", eslog.Err(err))
			return listener, err
		}
	}

	go func() {
		if tls {
			tlsOption := s.Option.TLSOption
			err = s.server.ServeTLS(listener, tlsOption.CertFile, tlsOption.KeyFile)
		} else {
			err = s.server.Serve(listener)
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			eslog.L().Panic("serve error:", eslog.Err(err))
		}
	}()
	// s.signalHandler(listener)
	return
}

func (s *Server) Stop(timeout int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		eslog.L().Warn("force shutdown", eslog.Field("Error", err))
	}
}

func (s *Server) Reload(listener net.Listener) error {
	tl, ok := listener.(*net.TCPListener)
	if !ok {
		return errors.New("listener is not tcp listener")
	}

	f, err := tl.File()
	if err != nil {
		return err
	}

	args := []string{"-graceful"}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// put socket FD at the first entry
	// cmd.ExtraFiles: If non-nil, entry i becomes file descriptor 3+i.
	cmd.ExtraFiles = []*os.File{f}
	return cmd.Start()
}

func (s *Server) signalHandler(listener net.Listener) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		sig := <-ch
		println("receive signal %s", sig)
		// mxlog.L().Info("receive signal", mxlog.Field("Signal", sig))
		// timeout context for shutdown
		timeout := s.Option.ShutdownTimeout
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			// stop
			eslog.L().Info("receive stop signal, will stop server")
			signal.Stop(ch)
			s.Stop(timeout)
			eslog.L().Info("graceful stop server")
			return
		case syscall.SIGUSR2:
			// Reload
			eslog.L().Info("receive Reload signal, will Reload server")
			err := s.Reload(listener)
			if err != nil {
				eslog.L().Error("graceful restart error", eslog.Field("Error", err))
			}
			s.Stop(timeout)
			eslog.L().Info("graceful Reload server")
		default:
			continue
		}
	}
}

func (s *Server) initContext(r *http.Request) *core.Context {
	ctx := core.NewContext()
	ctx.Request = r
	ctx.Reporter = s.Option.Reporter

	return ctx
}

func (s *Server) initTraceId(ctx *core.Context) {
	requestId := ""
	if reqId, exist := ctx.Params["RequestId"]; exist {
		requestId = reqId.(string)
	} else {
		requestId = uuid.New().String()
	}

	ctx.TraceId = requestId
	ctx.LogFields = map[string]interface{}{"RequestId": requestId}
}

func (s *Server) defaultEntrypoint(w http.ResponseWriter, r *http.Request) {
	// parse -> dispatch -> validate -> process -> output

	ctx := s.initContext(r)
	resp := ServerResponse{ctx: ctx, writer: w, translator: s.translator}
	defer recovery.Recover(ctx, func() {
		if s.collector != nil {
			// panic count
			s.collector.ServerPanicCounter.Inc()
		}
		resp.WithError(eserrors.InternalError()).Reply()
	})

	r.Body = http.MaxBytesReader(w, r.Body, s.Option.MaxBodySize)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			resp.WithError(
				eserrors.InvalidParameterEx(eserrors.InvalidParameterBodyTooLargeCode, nil),
			).Reply()
		} else {
			eslog.L().Error("read http body error", eslog.Err(err))
			resp.WithError(eserrors.InternalError()).Reply()
		}
		return
	}
	eslog.L().Info("receive", eslog.Field("body", string(body)))

	params, parseError := s.parser.PreParseRequest(body)
	if parseError != nil {
		resp.WithError(parseError).Reply()
		return
	}
	ctx.Params = params
	s.initTraceId(ctx)

	value, ok := params["Action"]
	if !ok {
		resp.WithError(eserrors.MissingParameter("Action")).Reply()
		return
	}
	action, ok := value.(string)
	if !ok {
		resp.WithError(eserrors.InvalidParameterValueEx(
			eserrors.InvalidParameterValueTypeCode,
			map[string]interface{}{
				"parameter":  "Action",
				"actualType": reflect.TypeOf(value).Kind(),
				"expectType": "string",
			})).Reply()
		return
	}
	ctx.Action = action

	actionController := s.router.Dispatch(action)
	if actionController == nil {
		resp.WithError(eserrors.InvalidAction(action)).Reply()
		return
	}

	description := actionController.GetDescription()

	// 检查未使用字段和类型错误
	if err := s.parser.CheckParams(ctx, params, description); err != nil {
		eslog.C(ctx).Warn("check params failed", eslog.Err(err))
		resp.WithError(err).Reply()
		return
	}

	//检查验证器
	if err := s.validator.ValidateParameters(ctx, description); err != nil {
		eslog.C(ctx).Warn("validate params failed", eslog.Err(err))
		resp.WithError(err).Reply()
		return
	}

	eslog.C(ctx).Info("run controller", eslog.Field("Action", action), eslog.Field("RequestBody", params))

	ctx.Use(NewLatencyMiddleware(s.collector))
	ctx.Use(s.Option.Middlewares...)
	ctx.Use(NewResultMiddleware(actionController, &resp, s.collector))
	_ = ctx.Next()
}

func (s *Server) Collectors() []prometheus.Collector {
	collectors := []prometheus.Collector{mq.DefaultCollector(), storage.DefaultCollector()}
	if s.collector != nil {
		collectors = append(collectors, s.collector)
		return collectors
	}
	return collectors
}

func (s *Server) RegisterFilterValidator(set string, key string, validateFunc FilterValidateFunc) {
	s.validator.RegisterFilterValidator(set, key, validateFunc)
}

func (s *Server) RegisterCustomValidatorTag(tag string, fn validator.Func) error {
	return s.validator.RegisterCustomValidatorTag(tag, fn)
}
