package app

import (
	"context"
	"mashaghel/internal/config"
	"mashaghel/internal/tasks"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Application interface {
	Setup()
}

type application struct {
	ctx    context.Context
	config *config.Config
}

func NewApplication(ctx context.Context, config *config.Config) Application {
	return &application{ctx: ctx, config: config}
}

// bootstrap

func (a *application) Setup() {
	app := fx.New(
		fx.Provide(
			// a.InitRouter,
			// a.InitFramework,
			// a.InitController,
			// a.InitServices,
			// a.InitRepositories,
			// a.InitRedis,
			// a.InitArangoDB,
			a.InitScyllaDB,
			a.InitLogger,
			a.InitTracerProvider,
			// a.InitGRPCServer,
			// a.InitNats,
			a.InitTask,
		),

		// fx.Invoke(func(lc fx.Lifecycle, connection nats.NatsConnection, logger *zap.Logger) {
		// 	lc.Append(fx.Hook{
		// 		OnStart: func(_ context.Context) error {
		// 			logger.Info("Starting nats connection")
		// 			return nil
		// 		},
		// 		OnStop: func(ctx context.Context) error {
		// 			connection.Close()
		// 			return nil
		// 		},
		// 	})
		// }),

		// fx.Invoke(func(lc fx.Lifecycle, nats nats.NatsConnection, logger *zap.Logger) {
		// 	lc.Append(fx.Hook{
		// 		OnStart: func(_ context.Context) error {
		// 			return nil
		// 		},
		// 		OnStop: func(ctx context.Context) error {
		// 			logger.Info("Closing nats connection ...")
		// 			nats.Close()
		// 			return nil
		// 		},
		// 	})
		// }),

		// fx.Invoke(func(lc fx.Lifecycle, grpcServer *grpc.Server, logger *zap.Logger) {
		// 	logger.Info("Initializing gRPC server")
		// 	lc.Append(fx.Hook{
		// 		OnStart: func(_ context.Context) error {
		// 			listener, err := net.Listen("tcp", a.config.GRPC.Host+":"+a.config.GRPC.Port)
		// 			if err != nil {
		// 				logger.Info("Failed to listen on gRPC port", zap.Error(err))
		// 				return err
		// 			}
		// 			logger.Info("starting gRPC Server",
		// 				zap.String("host", a.config.GRPC.Host),
		// 				zap.String("port", a.config.GRPC.Port),
		// 			)
		// 			go func() {
		// 				if err := grpcServer.Serve(listener); err != nil {
		// 					logger.Info("Failed to serve gRPC", zap.Error(err))
		// 				}
		// 			}()
		// 			// log.Println("gRPC server started on", a.config.GRPC.Host+":"+a.config.GRPC.Port)
		// 			return nil
		// 		},
		// 		OnStop: func(_ context.Context) error {
		// 			grpcServer.Stop()
		// 			logger.Info("gRPC server stopped")
		// 			logger.Sync()
		// 			return nil
		// 		},
		// 	})
		// }),

		// fx.Invoke(func(lc fx.Lifecycle, app *fiber.App, logger *zap.Logger) {
		// 	// Start Fiber server in a separate goroutine
		// 	lc.Append(fx.Hook{
		// 		OnStart: func(_ context.Context) error {
		// 			logger.Info("Starting Fiber server")
		// 			go func() {
		// 				if err := app.Listen(a.config.Server.Host + ":" + a.config.Server.Port); err != nil {
		// 					logger.Error("Failed to start Fiber server", zap.Error(err))
		// 				}
		// 			}()
		// 			return nil
		// 		},
		// 		OnStop: func(ctx context.Context) error {
		// 			logger.Sync()
		// 			return nil
		// 		},
		// 	})
		// }),

		// fx.Invoke(func(app *fiber.App, router routers.Router) {
		// 	router.AddRoutes(app.Group(""))
		// }),

		fx.Invoke(func(lc fx.Lifecycle, t tasks.Task, logger *zap.Logger) {
			// Start workerpool
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					if err := t.InitWorkerPool(); err != nil {
						logger.Error("Failed to initialize workerpool", zap.Error(err))
						panic(err)
					}
					t.Start()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					t.Stop()
					return nil
				},
			})
		}),
	)
	app.Run()
}
