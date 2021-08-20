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
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/score"
	"github.com/pkg/profile"

	// TODO: isucon11-portal に差し替える (isucon/isucon11-portal#167)
	"github.com/isucon/isucon10-portal/bench-tool.go/benchrun"
	isuxportalResources "github.com/isucon/isucon10-portal/proto.go/isuxportal/resources"

	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/scenario"
)

const (
	// FAIL になるエラー回数
	FAIL_ERROR_COUNT int64 = 100
	//load context
	LOAD_TIMEOUT time.Duration = 60 * time.Second
)

var (
	allowedTargetFQDN = []string{
		"isucondition-1.t.isucon.dev",
		"isucondition-2.t.isucon.dev",
		"isucondition-3.t.isucon.dev",
	}
)

var (
	COMMIT              string
	targetAddress       string
	targetableAddresses []string
	profileFile         string
	memProfileDir       string
	jiaServiceURL       *url.URL
	useTLS              bool
	exitStatusOnFail    bool
	noLoad              bool
	promOut             string
	showVersion         bool

	initializeTimeout time.Duration
	reporter          benchrun.Reporter
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

	var targetableAddressesStr string

	flag.StringVar(&targetAddress, "target", benchrun.GetTargetAddress(), "ex: localhost:9292")
	// TODO: benchrun.GetAllAddresses で環境変数を読み込む (isucon/isucon11-portal#167)
	flag.StringVar(&targetableAddressesStr, "all-addresses", getEnv("ISUXBENCH_ALL_ADDRESSES", ""), `ex: "192.168.0.1,192.168.0.2,192.168.0.3" (comma separated, limit 3)`)
	flag.StringVar(&profileFile, "profile", "", "ex: cpu.out")
	flag.StringVar(&memProfileDir, "mem-profile", "", "path of output heap profile at max memStats.sys allocated. ex: memprof")
	flag.BoolVar(&exitStatusOnFail, "exit-status", false, "set exit status non-zero when a benchmark result is failing")
	flag.BoolVar(&useTLS, "tls", false, "true if target server is a tls")
	flag.BoolVar(&noLoad, "no-load", false, "exit on finished prepare")
	flag.StringVar(&promOut, "prom-out", "", "Prometheus textfile output path")
	flag.BoolVar(&showVersion, "version", false, "show version and exit 1")

	var jiaServiceURLStr, timeoutDuration, initializeTimeoutDuration string
	flag.StringVar(&jiaServiceURLStr, "jia-service-url", getEnv("JIA_SERVICE_URL", "http://apitest:5000"), "jia service url")
	flag.StringVar(&timeoutDuration, "timeout", "1s", "request timeout duration")
	flag.StringVar(&initializeTimeoutDuration, "initialize-timeout", "20s", "request timeout duration of POST /initialize")

	flag.Parse()

	// validate target
	if targetAddress == "" {
		targetAddress = "localhost:9292"
	}
	// validate targetable-addresses
	// useTLS な場合のみ IPアドレスと FQDN のペアが必要になる
	targetableAddresses = strings.Split(targetableAddressesStr, ",")
	if !(1 <= len(targetableAddresses) && len(targetableAddresses) <= 3) || targetableAddresses[0] == "" {
		panic("invalid targetableAddresses: length must be 1~3")
	}
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

type PromTags []string

func (p PromTags) writePromFile() {
	if len(promOut) == 0 {
		return
	}

	promOutNew := fmt.Sprintf("%s.new", promOut)
	err := ioutil.WriteFile(promOutNew, []byte(strings.Join(p, "")), 0644)
	if err != nil {
		logger.AdminLogger.Printf("Failed to write prom file: %s", err)
		return
	}
}

func (p PromTags) commit() {
	if len(promOut) == 0 {
		return
	}

	promOutNew := fmt.Sprintf("%s.new", promOut)
	err := os.Rename(promOutNew, promOut)
	if err != nil {
		logger.AdminLogger.Printf("Failed to write prom file: %s", err)
		return
	}
}

func checkError(err error) (critical bool, timeout bool, deduction bool) {
	return scenario.CheckError(err)
}

func sendResult(s *scenario.Scenario, result *isucandar.BenchmarkResult, finish bool, writeScoreToAdminLogger bool) bool {
	defer func() {
		if finish {
			logger.AdminLogger.Println("<=== sendResult finish")
		}
	}()

	passed := true
	reason := "pass"
	errors := result.Errors.All()

	scoreRaw := result.Score.Sum()
	deduction := int64(0)
	timeoutCount := int64(0)

	type TagCountPair struct {
		Tag   score.ScoreTag
		Count int64
	}
	tagCountPair := make([]TagCountPair, 0)
	promTags := PromTags{}
	scoreTable := result.Score.Breakdown()
	scenario.SetScoreTags(scoreTable)
	for tag, count := range scoreTable {
		tagCountPair = append(tagCountPair, TagCountPair{Tag: tag, Count: count})
	}
	sort.Slice(tagCountPair, func(i, j int) bool {
		return tagCountPair[i].Tag < tagCountPair[j].Tag
	})
	for _, p := range tagCountPair {
		if writeScoreToAdminLogger {
			logger.AdminLogger.Printf("SCORE: %s: %d", p.Tag, p.Count)
		}
		promTags = append(promTags, fmt.Sprintf("xsuconbench_score_breakdown{name=\"%s\"} %d\n", strings.TrimRight(string(p.Tag), " "), p.Count))
	}
	if finish {
		for _, p := range tagCountPair {
			if p.Tag[0] == '_' {
				break //詳細のタグはコンテスタントには見せない
			}
			logger.ContestantLogger.Printf("SCORE: %s: %d", p.Tag, p.Count)
		}
	}

	for _, err := range errors {
		isCritical, isTimeout, isDeduction := checkError(err)

		switch true {
		case isCritical:
			passed = false
			reason = "Critical error"
			logger.AdminLogger.Printf("Critical error because: %+v\n", err)
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
	deductionTotal := deduction + timeoutCount/10

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

	promTags = append(promTags,
		fmt.Sprintf("xsuconbench_score_total{} %d\n", score),
		fmt.Sprintf("xsuconbench_score_raw{} %d\n", scoreRaw),
		fmt.Sprintf("xsuconbench_score_deduction{} %d\n", deductionTotal),
		fmt.Sprintf("xsuconbench_score_error_count{name=\"deduction\"} %d\n", deduction),
		fmt.Sprintf("xsuconbench_score_error_count{name=\"timeout\"} %d\n", timeoutCount),
	)

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

	if passed {
		promTags = append(promTags, "xsuconbench_passed{} 1\n")
	} else {
		promTags = append(promTags, "xsuconbench_passed{} 0\n")
	}

	promTags.writePromFile()
	if finish {
		promTags.commit()
	}

	return passed
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

	if memProfileDir != "" {
		var maxMemStats runtime.MemStats
		go func() {
			for {
				time.Sleep(5 * time.Second)

				var ms runtime.MemStats
				runtime.ReadMemStats(&ms)
				logger.AdminLogger.Printf("system: %d Kb, heap: %d Kb", ms.Sys/1024, ms.HeapAlloc/1024)

				if ms.Sys > maxMemStats.Sys {
					profile.Start(profile.MemProfile, profile.ProfilePath(memProfileDir)).Stop()
					maxMemStats = ms
				}
			}
		}()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// for Scenario
	s, err := scenario.NewScenario(jiaServiceURL, LOAD_TIMEOUT)
	if err != nil {
		panic(err)
	}
	s = s.WithInitializeTimeout(initializeTimeout)

	// IPAddr と FQDN の相互参照可能なmapをシナリオに登録
	var addrAndFqdn []string
	for idx, addr := range targetableAddresses {
		// len(targetableAddresses) は 3 以下なので out-of-range はしない
		addrAndFqdn = append(addrAndFqdn, addr, allowedTargetFQDN[idx])
	}
	if err := s.SetIPAddrAndFqdn(addrAndFqdn...); err != nil {
		panic(err)
	}

	s.NoLoad = noLoad
	s.UseTLS = useTLS

	if useTLS {
		s.BaseURL = fmt.Sprintf("https://%s/", targetAddress)
		// isucandar から送信する HTTPS リクエストの Hosts に isucondition-[1-3].t.isucon.dev を設定する
		var ok bool
		targetAddressWithoutPort := strings.Split(targetAddress, ":")[0]
		agent.DefaultTLSConfig.ServerName, ok = s.GetFqdnFromIPAddr(targetAddressWithoutPort)
		if !ok {
			panic("targetAddress が targetableAddresses に含まれていません")
		}
	} else {
		s.BaseURL = fmt.Sprintf("http://%s/", targetAddress)
	}

	// JIA API
	go s.JiaAPIService(ctx)

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

		count := 0
		for {
			// 途中経過を3秒毎に送信
			timer := time.After(3 * time.Second)
			sendResult(s, step.Result(), false, count%5 == 0)

			select {
			case <-timer:
			case <-ctx.Done():
				return nil
			}
			count++
		}
	})

	result := b.Start(ctx)

	wg.Wait()

	if !sendResult(s, result, true, true) && exitStatusOnFail {
		os.Exit(1)
	}
}
