package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"webFrame/dao/mysql"
	"webFrame/dao/redis"
	"webFrame/logger"
	"webFrame/routes"
	"webFrame/settings"

	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("need input file.eg: conf.yaml")
		return
	}
	// 加载配置
	err := settings.Init(os.Args[1])
	if err != nil {
		fmt.Printf("load conf failed, err:%v\n", err)
		return
	}
	err = logger.Init(settings.Conf.LogConfig, settings.Conf.Mode)
	if err != nil {
		fmt.Println("logger failed")
		return
	}
	defer zap.L().Sync()
	err = mysql.Init(settings.Conf.MySQLConfig)
	if err != nil {
		fmt.Println("mysql failed")
		return
	}
	defer mysql.Close()
	err = redis.Init(settings.Conf.RedisConfig)
	if err != nil {
		fmt.Println("redis failed")
		return
	}
	defer redis.Close()
	router := routes.Setup(settings.Conf.Mode)
	// 优雅重启
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.Conf.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("listen:", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown:", zap.Error(err))
	}

	zap.L().Info("Server exiting")
}
