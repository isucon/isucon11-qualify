package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"

	// TODO: isucon11-portal に差し替え
	"github.com/isucon/isucon10-portal/bench-tool.go/benchrun"
	isuxportalResources "github.com/isucon/isucon10-portal/proto.go/isuxportal/resources"

	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/scenario"
)

const (
	// FAIL になるエラー回数
	FAIL_ERROR_COUNT int64 = 100 //TODO:ちゃんと決める
	//load context
	LOAD_TIMEOUT time.Duration = 70 * time.Second
)

var (
	COMMIT             string
	targetAddress      string
	profileFile        string
	hostAdvertise      string
	jiaServiceURL      *url.URL
	tlsCertificatePath string
	tlsKeyPath         string
	useTLS             bool
	exitStatusOnFail   bool
	noLoad             bool
	promOut            string
	showVersion        bool

	initializeTimeout time.Duration
	// TODO: isucon11-portal に差し替え
	reporter benchrun.Reporter
)

func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defaultValue
}

func init() {
	certs, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}

	// https://github.com/golang/go/issues/16012#issuecomment-224948823
	// "connect: cannot assign requested address" 対策
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100

	agent.DefaultTLSConfig.ClientCAs = certs
	agent.DefaultTLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
	agent.DefaultTLSConfig.MinVersion = tls.VersionTLS12
	agent.DefaultTLSConfig.InsecureSkipVerify = false

	flag.StringVar(&targetAddress, "target", benchrun.GetTargetAddress(), "ex: localhost:9292")
	flag.StringVar(&profileFile, "profile", "", "ex: cpu.out")
	flag.BoolVar(&exitStatusOnFail, "exit-status", false, "set exit status non-zero when a benchmark result is failing")
	flag.BoolVar(&noLoad, "no-load", false, "exit on finished prepare")
	flag.StringVar(&promOut, "prom-out", "", "Prometheus textfile output path")
	flag.BoolVar(&showVersion, "version", false, "show version and exit 1")

	var jiaServiceURLStr, timeoutDuration, initializeTimeoutDuration string
	flag.StringVar(&jiaServiceURLStr, "jia-service-url", getEnv("JIA_SERVICE_URL", "http://apitest:5000"), "jia service url")
	flag.StringVar(&timeoutDuration, "timeout", "1s", "request timeout duration")
	flag.StringVar(&initializeTimeoutDuration, "initialize-timeout", "20s", "request timeout duration of POST /initialize")

	flag.Parse()

	// validate jia-service-url
	jiaServiceURL, err = url.Parse(jiaServiceURLStr)
	if err != nil {
		panic(err)
	}
	// validate timeout
	timeout, err := time.ParseDuration(timeoutDuration)
	if err != nil {
		panic(err)
	}
	agent.DefaultRequestTimeout = timeout

	// validate initialize-timeout
	initializeTimeout, err = time.ParseDuration(initializeTimeoutDuration)
	if err != nil {
		panic(err)
	}
}

func checkError(err error) (critical bool, timeout bool, deduction bool) {
	return scenario.CheckError(err)
}

func sendResult(s *scenario.Scenario, result *isucandar.BenchmarkResult, finish bool) bool {
	passed := true
	reason := "pass"
	errors := result.Errors.All()

	scoreRaw := result.Score.Sum()
	deduction := int64(0)
	timeoutCount := int64(0)

	for tag, count := range result.Score.Breakdown() {
		if finish {
			logger.ContestantLogger.Printf("SCORE: %s: %d", tag, count)
		} else {
			logger.AdminLogger.Printf("SCORE: %s: %d", tag, count)
		}
	}

	for _, err := range errors {
		isCritical, isTimeout, isDeduction := checkError(err)

		switch true {
		case isCritical:
			passed = false
			reason = "Critical error"
			logger.AdminLogger.Printf("Critical error because: %+v\n", err) //TODO: Contestantでも良いかも
		case isTimeout:
			timeoutCount++
		case isDeduction:
			if scenario.IsValidation(err) {
				deduction += 50
			} else {
				deduction++
			}
		}
	}
	deductionTotal := deduction + timeoutCount/10 //TODO:

	if passed && deduction > FAIL_ERROR_COUNT {
		passed = false
		reason = fmt.Sprintf("Error count over %d", FAIL_ERROR_COUNT)
	}

	score := scoreRaw - deductionTotal
	if passed && !s.NoLoad && score <= 0 {
		passed = false
		reason = "Score"
	}

	if !passed {
		score = 0
	}

	logger.ContestantLogger.Printf("score: %d(%d - %d) : %s", score, scoreRaw, deductionTotal, reason)
	logger.ContestantLogger.Printf("deduction: %d / timeout: %d", deduction, timeoutCount)

	err := reporter.Report(&isuxportalResources.BenchmarkResult{
		SurveyResponse: &isuxportalResources.SurveyResponse{
			Language: s.Language,
		},
		Finished: finish,
		Passed:   passed,
		Score:    score,
		ScoreBreakdown: &isuxportalResources.BenchmarkResult_ScoreBreakdown{
			Raw:       scoreRaw,
			Deduction: deductionTotal,
		},
		Execution: &isuxportalResources.BenchmarkResult_Execution{
			Reason: reason,
		},
	})
	if err != nil {
		panic(err)
	}

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
	logger.AdminLogger.Printf("ISUCON11 benchmarker %s", COMMIT)

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

	s, err := scenario.NewScenario(jiaServiceURL, LOAD_TIMEOUT)
	if err != nil {
		panic(err)
	}
	s = s.WithInitializeTimeout(initializeTimeout)
	scheme := "http"
	if useTLS {
		scheme = "https"
	}
	s.BaseURL = fmt.Sprintf("%s://%s/", scheme, targetAddress)
	s.NoLoad = noLoad

	// JIA API
	go func() {
		s.JiaCancel = cancel
		s.JiaAPIService(ctx)
	}()

	// Benchmarker
	b, err := isucandar.NewBenchmark(isucandar.WithoutPanicRecover())
	if err != nil {
		panic(err)
	}

	reporter, err = benchrun.NewReporter(false)
	if err != nil {
		panic(err)
	}

	errorCount := int64(0)
	b.OnError(func(err error, step *isucandar.BenchmarkStep) {
		critical, timeout, deduction := checkError(err)

		// Load 中の timeout のみログから除外
		if timeout && failure.IsCode(err, isucandar.ErrLoad) {
			return
		}

		if critical || (deduction && atomic.AddInt64(&errorCount, 1) > FAIL_ERROR_COUNT) {
			step.Cancel()
		}

		logger.ContestantLogger.Printf("ERR: %v", err)
	})

	b.AddScenario(s)

	wg := sync.WaitGroup{}
	b.Load(func(parent context.Context, step *isucandar.BenchmarkStep) error {
		//このWaitGroupで、sendResult(,,true)が呼ばれた後sendResult(,,false)が呼ばれないことを保証する
		//isucandarのparallelはcontextが終了した場合に、スレッドの終了を待たずにWaitを終了する
		//そこで、wg.Done()が呼ばれsendResult(,,false)の送信を終了した後に、sendResult(,,true)の処理に移る
		//
		//コーナーケースとして、wg.Add(1)する前にb.Start(ctx)が終了しwg.Wait()を突破する可能性がある（loadの開始直後にCriticalErrorの場合など）
		//その場合でも、この関数はctx.Done()を検出して早期returnすることでsendResult(,,false)を実行しないため、保証できている
		//
		//補足：
		//　wg.Addをgoroutine内で呼ぶとこのコーナーケースを引き起こすので一般にはgoroutine生成直前にAddするべき
		//　今回は生成直前がisucandar内にあり、そのタイミングでのAddが出来ないためここに記述
		//　b.Startの前にAddする実装も考えたが、Prepareフェーズでstep.Cancel()された場合に、
		//　Load自体がスキップされwg.Doneが実行されずにデッドロックを起こしたためボツ
		wg.Add(1)
		defer wg.Done()
		if s.NoLoad {
			return nil
		}

		ctx, cancel := context.WithTimeout(parent, s.LoadTimeout)
		defer cancel()
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// 初期実装だと fail してしまうため下駄をはかせる
		step.AddScore(scenario.ScoreStartBenchmark)

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
