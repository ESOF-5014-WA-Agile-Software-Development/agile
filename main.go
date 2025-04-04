package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/api/emails"
	"github.com/mylakehead/agile/api/me"
	"github.com/mylakehead/agile/api/metamask"
	"github.com/mylakehead/agile/api/users"
	"github.com/mylakehead/agile/runtime"
)

func httpServer(rt *runtime.Runtime) *http.Server {
	if rt.Config.Mode.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, nil)
	})

	r := router.Group("/api")
	{
		r.GET("/exists/users/:name", api.Wrap(users.Exists, rt, false, api.WithDataType(api.DataTypeJson)))
		r.GET("/exists/emails/:email", api.Wrap(emails.Exists, rt, false, api.WithDataType(api.DataTypeJson)))
		r.GET("/exists/metamask/:address", api.Wrap(metamask.Exists, rt, false, api.WithDataType(api.DataTypeJson)))
		r.GET("/metamask/:address/nonce", api.Wrap(metamask.Nonce, rt, false, api.WithDataType(api.DataTypeJson)))
		r.POST("/verify/:type", api.Wrap(emails.Verify, rt, false, api.WithDataType(api.DataTypeJson)))
		r.POST("/sign-up/:type", api.Wrap(users.SignUp, rt, false, api.WithDataType(api.DataTypeJson)))
		r.POST("/sign-in/:type", api.Wrap(users.SignIn, rt, false, api.WithDataType(api.DataTypeJson)))

		r.GET("/me/ongoing", api.Wrap(me.Ongoing, rt, true, api.WithDataType(api.DataTypeJson)))
		r.GET("/me/prediction/:id", api.Wrap(me.GetPrediction, rt, true, api.WithDataType(api.DataTypeJson)))
		r.POST("/me/prediction/:id", api.Wrap(me.UpdatePrediction, rt, true, api.WithDataType(api.DataTypeJson)))
	}

	addr := fmt.Sprintf("%s:%d", rt.Config.HTTP.Host, rt.Config.HTTP.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return server
}

func grpcServer(_ *runtime.Runtime) *http.Server {
	return nil
}

func onExit(rt *runtime.Runtime, httpServer *http.Server) {
	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal)
	defer close(quit)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can't be caught
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit

	println()
	log.Printf("[exit] system signal: %s received, shutting down gracefully", s.String())

	// shut down the http server with a timeout of 5 seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("[exit] close http server error: %v", err)
		os.Exit(1)
	}
	cancel()

	select {
	case <-ctx.Done():
		if deadline, ok := ctx.Deadline(); ok {
			log.Printf("[exit] close http server in %v.", 5*time.Second-time.Until(deadline))
		}
	}

	// shutdown the runtime with a timeout of 5 seconds
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rt.Close(ctx); err != nil {
		log.Printf("[exit] close runtime error: %v", err)
		os.Exit(1)
	}
	cancel()

	select {
	case <-ctx.Done():
		if deadline, ok := ctx.Deadline(); ok {
			log.Printf("[exit] close runtime in %v.", 5*time.Second-time.Until(deadline))
		}
	}

	log.Print("[exit] shutdown successfully")
	os.Exit(0)
}

func main() {
	rt, err := runtime.New()
	if err != nil {
		log.Fatal(err)
	}

	hs := httpServer(rt)
	go func() {
		if rt.Config.HTTP.TLS {
			log.Printf("https server listening on: %s", hs.Addr)
			if err := hs.ListenAndServeTLS(rt.Config.HTTP.Crt, rt.Config.HTTP.Key); err != nil {
				log.Fatalf("liste TLS: %s\n", err)
			}
		} else {
			log.Printf("http server listening on: %s", hs.Addr)
			if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("listen: %s\n", err)
			}
		}
	}()

	_ = grpcServer(rt)
	go func() {
		return
	}()

	onExit(rt, hs)
}
