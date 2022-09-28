package poll

import (
	"time"

	apiwait "k8s.io/apimachinery/pkg/util/wait"
)

type PollFunc func(interval, timeout time.Duration, condition apiwait.ConditionFunc) error
