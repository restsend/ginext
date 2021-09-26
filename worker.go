package ginext

import (
	"context"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

const NotImplementHandle = `{"msg":"not implement"}`
const UnmarshalFailHandle = `{"msg":"unmarshal fail"}`

type WorkHandle func(*GinTask) (string, error)

type Worker struct {
	db       *gorm.DB
	WorkerID uint
	Name     string
	Handlers map[string]WorkHandle

	PullInterval  time.Duration
	PullTaskNum   int
	TaskNum       int
	masterContext context.Context
	pullCancel    context.CancelFunc
}

func NewWorker(db *gorm.DB, name string) *Worker {
	w := &Worker{
		db:           db.Session(&gorm.Session{}),
		Name:         name,
		PullInterval: 1 * time.Second,
		PullTaskNum:  20,
		TaskNum:      4,
		Handlers:     make(map[string]WorkHandle),
	}
	w.masterContext = context.Background()
	return w
}

func (w *Worker) AddHandle(taskType string, h WorkHandle) {
	w.Handlers[taskType] = h
}

func (w *Worker) Init() (err error) {
	go w.Pull()
	return nil
}

func (w *Worker) Shutdown() {
	if w.pullCancel != nil {
		w.pullCancel()
		w.pullCancel = nil
	}
}

func (w *Worker) Pull() {
	var pullContext context.Context
	pullContext, w.pullCancel = context.WithCancel(w.masterContext)
	ticker := time.NewTicker(w.PullInterval)
	for {
		select {
		case <-ticker.C:
			SafeCall(w.pullTasks, nil)
		case <-pullContext.Done():
			return
		}
	}
}

func (w *Worker) pullTasks() error {
	var ts []GinTask

	tx := w.db.Model(&GinTask{}).Where("done", false).Limit(w.PullTaskNum)
	result := tx.Order("start_time").Find(&ts)

	if result.Error != nil {
		log.Println("query tasks fail", result.Error)
		return result.Error
	}

	for _, t := range ts {
		if t.StartTime != nil {
			if time.Since(*t.StartTime) < 0 {
				continue
			}
		}
		w.execute(func() {
			err := w.DoTask(&t)
			if err != nil {
				log.Printf("task fail taskid:%d type:%s err:%v", t.ID, t.TaskType, err)
			}
		}, 60*time.Second)
	}
	return nil
}

func (w *Worker) execute(h func(), timeout time.Duration) {
	h()
}

func (w *Worker) DoTask(t *GinTask) error {
	handle, ok := w.Handlers[t.TaskType]
	if !ok {
		return errors.New("unknown task type")
	}
	now := time.Now()
	vals := map[string]interface{}{
		"ExecTime": &now,
	}

	var handleResult string
	var err error
	err = SafeCall(func() error {
		var e error
		handleResult, e = handle(t)
		return e
	}, func(e error) {
		err = e
	})

	if err != nil {
		vals["Failed"] = true
	}
	now = time.Now()
	vals["EndTime"] = &now
	vals["Result"] = handleResult
	vals["Done"] = true
	w.db.Model(&t).UpdateColumns(vals)
	return err
}

type WorkerManager struct {
	db  *gorm.DB
	ext *GinExt
}

var defaultWorkerManagerInst *WorkerManager

func DefaultWorkerManager() *WorkerManager {
	return defaultWorkerManagerInst
}

func NewWorkerManager(c *GinExt) *WorkerManager {
	v := &WorkerManager{
		ext: c,
		db:  c.DbInstance,
	}
	defaultWorkerManagerInst = v
	return defaultWorkerManagerInst
}

func (wm *WorkerManager) Migrate() (err error) {
	err = wm.db.AutoMigrate(&GinTask{})
	if err != nil {
		log.Panicf("Migrate GinTask Fail %v", err)
		return err
	}
	return nil
}

func (wm *WorkerManager) Init() (err error) {
	return wm.Migrate()
}

func (wm *WorkerManager) Add(objectID int64, taskType, context string, delays time.Duration) error {
	o := GinTask{
		CreatedAt: time.Now(),
		TaskType:  taskType,
		ObjectID:  objectID,
		Done:      false,
		Context:   context,
		Result:    "",
	}
	if delays.Seconds() > 0 {
		now := time.Now().Add(delays)
		o.StartTime = &now
	}
	if wm == nil {
		log.Panic("wm is nil", wm)
	}
	if wm.db == nil {
		log.Panic("wm.db is nil", wm)
	}
	result := wm.db.Create(&o)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (wm *WorkerManager) CancelAll(objectID int64) int64 {
	tx := wm.db.Model(&GinTask{}).Where("object_id", objectID).Where("done", false)
	result := tx.UpdateColumn("done", true)
	return result.RowsAffected
}

// For Worker
func (wm *WorkerManager) Tidyup(maxCount int) {
	tx := wm.db.Where("done", true).Where("failed", false).Limit(maxCount).Order("created_at")
	tx.Delete(GinTask{})
}
