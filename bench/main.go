package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"

	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/scenario"
)

var (
	COMMIT             string
	targetAddress      string
	profileFile        string
	hostAdvertise      string
	tlsCertificatePath string
	tlsKeyPath         string
	useTLS             bool
	exitStatusOnFail   bool
	noLoad             bool
	promOut            string
	showVersion        bool

	// TODO: isucon11-portal に差し替え
	//reporter benchrun.Reporter
)

func init() {
	certs, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}

	agent.DefaultTLSConfig.ClientCAs = certs
	agent.DefaultTLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
	agent.DefaultTLSConfig.MinVersion = tls.VersionTLS12
	agent.DefaultTLSConfig.InsecureSkipVerify = false

	// TODO: isucon11-portal に差し替え
	//flag.StringVar(&targetAddress, "target", benchrun.GetTargetAddress(), "ex: localhost:9292")
	flag.StringVar(&targetAddress, "target", "localhost:9292", "ex: localhost:9292")
	flag.StringVar(&profileFile, "profile", "", "ex: cpu.out")
	flag.StringVar(&hostAdvertise, "host-advertise", "local.t.isucon.dev", "hostname to advertise against target")
	flag.StringVar(&tlsCertificatePath, "tls-cert", "../secrets/cert.pem", "path to TLS certificate for a push service")
	flag.StringVar(&tlsKeyPath, "tls-key", "../secrets/key.pem", "path to private key of TLS certificate for a push service")
	flag.BoolVar(&exitStatusOnFail, "exit-status", false, "set exit status non-zero when a benchmark result is failing")
	flag.BoolVar(&noLoad, "no-load", false, "exit on finished prepare")
	flag.StringVar(&promOut, "prom-out", "", "Prometheus textfile output path")
	flag.BoolVar(&showVersion, "version", false, "show version and exit 1")

	timeoutDuration := ""
	flag.StringVar(&timeoutDuration, "timeout", "10s", "request timeout duration")

	flag.Parse()

	timeout, err := time.ParseDuration(timeoutDuration)
	if err != nil {
		panic(err)
	}
	agent.DefaultRequestTimeout = timeout
}

func checkError(err error) (critical bool, timeout bool, deduction bool) {
	critical = false  // TODO: クリティカルなエラー(起きたら即ベンチを止める)
	timeout = false   // TODO: リクエストタイムアウト(ある程度の数許容するかも)
	deduction = false // TODO: 減点対象になるエラー

	return
}

func sendResult(s *scenario.Scenario, result *isucandar.BenchmarkResult, finish bool) bool {
	passed := true
	reason := ""
	errors := result.Errors.All()

	deduction := int64(0)
	timeoutCount := int64(0)

	for _, err := range errors {
		isCritical, isTimeout, isDeduction := checkError(err)

		switch true {
		case isCritical:
			passed = false
			reason = "Critical error"
		case isTimeout:
			timeoutCount++
		case isDeduction:
			deduction++
		}
	}

	// TODO: isucon11-portal に差し替え
	/*
		err := reporter.Report(&isuxportalResources.BenchmarkResult{
			SurveyResponse: &isuxportalResources.SurveyResponse{
				Language: s.Language,
			},
			Finished: finish,
			Passed:   passed,
			Score:    0, // TODO: 加点 - 減点
			ScoreBreakdown: &isuxportalResources.BenchmarkResult_ScoreBreakdown{
				Raw:       0, // TODO: 加点
				Deduction: 0, // TODO: 減点
			},
			Execution: &isuxportalResources.BenchmarkResult_Execution{
				Reason: reason,
			},
		})
		if err != nil {
			panic(err)
		}
	*/
	// TODO: 以下は消す
	fmt.Println(reason)

	return passed
}

func writePromFile(promTags []string) {
	if len(promOut) == 0 {
		return
	}

	promOutNew := fmt.Sprintf("%s.new", promOut)
	err := ioutil.WriteFile(promOutNew, []byte(strings.Join(promTags, "")), 0644)
	if err != nil {
		logger.AdminLogger.Printf("Failed to write prom file: %s", err)
		return
	}
	err = os.Rename(promOutNew, promOut)
	if err != nil {
		logger.AdminLogger.Printf("Failed to write prom file: %s", err)
		return
	}

}

func main() {
	logger.AdminLogger.Printf("ISUCON10 benchmarker %s", COMMIT)

	if showVersion {
		os.Exit(1)
	}

	if profileFile != "" {
		fs, err := os.Create(profileFile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(fs)
		defer pprof.StopCPUProfile()
	}
	if targetAddress == "" {
		targetAddress = "localhost:9292"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := scenario.NewScenario()
	scheme := "http"
	if useTLS {
		scheme = "https"
	}
	s.BaseURL = fmt.Sprintf("%s://%s/", scheme, targetAddress)
	s.NoLoad = noLoad

	b, err := isucandar.NewBenchmark(isucandar.WithLoadTimeout(70 * time.Second))
	if err != nil {
		panic(err)
	}

	// TODO: isucon11-portal に差し替え
	/*
		reporter, err = benchrun.NewReporter(false)
		if err != nil {
			panic(err)
		}
	*/

	errorCount := int64(0)
	b.OnError(func(err error, step *isucandar.BenchmarkStep) {
		// Load 中の timeout のみログから除外
		if failure.IsCode(err, failure.TimeoutErrorCode) && failure.IsCode(err, isucandar.ErrLoad) {
			return
		}

		critical, _, deduction := checkError(err)

		if critical || (deduction && atomic.AddInt64(&errorCount, 1) >= 100) {
			step.Cancel()
		}

		logger.ContestantLogger.Printf("ERR: %v", err)
	})

	b.AddScenario(s)

	wg := sync.WaitGroup{}
	b.Load(func(ctx context.Context, step *isucandar.BenchmarkStep) error {
		if s.NoLoad {
			return nil
		}

		wg.Add(1)
		defer wg.Done()

		for {
			// 途中経過を3秒毎に送信
			timer := time.After(3 * time.Second)
			sendResult(s, step.Result(), false)

			select {
			case <-timer:
			case <-ctx.Done():
				return nil
			}
		}
	})

	result := b.Start(ctx)

	wg.Wait()

	if !sendResult(s, result, true) && exitStatusOnFail {
		os.Exit(1)
	}
}
