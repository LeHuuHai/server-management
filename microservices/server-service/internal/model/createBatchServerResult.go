package model

type CreateBatchServerResult struct {
	Success    []string
	Failed     []string
	SuccessCnt int
	FailedCnt  int
}
