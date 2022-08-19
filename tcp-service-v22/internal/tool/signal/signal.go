package signal

import (
  "log"
  "os"
  "os/signal"
)

// WaitForShutdown 等待退出
func WaitForShutdown() {
  // 等待系统信号
  chansignal := make(chan os.Signal, 1)
  // os.Interrupt == SIGINT（ctrl+c）
  signal.Notify(chansignal, os.Interrupt)
  // 如果没有收到 SIGINT 信号，就会阻塞在这里
  <-chansignal
  log.Println("get SIGINT signal, shutdown immediately...")
}
