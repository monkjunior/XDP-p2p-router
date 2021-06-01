package cmd

import (
	"fmt"
	"github.com/gizak/termui/v3/widgets"
	"log"

	"math"
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/spf13/cobra"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
	"github.com/vu-ngoc-son/XDP-p2p-router/database/geolite2"
	bpfLoader "github.com/vu-ngoc-son/XDP-p2p-router/internal/bpf-loader"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/compute"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/ip2location"
	limitBand "github.com/vu-ngoc-son/XDP-p2p-router/internal/limit-band"
	myWidget "github.com/vu-ngoc-son/XDP-p2p-router/internal/monitor/widgets"
	packetCapture "github.com/vu-ngoc-son/XDP-p2p-router/internal/packet-capture"
)

var (
	device string

	stderrLogger = log.New(os.Stderr, "", 0)

	grid           *ui.Grid
	ipStats        *myWidget.IPStats
	peerStatsPie   *widgets.PieChart
	peerStatsTable *widgets.Table
	whiteList      *widgets.Table
	basicInfo      *widgets.Paragraph

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
		stderrLogger.Fatalln("failed to connect to sqlite", err)
		return
	}

	hostInfo, err := geoDB.HostInfo()
	if err != nil {
		stderrLogger.Fatalln("failed to query host info", err)
		return
	}

	err = sqliteDB.CreateHost(hostInfo)
	if err != nil {
		stderrLogger.Fatalln("failed to add host info to database", err)
		return
	}

	m := bpfLoader.LoadModule(device)
	pktCapture, err := packetCapture.Start(device, m)
	if err != nil {
		stderrLogger.Fatalln("failed to start packet capture module", err)
	}
	defer packetCapture.Close(device, m)
	limiter, err := limitBand.NewLimiter(m)
	if err != nil {
		stderrLogger.Fatalln("failed to init limiter module")
	}
	defer limitBand.Close(device, m)

	locator := ip2location.NewLocator(pktCapture, sqliteDB, geoDB)
	calculator := compute.NewCalculator(sqliteDB)
	//watchDog := monitor.NewMonitor(pktCapture, limiter, sqliteDB)

	stderrLogger.Println("starting router ... Ctrl+C to stop.")


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
				stderrLogger.Println("calculator | failed to update peer limit ", err)
			}
		}
	}()
	go func() {
		for {
			time.Sleep(15 * time.Second)
			err := calculator.UpdatePeersLimit()
			if err != nil {
				stderrLogger.Println("calculator | failed to update peer limit ", err)
			}
		}
	}()
	go func() {
		for {
			time.Sleep(5 * time.Second)
			_, err := limiter.ExportMap()
			if err != nil {
				stderrLogger.Println("limiter | failed to export map ", err)
			}
		}
	}()

	if err := ui.Init(); err != nil {
		stderrLogger.Fatalln("failed to initialize termui: %v\n", err)
	}
	defer ui.Close()


	initWidgets()

	setupGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	fmt.Println(termHeight, termWidth)
	grid.SetRect(0, 0, termWidth-1, termHeight-1)
	ui.Render(grid)

	eventLoop()
}

func setupGrid() {
	grid = ui.NewGrid()

	grid.Set(
		ui.NewRow(1.0/2,
			ui.NewCol(1.0/2, ipStats),
			ui.NewCol(1.0/4, peerStatsPie),
			ui.NewCol(1.0/4, peerStatsTable),
		),
		ui.NewRow(1.0/2,
			ui.NewCol(1.0/2, basicInfo),
			ui.NewCol(1.0/2, whiteList),
		),
	)
}

func initWidgets() {
	ipStats = myWidget.NewIPStats(1)

	peerStatsPie = widgets.NewPieChart()
	peerStatsPie.Title = "Pie Chart"
	peerStatsPie.Data = []float64{.25, .25, .25, .25}
	peerStatsPie.AngleOffset = -.5 * math.Pi
	peerStatsPie.LabelFormatter = func(i int, v float64) string {
		return fmt.Sprintf("%.02f", v)
	}

	peerStatsTable = widgets.NewTable()
	peerStatsTable.Rows = [][]string{
		{"header1", "header2", "header3"},
		{"你好吗", "Go-lang is so cool", "Im working on Ruby"},
		{"2016", "10", "11"},
	}

	whiteList = widgets.NewTable()
	whiteList.Rows = [][]string{
		{"header1", "header2", "header3"},
		{"你好吗", "Go-lang is so cool", "Im working on Ruby"},
		{"2016", "10", "11"},
	}

	basicInfo = widgets.NewParagraph()
	basicInfo.Text = "Hahaha"
}

func eventLoop() {
	uiEvents := ui.PollEvents()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		}
	}
}
