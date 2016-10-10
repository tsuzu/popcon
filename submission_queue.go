package main

var SJQueue *SubmissionJudgeQueue

type SubmissionJudgeQueue struct {
	queue     chan int64
	used      map[int64]bool
	pushQueue chan int64
	popQueue  chan chan int64
}

func CreateSubmissionJudgeQueue() *SubmissionJudgeQueue {
	var sjq SubmissionJudgeQueue

	sjq.pushQueue = make(chan int64, 100)
	sjq.popQueue = make(chan chan int64, 50)

	return &sjq
}

func (sjq *SubmissionJudgeQueue) Push(sid int64) {
	if sid <= 0 {
		return
	}

	sjq.pushQueue <- sid
}

func (sjq *SubmissionJudgeQueue) Pop() int64 {
	ch := make(chan int64, 1)

	sjq.popQueue <- ch

	return <-ch
}

func (sjq *SubmissionJudgeQueue) Remove(sid int64) {
	if sid <= 0 {
		return
	}

	sjq.pushQueue <- -sid
}

func (sjq *SubmissionJudgeQueue) run() {
	sjq.queue = make(chan int64, 500)

	for {
		select {
		case query, has := <-sjq.pushQueue:
			if !has {
				break
			}

			if query < 0 {
				query = -query

				if _, has = sjq.used[query]; has {
					delete(sjq.used, query)
				}
			} else {
				if _, has = sjq.used[query]; has {
					continue
				}

				if len(sjq.queue) == 500 {
					q, has := <-sjq.popQueue

					if !has {
						goto end
					}
					q <- (<-sjq.queue)
				}
				sjq.queue <- query
			}
		case query, has := <-sjq.popQueue:
			if !has {
				goto end
			}
			if len(sjq.queue) == 0 {
				for {
					v, has := <-sjq.pushQueue

					if !has {
						goto end
					}

					if v < 0 {
						v = -v

						if _, has = sjq.used[v]; has {
							delete(sjq.used, v)
						}
						continue
					}

					if _, has = sjq.used[v]; has {
						continue
					}

					query <- v

					break
				}
			} else {
				query <- (<-sjq.queue)
			}
		}
	}
end:
}
