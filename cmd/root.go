// Copyright © 2018 Ken'ichiro Oyama <k1lowxb@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"context"
	"net"
	"os"

	"github.com/k1LoW/tcprxy/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os/signal"
	"syscall"
)

var listenAddr string
var remoteAddr string
var dumper string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tcprxy",
	Short: "tcprxy",
	Long:  `tcprxy`,
	Run: func(cmd *cobra.Command, args []string) {
		lAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		rAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		signalChan := make(chan os.Signal, 1)
		signal.Ignore()
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

		s := server.NewServer(context.Background(), lAddr, rAddr)

		log.Printf("Starting server. %s:%d <-> %s:%d\n", lAddr.IP, lAddr.Port, rAddr.IP, rAddr.Port)
		go s.Start()

		sc := <-signalChan

		switch sc {
		case syscall.SIGINT:
			log.Println("Shutting down server...")
			s.Shutdown()
			s.Wg.Wait()
			<-s.ClosedChan
		case syscall.SIGQUIT, syscall.SIGTERM:
			log.Println("Graceful Shutting down server...")
			s.GracefulShutdown()
			s.Wg.Wait()
			<-s.ClosedChan
		default:
			log.Println("Unexpected signal")
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&listenAddr, "listen", "l", "localhost:8080", "listen address")
	rootCmd.Flags().StringVarP(&remoteAddr, "remote", "r", "localhost:80", "remote address")
	rootCmd.Flags().StringVarP(&dumper, "dumper", "d", "hex", "dumper")
	viper.BindPFlag("dumper", rootCmd.Flags().Lookup("dumper"))
}
