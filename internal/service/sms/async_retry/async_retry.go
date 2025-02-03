package async_retry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"math"
	"strconv"
	"time"
	"webok/internal/service/sms"
)

var ErrSendFailedPendingRetry = errors.New("发送失败等待重试")

const AsyncQueueKey = "sync_queue:retry_sms"
const AsyncTaskMetaKeyPrefix = "async_retry_sms_task"

// AsyncRetrySMSService 对失败的请求采用异步重试的方式，重试协程数为1
type AsyncRetrySMSService struct {
	service                sms.Service
	rdb                    redis.Cmdable
	maxRetries             int32
	taskReTryInterval      int64
	goRoutineCheckInterval int64
	cancel                 context.CancelFunc
}

// Send 发送短信消息并在初次发送失败时处理重试逻辑。
// 它检查 Redis 中是否存在待重试的任务，如果存在则返回错误。
// 如果发送失败，它会存储任务元数据并将其添加到重试队列中。
func (a *AsyncRetrySMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	if len(number) == 0 {
		return nil
	}
	// 检查redis中是否存在重试，存在则直接返回对应错误
	taskID := a.key(number[0])
	exists, err := a.rdb.Exists(ctx, taskID).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return ErrSendFailedPendingRetry
	}

	//发送
	err = a.service.Send(ctx, tplId, number, args...)
	if err == nil {
		return nil
	}

	// 存储任务元数据
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("failed to marshal args: %w", err)
	}
	numbersJSON, err := json.Marshal(number)
	if err != nil {
		return fmt.Errorf("failed to marshal numbers: %w", err)
	}
	//TODO 后续使用lua脚本保证原子性
	pipe := a.rdb.TxPipeline()
	pipe.HSet(ctx, taskID, map[string]any{
		"tplId":      tplId,
		"args":       argsJSON,
		"numbers":    numbersJSON,
		"retries":    0,
		"maxRetries": a.maxRetries,
	})
	pipe.ZAdd(ctx, AsyncQueueKey, redis.Z{
		Score:  float64(time.Now().Unix() + a.taskReTryInterval),
		Member: taskID,
	})
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return ErrSendFailedPendingRetry
}

func (a *AsyncRetrySMSService) key(tel string) string {
	return fmt.Sprintf("%s:%s", AsyncTaskMetaKeyPrefix, tel)
}

func (a *AsyncRetrySMSService) AsyncRetryWorker() error {
	ctx := context.Background()

	// 获取当前时间戳
	now := time.Now().Unix()

	// 获取所有到期的任务ID
	taskIDs, err := a.rdb.ZRangeByScore(ctx, AsyncQueueKey, &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(now, 10),
	}).Result()

	if err != nil {
		log.Printf("获取任务失败: %v", err)
		return err
	}

	// 处理每个任务
	for _, taskID := range taskIDs {
		// 使用事务确保原子性
		tx := a.rdb.TxPipeline()

		//TODO 后续使用lua脚本保证原子性

		// 1. 获取任务元数据
		taskData := tx.HGetAll(ctx, taskID)

		// 2. 从Sorted Set中临时移除
		tx.ZRem(ctx, AsyncQueueKey, taskID)

		// 提交事务
		_, err := tx.Exec(ctx)
		if err != nil {
			log.Printf("获取任务元数据失败: %v", err)
			continue
		}

		// 解析元数据
		tplId, _ := taskData.Val()["tplId"]
		argsStr, _ := taskData.Val()["args"]
		numberStr, _ := taskData.Val()["numbers"]
		retries, _ := strconv.Atoi(taskData.Val()["retries"])
		maxRetries, _ := strconv.Atoi(taskData.Val()["maxRetries"])

		// 检查最大重试次数
		if retries >= maxRetries {
			log.Printf("任务达到最大重试次数: %s", taskID)
			a.rdb.Del(ctx, taskID)
			continue
		}

		// 反序列化请求
		var args []string
		jsonErr := json.Unmarshal([]byte(argsStr), &args)
		if jsonErr != nil {
			log.Printf("反序列化参数失败: %v\n", jsonErr)
			continue
		}
		var number []string
		jsonErr = json.Unmarshal([]byte(numberStr), &number)
		if jsonErr != nil {
			log.Printf("反序列化参数失败: %v\n", jsonErr)
			continue
		}

		// 尝试发送请求
		sendErr := a.service.Send(ctx, tplId, args, number...)
		if sendErr == nil {
			// 成功：删除任务
			_, delErr := a.rdb.Del(ctx, taskID).Result()
			if delErr != nil {
				log.Printf("重试成功, 删除任务失败: %v", delErr)
				continue
			}
		} else {
			log.Printf("发送失败: %v", sendErr)
			//TODO 后续使用lua脚本保证原子性

			// 失败：更新重试次数和下次重试时间
			nextRetryAt := time.Now().Add(time.Duration(int64(math.Pow(2, float64(retries)))*a.taskReTryInterval) * time.Second).Unix()
			_, hsetErr := a.rdb.HSet(ctx, "async_task:"+taskID, map[string]any{
				"retries": retries + 1,
			}).Result()
			if hsetErr != nil {
				log.Printf("更新任务元数据失败: %v", hsetErr)
				continue
			}
			// 重新加入队列
			_, zaddErr := a.rdb.ZAdd(ctx, AsyncQueueKey, redis.Z{
				Score:  float64(nextRetryAt),
				Member: taskID,
			}).Result()
			if zaddErr != nil {
				log.Printf("重新加入队列失败: %v", zaddErr)
				continue
			}
		}
	}
	return nil
}

func (a *AsyncRetrySMSService) StartAsyncRetryWorker() {
	ctx, cancel := context.WithCancel(context.Background())
	a.StopAsyncRetryWorker()
	a.cancel = cancel
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := a.AsyncRetryWorker()
			if err != nil {
				log.Printf("异步重试工作失败: %v", err)
				// 可以在这里添加重试逻辑或其他处理方式
			}
			time.Sleep(time.Duration(a.goRoutineCheckInterval) * time.Second)
		}
	}
}

func (a *AsyncRetrySMSService) StopAsyncRetryWorker() {
	if a.cancel != nil {
		a.cancel()
	}
}

func NewAsyncRetrySMSService(svc sms.Service, r redis.Cmdable, maxRetryTimes int32, reTryInterval int64, checkInterval int64) *AsyncRetrySMSService {
	service := AsyncRetrySMSService{
		service:                svc,
		rdb:                    r,
		maxRetries:             maxRetryTimes,
		taskReTryInterval:      reTryInterval,
		goRoutineCheckInterval: checkInterval,
	}
	return &service
}
