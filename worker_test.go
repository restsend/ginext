package ginext

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func NewTestWorkerManager() *WorkerManager {
	cfg := NewGinExt("..")
	cfg.Init()

	wm := NewWorkerManager(cfg)
	wm.Init()
	wm.db.Delete(&GinTask{}, "id > 0")
	return wm
}
func TestWorkerManager(t *testing.T) {
	wm := NewTestWorkerManager()
	{
		err := wm.Add(1, "hello", "{}", 2*time.Second)
		assert.Nil(t, err)
	}
	{
		err := wm.Add(1, "hello2", "{}", 0)
		assert.Nil(t, err)
	}
	{
		var ts []GinTask
		tx := wm.db.Model(&GinTask{}).Where("done", false)
		result := tx.Order("start_time").Find(&ts)
		assert.Nil(t, result.Error)
		assert.Equal(t, len(ts), 2)
		assert.Equal(t, ts[0].TaskType, "hello2")
	}
	{
		wm.Tidyup(1)
		var count int64
		wm.db.Model(&GinTask{}).Count(&count)
		assert.Equal(t, int(count), 2)
		wm.db.Debug().Model(&GinTask{}).Where("task_type", "hello").UpdateColumn("done", true)

		wm.Tidyup(1)
		wm.db.Model(&GinTask{}).Count(&count)
		assert.Equal(t, int(count), 1)
	}
	{
		c := wm.CancelAll(1)
		assert.Equal(t, 1, int(c))
	}
}

func TestWorker(t *testing.T) {
	wm := NewTestWorkerManager()
	var err error
	startTime := time.Now()
	{
		err = wm.Add(1, "hello", "{}", 1*time.Second)
		assert.Nil(t, err)
		err = wm.Add(1, "hello", "{}", 0)
		assert.Nil(t, err)
	}
	w := NewWorker(wm.db, "test worker")
	wg := sync.WaitGroup{}
	wg.Add(2)

	w.AddHandle("hello", func(t *GinTask) (string, error) {
		wg.Done()
		return "", nil
	})
	err = w.Init()
	assert.Nil(t, err)
	wg.Wait()
	w.Shutdown()

	{
		var ts []GinTask
		tx := wm.db.Model(&GinTask{}).Where("done", true).Order("exec_time desc")
		tx.Find(&ts)
		assert.Equal(t, len(ts), 2)

		assert.GreaterOrEqual(t, ts[0].ExecTime.Sub(startTime), 1*time.Second)
		assert.GreaterOrEqual(t, time.Since(*ts[0].ExecTime), 0*time.Second)
		assert.GreaterOrEqual(t, time.Since(*ts[1].ExecTime), 0*time.Second)
	}
}
func TestWorkerFail(t *testing.T) {
	wm := NewTestWorkerManager()
	var err error
	{
		err = wm.Add(1, "hello", "{}", 0)
		assert.Nil(t, err)
	}

	w := NewWorker(wm.db, "test worker")
	wg := sync.WaitGroup{}
	wg.Add(1)

	w.AddHandle("hello", func(t *GinTask) (string, error) {
		wg.Done()
		return "", errors.New("mock fail")
	})
	err = w.Init()
	assert.Nil(t, err)
	wg.Wait()
	w.Shutdown()

	{
		time.Sleep(500 * time.Millisecond)
		var ts []GinTask
		tx := wm.db.Model(&GinTask{}).Where("done", true).Order("exec_time desc")
		tx.Find(&ts)
		assert.Equal(t, 1, len(ts))
		assert.True(t, ts[0].Failed)
	}
}
