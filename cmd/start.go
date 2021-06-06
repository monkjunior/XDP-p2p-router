package cmd

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/spf13/cobra"
	"github.com/vu-ngoc-son/XDP-p2p-router/database"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
	"github.com/vu-ngoc-son/XDP-p2p-router/database/geolite2"
	bpfLoader "github.com/vu-ngoc-son/XDP-p2p-router/internal/bpf-loader"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/compute"
	limitBand "github.com/vu-ngoc-son/XDP-p2p-router/internal/limit-band"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/logger"
	myWidget "github.com/vu-ngoc-son/XDP-p2p-router/internal/monitor/widgets"
	packetCapture "github.com/vu-ngoc-son/XDP-p2p-router/internal/packet-capture"
)

var (
	device string

	myLogger = logger.GetLogger()

	hostPublicIP  string
	hostPrivateIP net.IP

	geoDB    *geolite2.GeoLite2
	sqliteDB *dbSqlite.SQLiteDB

	hostInfo *database.Hosts

	pktCapture *packetCapture.PacketCapture
	limiter    *limitBand.BandwidthLimiter

	// TODO: this should be configurable
	fakeData       bool
	updateInterval = time.Second

	grid           *ui.Grid
	ipStats        *myWidget.IPStats
	peerStatsPie   *myWidget.PeersPie
	peerStatsTable *myWidget.PeersTable
	whiteList      *myWidget.WhiteList
	basicInfo      *widgets.Paragraph
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the router",
	Run:   execStartCmd,
}

func init() {
	// TODO: these vars should be configurable
	asnDBPath := fmt.Sprintf("/home/ted/TheFirstProject/XDP-p2p-router/data/geolite2/GeoLite2-ASN_%s/GeoLite2-ASN.mmdb", "20210504")
	cityDBPath := fmt.Sprintf("/home/ted/TheFirstProject/XDP-p2p-router/data/geolite2/GeoLite2-City_%s/GeoLite2-City.mmdb", "20210427")
	countryDBPath := fmt.Sprintf("/home/ted/TheFirstProject/XDP-p2p-router/data/geolite2/GeoLite2-Country_%s/GeoLite2-Country.mmdb", "20210427")

	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&device, "device", "wlp8s0", "network interface that you want to attach this program to it")

	var err error

	hostPublicIP, err = common.GetMyPublicIP()
	if err != nil {
		myLogger.Fatal("failed to get host public ip", zap.Error(err))
	}

	hostPrivateIP, err = common.GetMyPrivateIP(device)
	if err != nil {
		myLogger.Fatal("failed to get host private ip", zap.Error(err))
	}

	geoDB, err = geolite2.NewGeoLite2(asnDBPath, cityDBPath, countryDBPath, hostPublicIP)
	if err != nil {
		myLogger.Fatal("failed to connect to geolite db", zap.Error(err))
	}

	sqliteDB, err = dbSqlite.NewSQLite()
	if err != nil {
		myLogger.Fatal("failed to connect to sqlite db", zap.Error(err))
		return
	}

	hostInfo, err = geoDB.HostInfo()
	if err != nil {
		myLogger.Fatal("failed to query host info", zap.Error(err))
		return
	}

	err = sqliteDB.CreateHost(hostInfo)
	if err != nil {
		myLogger.Fatal("failed to add host info to database", zap.Any("host", hostInfo), zap.Error(err))
		return
	}
}

func execStartCmd(_ *cobra.Command, _ []string) {
	var err error
	m := bpfLoader.LoadModule(hostPrivateIP)

	pktCapture, err = packetCapture.Start(device, m, sqliteDB, geoDB)
	if err != nil {
		myLogger.Fatal("failed to start packet capture module", zap.Error(err))
	}
	defer pktCapture.Close()

	limiter, err = limitBand.NewLimiter(m)
	if err != nil {
		myLogger.Fatal("failed to init limiter module")
	}
	defer limitBand.Close(device, m)

	calculator := compute.NewCalculator(sqliteDB)

	myLogger.Info("starting router ... Ctrl+C to stop.")

	go func() {
		for {
			time.Sleep(15 * time.Second)
			err := calculator.UpdatePeersLimit()
			if err != nil {
				myLogger.Error("calculator | failed to update peer limit ", zap.Error(err))
			}
		}
	}()

	if err := ui.Init(); err != nil {
		myLogger.Fatal("failed to initialize termui: %v\n", zap.Error(err))
	}
	defer ui.Close()

	setDefaultTermUIColors()
	fakeData = true
	initWidgets(fakeData)
	setupGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	fmt.Println(termHeight, termWidth)
	grid.SetRect(0, 0, termWidth, termHeight)
	ui.Render(grid)

	eventLoop()
}

func setDefaultTermUIColors() {
	ui.Theme.Block.Title = ui.NewStyle(ui.ColorCyan)
	ui.Theme.Block.Border = ui.NewStyle(ui.ColorGreen)
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

func initWidgets(fakeData bool) {
	ipStats = myWidget.NewIPStats(updateInterval, sqliteDB, pktCapture.Table, limiter.Table, false)
	peerStatsPie = myWidget.NewPeersPie(updateInterval, sqliteDB, false)
	peerStatsTable = myWidget.NewPeersTable(updateInterval, sqliteDB, fakeData)
	whiteList = myWidget.NewWhiteList(updateInterval, sqliteDB, fakeData)
	basicInfo = widgets.NewParagraph()
	basicInfo.Text = fmt.Sprintf(
		"Public IP\t: %s\nPrivate IP\t: %s\n",
		hostPublicIP, hostPrivateIP.String(),
	)

}

func eventLoop() {
	drawTicker := time.NewTicker(updateInterval).C
	uiEvents := ui.PollEvents()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		case <-drawTicker:
			termWidth, termHeight := ui.TerminalDimensions()
			grid.SetRect(0, 0, termWidth, termHeight)
			ui.Render(grid)
		}
	}
}
