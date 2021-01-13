package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func getEnv(name string) string {
	envValue, ok := os.LookupEnv(name)
	if ok {
		return envValue
	}
	panic(fmt.Sprintf("Missing environment variable: %s", name))
}

func getEnvDefault(name string, defaultVal string) string {
	envValue, ok := os.LookupEnv(name)
	if ok {
		return envValue
	} else {
		return defaultVal
	}
}

func setGauge(name string, help string, callback func() float64) {
	gaugeFunc := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "bitcoind",
		Subsystem: "blockchain",
		Name:      name,
		Help:      help,
	}, callback)
	prometheus.MustRegister(gaugeFunc)
}

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
	btcUser := getEnv("BTC_USER")
	btcPass := getEnv("BTC_PASS")
	btcHost := getEnv("BTC_HOST")
	listendAddr := getEnvDefault("HTTP_LISTENADDR", ":8080")
	config := &rpcclient.ConnConfig{
		Host:         btcHost,
		User:         btcUser,
		Pass:         btcPass,
		DisableTLS:   true,
		HTTPPostMode: true,
	}
	client, err := rpcclient.New(config, nil)
	if err != nil {
		panic(err)
	}
	defer client.Shutdown()
	setGauge("hashrate", "The local blockchain hashrate", func() float64 {
		minfo, err := client.GetMiningInfo()
		if err != nil {
			fmt.Println("hashrate update error", err)
			return -1
		}
		return float64(minfo.NetworkHashPS)
	})

	setGauge("difficulty", "The local blockchain difficulty", func() float64 {
		minfo, err := client.GetMiningInfo()
		if err != nil {
			fmt.Println("difficulty update error", err)
			return -1
		}
		return float64(minfo.Difficulty)
	})

	setGauge("block_count", "The local blockchain length", func() float64 {
		blockCount, err := client.GetBlockCount()
		if err != nil {
			fmt.Println("block count update error", err)
		}
		return float64(blockCount)
	})

	setGauge("raw_mempool_size", "The number of txes in rawmempool", func() float64 {
		hashes, err := client.GetRawMempool()
		if err != nil {
			panic(err)
		}
		return float64(len(hashes))
	})

	setGauge("connected_peers", "The number of connected peers", func() float64 {
		peerInfo, err := client.GetPeerInfo()
		if err != nil {
			panic(err)
		}
		return float64(len(peerInfo))
	})

	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())
	http.Handle("/metrics", promhttp.Handler())
	logrus.Info("Now listening on " + listendAddr)
	logrus.Fatal(http.ListenAndServe(listendAddr, nil))
}
