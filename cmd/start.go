package cmd

import (
	"fmt"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/compute"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/ip2location"
	limitBand "github.com/vu-ngoc-son/XDP-p2p-router/internal/limit-band"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/monitor"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
	"github.com/vu-ngoc-son/XDP-p2p-router/database/geolite2"
	bpfLoader "github.com/vu-ngoc-son/XDP-p2p-router/internal/bpf-loader"
	packetCapture "github.com/vu-ngoc-son/XDP-p2p-router/internal/packet-capture"
)

var (
	device string
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the router",
	Run:   execStartCmd,
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&device, "device", "wlp8s0", "network interface that you want to attach this program to it")
}

func execStartCmd(_ *cobra.Command, _ []string) {
	asnDBPath := fmt.Sprintf("/home/ted/TheFirstProject/XDP-p2p-router/data/geolite2/GeoLite2-ASN_%s/GeoLite2-ASN.mmdb", "20210504")
	cityDBPath := fmt.Sprintf("/home/ted/TheFirstProject/XDP-p2p-router/data/geolite2/GeoLite2-City_%s/GeoLite2-City.mmdb", "20210427")
	countryDBPath := fmt.Sprintf("/home/ted/TheFirstProject/XDP-p2p-router/data/geolite2/GeoLite2-Country_%s/GeoLite2-Country.mmdb", "20210427")
	sqliteDBPath := "/home/ted/TheFirstProject/XDP-p2p-router/data/sqlite/p2p-router.db"

	geoDB := geolite2.NewGeoLite2(asnDBPath, cityDBPath, countryDBPath)

	sqliteDB, err := dbSqlite.NewSQLite(sqliteDBPath)
	if err != nil {
		fmt.Println("failed to connect to sqlite", err)
		return
	}

	hostInfo, err := geoDB.HostInfo()
	if err != nil {
		fmt.Println("failed to query host info", err)
		return
	}

	err = sqliteDB.CreateHost(hostInfo)
	if err != nil {
		fmt.Println("failed to add host info to database", err)
		return
	}

	m := bpfLoader.LoadModule(device)
	pktCapture, err := packetCapture.Start(device, m)
	if err != nil {
		fmt.Println("failed to start packet capture module")
		os.Exit(1)
	}
	defer packetCapture.Close(device, m)
	limiter, err := limitBand.NewLimiter(m)
	if err != nil {
		fmt.Println("failed to init limiter module")
		os.Exit(1)
	}
	defer limitBand.Close(device, m)

	locator := ip2location.NewLocator(pktCapture, sqliteDB, geoDB)
	calculator := compute.NewCalculator(sqliteDB)
	watchDog := monitor.NewMonitor(pktCapture, limiter, sqliteDB)

	fmt.Println("starting router ... Ctrl+C to stop.")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)
	exitWatchdog := make(chan int)
	go func() {
		sig := <-signals
		fmt.Printf("\n%v\n", sig)
		exitWatchdog <- 1
		done <- true
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			locator.UpdatePeersToDB()
		}
	}()

	go func() {
		for {
			time.Sleep(15 * time.Second)
			err := calculator.UpdatePeersLimit()
			if err != nil {
				fmt.Println("calculator | failed to update peer limit ", err)
			}
		}
	}()
	go func() {
		for {
			time.Sleep(15 * time.Second)
			err := calculator.UpdatePeersLimit()
			if err != nil {
				fmt.Println("calculator | failed to update peer limit ", err)
			}
		}
	}()
	go func() {
		for {
			time.Sleep(5 * time.Second)
			_, err := limiter.ExportMap()
			if err != nil {
				fmt.Println("limiter | failed to export map ", err)
			}
		}
	}()

	watchDog.ExportThroughput(&exitWatchdog, 1)

	_ = <-done
	fmt.Println("shutting down gracefully ...")
}
