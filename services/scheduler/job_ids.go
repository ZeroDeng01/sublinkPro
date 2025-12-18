package scheduler

// 系统任务ID常量
// 使用负数ID区间，避免与用户订阅ID（正整数）冲突
const (
	// JobIDSpeedTest 节点测速定时任务ID
	JobIDSpeedTest = -100

	// JobIDHostCleanup Host过期清理任务ID
	JobIDHostCleanup = -101

	// 预留区间 -102 ~ -199 用于未来系统任务
	// 新增系统任务时按顺序递减分配ID
)

// 用户订阅任务使用 SubScheduler.ID（正整数，数据库自增）
// 无需在此处定义
