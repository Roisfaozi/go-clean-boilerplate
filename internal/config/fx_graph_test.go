package config

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	ws2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestCoreFxModuleGraphValid(t *testing.T) {
	err := fx.ValidateApp(fx.Supply(&AppConfig{}), CoreFxModule)
	require.NoError(t, err)
}

func TestCoreFxModuleMissingAppConfigFails(t *testing.T) {
	err := fx.ValidateApp(CoreFxModule)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "*config.AppConfig")
}

func TestLifecycleHookStartupAndShutdownOrder(t *testing.T) {
	var events []string

	app := fxtest.New(t,
		fx.Invoke(func(lc fx.Lifecycle) {
			registerLifecycleHooks(lc, lifecycleHooks{
				StartHTTP: func() {
					events = append(events, "http:start")
				},
				StopHTTP: func(context.Context) error {
					events = append(events, "http:stop")
					return nil
				},
				StartScheduler: func() {
					events = append(events, "scheduler:start")
				},
				StopScheduler: func() {
					events = append(events, "scheduler:stop")
				},
				StartProcessor: func() {
					events = append(events, "processor:start")
				},
				StopProcessor: func() {
					events = append(events, "processor:stop")
				},
				StartBroadcaster: func() {
					events = append(events, "broadcaster:start")
				},
				StopBroadcaster: func() {
					events = append(events, "broadcaster:stop")
				},
				StartWS: func() {
					events = append(events, "ws:start")
				},
				StopWS: func() {
					events = append(events, "ws:stop")
				},
				StopSSE: func() {
					events = append(events, "sse:stop")
				},
				CloseRedis: func() error {
					events = append(events, "redis:close")
					return nil
				},
				CloseDB: func() error {
					events = append(events, "db:close")
					return nil
				},
				ShutdownTracer: func(context.Context) error {
					events = append(events, "tracer:shutdown")
					return nil
				},
			})
		}),
	)

	app.RequireStart()
	require.Equal(t, []string{
		"ws:start",
		"broadcaster:start",
		"processor:start",
		"scheduler:start",
		"http:start",
	}, events)

	app.RequireStop()
	require.Equal(t, []string{
		"ws:start",
		"broadcaster:start",
		"processor:start",
		"scheduler:start",
		"http:start",
		"http:stop",
		"scheduler:stop",
		"processor:stop",
		"broadcaster:stop",
		"ws:stop",
		"sse:stop",
		"redis:close",
		"db:close",
		"tracer:shutdown",
	}, events)
}

type stubTaskProcessor struct {
	startedCh chan struct{}
	onStop    func()
}

func (s *stubTaskProcessor) Start() error {
	if s.startedCh != nil {
		close(s.startedCh)
	}
	return nil
}

func (s *stubTaskProcessor) Shutdown() {
	if s.onStop != nil {
		s.onStop()
	}
}

func TestRegisterAppLifecycleWithStubApplication(t *testing.T) {
	logger := logrus.New()
	startedCh := make(chan struct{})
	var stopOrder []string

	wsManager := ws2.NewWebSocketManager(NewDefaultWebSocketConfig().ToPkgConfig(), logger, nil, nil)
	sseManager := sse.NewManager()

	application := &Application{
		Log: logger,
		TaskProcessor: &stubTaskProcessor{
			startedCh: startedCh,
			onStop: func() {
				stopOrder = append(stopOrder, "processor:stop")
			},
		},
		BroadcasterCancel: func() {
			stopOrder = append(stopOrder, "broadcaster:stop")
		},
		TracerShutdown: func(context.Context) error {
			stopOrder = append(stopOrder, "tracer:shutdown")
			return nil
		},
		wsManager:  wsManager,
		sseManager: sseManager,
	}

	fxApp := fxtest.New(t,
		fx.Supply(application, logger),
		fx.Invoke(RegisterAppLifecycle),
	)

	fxApp.RequireStart()
	<-startedCh
	fxApp.RequireStop()

	require.Equal(t, []string{
		"processor:stop",
		"broadcaster:stop",
		"tracer:shutdown",
	}, stopOrder)
}
