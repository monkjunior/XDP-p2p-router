package cmd

import (
	"fmt"
	"github.com/gizak/termui/v3/widgets"
	"log"
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
	//"github.com/vu-ngoc-son/XDP-p2p-router/internal/monitor"
	myWidgets "github.com/vu-ngoc-son/XDP-p2p-router/internal/monitor/widgets"
	packetCapture "github.com/vu-ngoc-son/XDP-p2p-router/internal/packet-capture"
)

var (
	device string

	stderrLogger = log.New(os.Stderr, "", 0)
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

	done := make(chan bool)

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

	header := widgets.NewParagraph()
	header.Text = "Press q to quit, Press h or l to switch tabs"
	header.SetRect(0, 0, 50, 1)
	header.Border = true
	header.TextStyle.Fg = ui.ColorWhite

	p2 := widgets.NewParagraph()
	p2.Text = fmt.Sprintf(`
Public IP: %s
Private IP: %s
`, )
	p2.Title = "Basic information"
	p2.SetRect(5, 5, 40, 15)
	p2.BorderStyle.Fg = ui.ColorYellow

	bc := widgets.NewBarChart()
	bc.Title = "Bar Chart"
	bc.Data = []float64{3, 2, 5, 3, 9, 5, 3, 2, 5, 8, 3, 2, 4, 5, 3, 2, 5, 7, 5, 3, 2, 6, 7, 4, 6, 3, 6, 7, 8, 3, 6, 4, 5, 3, 2, 4, 6, 4, 8, 5, 9, 4, 3, 6, 5, 3, 6}
	bc.SetRect(5, 5, 35, 10)
	bc.Labels = []string{"S0", "S1", "S2", "S3", "S4", "S5"}

	tabpane := widgets.NewTabPane("basic info", "drugi", "limits", "żółw", "four", "five")
	tabpane.SetRect(0, 1, 50, 4)
	tabpane.Border = true

	limitsWidget := myWidgets.NewIPList(sqliteDB, time.Second)

	renderTab := func() {
		switch tabpane.ActiveTabIndex {
		case 0:
			ui.Render(p2)
		case 1:
			ui.Render(bc)
		case 2:
			ui.Render(limitsWidget)
		}
	}

	ui.Render(header, tabpane, p2)

	uiEvents := ui.PollEvents()

	go func() {
		for {
			e := <-uiEvents

			switch e.ID {
			case "q", "<C-c>":
				done <- true
			case "h":
				tabpane.FocusLeft()
				ui.Clear()
				ui.Render(header, tabpane)
				renderTab()
			case "l":
				tabpane.FocusRight()
				ui.Clear()
				ui.Render(header, tabpane)
				renderTab()
			}
		}
	}()

	finished := <-done
	stderrLogger.Println("stopping router done %v\n", finished)
}
