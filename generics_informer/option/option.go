package option

import "time"

// CollectorOption 配置文件
type CollectorOption struct {
	// MaxReQueueTime 最大重新入队次数
	MaxReQueueTime int
	// ResyncPeriod 重新调协间隔时间
	ResyncPeriod time.Duration
}

func NewCollectorOption(maxReQueueTime int, resyncPeriod time.Duration) *CollectorOption {
	return &CollectorOption{MaxReQueueTime: maxReQueueTime, ResyncPeriod: resyncPeriod}
}
