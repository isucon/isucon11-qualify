package scenario

// prepare.go
// シナリオの内、prepareフェーズの処理

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon11-qualify/bench/logger"
	"github.com/isucon/isucon11-qualify/bench/model"
	"github.com/isucon/isucon11-qualify/bench/service"
)

func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	logger.ContestantLogger.Printf("===> PREPARE")

	//initialize
	initializer, err := s.NewAgent(
		agent.WithNoCache(), agent.WithNoCookie(), agent.WithTimeout(20*time.Second),
	)
	if err != nil {
		return failure.NewError(ErrCritical, err)
	}
	initializer.Name = "benchmarker-initializer"

	initResponse, errs := initializeAction(ctx, initializer)
	for _, err := range errs {
		step.AddError(err)
	}
	if len(errs) > 0 {
		//return ErrScenarioCancel
		return ErrCritical
	}

	s.Language = initResponse.Language

	//各エンドポイントのチェック
	// TODO: 並列
	err = s.prepareCheckAuth(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareCheckSignout(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareGetMe(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareGetIsu(ctx, step)
	if err != nil {
		return err
	}
	err = s.preparePostIsu(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareGetIsuId(ctx, step)
	if err != nil {
		return err
	}
	err = s.preparePutIsu(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareDeleteIsu(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareGetIsuIcon(ctx, step)
	if err != nil {
		return err
	}
	err = s.preparePutIsuIcon(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareGetIsuSearch(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareGetCatalog(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareGetGraph(ctx, step)
	if err != nil {
		return err
	}
	err = s.preparePostCondition(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareGetIsuCondition(ctx, step)
	if err != nil {
		return err
	}
	err = s.prepareGetCondition(ctx, step)
	if err != nil {
		return err
	}

	// Prepare step でのエラーはすべて Critical の扱い
	if len(step.Result().Errors.All()) > 0 {
		//return ErrScenarioCancel
		return ErrCritical
	}
	return nil
}

//エンドポイント事の単体テスト

func (s *Scenario) prepareCheckAuth(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {

		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		if (index % 10) < authActionErrorNum {
			//各種ログイン失敗ケース
			errs := authActionError(ctx, agt, userID, index%10)
			for _, err := range errs {
				step.AddError(err)
			}
		} else {
			//ログイン成功
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
			}
		}
	}, worker.WithLoopCount(20))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	//w.Wait() //念のためもう一度止まってるか確認

	//作成済みユーザーへのログイン確認
	agt, err := s.NewAgent()
	if err != nil {
		step.AddError(failure.NewError(ErrCritical, err))
		return nil
	}
	userID, err := model.MakeRandomUserID()
	if err != nil {
		step.AddError(failure.NewError(ErrCritical, err))
		return nil
	}

	_, errs := authAction(ctx, agt, userID)
	for _, err := range errs {
		step.AddError(err)
	}
	agt.ClearCookie()
	//二回目のログイン
	_, errs = authAction(ctx, agt, userID)
	for _, err := range errs {
		step.AddError(err)
	}

	return nil
}

func (s *Scenario) prepareCheckSignout(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		switch index {
		// success
		case 0:
			// login
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
			}
			// signout
			_, err := signoutAction(ctx, agt)
			if err != nil {
				step.AddError(err)
			}
		// fail: not login
		case 1:
			text, res, err := signoutErrorAction(ctx, agt)
			if err != nil {
				step.AddError(err)
			}
			err = verifyNotSignedIn(res, text)
			if err != nil {
				step.AddError(err)
			}
		// fail: re-signout
		case 2:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
			}
			_, err := signoutAction(ctx, agt)
			if err != nil {
				step.AddError(err)
			}
			text, res, err := signoutErrorAction(ctx, agt)
			if err != nil {
				step.AddError(err)
			}
			err = verifyNotSignedIn(res, text)
			if err != nil {
				step.AddError(err)
			}
		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) prepareGetMe(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			me, res, err := getMeAction(ctx, agt)
			if err != nil {
				step.AddError(err)
				return
			}
			if me == nil {
				step.AddError(errorInvalidJSON(res))
			}
		// fail
		case 1:
			// not login
			text, res, err := getMeErrorAction(ctx, agt)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedIn(res, text)
			if err != nil {
				step.AddError(err)
			}
		}
	}, worker.WithLoopCount(2))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) prepareGetIsu(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			isuList, res, err := getIsuAction(ctx, agt)
			if err != nil {
				step.AddError(err)
				return
			}
			if isuList == nil {
				step.AddError(errorInvalidJSON(res))
			}
		// fail
		case 1:
			// not login
			text, res, err := getIsuErrorAction(ctx, agt)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedIn(res, text)
			if err != nil {
				step.AddError(err)
			}
		}
	}, worker.WithLoopCount(1))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) preparePostIsu(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		//TODO: デバッグ目的でランダム文字列を生成しているだけなので、本書きするときにIsuIDに直す
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		req := service.PostIsuRequest{
			JIAIsuUUID: userID,
			IsuName:    "po",
		}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			isu, res, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			if isu == nil || isu.JIAIsuUUID != req.JIAIsuUUID || isu.Name != req.IsuName {
				step.AddError(errorInvalidJSON(res))
			}
		// fail: not login
		case 1:
			text, res, err := postIsuErrorAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedIn(res, text)
			if err != nil {
				step.AddError(err)
			}
		// fail: bad request body
		// case 2:
		// 	_, errs := authAction(ctx, agt, userID)
		// 	for _, err := range errs {
		// 		step.AddError(err)
		// 		return
		// 	}
		// 	text, res, err := postIsuErrorAction(ctx, agt, service.GetIsuConditionResponse{})
		// 	if err != nil {
		// 		step.AddError(err)
		// 		return
		// 	}
		// 	err = verifyBadReqBody(res, text)
		// 	if err != nil {
		// 		step.AddError(err)
		// 	}
		// fail: 404
		case 3:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			// uuid = a だけは外部APIの方で404を返すようにした
			req.JIAIsuUUID = "a"
			text, res, err := postIsuErrorAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyStatusCode(res, http.StatusNotFound)
			if err != nil {
				step.AddError(err)
			}
			err = verifyText(res, text, "JIAService returned error")
			if err != nil {
				step.AddError(err)
			}
		case 4:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			_, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			text, res, err := postIsuErrorAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyStatusCode(res, http.StatusConflict)
			if err != nil {
				step.AddError(err)
			}
			err = verifyText(res, text, "duplicated isu")
			if err != nil {
				step.AddError(err)
			}
		}
	}, worker.WithLoopCount(5))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) prepareGetIsuId(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		req := service.PostIsuRequest{JIAIsuUUID: userID, IsuName: "po"}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			expectedIsu, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			isu, res, err := getIsuIdAction(ctx, agt, expectedIsu.JIAIsuUUID)
			if err != nil {
				step.AddError(err)
				return
			}
			isEqual := AssertEqual("getIsuIdAction", expectedIsu, isu)
			if !isEqual {
				step.AddError(errorInvalidJSON(res))
			}
		// fail: not login(存在よりログインチェックが先)
		case 1:
			text, res, err := getIsuIdErrorAction(ctx, agt, userID)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedInTODO(res, text)
			if err != nil {
				step.AddError(err)
			}
		// fail: not found
		case 2:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			text, res, err := getIsuIdErrorAction(ctx, agt, userID)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyIsuNotFound(res, text)
			if err != nil {
				step.AddError(err)
			}

		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) preparePutIsu(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		req := service.PostIsuRequest{JIAIsuUUID: userID, IsuName: "po"}
		putReq := service.PutIsuRequest{Name: "pi"}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			expectedIsu, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			isu, res, err := putIsuAction(ctx, agt, expectedIsu.JIAIsuUUID, putReq)
			if err != nil {
				step.AddError(err)
				return
			}
			if isu.Name != putReq.Name {
				step.AddError(errorInvalidJSON(res))
			}
		// fail: not login(存在よりログインチェックが先)
		case 1:
			text, res, err := putIsuErrorAction(ctx, agt, userID, putReq)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedInTODO(res, text)
			if err != nil {
				step.AddError(err)
			}
		// fail: not found
		case 2:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			text, res, err := putIsuErrorAction(ctx, agt, userID, putReq)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyIsuNotFound(res, text)
			if err != nil {
				step.AddError(err)
			}

		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) prepareDeleteIsu(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		req := service.PostIsuRequest{JIAIsuUUID: userID, IsuName: "po"}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			isu, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			_, err = deleteIsuAction(ctx, agt, isu.JIAIsuUUID)
			if err != nil {
				step.AddError(err)
				return
			}
			text, res, err := getIsuIdErrorAction(ctx, agt, isu.JIAIsuUUID)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyIsuNotFound(res, text)
			if err != nil {
				step.AddError(err)
				return
			}
			// 二回目は失敗
			_, _, err = deleteIsuErrorAction(ctx, agt, isu.JIAIsuUUID)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyIsuNotFound(res, text)
			if err != nil {
				step.AddError(err)
			}
		// fail: not login(存在よりログインチェックが先)
		case 1:
			text, res, err := deleteIsuErrorAction(ctx, agt, userID)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedIn(res, text)
			if err != nil {
				step.AddError(err)
			}
		// fail: not found
		case 2:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			text, res, err := deleteIsuErrorAction(ctx, agt, userID)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyIsuNotFound(res, text)
			if err != nil {
				step.AddError(err)
			}

		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) prepareGetIsuIcon(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		req := service.PostIsuRequest{JIAIsuUUID: userID, IsuName: "po"}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			isu, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			_, _, err = getIsuIconAction(ctx, agt, isu.JIAIsuUUID)
			if err != nil {
				step.AddError(err)
				return
			}
		// fail: not login(存在よりログインチェックが先)
		case 1:
			text, res, err := getIsuIconErrorAction(ctx, agt, userID)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedInTODO(res, text)
			if err != nil {
				step.AddError(err)
			}
		// fail: not found
		case 2:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			text, res, err := getIsuIconErrorAction(ctx, agt, userID)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyIsuNotFound(res, text)
			if err != nil {
				step.AddError(err)
			}

		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) preparePutIsuIcon(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		req := service.PostIsuRequest{JIAIsuUUID: userID, IsuName: "po"}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			isu, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			// TODO: ちゃんとする。デバッグ用に手元のa.pngを読んでる
			file, _ := os.Open("a.png")
			_, err = putIsuIconAction(ctx, agt, isu.JIAIsuUUID, file)
			if err != nil {
				step.AddError(err)
				return
			}
		// fail: not login(存在よりログインチェックが先)
		case 1:
			// TODO: ちゃんとする。デバッグ用に手元のa.pngを読んでる
			file, _ := os.Open("a.png")
			text, res, err := putIsuIconErrorAction(ctx, agt, userID, file)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedInTODO(res, text)
			if err != nil {
				step.AddError(err)
			}
		// fail: not found
		case 2:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			// TODO: ちゃんとする。デバッグ用に手元のa.pngを読んでる
			file, _ := os.Open("a.png")
			text, res, err := putIsuIconErrorAction(ctx, agt, userID, file)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyIsuNotFound(res, text)
			if err != nil {
				step.AddError(err)
			}

		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) prepareGetIsuSearch(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		req := service.PostIsuRequest{JIAIsuUUID: userID, IsuName: "po"}
		searchReq := service.GetIsuSearchRequest{}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			_, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			_, _, err = getIsuSearchAction(ctx, agt, searchReq)
			if err != nil {
				step.AddError(err)
				return
			}
		// fail: not login(存在よりログインチェックが先)
		case 1:
			text, res, err := getIsuSearchErrorAction(ctx, agt, searchReq)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedInTODO(res, text)
			if err != nil {
				step.AddError(err)
			}
			// パラメーターを変える (Bad Request) のは型変えないとだめそうでめんどくさかったから未確認
		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) prepareGetCatalog(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		validId := "550e8400-e29b-41d4-a716-446655440000"
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			_, _, err := getCatalogAction(ctx, agt, validId)
			if err != nil {
				step.AddError(err)
				return
			}
		// fail: not login(存在よりログインチェックが先)
		case 1:
			text, res, err := getCatalogErrorAction(ctx, agt, "invalid")
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedIn(res, text)
			if err != nil {
				step.AddError(err)
			}
			// fail: not login(存在よりログインチェックが先)
		case 2:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			text, res, err := getCatalogErrorAction(ctx, agt, "invalid")
			if err != nil {
				step.AddError(err)
				return
			}
			if res.StatusCode != http.StatusNotFound || text != "invalid jia_catalog_id" {
				step.AddError(errorInvalidStatusCode(res, http.StatusNotFound))
			}
		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) prepareGetGraph(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		req := service.PostIsuRequest{JIAIsuUUID: userID, IsuName: "po"}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			isu, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			_, _, err = getIsuGraphAction(ctx, agt, isu.JIAIsuUUID, 0)
			if err != nil {
				step.AddError(err)
				return
			}
		// fail: not login(存在よりログインチェックが先)
		case 1:
			text, res, err := getIsuGraphErrorAction(ctx, agt, "invalid", 0)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedInTODO(res, text)
			if err != nil {
				step.AddError(err)
			}
			// fail: not found(存在よりログインチェックが先)
			// TODO: エラーチェックがされてないからされたら試す
			// case 2:
			// 	_, errs := authAction(ctx, agt, userID)
			// 	for _, err := range errs {
			// 		step.AddError(err)
			// 		return
			// 	}
			// 	text, res, err := getIsuGraphErrorAction(ctx, agt, "invalid", 0)
			// 	if err != nil {
			// 		step.AddError(err)
			// 		return
			// 	}
			// 	err = verifyIsuNotFound(res, text)
			// 	if err != nil {
			// 		step.AddError(err)
			// 	}
		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) preparePostCondition(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		req := service.PostIsuRequest{JIAIsuUUID: userID, IsuName: "po"}
		conditionReq := service.PostIsuConditionRequest{IsSitting: true, Message: "hello", Timestamp: 0, Condition: "sleep=true"}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			isu, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			_, err = postIsuConditionAction(ctx, agt, isu.JIAIsuUUID, conditionReq)
			if err != nil {
				step.AddError(err)
				return
			}
		// fail: not found
		case 1:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			text, res, err := postIsuConditionErrorAction(ctx, agt, "invalid", conditionReq)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyIsuNotFound(res, text)
			if err != nil {
				step.AddError(err)
			}
		case 2:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			isu, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			invalidCondition := "sleep=hige"
			conditionReq.Condition = invalidCondition
			text, res, err := postIsuConditionErrorAction(ctx, agt, isu.JIAIsuUUID, conditionReq)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyBadReqBody(res, text)
			if err != nil {
				step.AddError(err)
			}
		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) prepareGetIsuCondition(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		req := service.PostIsuRequest{JIAIsuUUID: userID, IsuName: "po"}
		conditionReq := service.GetIsuConditionRequest{CursorEndTime: 0, CursorJIAIsuUUID: "z", ConditionLevel: "info"}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			isu, _, err := postIsuAction(ctx, agt, req)
			if err != nil {
				step.AddError(err)
				return
			}
			_, _, err = getIsuConditionAction(ctx, agt, isu.JIAIsuUUID, conditionReq)
			if err != nil {
				logger.AdminLogger.Println()
				step.AddError(err)
				return
			}
		// fail: not login
		case 1:
			text, res, err := getIsuConditionErrorAction(ctx, agt, "invalid", conditionReq)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedInTODO(res, text)
			if err != nil {
				step.AddError(err)
			}
		// fail: not found
		case 2:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			text, res, err := getIsuConditionErrorAction(ctx, agt, "invalid", conditionReq)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyIsuNotFound(res, text)
			if err != nil {
				step.AddError(err)
			}
		}
	}, worker.WithLoopCount(3))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}

func (s *Scenario) prepareGetCondition(ctx context.Context, step *isucandar.BenchmarkStep) error {

	//TODO: ユーザープール
	//とりあえずは使い捨てのユーザーを使う

	w, err := worker.NewWorker(func(ctx context.Context, index int) {
		agt, err := s.NewAgent()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		userID, err := model.MakeRandomUserID()
		if err != nil {
			step.AddError(failure.NewError(ErrCritical, err))
			return
		}
		conditionReq := service.GetIsuConditionRequest{CursorEndTime: 0, CursorJIAIsuUUID: "z", ConditionLevel: "info"}
		switch index {
		// success
		case 0:
			_, errs := authAction(ctx, agt, userID)
			for _, err := range errs {
				step.AddError(err)
				return
			}
			_, _, err = getConditionAction(ctx, agt, conditionReq)
			if err != nil {
				logger.AdminLogger.Println()
				step.AddError(err)
				return
			}
		// fail: not login
		case 1:
			text, res, err := getConditionErrorAction(ctx, agt, conditionReq)
			if err != nil {
				step.AddError(err)
				return
			}
			err = verifyNotSignedInTODO(res, text)
			if err != nil {
				step.AddError(err)
			}
		}
	}, worker.WithLoopCount(2))

	if err != nil {
		return failure.NewError(ErrCritical, err)
	}

	w.Process(ctx)
	w.Wait()

	return nil
}
