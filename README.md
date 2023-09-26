# Redis streams implements message queue demo
just demo 
是否能完成还不一定 
## example
```bash
./myredmq.exe run
create topic my_test_topic 1m # 一分钟后进入到延迟队列中
create topic my_test_dead_topic # 死信队列
bind topic my_test_topic to my_test_dead_topic # 绑定延迟队列和死信队列
send <msg> to my_test_topic # 发送消息给延迟队列
```


## How implement Delay Queue?

1. 为每一个消息维护一个协程定时器，如果这个定时器达到了时间，会主动ack这个msg
2. 为每一个 topic 维护一个 定时器，但是当send数据的时候，一定要记录下send数据的时间了，这个定时器不断的监控第一个到达时间的消息

方案1: 消耗资源过多，10w数据到达，维护10w个协程，即使协程比线程更加轻量级，但为每一个msg维护一个协程也没必要,采用方案2吧

当数据到来的时候，维护一个 msg Id 和 create time的map

定时任务 go api
```go
    // 创建 嘀嗒器 500ms 检查一次topic
    ticker := time.NewTicker(500 * time.Millisecond)
    done := make(chan bool)
    go func() {
        for {
            select {
            case <-done:
                return
            case t := <-ticker.C:
				// 检查 topic 第一个 消息id
				// 并通过和 map 对比 是否超时，假设超时，ack 并移除
				// 假设
				// 假设并未超时 啥也不干
                fmt.Println("Tick at", t)
            }
        }
    }()

```




